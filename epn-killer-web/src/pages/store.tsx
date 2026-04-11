import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ShoppingBag, Search, Globe, Gamepad2, X, Copy, Check, Loader2,
  AlertTriangle, QrCode, Key, ChevronRight, ChevronLeft, ArrowLeft,
  Download, FileText, Smartphone, Wifi
} from 'lucide-react';
import { DashboardLayout } from '../components/dashboard-layout';
import {
  getStoreCatalog, purchaseProduct,
  getESIMDestinations, getESIMPlans, orderESIM,
  type StoreProduct, type StoreCategory, type ESIMDestination, type ESIMPlan, type ESIMOrderResult,
} from '../api/store';

// ── Skeleton Loader ──
const CardSkeleton = () => (
  <div className="glass-card p-5 animate-pulse">
    <div className="h-5 bg-white/10 rounded w-2/3 mb-3" />
    <div className="h-3 bg-white/5 rounded w-full mb-2" />
    <div className="h-3 bg-white/5 rounded w-1/2 mb-4" />
    <div className="flex items-center justify-between">
      <div className="h-6 bg-white/10 rounded w-20" />
      <div className="h-8 bg-white/5 rounded-lg w-24" />
    </div>
  </div>
);

// ── Country flag emoji ──
const countryFlag = (code: string) => {
  if (!code || code.length < 2 || code === 'GLOBAL') return '\u{1F30D}';
  const codePoints = code.toUpperCase().split('').map(c => 0x1f1e6 + c.charCodeAt(0) - 65);
  return String.fromCodePoint(...codePoints);
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
// eSIM Activation Modal (QR + LPA + PDF download)
// ══════════════════════════════════════════════════════════════
const ESIMActivationModal = ({
  result, planName, onClose,
}: {
  result: ESIMOrderResult;
  planName: string;
  onClose: () => void;
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
    navigator.clipboard.writeText(text).then(() => {
      setCopied(label);
      setTimeout(() => setCopied(''), 2000);
    });
  };

  const downloadPDF = () => {
    const content = `
XPLR eSIM — Инструкция по активации
=====================================

Товар: ${planName}
Цена: $${result.price_usd}
ICCID: ${result.iccid || 'N/A'}
Дата: ${new Date().toLocaleString('ru-RU')}

─────────────────────────────────────

СПОСОБ 1: QR-код
Откройте Настройки → Сотовая связь → Добавить eSIM → Сканировать QR-код.
Отсканируйте QR-код из приложения XPLR (Магазин → Мои покупки).

СПОСОБ 2: Ручная установка
Откройте Настройки → Сотовая связь → Добавить eSIM → Ввести данные вручную.

SM-DP+ адрес: ${result.smdp || 'N/A'}
Код активации: ${result.matching_id || 'N/A'}
LPA строка: ${result.lpa || result.qr_data || 'N/A'}

─────────────────────────────────────

Поддержка: https://xplr.pro/support
    `.trim();

    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `XPLR_eSIM_${result.iccid || 'activation'}.txt`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  };

  const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(result.qr_data || result.lpa || '')}`;
  const lpaString = result.lpa || result.qr_data || '';
  const smdp = result.smdp || (lpaString.includes('$') ? lpaString.split('$')[1] : '');
  const matchingId = result.matching_id || (lpaString.includes('$') ? lpaString.split('$')[2] : '');

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div
        className="relative w-full max-w-lg max-h-[92vh] overflow-y-auto rounded-2xl bg-[#111118] border border-white/10 shadow-2xl"
        onClick={e => e.stopPropagation()}
      >
        <button onClick={onClose} className="absolute top-3 right-3 z-10 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all">
          <X className="w-4 h-4" />
        </button>

        <div className="p-6">
          {/* Success header */}
          <div className="text-center mb-6">
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-green-500/20 to-emerald-500/20 border border-green-500/30 flex items-center justify-center mx-auto mb-3">
              <Check className="w-8 h-8 text-green-400" />
            </div>
            <h2 className="text-lg font-bold text-white mb-1">eSIM активирована!</h2>
            <p className="text-sm text-slate-400">{planName} — ${result.price_usd}</p>
          </div>

          {/* QR Code */}
          <div ref={qrRef} className="text-center mb-6">
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium mb-4">
              <QrCode className="w-3.5 h-3.5" />
              Отсканируйте для установки
            </div>
            <div className="bg-white rounded-2xl p-5 inline-block mx-auto shadow-lg">
              <img src={qrUrl} alt="eSIM QR" className="w-[220px] h-[220px]" />
            </div>
          </div>

          {/* Manual installation instructions */}
          <div className="space-y-3 mb-6">
            <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
              <Smartphone className="w-4 h-4 text-blue-400" />
              Как установить вручную
            </div>

            <div className="text-xs text-slate-500 space-y-1 mb-3">
              <p>Настройки → Сотовая связь → Добавить eSIM → Ввести данные вручную</p>
            </div>

            {/* SM-DP+ */}
            {smdp && (
              <div className="bg-white/5 border border-white/10 rounded-xl p-3">
                <p className="text-[10px] text-slate-500 mb-1.5 uppercase tracking-wider">SM-DP+ адрес</p>
                <div className="flex items-center justify-between gap-2">
                  <code className="text-xs text-white font-mono break-all flex-1">{smdp}</code>
                  <button onClick={() => copyText(smdp, 'smdp')} className="shrink-0 p-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 hover:bg-blue-500/20 transition-all">
                    {copied === 'smdp' ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            )}

            {/* Matching ID / Activation Code */}
            {matchingId && (
              <div className="bg-white/5 border border-white/10 rounded-xl p-3">
                <p className="text-[10px] text-slate-500 mb-1.5 uppercase tracking-wider">Код активации</p>
                <div className="flex items-center justify-between gap-2">
                  <code className="text-xs text-white font-mono break-all flex-1">{matchingId}</code>
                  <button onClick={() => copyText(matchingId, 'mid')} className="shrink-0 p-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 hover:bg-blue-500/20 transition-all">
                    {copied === 'mid' ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            )}

            {/* Full LPA string */}
            {lpaString && (
              <div className="bg-white/5 border border-white/10 rounded-xl p-3">
                <p className="text-[10px] text-slate-500 mb-1.5 uppercase tracking-wider">LPA строка (полная)</p>
                <div className="flex items-center justify-between gap-2">
                  <code className="text-[10px] text-slate-300 font-mono break-all flex-1">{lpaString}</code>
                  <button onClick={() => copyText(lpaString, 'lpa')} className="shrink-0 p-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 hover:bg-blue-500/20 transition-all">
                    {copied === 'lpa' ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Action buttons */}
          <div className="flex gap-3">
            <button
              onClick={downloadPDF}
              className="flex-1 py-3 rounded-xl bg-white/5 border border-white/10 text-slate-300 text-sm font-medium hover:bg-white/10 transition-all flex items-center justify-center gap-2"
            >
              <Download className="w-4 h-4" />
              Скачать инструкцию
            </button>
            <button
              onClick={onClose}
              className="flex-1 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-600 text-white text-sm font-medium hover:opacity-90 transition-all"
            >
              Готово
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// Digital Product Result Modal
// ══════════════════════════════════════════════════════════════
const DigitalResultModal = ({
  productName, priceUsd, activationKey, onClose,
}: {
  productName: string; priceUsd: string; activationKey: string; onClose: () => void;
}) => {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', h);
    document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div className="relative w-full max-w-md rounded-2xl bg-[#111118] border border-white/10 shadow-2xl" onClick={e => e.stopPropagation()}>
        <button onClick={onClose} className="absolute top-3 right-3 z-10 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all">
          <X className="w-4 h-4" />
        </button>
        <div className="p-6 text-center">
          <div className="w-16 h-16 rounded-full bg-gradient-to-br from-green-500/20 to-emerald-500/20 border border-green-500/30 flex items-center justify-center mx-auto mb-4">
            <Check className="w-8 h-8 text-green-400" />
          </div>
          <h2 className="text-lg font-bold text-white mb-1">Покупка успешна!</h2>
          <p className="text-sm text-slate-400 mb-6">{productName} — ${priceUsd}</p>

          {activationKey && (
            <div className="mb-6">
              <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400 text-xs font-medium mb-4">
                <Key className="w-3.5 h-3.5" />
                Ключ активации
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-4 flex items-center justify-between gap-3">
                <code className="text-white font-mono text-sm break-all flex-1 text-left">{activationKey}</code>
                <button
                  onClick={() => { navigator.clipboard.writeText(activationKey); setCopied(true); setTimeout(() => setCopied(false), 2000); }}
                  className="shrink-0 p-2.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 hover:bg-blue-500/20 transition-all"
                >
                  {copied ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
                </button>
              </div>
            </div>
          )}

          <button onClick={onClose} className="w-full py-3 rounded-xl bg-white/5 border border-white/10 text-slate-300 text-sm font-medium hover:bg-white/10 transition-all">
            Закрыть
          </button>
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// eSIM Confirm Purchase Modal
// ══════════════════════════════════════════════════════════════
const ESIMConfirmModal = ({
  plan, loading, onConfirm, onClose,
}: {
  plan: ESIMPlan; loading: boolean; onConfirm: () => void; onClose: () => void;
}) => {
  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape' && !loading) onClose(); };
    document.addEventListener('keydown', h);
    document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose, loading]);

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={() => !loading && onClose()}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div className="relative w-full max-w-md rounded-2xl bg-[#111118] border border-white/10 shadow-2xl" onClick={e => e.stopPropagation()}>
        <button onClick={onClose} disabled={loading} className="absolute top-3 right-3 z-10 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all disabled:opacity-30">
          <X className="w-4 h-4" />
        </button>
        <div className="p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 border border-blue-500/30 flex items-center justify-center">
              <Wifi className="w-6 h-6 text-blue-400" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">Подтвердите покупку eSIM</h2>
              <p className="text-xs text-slate-500">Средства будут списаны с кошелька XPLR</p>
            </div>
          </div>
          <div className="space-y-3 mb-6">
            <div className="flex justify-between items-center py-2.5 border-b border-white/5">
              <span className="text-sm text-slate-400">План</span>
              <span className="text-sm text-white font-medium">{plan.name}</span>
            </div>
            <div className="flex justify-between items-center py-2.5 border-b border-white/5">
              <span className="text-sm text-slate-400">Страна</span>
              <span className="text-sm text-white font-medium">{countryFlag(plan.country_code)} {plan.country}</span>
            </div>
            <div className="flex justify-between items-center py-2.5 border-b border-white/5">
              <span className="text-sm text-slate-400">Объём</span>
              <span className="text-sm text-white font-medium">{plan.data_gb} ГБ</span>
            </div>
            <div className="flex justify-between items-center py-2.5 border-b border-white/5">
              <span className="text-sm text-slate-400">Срок</span>
              <span className="text-sm text-white font-medium">{plan.validity_days} дней</span>
            </div>
            <div className="flex justify-between items-center py-3 bg-blue-500/5 rounded-xl px-3 border border-blue-500/10">
              <span className="text-sm text-slate-300 font-medium">Итого</span>
              <span className="text-xl text-blue-400 font-bold">${plan.price_usd.toFixed(2)}</span>
            </div>
          </div>
          <div className="flex gap-3">
            <button onClick={onClose} disabled={loading} className="flex-1 py-3 rounded-xl bg-white/5 border border-white/10 text-slate-400 text-sm font-medium hover:bg-white/10 transition-all disabled:opacity-30">
              Отмена
            </button>
            <button onClick={onConfirm} disabled={loading} className="flex-1 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-600 text-white text-sm font-medium hover:opacity-90 transition-all disabled:opacity-60 flex items-center justify-center gap-2">
              {loading ? <><Loader2 className="w-4 h-4 animate-spin" /> Покупка...</> : 'Купить eSIM'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// Digital Product Confirm Modal
// ══════════════════════════════════════════════════════════════
const DigitalConfirmModal = ({
  product, loading, onConfirm, onClose,
}: {
  product: StoreProduct; loading: boolean; onConfirm: () => void; onClose: () => void;
}) => {
  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape' && !loading) onClose(); };
    document.addEventListener('keydown', h);
    document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose, loading]);

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={() => !loading && onClose()}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div className="relative w-full max-w-md rounded-2xl bg-[#111118] border border-white/10 shadow-2xl" onClick={e => e.stopPropagation()}>
        <button onClick={onClose} disabled={loading} className="absolute top-3 right-3 z-10 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all disabled:opacity-30">
          <X className="w-4 h-4" />
        </button>
        <div className="p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 border border-purple-500/30 flex items-center justify-center">
              <ShoppingBag className="w-6 h-6 text-purple-400" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">Подтвердите покупку</h2>
              <p className="text-xs text-slate-500">Средства будут списаны с кошелька XPLR</p>
            </div>
          </div>
          <div className="space-y-3 mb-6">
            <div className="flex justify-between items-center py-2.5 border-b border-white/5">
              <span className="text-sm text-slate-400">Товар</span>
              <span className="text-sm text-white font-medium text-right max-w-[200px] truncate">{product.name}</span>
            </div>
            {product.description && (
              <div className="flex justify-between items-center py-2.5 border-b border-white/5">
                <span className="text-sm text-slate-400">Описание</span>
                <span className="text-xs text-slate-300 text-right max-w-[200px]">{product.description}</span>
              </div>
            )}
            <div className="flex justify-between items-center py-3 bg-blue-500/5 rounded-xl px-3 border border-blue-500/10">
              <span className="text-sm text-slate-300 font-medium">Итого</span>
              <span className="text-xl text-blue-400 font-bold">${product.price_usd}</span>
            </div>
          </div>
          <div className="flex gap-3">
            <button onClick={onClose} disabled={loading} className="flex-1 py-3 rounded-xl bg-white/5 border border-white/10 text-slate-400 text-sm font-medium hover:bg-white/10 transition-all disabled:opacity-30">
              Отмена
            </button>
            <button onClick={onConfirm} disabled={loading} className="flex-1 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-600 text-white text-sm font-medium hover:opacity-90 transition-all disabled:opacity-60 flex items-center justify-center gap-2">
              {loading ? <><Loader2 className="w-4 h-4 animate-spin" /> Покупка...</> : 'Купить'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

// ══════════════════════════════════════════════════════════════
// Store Page
// ══════════════════════════════════════════════════════════════

export const StorePage = () => {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<'esim' | 'digital'>('esim');
  const [error, setError] = useState('');

  // eSIM state
  const [destinations, setDestinations] = useState<ESIMDestination[]>([]);
  const [destsLoading, setDestsLoading] = useState(true);
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
  const [digitalLoading, setDigitalLoading] = useState(true);
  const [confirmDigital, setConfirmDigital] = useState<StoreProduct | null>(null);
  const [digitalPurchasing, setDigitalPurchasing] = useState(false);
  const [digitalResult, setDigitalResult] = useState<{ productName: string; priceUsd: string; activationKey: string } | null>(null);

  // Load eSIM destinations
  const loadDestinations = useCallback(async () => {
    setDestsLoading(true);
    try {
      const res = await getESIMDestinations(destsSearch || undefined);
      setDestinations(res.destinations);
    } catch {
      setError('Не удалось загрузить направления');
    } finally {
      setDestsLoading(false);
    }
  }, [destsSearch]);

  useEffect(() => {
    if (activeTab !== 'esim') return;
    const t = setTimeout(() => loadDestinations(), destsSearch ? 300 : 0);
    return () => clearTimeout(t);
  }, [loadDestinations, destsSearch, activeTab]);

  // Load eSIM plans for selected country
  const loadPlans = useCallback(async (cc: string) => {
    setPlansLoading(true);
    try {
      const res = await getESIMPlans(cc);
      setPlans(res.plans);
    } catch {
      setError('Не удалось загрузить тарифы');
    } finally {
      setPlansLoading(false);
    }
  }, []);

  useEffect(() => {
    if (selectedCountry) loadPlans(selectedCountry.country_code);
  }, [selectedCountry, loadPlans]);

  // Load digital products
  const loadDigital = useCallback(async () => {
    setDigitalLoading(true);
    try {
      const res = await getStoreCatalog({ category: 'digital' });
      setDigitalProducts(res.products);
    } catch {
      setError('Не удалось загрузить товары');
    } finally {
      setDigitalLoading(false);
    }
  }, []);

  useEffect(() => {
    if (activeTab === 'digital') loadDigital();
  }, [activeTab, loadDigital]);

  // eSIM purchase handler
  const handleESIMPurchase = async () => {
    if (!confirmPlan) return;
    setEsimPurchasing(true);
    setError('');
    try {
      const res = await orderESIM(confirmPlan);
      const planName = confirmPlan.name;
      setConfirmPlan(null);
      setEsimResult({ result: res, planName });
    } catch (err: any) {
      setConfirmPlan(null);
      const msg = err?.response?.data?.error || 'Ошибка при покупке eSIM';
      setError(typeof msg === 'string' ? msg : 'Ошибка при покупке');
    } finally {
      setEsimPurchasing(false);
    }
  };

  // Digital purchase handler
  const handleDigitalPurchase = async () => {
    if (!confirmDigital) return;
    setDigitalPurchasing(true);
    setError('');
    try {
      const res = await purchaseProduct(confirmDigital.id);
      setConfirmDigital(null);
      setDigitalResult({ productName: res.product_name, priceUsd: res.price_usd, activationKey: res.activation_key });
    } catch (err: any) {
      setConfirmDigital(null);
      const msg = err?.response?.data?.error || 'Ошибка при покупке';
      setError(typeof msg === 'string' ? msg : 'Ошибка при покупке');
    } finally {
      setDigitalPurchasing(false);
    }
  };

  const tabs = [
    { slug: 'esim' as const, label: 'eSIM — Весь мир', icon: <Globe className="w-4 h-4" /> },
    { slug: 'digital' as const, label: 'Цифровые товары', icon: <Gamepad2 className="w-4 h-4" /> },
  ];

  return (
    <DashboardLayout>
      <div className="max-w-5xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center">
              <ShoppingBag className="w-5 h-5 text-blue-400" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">Магазин</h1>
              <p className="text-xs text-slate-400">eSIM, игровые ключи, подписки</p>
            </div>
          </div>
          <button
            onClick={() => navigate('/purchases')}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-white/5 border border-white/10 text-slate-400 text-xs font-medium hover:bg-white/10 hover:text-white transition-all"
          >
            <FileText className="w-4 h-4" />
            Мои покупки
          </button>
        </div>

        {/* Tabs */}
        <div className="flex gap-2">
          {tabs.map(tab => (
            <button
              key={tab.slug}
              onClick={() => { setActiveTab(tab.slug); setSelectedCountry(null); setDestsSearch(''); }}
              className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium border transition-all ${
                activeTab === tab.slug
                  ? 'bg-blue-500/10 border-blue-500/30 text-blue-400'
                  : 'bg-white/5 border-white/10 text-slate-400 hover:bg-white/10 hover:text-white'
              }`}
            >
              {tab.icon}
              <span className="hidden sm:inline">{tab.label}</span>
              <span className="sm:hidden">{tab.slug === 'esim' ? 'eSIM' : 'Товары'}</span>
            </button>
          ))}
        </div>

        {/* ═══════ eSIM Tab ═══════ */}
        {activeTab === 'esim' && !selectedCountry && (
          <>
            {/* Country search */}
            <div className="relative">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
              <input
                type="text"
                placeholder="Поиск по стране (Турция, USA, Japan...)"
                value={destsSearch}
                onChange={e => setDestsSearch(e.target.value)}
                className="w-full pl-11 pr-10 py-3 bg-white/5 border border-white/10 rounded-xl text-sm text-white placeholder-slate-500 outline-none focus:border-blue-500/50 transition-colors"
              />
              {destsSearch && (
                <button onClick={() => setDestsSearch('')} className="absolute right-3 top-1/2 -translate-y-1/2 p-1 rounded-lg hover:bg-white/10 text-slate-500 hover:text-white transition-all">
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>

            {/* Destinations grid — SIM chip style */}
            {destsLoading ? (
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
                {[...Array(8)].map((_, i) => (
                  <div key={i} className="relative rounded-2xl overflow-hidden animate-pulse" style={{ aspectRatio: '1.6/1' }}>
                    <div className="absolute inset-0 bg-gradient-to-br from-white/[0.06] to-white/[0.02]" />
                    <div className="absolute top-4 left-4 w-10 h-7 rounded bg-yellow-500/10" />
                    <div className="absolute bottom-4 left-4 right-4">
                      <div className="h-4 bg-white/10 rounded w-2/3 mb-1.5" />
                      <div className="h-3 bg-white/5 rounded w-1/3" />
                    </div>
                  </div>
                ))}
              </div>
            ) : destinations.length === 0 ? (
              <div className="glass-card p-12 text-center">
                <Globe className="w-12 h-12 text-slate-600 mx-auto mb-3" />
                <p className="text-slate-400 text-sm">{destsSearch ? 'Страна не найдена' : 'Направления загружаются...'}</p>
                {destsSearch && (
                  <button onClick={() => setDestsSearch('')} className="mt-3 text-xs text-blue-400 hover:text-blue-300 transition-colors">
                    Сбросить поиск
                  </button>
                )}
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
                {destinations.map(dest => (
                  <button
                    key={dest.country_code}
                    onClick={() => setSelectedCountry(dest)}
                    className="relative rounded-2xl overflow-hidden border border-white/[0.08] hover:border-blue-500/30 transition-all group cursor-pointer text-left"
                    style={{ aspectRatio: '1.6/1', background: 'linear-gradient(145deg, rgba(30,30,50,0.95), rgba(15,15,30,0.98))' }}
                  >
                    {/* SIM chip accent */}
                    <div className="absolute top-3.5 left-3.5 w-9 h-6 rounded-[4px] border border-yellow-500/30 bg-gradient-to-br from-yellow-400/20 to-yellow-600/10 flex items-center justify-center">
                      <div className="w-5 h-3 rounded-[2px] border border-yellow-500/20 bg-gradient-to-br from-yellow-400/10 to-transparent" />
                    </div>
                    {/* Large centered flag */}
                    <div className="absolute inset-0 flex items-center justify-center pointer-events-none opacity-90 group-hover:opacity-100 group-hover:scale-110 transition-all duration-300">
                      <span className="text-6xl sm:text-7xl drop-shadow-lg select-none">{dest.flag_emoji || countryFlag(dest.country_code)}</span>
                    </div>
                    {/* Gradient overlay at bottom */}
                    <div className="absolute inset-x-0 bottom-0 h-2/3 bg-gradient-to-t from-black/80 via-black/40 to-transparent pointer-events-none" />
                    {/* Country name + plan count */}
                    <div className="absolute bottom-3 left-3.5 right-3.5">
                      <h3 className="text-sm font-bold text-white truncate drop-shadow-md">{dest.country_name}</h3>
                      <p className="text-[10px] text-slate-400 mt-0.5">{dest.plan_count} {dest.plan_count === 1 ? 'тариф' : 'тарифов'}</p>
                    </div>
                    {/* Hover arrow */}
                    <div className="absolute top-3.5 right-3.5 opacity-0 group-hover:opacity-100 transition-opacity">
                      <ChevronRight className="w-4 h-4 text-blue-400 drop-shadow" />
                    </div>
                    {/* Subtle shine on hover */}
                    <div className="absolute inset-0 bg-gradient-to-br from-blue-500/0 to-purple-500/0 group-hover:from-blue-500/5 group-hover:to-purple-500/5 transition-all duration-300 pointer-events-none" />
                  </button>
                ))}
              </div>
            )}
          </>
        )}

        {/* ═══════ eSIM Plans (selected country) ═══════ */}
        {activeTab === 'esim' && selectedCountry && (
          <>
            <button
              onClick={() => { setSelectedCountry(null); setPlans([]); }}
              className="flex items-center gap-2 text-sm text-slate-400 hover:text-white transition-colors"
            >
              <ArrowLeft className="w-4 h-4" />
              Все страны
            </button>

            <div className="flex items-center gap-3 mb-2">
              <span className="text-3xl">{selectedCountry.flag_emoji || countryFlag(selectedCountry.country_code)}</span>
              <div>
                <h2 className="text-lg font-bold text-white">{selectedCountry.country_name}</h2>
                <p className="text-xs text-slate-400">Выберите тариф eSIM</p>
              </div>
            </div>

            {plansLoading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="relative rounded-2xl overflow-hidden animate-pulse h-44 border border-white/[0.06]" style={{ background: 'linear-gradient(145deg, rgba(30,30,50,0.95), rgba(15,15,30,0.98))' }}>
                    <div className="absolute top-4 left-4 w-10 h-7 rounded bg-yellow-500/10" />
                    <div className="absolute top-4 right-4 w-12 h-12 rounded-xl bg-white/5" />
                    <div className="absolute bottom-4 left-4 right-4 space-y-2">
                      <div className="h-4 bg-white/10 rounded w-1/2" />
                      <div className="h-6 bg-white/10 rounded w-1/3" />
                    </div>
                  </div>
                ))}
              </div>
            ) : plans.length === 0 ? (
              <div className="glass-card p-12 text-center">
                <Wifi className="w-12 h-12 text-slate-600 mx-auto mb-3" />
                <p className="text-slate-400 text-sm">Нет доступных тарифов</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {plans.map(plan => (
                  <article
                    key={plan.plan_id}
                    onClick={() => plan.in_stock && setConfirmPlan(plan)}
                    className={`relative rounded-2xl overflow-hidden border transition-all group ${
                      plan.in_stock
                        ? 'cursor-pointer border-white/[0.08] hover:border-blue-500/30 hover:shadow-lg hover:shadow-blue-500/5'
                        : 'opacity-50 cursor-not-allowed border-white/[0.05]'
                    }`}
                    style={{ background: 'linear-gradient(145deg, rgba(30,30,50,0.95), rgba(15,15,30,0.98))' }}
                  >
                    <div className="p-5">
                      {/* Top row: SIM chip + flag */}
                      <div className="flex items-start justify-between mb-5">
                        <div className="w-10 h-7 rounded-[4px] border border-yellow-500/30 bg-gradient-to-br from-yellow-400/20 to-yellow-600/10 flex items-center justify-center">
                          <div className="w-6 h-3.5 rounded-[2px] border border-yellow-500/20 bg-gradient-to-br from-yellow-400/10 to-transparent" />
                        </div>
                        <span className="text-3xl select-none">{countryFlag(selectedCountry.country_code)}</span>
                      </div>
                      {/* Plan name + badges */}
                      <h3 className="text-sm font-bold text-white mb-2.5">{plan.name}</h3>
                      <div className="flex items-center gap-2 mb-4">
                        <span className="px-2.5 py-1 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-[11px] font-medium flex items-center gap-1">
                          <Wifi className="w-3 h-3" /> {plan.data_gb} ГБ
                        </span>
                        <span className="px-2.5 py-1 rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400 text-[11px] font-medium">
                          {plan.validity_days} дн.
                        </span>
                      </div>
                      {/* Price row */}
                      <div className="flex items-center justify-between pt-3 border-t border-white/[0.06]">
                        <div className="flex items-baseline gap-2">
                          <span className="text-xl font-bold text-white">${plan.price_usd.toFixed(2)}</span>
                          {plan.old_price > 0 && plan.old_price > plan.price_usd && (
                            <span className="text-xs text-slate-500 line-through">${plan.old_price.toFixed(2)}</span>
                          )}
                        </div>
                        {plan.in_stock ? (
                          <span className="flex items-center gap-1 text-xs text-blue-400/60 group-hover:text-blue-400 transition-colors font-medium">
                            Купить <ChevronRight className="w-3.5 h-3.5" />
                          </span>
                        ) : (
                          <span className="flex items-center gap-1 text-xs text-red-400/60 font-medium">
                            <AlertTriangle className="w-3 h-3" /> Нет в наличии
                          </span>
                        )}
                      </div>
                    </div>
                    {/* Hover shine */}
                    <div className="absolute inset-0 bg-gradient-to-br from-blue-500/0 to-purple-500/0 group-hover:from-blue-500/5 group-hover:to-purple-500/5 transition-all duration-300 pointer-events-none" />
                  </article>
                ))}
              </div>
            )}
          </>
        )}

        {/* ═══════ Digital Goods Tab ═══════ */}
        {activeTab === 'digital' && (
          <>
            {digitalLoading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="relative rounded-2xl overflow-hidden animate-pulse h-40 border border-white/[0.06]" style={{ background: 'linear-gradient(145deg, rgba(30,30,50,0.95), rgba(15,15,30,0.98))' }}>
                    <div className="absolute top-4 left-4 w-10 h-10 rounded-xl bg-white/5" />
                    <div className="absolute bottom-4 left-4 right-4 space-y-2">
                      <div className="h-4 bg-white/10 rounded w-2/3" />
                      <div className="h-3 bg-white/5 rounded w-1/2" />
                    </div>
                  </div>
                ))}
              </div>
            ) : digitalProducts.length === 0 ? (
              <div className="glass-card p-12 text-center">
                <Gamepad2 className="w-12 h-12 text-slate-600 mx-auto mb-3" />
                <p className="text-slate-400 text-sm">Товаров пока нет</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {digitalProducts.map(product => {
                  // Brand color accent based on product name
                  const brandKey = product.external_id?.split('-')[0] || product.name.split(' ')[0].toLowerCase();
                  const brandColors: Record<string, { bg: string; border: string; text: string }> = {
                    steam:     { bg: 'from-[#1b2838]/40 to-[#1b2838]/20', border: 'border-[#66c0f4]/20 hover:border-[#66c0f4]/40', text: 'text-[#66c0f4]' },
                    psn:       { bg: 'from-[#003087]/30 to-[#003087]/10', border: 'border-blue-500/20 hover:border-blue-500/40', text: 'text-blue-400' },
                    xbox:      { bg: 'from-[#107c10]/30 to-[#107c10]/10', border: 'border-green-500/20 hover:border-green-500/40', text: 'text-green-400' },
                    nintendo:  { bg: 'from-[#e60012]/20 to-[#e60012]/10', border: 'border-red-500/20 hover:border-red-500/40', text: 'text-red-400' },
                    spotify:   { bg: 'from-[#1DB954]/20 to-[#1DB954]/10', border: 'border-[#1DB954]/20 hover:border-[#1DB954]/40', text: 'text-[#1DB954]' },
                    netflix:   { bg: 'from-[#E50914]/20 to-[#E50914]/10', border: 'border-[#E50914]/20 hover:border-[#E50914]/40', text: 'text-[#E50914]' },
                  };
                  const brand = brandColors[brandKey] || { bg: 'from-purple-500/20 to-purple-500/10', border: 'border-purple-500/20 hover:border-purple-500/40', text: 'text-purple-400' };

                  return (
                    <article
                      key={product.id}
                      onClick={() => product.in_stock && setConfirmDigital(product)}
                      className={`relative rounded-2xl overflow-hidden border transition-all group ${
                        product.in_stock
                          ? `cursor-pointer ${brand.border} hover:shadow-lg`
                          : 'opacity-50 cursor-not-allowed border-white/[0.05]'
                      }`}
                      style={{ background: 'linear-gradient(145deg, rgba(30,30,50,0.95), rgba(15,15,30,0.98))' }}
                    >
                      {/* Brand gradient overlay */}
                      <div className={`absolute inset-0 bg-gradient-to-br ${brand.bg} pointer-events-none`} />
                      <div className="relative p-5">
                        {/* Logo + title row */}
                        <div className="flex items-start gap-3 mb-3">
                          {product.image_url ? (
                            <div className="w-10 h-10 rounded-xl bg-white/[0.06] border border-white/[0.08] flex items-center justify-center shrink-0 p-2">
                              <img src={product.image_url} alt="" className="w-full h-full object-contain" />
                            </div>
                          ) : (
                            <div className="w-10 h-10 rounded-xl bg-white/[0.06] border border-white/[0.08] flex items-center justify-center shrink-0">
                              <Gamepad2 className={`w-5 h-5 ${brand.text}`} />
                            </div>
                          )}
                          <div className="flex-1 min-w-0">
                            <h3 className="text-sm font-bold text-white mb-0.5 truncate">{product.name}</h3>
                            {product.description && (
                              <p className="text-[11px] text-slate-500 line-clamp-2">{product.description}</p>
                            )}
                          </div>
                        </div>
                        {/* Price row */}
                        <div className="flex items-center justify-between pt-3 border-t border-white/[0.06]">
                          <div className="flex items-baseline gap-2">
                            <span className="text-xl font-bold text-white">${product.price_usd}</span>
                            {product.old_price && parseFloat(product.old_price) > parseFloat(product.price_usd) && (
                              <span className="text-xs text-slate-500 line-through">${parseFloat(product.old_price).toFixed(2)}</span>
                            )}
                          </div>
                          {product.in_stock ? (
                            <span className={`flex items-center gap-1 text-xs ${brand.text} opacity-60 group-hover:opacity-100 transition-all font-medium`}>
                              Купить <ChevronRight className="w-3.5 h-3.5" />
                            </span>
                          ) : (
                            <span className="flex items-center gap-1 text-xs text-red-400/60 font-medium">
                              <AlertTriangle className="w-3 h-3" /> Нет в наличии
                            </span>
                          )}
                        </div>
                      </div>
                    </article>
                  );
                })}
              </div>
            )}
          </>
        )}
      </div>

      {/* ═══════ Modals ═══════ */}
      {confirmPlan && (
        <ESIMConfirmModal plan={confirmPlan} loading={esimPurchasing} onConfirm={handleESIMPurchase} onClose={() => !esimPurchasing && setConfirmPlan(null)} />
      )}
      {esimResult && (
        <ESIMActivationModal result={esimResult.result} planName={esimResult.planName} onClose={() => setEsimResult(null)} />
      )}
      {confirmDigital && (
        <DigitalConfirmModal product={confirmDigital} loading={digitalPurchasing} onConfirm={handleDigitalPurchase} onClose={() => !digitalPurchasing && setConfirmDigital(null)} />
      )}
      {digitalResult && (
        <DigitalResultModal productName={digitalResult.productName} priceUsd={digitalResult.priceUsd} activationKey={digitalResult.activationKey} onClose={() => setDigitalResult(null)} />
      )}
      {error && <ErrorToast message={error} onClose={() => setError('')} />}
    </DashboardLayout>
  );
};
