import { useState } from 'react';
import { X } from 'lucide-react';
import { useRates } from '../store/rates-context';
import { ModalPortal } from './modal-portal';

// Inline SVG bank logos
const SbpLogo = () => (
  <svg viewBox="0 0 40 40" className="w-7 h-7">
    <defs>
      <linearGradient id="vt-sbp" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0%" stopColor="#5B57A2" />
        <stop offset="35%" stopColor="#D90751" />
        <stop offset="65%" stopColor="#FAB718" />
        <stop offset="100%" stopColor="#0FA8D6" />
      </linearGradient>
    </defs>
    <rect rx="8" width="40" height="40" fill="url(#vt-sbp)" />
    <text x="20" y="26" textAnchor="middle" fill="white" fontSize="14" fontWeight="700" fontFamily="system-ui">СБП</text>
  </svg>
);

const bankIcons = [
  { id: 'sber', el: (
    <svg viewBox="0 0 40 40" className="w-full h-full">
      <circle cx="20" cy="20" r="20" fill="#21A038" />
      <path d="M20 8 L20 20 L30 20" stroke="white" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round" fill="none" />
      <circle cx="20" cy="20" r="11" stroke="white" strokeWidth="2.5" fill="none" />
    </svg>
  )},
  { id: 'tbank', el: (
    <svg viewBox="0 0 40 40" className="w-full h-full">
      <rect rx="8" width="40" height="40" fill="#FFDD2D" />
      <text x="20" y="27" textAnchor="middle" fill="#333" fontSize="20" fontWeight="800" fontFamily="system-ui">T</text>
    </svg>
  )},
  { id: 'alfa', el: (
    <svg viewBox="0 0 40 40" className="w-full h-full">
      <rect rx="8" width="40" height="40" fill="#EF3124" />
      <text x="20" y="28" textAnchor="middle" fill="white" fontSize="22" fontWeight="800" fontFamily="system-ui">A</text>
    </svg>
  )},
  { id: 'vtb', el: (
    <svg viewBox="0 0 40 40" className="w-full h-full">
      <rect rx="8" width="40" height="40" fill="#002882" />
      <rect x="8" y="12" width="24" height="3.5" rx="1.5" fill="white" />
      <rect x="8" y="18.5" width="24" height="3.5" rx="1.5" fill="white" />
      <rect x="8" y="25" width="24" height="3.5" rx="1.5" fill="white" />
    </svg>
  )},
];

interface VaultTopUpModalProps {
  onClose: () => void;
}

export const VaultTopUpModal = ({ onClose }: VaultTopUpModalProps) => {
  const { rates } = useRates();
  const [selectedCurrency, setSelectedCurrency] = useState<'USD' | 'EUR'>('USD');
  const [foreignAmount, setForeignAmount] = useState('');
  const [activeRubPreset, setActiveRubPreset] = useState<number | null>(null);

  const currentRate = selectedCurrency === 'USD' ? rates.usd : rates.eur;
  const currencySymbol = selectedCurrency === 'USD' ? '$' : '€';
  const rubAmount = foreignAmount && !isNaN(parseFloat(foreignAmount))
    ? (parseFloat(foreignAmount) * currentRate).toFixed(0)
    : '';

  const rubPresets = [1000, 2000, 5000, 10000, 20000, 50000, 100000];

  const handleRubPreset = (rub: number) => {
    setActiveRubPreset(rub);
    const converted = (rub / currentRate).toFixed(2);
    setForeignAmount(converted);
  };

  const handleForeignInput = (val: string) => {
    setForeignAmount(val);
    setActiveRubPreset(null);
  };

  return (
    <ModalPortal>
      <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-0 sm:p-4">
        <div className="absolute inset-0 bg-black/60 backdrop-blur-md" onClick={onClose} />

        <div className="relative bg-[#0b0b14]/95 border border-white/[0.10] rounded-t-2xl sm:rounded-2xl w-full max-w-[400px] max-h-[90vh] overflow-hidden flex flex-col animate-scale-in shadow-[0_24px_80px_-12px_rgba(0,0,0,0.8)]">
          {/* Glass accent */}
          <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-amber-400/40 to-transparent rounded-t-2xl" />

          {/* Header */}
          <div className="shrink-0 px-5 py-4 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-amber-500/20 to-orange-500/20 border border-amber-500/30 flex items-center justify-center">
                <svg viewBox="0 0 20 20" className="w-4 h-4 text-amber-400" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M10 2v16M6 6l4-4 4 4M4 10h12" />
                </svg>
              </div>
              <div>
                <h3 className="text-base font-semibold text-white leading-tight">Пополнить Сейф</h3>
                <p className="text-[11px] text-slate-400">Через Систему быстрых платежей</p>
              </div>
            </div>
            <button onClick={onClose} className="p-1.5 hover:bg-white/10 rounded-lg transition-colors">
              <X className="w-5 h-5 text-slate-400" />
            </button>
          </div>

          {/* Body — no scroll */}
          <div className="shrink-0 px-5 pb-3 space-y-3">
            {/* Currency toggle */}
            <div className="flex gap-2">
              <button
                onClick={() => { setSelectedCurrency('USD'); setActiveRubPreset(null); setForeignAmount(''); }}
                className={`flex-1 py-2 rounded-xl text-sm font-semibold transition-all text-center ${
                  selectedCurrency === 'USD'
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                    : 'bg-white/[0.04] text-slate-400 border border-white/[0.08] hover:bg-white/[0.08]'
                }`}
              >$ USD</button>
              <button
                onClick={() => { setSelectedCurrency('EUR'); setActiveRubPreset(null); setForeignAmount(''); }}
                className={`flex-1 py-2 rounded-xl text-sm font-semibold transition-all text-center ${
                  selectedCurrency === 'EUR'
                    ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                    : 'bg-white/[0.04] text-slate-400 border border-white/[0.08] hover:bg-white/[0.08]'
                }`}
              >€ EUR</button>
            </div>

            {/* Amount input */}
            <div>
              <label className="block text-xs text-slate-400 mb-1">Сумма в валюте</label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-blue-400 text-lg font-bold">{currencySymbol}</span>
                <input
                  type="number"
                  inputMode="decimal"
                  placeholder="0.00"
                  value={foreignAmount}
                  onChange={(e) => handleForeignInput(e.target.value)}
                  className="w-full h-12 pl-12 pr-4 bg-white/[0.04] border border-white/[0.10] rounded-xl text-white text-xl font-bold focus:outline-none focus:border-blue-400 focus:ring-1 focus:ring-blue-400/50 transition-colors placeholder:text-slate-600"
                />
              </div>
            </div>

            {/* RUB quick-select presets */}
            <div>
              <label className="block text-xs text-slate-400 mb-1.5">Быстрый выбор в ₽</label>
              <div className="flex flex-wrap gap-1.5">
                {rubPresets.map(rub => (
                  <button
                    key={rub}
                    onClick={() => handleRubPreset(rub)}
                    className={`px-2.5 py-1 rounded-lg text-xs font-medium transition-all ${
                      activeRubPreset === rub
                        ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                        : 'bg-white/[0.04] text-slate-400 border border-white/[0.08] hover:bg-white/[0.08]'
                    }`}
                  >
                    {rub >= 1000 ? `${rub / 1000}k` : rub} ₽
                  </button>
                ))}
              </div>
            </div>

            {/* RUB result + rate */}
            <div className="flex items-center gap-2">
              <div className="flex-1 relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-white/50 text-sm font-bold">₽</span>
                <input
                  type="text"
                  readOnly
                  value={rubAmount ? Number(rubAmount).toLocaleString('ru-RU') : ''}
                  placeholder="0"
                  className="w-full h-11 pl-9 pr-3 bg-white/[0.04] border border-white/[0.08] rounded-xl text-white text-base font-semibold placeholder:text-slate-600 cursor-default"
                />
              </div>
              <div className="px-3 py-2 rounded-xl bg-white/[0.04] border border-white/[0.06] text-[11px] text-slate-400 whitespace-nowrap">
                1 {selectedCurrency} = {currentRate.toFixed(2)} ₽
              </div>
            </div>
          </div>

          {/* Footer — СБП button + bank icons */}
          <div className="shrink-0 px-5 pt-2 pb-5 flex flex-col gap-2.5">
            <button
              onClick={onClose}
              disabled={!foreignAmount || parseFloat(foreignAmount) <= 0}
              className="w-full py-4 bg-gradient-to-r from-blue-500 to-indigo-600 hover:from-blue-400 hover:to-indigo-500 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/20 hover:shadow-blue-500/30 disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2.5 text-base"
            >
              <SbpLogo />
              <span>Пополнить через СБП{rubAmount ? ` — ${Number(rubAmount).toLocaleString('ru-RU')} ₽` : ''}</span>
            </button>
            {/* Bank icons row */}
            <div className="flex items-center justify-center gap-3">
              <span className="text-[10px] text-slate-500">Поддерживаемые банки:</span>
              <div className="flex items-center gap-1.5">
                {bankIcons.map((b) => (
                  <div key={b.id} className="w-5 h-5 rounded-[4px] overflow-hidden opacity-60">
                    {b.el}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>

      <style>{`
        @keyframes scale-in { from { opacity: 0; transform: scale(0.95); } to { opacity: 1; transform: scale(1); } }
        .animate-scale-in { animation: scale-in 0.15s ease-out forwards; }
      `}</style>
    </ModalPortal>
  );
};
