import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ShoppingBag, Search, Globe, Gamepad2, X, Copy, Check, Loader2,
  AlertTriangle, QrCode, Key, ChevronRight, ArrowLeft,
  Download, Smartphone, Wifi, CreditCard
} from 'lucide-react';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import {
  getStoreCatalog, purchaseProduct,
  getESIMDestinations, getESIMPlans, orderESIM,
  type StoreProduct, type ESIMDestination, type ESIMPlan, type ESIMOrderResult,
} from '../services/store';
import { getUserCards, type Card } from '../services/cards';

// ── Country flag (SVG via flagcdn — cross-platform) ──
const CountryFlag = ({ code, size = 24 }: { code: string; size?: number }) => {
  if (!code || code.length < 2 || code === 'GLOBAL') {
    return <Globe className="inline-block text-slate-400" style={{ width: size, height: size }} />;
  }
  return (
    <img
      src={`https://flagcdn.com/w80/${code.toLowerCase()}.png`}
      srcSet={`https://flagcdn.com/w160/${code.toLowerCase()}.png 2x`}
      alt={code}
      style={{ width: size, height: Math.round(size * 0.75) }}
      className="inline-block rounded-[3px] object-cover shadow-sm"
      loading="lazy"
      onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
    />
  );
};

// ── SVG Brand Icons (inline, no external deps) ──
const SteamIcon = ({ className = 'w-8 h-8' }: { className?: string }) => (
  <svg viewBox="0 0 256 256" className={className} fill="none">
    <defs><linearGradient id="steam-g" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" stopColor="#111D2E"/><stop offset="50%" stopColor="#0A1B48"/><stop offset="100%" stopColor="#1387B8"/></linearGradient></defs>
    <circle cx="128" cy="128" r="128" fill="url(#steam-g)"/>
    <path d="M128.079 57.604c28.299 0 51.364 22.544 52.253 50.664l.033 1.58-30.474 17.747a37.406 37.406 0 0 0-14.6-2.958c-1.238 0-2.46.063-3.667.182l-20.378-29.473-.002-.514c0-20.672 16.78-37.228 37.43-37.228h-.595Zm22.263 56.444 16.68-9.72c-2.74-14.14-15.21-24.858-30.148-24.858-16.899 0-30.6 13.632-30.728 30.474l20.607 14.73a26.468 26.468 0 0 1 10.198-2.041c4.84 0 9.335 1.286 13.391 4.415ZM91.51 151.773l-14.682-6.065c2.657 7.853 8.788 14.306 16.573 17.498 16.346 6.695 35.093-1.171 41.811-17.546a30.49 30.49 0 0 0 1.758-15.395 30.567 30.567 0 0 0-6.298-14.44l-16.396 9.596a19.478 19.478 0 0 1-7.25 21.516 19.39 19.39 0 0 1-15.516 4.836Z" fill="white"/>
  </svg>
);
const PlayStationIcon = ({ className = 'w-8 h-8' }: { className?: string }) => (
  <svg viewBox="0 0 256 256" className={className}><rect width="256" height="256" rx="40" fill="#003791"/><path d="M99.8 166.4V85.6l36.4-12v119.2l-36.4 12.8V166.4Zm77.6-22.8c-6-3.6-13.6-5.2-24-3.6l-17.2 4.8v29.6l12-3.6c6.4-2 12-1.2 12 4.8s-5.6 10.8-12 12.8l-12 4V209l11.6-4c13.6-4.8 24.8-11.6 30.4-20.8 6-9.6 5.6-22-.8-30l-.8.4h.8ZM72.6 194.8l36.4-12.8v-20.4l-22 7.6c-6.4 2-12 1.2-12-4.8s5.6-10.8 12-12.8l22-7.6v-18l-12 4c-13.6 4.8-24.8 11.6-30.4 20.8-5.6 10-5.2 22.4 1.2 30.8l4.8 2.8v10.4Z" fill="white"/></svg>
);
const XboxIcon = ({ className = 'w-8 h-8' }: { className?: string }) => (
  <svg viewBox="0 0 256 256" className={className}><rect width="256" height="256" rx="40" fill="#107C10"/><path d="M128 44c-8.4 0-16.4 1.2-24 3.4 8-2 19.2 4.4 24 8.4 4.8-4 16-10.4 24-8.4-7.6-2.2-15.6-3.4-24-3.4Zm-50.4 20.8C65.2 76.4 56 94.8 52.4 114c-3.2 17.2-1.6 34.4 4 48.8 12-20 28.8-40 47.2-56.8-8.4-12-18.4-28-26-41.2h.4Zm100.8 0c-7.6 13.2-17.6 29.2-26 41.2 18.4 16.8 35.2 36.8 47.2 56.8 5.6-14.4 7.2-31.6 4-48.8-3.6-19.2-12.8-37.6-25.2-49.2Zm-50.4 52.8c-20.4 18.8-38.8 41.2-50 62.4 14.4 14.8 34 24.4 56 25.6V208c-2 0-3.6-.4-5.6-.4-17.6-1.6-34-10.4-44.8-24.4 12.8-20.4 28.4-41.6 44.4-56.8v-8.8Zm0 0c16 15.2 31.6 36.4 44.4 56.8-10.8 14-27.2 22.8-44.8 24.4-2 0-3.6.4-5.6.4v-2.4c22-1.2 41.6-10.8 56-25.6-11.2-21.2-29.6-43.6-50-53.6Z" fill="white"/></svg>
);
const NintendoIcon = ({ className = 'w-8 h-8' }: { className?: string }) => (
  <svg viewBox="0 0 256 256" className={className}><rect width="256" height="256" rx="40" fill="#E60012"/><path d="M92 60h-16c-17.6 0-32 14.4-32 32v72c0 17.6 14.4 32 32 32h16c4.4 0 8-3.6 8-8V68c0-4.4-3.6-8-8-8Zm-16 112c-8.8 0-16-7.2-16-16s7.2-16 16-16 16 7.2 16 16-7.2 16-16 16Zm104-112h-16c-4.4 0-8 3.6-8 8v120c0 4.4 3.6 8 8 8h16c17.6 0 32-14.4 32-32V92c0-17.6-14.4-32-32-32Z" fill="white"/></svg>
);
const SpotifyIcon = ({ className = 'w-8 h-8' }: { className?: string }) => (
  <svg viewBox="0 0 256 256" className={className}><circle cx="128" cy="128" r="128" fill="#1DB954"/><path d="M186.8 118c-23.2-13.8-61.4-15-83.6-8.3-3.6 1.1-7.3-.9-8.4-4.5s.9-7.3 4.5-8.4c25.5-7.7 67.9-6.2 94.7 9.6 3.2 1.9 4.2 6 2.3 9.2-1.9 3.1-6 4.2-9.2 2.3l-.3.1Zm-2-21.7c-27.4-16.3-68.5-17.7-98.6-9.8-4.2 1.1-8.5-1.3-9.7-5.5-1.1-4.2 1.3-8.5 5.5-9.7 34.4-9 86.3-7.3 117.6 11.4 3.7 2.2 5 7.1 2.8 10.8-2.2 3.7-7 5-10.8 2.8h1.2Zm-10.7 41.2c-19-11.3-50.4-12.3-68.5-6.8-2.9.9-6-.7-6.9-3.7-.9-2.9.7-6 3.7-6.9 20.8-6.3 55.5-5.1 77.3 7.9 2.6 1.6 3.5 5 1.9 7.6-1.6 2.6-5 3.4-7.5 1.9Z" fill="white"/></svg>
);
const NetflixIcon = ({ className = 'w-8 h-8' }: { className?: string }) => (
  <svg viewBox="0 0 256 256" className={className}><rect width="256" height="256" rx="40" fill="#E50914"/><path d="M88 60v136l36-60V60H88Zm44 0v136l36-60V60h-36Zm0 76v60l36-60v-60l-36 60Z" fill="white"/></svg>
);
const brandIcons: Record<string, React.FC<{ className?: string }>> = {
  steam: SteamIcon, playstation: PlayStationIcon, xbox: XboxIcon,
  nintendo: NintendoIcon, spotify: SpotifyIcon, netflix: NetflixIcon,
};
const brandIconForProduct = (name: string, externalId?: string): React.FC<{ className?: string }> | null => {
  const key = (externalId?.split('-')[0] || name.split(' ')[0]).toLowerCase();
  const map: Record<string, string> = {
    steam: 'steam', psn: 'playstation', playstation: 'playstation', xbox: 'xbox',
    nintendo: 'nintendo', spotify: 'spotify', netflix: 'netflix',
  };
  return brandIcons[map[key] || ''] || null;
};

// ── Error toast ──
const ErrorToast = ({ message, onClose }: { message: string; onClose: () => void }) => (
  <div className="fixed bottom-24 left-1/2 -translate-x-1/2 z-[110] max-w-sm w-full mx-4">
    <div className="flex items-center gap-3 px-4 py-3 rounded-xl bg-red-500/10 border border-red-500/30 backdrop-blur-xl shadow-2xl">
      <AlertTriangle className="w-5 h-5 text-red-400 shrink-0" />
      <p className="text-sm text-red-300 flex-1">{message}</p>
      <button onClick={onClose} className="shrink-0 p-1 rounded-lg hover:bg-white/5 text-slate-500 hover:text-white transition-all">
        <X className="w-4 h-4" />
      </button>
    </div>
  </div>
);

// ══════════════════════════════════════════════════════════════
// eSIM Activation Modal (QR + LPA + download)
// ══════════════════════════════════════════════════════════════
const ESIMActivationModal = ({
  result, planName, onClose,
}: {
  result: ESIMOrderResult; planName: string; onClose: () => void;
}) => {
  const [copied, setCopied] = useState('');
  const qrRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', h);
    document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose]);

  const copyText = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => { setCopied(label); setTimeout(() => setCopied(''), 2000); });
  };

  const downloadInstructions = () => {
    const content = `XPLR eSIM — Инструкция по активации\n=====================================\n\nТовар: ${planName}\nЦена: $${result.price_usd}\nICCID: ${result.iccid || 'N/A'}\nДата: ${new Date().toLocaleString('ru-RU')}\n\nСПОСОБ 1: QR-код\nНастройки → Сотовая связь → Добавить eSIM → Сканировать QR-код.\n\nСПОСОБ 2: Ручная установка\nSM-DP+: ${result.smdp || 'N/A'}\nКод активации: ${result.matching_id || 'N/A'}\nLPA: ${result.lpa || result.qr_data || 'N/A'}\n\nПоддержка: https://xplr.pro/support`;
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a'); a.href = url; a.download = `XPLR_eSIM_${result.iccid || 'activation'}.txt`;
    document.body.appendChild(a); a.click(); a.remove(); URL.revokeObjectURL(url);
  };

  const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(result.qr_data || result.lpa || '')}`;
  const lpaString = result.lpa || result.qr_data || '';
  const smdp = result.smdp || (lpaString.includes('$') ? lpaString.split('$')[1] : '');
  const matchingId = result.matching_id || (lpaString.includes('$') ? lpaString.split('$')[2] : '');

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div className="relative z-10 w-full max-w-lg max-h-[92vh] overflow-y-auto rounded-2xl bg-[#0F1629] border border-white/[0.08] shadow-2xl animate-scale-in" onClick={e => e.stopPropagation()}>
        <button onClick={onClose} className="absolute top-3 right-3 z-10 p-2 rounded-xl bg-black/50 border border-white/[0.08] text-white/40 hover:text-white transition-all"><X className="w-4 h-4" /></button>
        <div className="p-6 sm:p-8">
          <div className="text-center mb-6">
            <div className="flex items-center justify-center w-16 h-16 rounded-2xl bg-emerald-500/10 mx-auto mb-3">
              <Check className="w-8 h-8 text-emerald-400" />
            </div>
            <h2 className="text-lg font-bold text-white mb-1">eSIM активирована!</h2>
            <p className="text-sm text-white/40">{planName} — ${result.price_usd}</p>
          </div>
          <div ref={qrRef} className="text-center mb-6">
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-[#38BDF8]/10 border border-[#38BDF8]/20 text-[#38BDF8] text-xs font-medium mb-4">
              <QrCode className="w-3.5 h-3.5" /> Отсканируйте для установки
            </div>
            <div className="bg-white rounded-2xl p-5 inline-block mx-auto shadow-lg"><img src={qrUrl} alt="QR" className="w-[200px] h-[200px]" /></div>
          </div>
          <div className="space-y-3 mb-6">
            <div className="flex items-center gap-2 text-sm font-medium text-white/70"><Smartphone className="w-4 h-4 text-[#38BDF8]" /> Ручная установка</div>
            <p className="text-xs text-white/30">Настройки → Сотовая связь → Добавить eSIM → Ввести данные вручную</p>
            {smdp && (
              <div className="bg-white/[0.04] border border-white/[0.05] rounded-xl p-3">
                <p className="text-[10px] text-white/30 mb-1.5 uppercase tracking-wider">SM-DP+ адрес</p>
                <div className="flex items-center justify-between gap-2">
                  <code className="text-xs text-white font-mono break-all flex-1">{smdp}</code>
                  <button onClick={() => copyText(smdp, 'smdp')} className="shrink-0 p-1.5 rounded-lg bg-[#38BDF8]/10 border border-[#38BDF8]/20 text-[#38BDF8] hover:bg-[#38BDF8]/20 transition-all">
                    {copied === 'smdp' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            )}
            {matchingId && (
              <div className="bg-white/[0.04] border border-white/[0.05] rounded-xl p-3">
                <p className="text-[10px] text-white/30 mb-1.5 uppercase tracking-wider">Код активации</p>
                <div className="flex items-center justify-between gap-2">
                  <code className="text-xs text-white font-mono break-all flex-1">{matchingId}</code>
                  <button onClick={() => copyText(matchingId, 'mid')} className="shrink-0 p-1.5 rounded-lg bg-[#38BDF8]/10 border border-[#38BDF8]/20 text-[#38BDF8] hover:bg-[#38BDF8]/20 transition-all">
                    {copied === 'mid' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            )}
            {lpaString && (
              <div className="bg-white/[0.04] border border-white/[0.05] rounded-xl p-3">
                <p className="text-[10px] text-white/30 mb-1.5 uppercase tracking-wider">LPA строка</p>
                <div className="flex items-center justify-between gap-2">
                  <code className="text-[10px] text-white/60 font-mono break-all flex-1">{lpaString}</code>
                  <button onClick={() => copyText(lpaString, 'lpa')} className="shrink-0 p-1.5 rounded-lg bg-[#38BDF8]/10 border border-[#38BDF8]/20 text-[#38BDF8] hover:bg-[#38BDF8]/20 transition-all">
                    {copied === 'lpa' ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            )}
          </div>
          <div className="flex gap-3">
            <button onClick={downloadInstructions} className="flex-1 py-3 rounded-xl bg-white/[0.04] border border-white/[0.06] text-white/60 text-sm font-medium hover:bg-white/[0.08] transition-all flex items-center justify-center gap-2"><Download className="w-4 h-4" /> Скачать</button>
            <button onClick={onClose} className="flex-1 py-3 rounded-xl bg-gradient-to-r from-[#38BDF8] to-[#A78BFA] text-white text-sm font-semibold hover:opacity-90 transition-all">Готово</button>
          </div>
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// Digital Product Result Modal
// ══════════════════════════════════════════════════════════════
const DigitalResultModal = ({ productName, priceUsd, activationKey, onClose }: { productName: string; priceUsd: string; activationKey: string; onClose: () => void }) => {
  const [copied, setCopied] = useState(false);
  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', h); document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div className="relative z-10 w-full max-w-sm rounded-2xl bg-[#0F1629] border border-white/[0.08] p-6 sm:p-8 animate-scale-in" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-center w-14 h-14 rounded-2xl bg-emerald-500/10 mb-5 mx-auto">
          <Check className="w-7 h-7 text-emerald-400" />
        </div>
        <h3 className="text-lg font-bold text-white text-center mb-1">Покупка успешна!</h3>
        <p className="text-sm text-white/40 text-center mb-6">{productName} — ${priceUsd}</p>
        {activationKey && (
          <div className="mb-6">
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-[#A78BFA]/10 border border-[#A78BFA]/20 text-[#A78BFA] text-xs font-medium mb-3"><Key className="w-3.5 h-3.5" /> Ключ активации</div>
            <div className="bg-white/[0.04] border border-white/[0.05] rounded-xl p-4 flex items-center justify-between gap-3">
              <code className="text-white font-mono text-sm break-all flex-1 text-left">{activationKey}</code>
              <button onClick={() => { navigator.clipboard.writeText(activationKey); setCopied(true); setTimeout(() => setCopied(false), 2000); }} className="shrink-0 p-2.5 rounded-lg bg-[#38BDF8]/10 border border-[#38BDF8]/20 text-[#38BDF8] hover:bg-[#38BDF8]/20 transition-all">
                {copied ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
              </button>
            </div>
          </div>
        )}
        <button onClick={onClose} className="w-full py-3 rounded-xl bg-white/[0.04] border border-white/[0.06] text-white/60 text-sm font-medium hover:bg-white/[0.08] transition-all">Закрыть</button>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// Confirm Purchase Modal (unified for eSIM + Digital)
// ══════════════════════════════════════════════════════════════
const ConfirmPurchaseModal = ({
  title, itemLabel, priceLabel, loading, onConfirm, onClose, allCards, onCardChange,
}: {
  title: string; itemLabel: string; priceLabel: string; loading: boolean;
  onConfirm: () => void; onClose: () => void;
  allCards: Card[]; onCardChange: (card: Card) => void;
}) => {
  const [selectedIdx, setSelectedIdx] = useState(0);
  const card = allCards[selectedIdx] || null;

  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape' && !loading) onClose(); };
    document.addEventListener('keydown', h); document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose, loading]);

  const handleCardSwitch = (idx: number) => {
    setSelectedIdx(idx);
    if (allCards[idx]) onCardChange(allCards[idx]);
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={() => !loading && onClose()}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div className="relative z-10 w-full max-w-sm rounded-2xl bg-[#0F1629] border border-white/[0.08] p-6 sm:p-8 animate-scale-in" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-center w-14 h-14 rounded-2xl bg-emerald-500/10 mb-5 mx-auto">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" className="text-emerald-400"><path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/></svg>
        </div>
        <h3 className="text-lg font-bold text-white text-center mb-4">{title}</h3>
        <div className="rounded-xl bg-white/[0.04] border border-white/[0.05] p-4 mb-4 space-y-2.5">
          <div className="flex justify-between items-center">
            <span className="text-sm text-white/40">Товар</span>
            <span className="text-sm text-white font-medium text-right max-w-[180px] truncate">{itemLabel}</span>
          </div>
          {/* Card selector */}
          <div className="flex justify-between items-center">
            <span className="text-sm text-white/40">Карта</span>
            {allCards.length <= 1 && card ? (
              <span className="text-sm text-white font-medium flex items-center gap-1.5">
                <CreditCard className="w-3.5 h-3.5 text-[#38BDF8]" />
                {card.nickname ? `${card.nickname} ` : ''}•••• {card.last_4_digits}
              </span>
            ) : (
              <select
                value={selectedIdx}
                onChange={e => handleCardSwitch(Number(e.target.value))}
                disabled={loading}
                className="bg-white/[0.06] border border-white/[0.08] rounded-lg px-2 py-1 text-sm text-white font-medium outline-none focus:border-[#38BDF8]/50 transition-colors cursor-pointer max-w-[200px] truncate"
              >
                {allCards.map((c, i) => (
                  <option key={c.id} value={i} className="bg-[#0F1629] text-white">
                    {c.nickname ? `${c.nickname} ` : ''}•••• {c.last_4_digits}
                  </option>
                ))}
              </select>
            )}
          </div>
          <div className="border-t border-white/[0.05] pt-2.5 flex justify-between items-center">
            <span className="text-sm text-white/60 font-medium">Итого</span>
            <span className="text-lg font-bold gradient-text">{priceLabel}</span>
          </div>
        </div>
        <p className="text-[11px] text-white/30 mb-4 text-center leading-relaxed">При нехватке средств — авто-пополнение из Кошелька XPLR</p>
        <button onClick={onConfirm} disabled={loading} className="w-full py-3 rounded-xl font-semibold text-sm transition-all duration-200 bg-gradient-to-r from-[#38BDF8] to-[#A78BFA] text-white hover:opacity-90 active:scale-[0.98] disabled:opacity-60 flex items-center justify-center gap-2">
          {loading ? <><Loader2 className="w-4 h-4 animate-spin" /> Оплата...</> : 'Подтвердить оплату'}
        </button>
        <button onClick={onClose} disabled={loading} className="w-full py-2.5 mt-2 text-sm text-white/40 hover:text-white/60 transition-colors disabled:opacity-30">Отмена</button>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// No Card Warning Modal
// ══════════════════════════════════════════════════════════════
const NoCardWarningModal = ({ onClose, onGoToCards }: { onClose: () => void; onGoToCards: () => void }) => {
  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', h); document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div className="relative z-10 w-full max-w-sm rounded-2xl bg-[#0F1629] border border-white/[0.08] p-6 sm:p-8 animate-scale-in" onClick={e => e.stopPropagation()}>
        <div className="flex items-center justify-center w-14 h-14 rounded-2xl bg-amber-500/10 mb-5 mx-auto">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="none" className="text-amber-400"><path d="M12 9V13M12 17H12.01M10.29 3.86L1.82 18A2 2 0 003.54 21H20.46A2 2 0 0022.18 18L13.71 3.86A2 2 0 0010.29 3.86Z" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/></svg>
        </div>
        <h3 className="text-lg font-bold text-white text-center mb-2">Требуется карта XPLR</h3>
        <p className="text-sm text-white/50 text-center leading-relaxed mb-6">
          Для покупки необходима карта для подписок или премиальная карта. Карта для путешествий не подходит.
        </p>
        <button onClick={onGoToCards} className="w-full py-3 rounded-xl font-semibold text-sm transition-all duration-200 bg-gradient-to-r from-[#38BDF8] to-[#A78BFA] text-white hover:opacity-90 active:scale-[0.98]">Открыть карты</button>
        <button onClick={onClose} className="w-full py-2.5 mt-2 text-sm text-white/40 hover:text-white/60 transition-colors">Отмена</button>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// Store Page — Store 2.0
// ══════════════════════════════════════════════════════════════
export const StorePage = () => {
  const navigate = useNavigate();
  const [storeView, setStoreView] = useState<'hub' | 'esim' | 'digital'>('hub');
  const [error, setError] = useState('');

  // eSIM state
  const [destinations, setDestinations] = useState<ESIMDestination[]>([]);
  const [destsLoading, setDestsLoading] = useState(false);
  const [destsSearch, setDestsSearch] = useState('');
  const [selectedCountry, setSelectedCountry] = useState<ESIMDestination | null>(null);
  const [plans, setPlans] = useState<ESIMPlan[]>([]);
  const [plansLoading, setPlansLoading] = useState(false);

  // eSIM purchase flow
  const [confirmPlan, setConfirmPlan] = useState<ESIMPlan | null>(null);
  const [esimPurchasing, setEsimPurchasing] = useState(false);
  const [esimResult, setEsimResult] = useState<{ result: ESIMOrderResult; planName: string } | null>(null);

  // Digital state
  const [digitalProducts, setDigitalProducts] = useState<StoreProduct[]>([]);
  const [digitalLoading, setDigitalLoading] = useState(false);
  const [confirmDigital, setConfirmDigital] = useState<StoreProduct | null>(null);
  const [digitalPurchasing, setDigitalPurchasing] = useState(false);
  const [digitalResult, setDigitalResult] = useState<{ productName: string; priceUsd: string; activationKey: string } | null>(null);

  // Card state — purchases only via active non-travel cards
  const [activeCards, setActiveCards] = useState<Card[]>([]);
  const [selectedCard, setSelectedCard] = useState<Card | null>(null);
  const [showNoCardWarning, setShowNoCardWarning] = useState(false);

  useEffect(() => {
    getUserCards()
      .then(cards => {
        // Filter: only ACTIVE, non-travel cards; sort subscription first
        const slugOrder: Record<string, number> = { subscriptions: 0, services: 0, premium: 1 };
        const eligible = (cards || [])
          .filter(c => c.card_status === 'ACTIVE' && c.service_slug !== 'travel')
          .sort((a, b) => (slugOrder[a.service_slug || ''] ?? 9) - (slugOrder[b.service_slug || ''] ?? 9));
        setActiveCards(eligible);
        if (eligible.length > 0) setSelectedCard(eligible[0]);
      })
      .catch(() => setActiveCards([]));
  }, []);

  const tryBuyESIM = (plan: ESIMPlan) => { if (activeCards.length === 0) { setShowNoCardWarning(true); return; } setConfirmPlan(plan); };
  const tryBuyDigital = (product: StoreProduct) => { if (activeCards.length === 0) { setShowNoCardWarning(true); return; } setConfirmDigital(product); };

  // Load eSIM destinations
  const loadDestinations = useCallback(async () => {
    setDestsLoading(true);
    try { const res = await getESIMDestinations(destsSearch || undefined); setDestinations(res.destinations); }
    catch { setError('Не удалось загрузить направления'); }
    finally { setDestsLoading(false); }
  }, [destsSearch]);

  useEffect(() => {
    if (storeView !== 'esim') return;
    const t = setTimeout(() => loadDestinations(), destsSearch ? 300 : 0);
    return () => clearTimeout(t);
  }, [loadDestinations, destsSearch, storeView]);

  const loadPlans = useCallback(async (cc: string) => {
    setPlansLoading(true);
    try { const res = await getESIMPlans(cc); setPlans(res.plans); }
    catch { setError('Не удалось загрузить тарифы'); }
    finally { setPlansLoading(false); }
  }, []);

  useEffect(() => { if (selectedCountry) loadPlans(selectedCountry.country_code); }, [selectedCountry, loadPlans]);

  const loadDigital = useCallback(async () => {
    setDigitalLoading(true);
    try { const res = await getStoreCatalog({ category: 'digital' }); setDigitalProducts(res.products); }
    catch { setError('Не удалось загрузить товары'); }
    finally { setDigitalLoading(false); }
  }, []);

  useEffect(() => { if (storeView === 'digital') loadDigital(); }, [storeView, loadDigital]);

  // Purchase handlers
  const handleESIMPurchase = async () => {
    if (!confirmPlan) return;
    setEsimPurchasing(true); setError('');
    try {
      const res = await orderESIM(confirmPlan);
      const planName = confirmPlan.name;
      setConfirmPlan(null); setEsimResult({ result: res, planName });
    } catch (err: any) {
      setConfirmPlan(null);
      if (err?.response?.data?.code === 'NO_ACTIVE_CARD') setShowNoCardWarning(true);
      else { const msg = err?.response?.data?.error || 'Ошибка при покупке eSIM'; setError(typeof msg === 'string' ? msg : 'Ошибка при покупке'); }
    } finally { setEsimPurchasing(false); }
  };

  const handleDigitalPurchase = async () => {
    if (!confirmDigital) return;
    setDigitalPurchasing(true); setError('');
    try {
      const res = await purchaseProduct(confirmDigital.id);
      setConfirmDigital(null); setDigitalResult({ productName: res.product_name, priceUsd: res.price_usd, activationKey: res.activation_key });
    } catch (err: any) {
      setConfirmDigital(null);
      if (err?.response?.data?.code === 'NO_ACTIVE_CARD') setShowNoCardWarning(true);
      else { const msg = err?.response?.data?.error || 'Ошибка при покупке'; setError(typeof msg === 'string' ? msg : 'Ошибка при покупке'); }
    } finally { setDigitalPurchasing(false); }
  };

  // Group digital products by brand name (first word)
  const groupedDigital = digitalProducts.reduce<Record<string, StoreProduct[]>>((acc, p) => {
    const brand = p.name.split(' ')[0];
    if (!acc[brand]) acc[brand] = [];
    acc[brand].push(p);
    return acc;
  }, {});

  return (
    <DashboardLayout>
      <div className="max-w-5xl mx-auto">
        <BackButton />

        {/* ═══════ Hub — Full-Bleed Category Cards ═══════ */}
        {storeView === 'hub' && (
          <div className="stagger-fade-in">
            {/* Header */}
            <div className="flex items-center justify-between px-1 mb-5">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[#38BDF8] to-[#A78BFA] flex items-center justify-center">
                  <ShoppingBag className="w-4 h-4 text-white" />
                </div>
                <span className="text-base sm:text-lg font-bold text-white tracking-tight">Магазин</span>
              </div>
            </div>

            {/* Мои покупки — premium wide button */}
            <button
              onClick={() => navigate('/purchases')}
              className="w-full flex items-center justify-center gap-3 px-6 py-3.5 mb-5 rounded-2xl bg-gradient-to-r from-[#4F46E5] to-[#7C3AED] border border-white/10 text-white font-semibold text-sm tracking-wide hover:opacity-90 active:scale-[0.98] transition-all shadow-lg shadow-purple-500/15"
            >
              <ShoppingBag className="w-5 h-5" />
              Мои покупки
            </button>

            {/* Cards */}
            <div className="flex flex-col lg:flex-row gap-4 sm:gap-5">
              <button onClick={() => setStoreView('esim')} className="relative flex-1 min-h-[240px] sm:min-h-[280px] rounded-2xl overflow-hidden group cursor-pointer">
                <img src="/store/esim-card.png" alt="eSIM" className="absolute inset-0 w-full h-full object-cover transition-transform duration-700 group-hover:scale-105" />
                <div className="absolute inset-0 bg-gradient-to-t from-[#0a0a0f] via-[#0a0a0f]/60 to-transparent" />
                <div className="relative z-10 flex flex-col justify-end h-full p-6 sm:p-8">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-2 h-2 rounded-full bg-[#38BDF8] animate-pulse" />
                    <span className="text-xs sm:text-sm font-medium tracking-widest uppercase text-[#38BDF8]/80">eSIM</span>
                  </div>
                  <h2 className="text-2xl sm:text-3xl lg:text-4xl font-extrabold text-white text-left leading-tight">eSIM и Сим-карты</h2>
                  <p className="text-sm sm:text-base text-white/50 mt-2 text-left max-w-md">Мобильный интернет в любой точке мира. Мгновенная активация.</p>
                </div>
              </button>

              <button onClick={() => setStoreView('digital')} className="relative flex-1 min-h-[240px] sm:min-h-[280px] rounded-2xl overflow-hidden group cursor-pointer">
                <img src="/store/digital-products.png" alt="Digital" className="absolute inset-0 w-full h-full object-cover transition-transform duration-700 group-hover:scale-105" />
                <div className="absolute inset-0 bg-gradient-to-t from-[#0a0a0f] via-[#0a0a0f]/60 to-transparent" />
                <div className="relative z-10 flex flex-col justify-end h-full p-6 sm:p-8">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="w-2 h-2 rounded-full bg-[#A78BFA] animate-pulse" />
                    <span className="text-xs sm:text-sm font-medium tracking-widest uppercase text-[#A78BFA]/80">Цифровые</span>
                  </div>
                  <h2 className="text-2xl sm:text-3xl lg:text-4xl font-extrabold text-white text-left leading-tight">Цифровые товары</h2>
                  <p className="text-sm sm:text-base text-white/50 mt-2 text-left max-w-md">Подарочные карты, подписки и пополнения по лучшим ценам.</p>
                </div>
              </button>
            </div>
          </div>
        )}

        {/* ═══════ eSIM — Country List ═══════ */}
        {storeView === 'esim' && !selectedCountry && (
          <div className="max-w-2xl mx-auto stagger-fade-in">
            <div className="flex items-center gap-4 mb-6 sm:mb-8">
              <button onClick={() => { setStoreView('hub'); setDestsSearch(''); }} className="flex items-center justify-center w-10 h-10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] transition-colors">
                <ArrowLeft className="w-5 h-5 text-white" />
              </button>
              <div>
                <h1 className="text-xl sm:text-2xl font-bold text-white">eSIM и Сим-карты</h1>
                <p className="text-sm text-white/40 mt-0.5">Выберите страну</p>
              </div>
            </div>

            <div className="relative mb-4">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-white/30" />
              <input type="text" placeholder="Поиск по стране..." value={destsSearch} onChange={e => setDestsSearch(e.target.value)}
                className="w-full pl-11 pr-10 py-3 bg-white/[0.04] rounded-xl text-sm text-white placeholder-white/30 outline-none focus:bg-white/[0.06] border border-white/[0.05] transition-colors" />
              {destsSearch && <button onClick={() => setDestsSearch('')} className="absolute right-3 top-1/2 -translate-y-1/2 p-1 rounded-lg hover:bg-white/10 text-white/30 hover:text-white transition-all"><X className="w-4 h-4" /></button>}
            </div>

            {destsLoading ? (
              <div className="rounded-2xl bg-white/[0.02] border border-white/[0.05] overflow-hidden">
                {[...Array(8)].map((_, i) => (
                  <div key={i} className="flex items-center gap-4 px-4 sm:px-5 py-3.5 sm:py-4 border-b border-white/[0.05] animate-pulse">
                    <div className="w-10 h-7 rounded bg-white/[0.06]" /><div className="flex-1"><div className="h-4 bg-white/[0.05] rounded w-32" /></div><div className="h-3 bg-white/[0.04] rounded w-14" />
                  </div>
                ))}
              </div>
            ) : destinations.length === 0 ? (
              <div className="py-16 text-center"><Globe className="w-10 h-10 text-white/10 mx-auto mb-3" /><p className="text-white/30 text-sm">{destsSearch ? 'Страна не найдена' : 'Направления загружаются...'}</p></div>
            ) : (
              <div className="rounded-2xl bg-white/[0.02] border border-white/[0.05] overflow-hidden">
                {destinations.map((dest, i) => (
                  <button key={dest.country_code} onClick={() => setSelectedCountry(dest)}
                    className={`w-full flex items-center gap-4 px-4 sm:px-5 py-3.5 sm:py-4 hover:bg-white/[0.04] transition-colors text-left ${i !== destinations.length - 1 ? 'border-b border-white/[0.05]' : ''}`}>
                    <CountryFlag code={dest.country_code} size={32} />
                    <span className="text-sm sm:text-[15px] font-medium text-white flex-1 truncate">{dest.country_name}</span>
                    <span className="text-xs sm:text-sm text-white/30 flex-shrink-0 tabular-nums">{dest.plan_count} тарифа</span>
                    <ChevronRight className="w-4 h-4 text-white/20" />
                  </button>
                ))}
              </div>
            )}
          </div>
        )}

        {/* ═══════ eSIM Plans (selected country) ═══════ */}
        {storeView === 'esim' && selectedCountry && (
          <div className="max-w-2xl mx-auto stagger-fade-in">
            <div className="flex items-center gap-4 mb-6 sm:mb-8">
              <button onClick={() => { setSelectedCountry(null); setPlans([]); }} className="flex items-center justify-center w-10 h-10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] transition-colors">
                <ArrowLeft className="w-5 h-5 text-white" />
              </button>
              <CountryFlag code={selectedCountry.country_code} size={36} />
              <div>
                <h1 className="text-xl sm:text-2xl font-bold text-white">{selectedCountry.country_name}</h1>
                <p className="text-sm text-white/40 mt-0.5">Выберите тариф eSIM</p>
              </div>
            </div>

            {plansLoading ? (
              <div className="rounded-2xl bg-white/[0.02] border border-white/[0.05] overflow-hidden">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="flex items-center gap-4 px-4 sm:px-5 py-4 border-b border-white/[0.05] animate-pulse">
                    <div className="flex-1 space-y-1.5"><div className="h-4 bg-white/[0.05] rounded w-40" /><div className="h-3 bg-white/[0.04] rounded w-24" /></div>
                    <div className="h-5 bg-white/[0.05] rounded w-16" /><div className="h-8 bg-white/[0.04] rounded w-20" />
                  </div>
                ))}
              </div>
            ) : plans.length === 0 ? (
              <div className="py-16 text-center"><Wifi className="w-10 h-10 text-white/10 mx-auto mb-3" /><p className="text-white/30 text-sm">Нет доступных тарифов</p></div>
            ) : (
              <div className="rounded-2xl bg-white/[0.02] border border-white/[0.05] overflow-hidden">
                {plans.map((plan, idx) => (
                  <div key={plan.plan_id} className={`flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-3.5 sm:py-4 ${plan.in_stock ? '' : 'opacity-40'} ${idx < plans.length - 1 ? 'border-b border-white/[0.05]' : ''}`}>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm sm:text-[15px] font-medium text-white truncate">{plan.name}</p>
                      <p className="text-xs text-white/35 mt-0.5">{plan.data_gb} GB / {plan.validity_days} дней</p>
                    </div>
                    <div className="flex items-center gap-3 flex-shrink-0">
                      <div className="text-right">
                        <span className="text-sm sm:text-[15px] font-bold gradient-text tabular-nums">${plan.price_usd.toFixed(2)}</span>
                        {plan.old_price > 0 && plan.old_price > plan.price_usd && <span className="block text-[10px] text-white/20 line-through">${plan.old_price.toFixed(2)}</span>}
                      </div>
                      {plan.in_stock ? (
                        <button onClick={() => tryBuyESIM(plan)} className="px-3.5 sm:px-4 py-1.5 sm:py-2 rounded-lg text-xs sm:text-sm font-semibold bg-white/[0.06] hover:bg-white/[0.1] text-white transition-all duration-200 active:scale-[0.96]">Купить</button>
                      ) : <span className="text-xs text-white/20">Нет в наличии</span>}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* ═══════ Digital — Products (grouped by brand) ═══════ */}
        {storeView === 'digital' && (
          <div className="max-w-2xl mx-auto stagger-fade-in">
            <div className="flex items-center gap-4 mb-6 sm:mb-8">
              <button onClick={() => setStoreView('hub')} className="flex items-center justify-center w-10 h-10 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] transition-colors">
                <ArrowLeft className="w-5 h-5 text-white" />
              </button>
              <div>
                <h1 className="text-xl sm:text-2xl font-bold text-white">Цифровые товары</h1>
                <p className="text-sm text-white/40 mt-0.5">Подарочные карты и подписки</p>
              </div>
            </div>

            {digitalLoading ? (
              <div className="space-y-4">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="rounded-2xl bg-white/[0.02] border border-white/[0.05] p-4 animate-pulse space-y-3">
                    <div className="h-4 bg-white/[0.05] rounded w-24" /><div className="h-4 bg-white/[0.04] rounded w-48" />
                  </div>
                ))}
              </div>
            ) : digitalProducts.length === 0 ? (
              <div className="py-16 text-center"><Gamepad2 className="w-10 h-10 text-white/10 mx-auto mb-3" /><p className="text-white/30 text-sm">Товаров пока нет</p></div>
            ) : (
              <div className="space-y-4">
                {Object.entries(groupedDigital).map(([brand, items]) => (
                  <div key={brand} className="rounded-2xl bg-white/[0.02] border border-white/[0.05] overflow-hidden">
                    {items.map((product, i) => {
                      const BrandIcon = brandIconForProduct(product.name, product.external_id);
                      return (
                        <div key={product.id} className={`flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-3.5 sm:py-4 ${product.in_stock ? '' : 'opacity-40'} ${i !== items.length - 1 ? 'border-b border-white/[0.05]' : ''}`}>
                          {BrandIcon ? <BrandIcon className="w-8 h-8 sm:w-9 sm:h-9 flex-shrink-0" /> : (
                            <div className="w-8 h-8 sm:w-9 sm:h-9 rounded-lg bg-white/10 flex items-center justify-center flex-shrink-0">
                              <Gamepad2 className="w-4 h-4 text-white/30" />
                            </div>
                          )}
                          <div className="flex-1 min-w-0">
                            <div className="text-sm sm:text-[15px] font-medium text-white truncate">{product.name}</div>
                            {product.description && <div className="text-xs text-white/35 mt-0.5 truncate">{product.description}</div>}
                          </div>
                          <div className="flex items-center gap-3 flex-shrink-0">
                            <span className="text-sm sm:text-[15px] font-bold gradient-text tabular-nums">${product.price_usd}</span>
                            {product.in_stock ? (
                              <button onClick={() => tryBuyDigital(product)} className="px-3.5 sm:px-4 py-1.5 sm:py-2 rounded-lg text-xs sm:text-sm font-semibold bg-white/[0.06] hover:bg-white/[0.1] text-white transition-all duration-200 active:scale-[0.96]">Купить</button>
                            ) : <span className="text-xs text-white/20">Нет</span>}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* ═══════ Modals ═══════ */}
      {showNoCardWarning && <NoCardWarningModal onClose={() => setShowNoCardWarning(false)} onGoToCards={() => navigate('/cards')} />}
      {confirmPlan && (
        <ConfirmPurchaseModal
          title="Подтвердите покупку eSIM"
          itemLabel={`${confirmPlan.name} — ${confirmPlan.country}`}
          priceLabel={`$${confirmPlan.price_usd.toFixed(2)}`}
          loading={esimPurchasing} onConfirm={handleESIMPurchase} onClose={() => !esimPurchasing && setConfirmPlan(null)} allCards={activeCards} onCardChange={setSelectedCard}
        />
      )}
      {esimResult && <ESIMActivationModal result={esimResult.result} planName={esimResult.planName} onClose={() => setEsimResult(null)} />}
      {confirmDigital && (
        <ConfirmPurchaseModal
          title="Подтвердите покупку"
          itemLabel={confirmDigital.name}
          priceLabel={`$${confirmDigital.price_usd}`}
          loading={digitalPurchasing} onConfirm={handleDigitalPurchase} onClose={() => !digitalPurchasing && setConfirmDigital(null)} allCards={activeCards} onCardChange={setSelectedCard}
        />
      )}
      {digitalResult && <DigitalResultModal productName={digitalResult.productName} priceUsd={digitalResult.priceUsd} activationKey={digitalResult.activationKey} onClose={() => setDigitalResult(null)} />}
      {error && <ErrorToast message={error} onClose={() => setError('')} />}
    </DashboardLayout>
  );
};
