import { useState } from 'react';
import { X, Crown, Check, CreditCard } from 'lucide-react';
import { ModalPortal } from './modal-portal';
import { upgradeTier, type TierInfo } from '../api/tier';

interface TierUpgradeModalProps {
  tierInfo: TierInfo;
  walletBalance: number;
  onClose: () => void;
  onSuccess: () => void;
}

export const TierUpgradeModal = ({ tierInfo, walletBalance, onClose, onSuccess }: TierUpgradeModalProps) => {
  const [isUpgrading, setIsUpgrading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const goldPrice = parseFloat(tierInfo.gold_price);
  const canAfford = walletBalance >= goldPrice;

  const handleUpgrade = async () => {
    if (!canAfford || isUpgrading) return;

    setIsUpgrading(true);
    setError(null);

    try {
      await upgradeTier();
      onSuccess();
      onClose();
    } catch (err: any) {
      setError(err.response?.data || 'Failed to upgrade tier');
      setIsUpgrading(false);
    }
  };

  return (
    <ModalPortal>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
        
        <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 rounded-2xl w-full max-w-[480px] animate-scale-in shadow-2xl shadow-black/60">
          {/* Header */}
          <div className="p-6 pb-4 border-b border-white/[0.06]">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-yellow-500/20 to-orange-500/20 border border-yellow-500/30 flex items-center justify-center">
                  <Crown className="w-6 h-6 text-yellow-400" />
                </div>
                <div>
                  <h2 className="text-xl font-bold text-white">Upgrade to GOLD</h2>
                  <p className="text-xs text-slate-400">Premium tier with extended limits</p>
                </div>
              </div>
              <button onClick={onClose} className="p-1.5 hover:bg-white/10 rounded-lg transition-colors">
                <X className="w-5 h-5 text-slate-400" />
              </button>
            </div>
          </div>

          {/* Body */}
          <div className="p-6 space-y-4">
            {/* Comparison */}
            <div className="grid grid-cols-2 gap-3">
              {/* Standard */}
              <div className="p-4 bg-white/5 rounded-xl border border-white/10">
                <h3 className="text-sm font-semibold text-slate-400 mb-3">Standard</h3>
                <ul className="space-y-2">
                  <li className="flex items-center gap-2 text-xs text-slate-400">
                    <div className="w-1.5 h-1.5 rounded-full bg-slate-500" />
                    Up to 3 cards
                  </li>
                  <li className="flex items-center gap-2 text-xs text-slate-400">
                    <div className="w-1.5 h-1.5 rounded-full bg-slate-500" />
                    Standard support
                  </li>
                </ul>
              </div>

              {/* Gold */}
              <div className="p-4 bg-gradient-to-br from-yellow-500/10 to-orange-500/10 rounded-xl border border-yellow-500/30">
                <h3 className="text-sm font-semibold text-yellow-400 mb-3 flex items-center gap-1">
                  <Crown className="w-4 h-4" />
                  Gold
                </h3>
                <ul className="space-y-2">
                  <li className="flex items-center gap-2 text-xs text-white">
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    Up to 15 cards
                  </li>
                  <li className="flex items-center gap-2 text-xs text-white">
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    Priority support
                  </li>
                  <li className="flex items-center gap-2 text-xs text-white">
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    Extended limits
                  </li>
                </ul>
              </div>
            </div>

            {/* Pricing */}
            <div className="p-4 bg-blue-500/10 rounded-xl border border-blue-500/30">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-slate-400">Price</span>
                <span className="text-2xl font-bold text-white">${goldPrice.toFixed(2)}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Duration</span>
                <span className="text-white">{tierInfo.gold_duration} days</span>
              </div>
            </div>

            {/* Wallet Balance */}
            <div className="p-3 bg-white/5 rounded-xl border border-white/10">
              <div className="flex items-center justify-between">
                <span className="text-xs text-slate-400">Your Wallet Balance</span>
                <span className={`text-sm font-bold ${canAfford ? 'text-emerald-400' : 'text-red-400'}`}>
                  ${walletBalance.toFixed(2)}
                </span>
              </div>
            </div>

            {/* Error */}
            {error && (
              <div className="p-3 bg-red-500/10 rounded-xl border border-red-500/30">
                <p className="text-sm text-red-400">{error}</p>
              </div>
            )}

            {/* Insufficient funds warning */}
            {!canAfford && (
              <div className="p-3 bg-orange-500/10 rounded-xl border border-orange-500/30">
                <p className="text-sm text-orange-400">
                  Insufficient funds. Please top up your wallet to upgrade.
                </p>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="p-6 pt-3 border-t border-white/[0.06]">
            <button 
              onClick={handleUpgrade}
              disabled={!canAfford || isUpgrading}
              className="w-full py-3.5 bg-gradient-to-r from-yellow-500 to-orange-600 hover:from-yellow-400 hover:to-orange-500 text-white font-semibold rounded-xl transition-all shadow-lg shadow-yellow-500/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {isUpgrading ? (
                'Processing...'
              ) : (
                <>
                  <CreditCard className="w-5 h-5" />
                  Upgrade for ${goldPrice.toFixed(2)}
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
};
