import { useState, useEffect } from 'react';
import { Crown, ArrowUpRight } from 'lucide-react';
import { getTierInfo, type TierInfo } from '../api/tier';
import { TierUpgradeModal } from './tier-upgrade-modal';
import { getWallet, type InternalBalance } from '../api/wallet';

export const TierBadge = () => {
  const [tierInfo, setTierInfo] = useState<TierInfo | null>(null);
  const [wallet, setWallet] = useState<InternalBalance | null>(null);
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);
  const [loading, setLoading] = useState(true);

  const fetchData = async () => {
    try {
      const [tierData, walletData] = await Promise.all([
        getTierInfo(),
        getWallet()
      ]);
      setTierInfo(tierData);
      setWallet(walletData);
    } catch (error) {
      console.error('Failed to fetch tier info:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  if (loading || !tierInfo) {
    return (
      <div className="p-3 bg-white/5 rounded-xl border border-white/10 animate-pulse">
        <div className="h-6 bg-white/10 rounded w-24" />
      </div>
    );
  }

  const isGold = tierInfo.tier === 'gold';
  const isExpired = tierInfo.tier_expires_at && new Date(tierInfo.tier_expires_at) < new Date();

  return (
    <>
      {showUpgradeModal && wallet && (
        <TierUpgradeModal
          tierInfo={tierInfo}
          walletBalance={typeof wallet.master_balance === 'string' ? parseFloat(wallet.master_balance) : wallet.master_balance}
          onClose={() => setShowUpgradeModal(false)}
          onSuccess={fetchData}
        />
      )}

      <div className={`p-3 rounded-xl border ${
        isGold && !isExpired
          ? 'bg-gradient-to-r from-yellow-500/10 to-orange-500/10 border-yellow-500/30'
          : 'bg-white/5 border-white/10'
      }`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {isGold && !isExpired && <Crown className="w-4 h-4 text-yellow-400" />}
            <div>
              <p className="text-xs text-slate-400">Your Tier</p>
              <p className={`text-sm font-bold ${
                isGold && !isExpired ? 'text-yellow-400' : 'text-white'
              }`}>
                {isGold && !isExpired ? 'GOLD' : 'STANDARD'}
              </p>
            </div>
          </div>

          <div className="text-right">
            <p className="text-xs text-slate-400">Cards</p>
            <p className="text-sm font-bold text-white">
              {tierInfo.current_cards}/{tierInfo.card_limit}
            </p>
          </div>
        </div>

        {(!isGold || isExpired) && (
          <button
            onClick={() => setShowUpgradeModal(true)}
            className="w-full mt-3 py-2 bg-gradient-to-r from-yellow-500/20 to-orange-500/20 hover:from-yellow-500/30 hover:to-orange-500/30 text-yellow-400 rounded-lg transition-all text-xs font-semibold border border-yellow-500/30 flex items-center justify-center gap-1"
          >
            <Crown className="w-3.5 h-3.5" />
            Upgrade to GOLD
            <ArrowUpRight className="w-3.5 h-3.5" />
          </button>
        )}

        {isGold && !isExpired && tierInfo.tier_expires_at && (
          <p className="text-xs text-slate-400 mt-2 text-center">
            Expires: {new Date(tierInfo.tier_expires_at).toLocaleDateString()}
          </p>
        )}
      </div>
    </>
  );
};
