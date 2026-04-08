import { useState, useEffect } from 'react';
import { Crown, ArrowUpRight, RefreshCw } from 'lucide-react';
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
      console.error('[TierBadge] Failed to fetch tier info:', error);
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

  // Safe date parsing — handles ISO string, NullTime object {Time, Valid}, or null
  const parseExpiry = (raw: any): Date | null => {
    if (!raw) return null;
    // Handle Go sql.NullTime leak: {Time: "...", Valid: true}
    if (typeof raw === 'object' && raw.Valid === true && raw.Time) {
      const d = new Date(raw.Time);
      return isNaN(d.getTime()) ? null : d;
    }
    if (typeof raw === 'string') {
      const d = new Date(raw);
      return isNaN(d.getTime()) ? null : d;
    }
    return null;
  };

  const isGold = tierInfo.tier === 'gold';
  const expiryDate = parseExpiry(tierInfo.tier_expires_at);
  const isExpired = expiryDate ? expiryDate < new Date() : true;
  const isActiveGold = isGold && !isExpired && !!expiryDate;

  // Calculate days remaining for Gold users
  let daysLeft = 0;
  if (isActiveGold && expiryDate) {
    daysLeft = Math.ceil((expiryDate.getTime() - Date.now()) / (1000 * 60 * 60 * 24));
  }

  // Color logic: <5 days = red, <30 days = orange, else normal
  const expiryColor = daysLeft <= 5 ? 'text-red-400' : daysLeft <= 30 ? 'text-orange-400' : 'text-slate-400';
  const showExtendButton = isActiveGold && daysLeft <= 30;

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
        isActiveGold
          ? 'bg-gradient-to-r from-yellow-500/10 to-orange-500/10 border-yellow-500/30'
          : 'bg-white/5 border-white/10'
      }`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {isActiveGold && <Crown className="w-4 h-4 text-yellow-400" />}
            <div>
              <p className="text-xs text-slate-400">Ваш уровень</p>
              <p className={`text-sm font-bold ${
                isActiveGold ? 'text-yellow-400' : 'text-white'
              }`}>
                {isActiveGold ? 'GOLD' : 'СТАНДАРТ'}
              </p>
            </div>
          </div>

          <div className="text-right">
            <p className="text-xs text-slate-400">Карт</p>
            <p className="text-sm font-bold text-white">
              {tierInfo.current_cards}/{tierInfo.card_limit}
            </p>
          </div>
        </div>

        {/* Gold expiry date with color coding */}
        {isActiveGold && expiryDate && (
          <div className="mt-2 text-center">
            <p className={`text-xs font-medium ${expiryColor}`}>
              Активен до: {expiryDate.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' })}
              {daysLeft <= 30 && (
                <span className="ml-1">({daysLeft} {daysLeft === 1 ? 'день' : daysLeft >= 2 && daysLeft <= 4 ? 'дня' : 'дней'})</span>
              )}
            </p>
          </div>
        )}

        {/* Extend button for active Gold users nearing expiry */}
        {showExtendButton && (
          <button
            onClick={() => setShowUpgradeModal(true)}
            className="w-full mt-2 py-2 bg-gradient-to-r from-orange-500/20 to-red-500/20 hover:from-orange-500/30 hover:to-red-500/30 text-orange-400 rounded-lg transition-all text-xs font-semibold border border-orange-500/30 flex items-center justify-center gap-1"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            Продлить Gold
          </button>
        )}

        {/* Upgrade button for Standard users */}
        {(!isGold || isExpired) && (
          <button
            onClick={() => setShowUpgradeModal(true)}
            className="w-full mt-3 py-2 bg-gradient-to-r from-yellow-500/20 to-orange-500/20 hover:from-yellow-500/30 hover:to-orange-500/30 text-yellow-400 rounded-lg transition-all text-xs font-semibold border border-yellow-500/30 flex items-center justify-center gap-1"
          >
            <Crown className="w-3.5 h-3.5" />
            Улучшить до GOLD
            <ArrowUpRight className="w-3.5 h-3.5" />
          </button>
        )}
      </div>
    </>
  );
};
