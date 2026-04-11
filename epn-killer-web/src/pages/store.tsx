import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ShoppingBag, Search, Globe, Gamepad2, X, Copy, Check, Loader2,
  AlertTriangle, QrCode, Key, ChevronRight, ArrowLeft,
  Download, FileText, Smartphone, Wifi
} from 'lucide-react';
import { DashboardLayout } from '../components/dashboard-layout';
import {
  getStoreCatalog, purchaseProduct,
  getESIMDestinations, getESIMPlans, orderESIM,
  type StoreProduct, type ESIMDestination, type ESIMPlan, type ESIMOrderResult,
} from '../api/store';

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
    if (storeView !== 'esim') return;
    const t = setTimeout(() => loadDestinations(), destsSearch ? 300 : 0);
    return () => clearTimeout(t);
  }, [loadDestinations, destsSearch, storeView]);

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
    if (storeView === 'digital') loadDigital();
  }, [storeView, loadDigital]);

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

  return (
    <DashboardLayout>
      <div className="max-w-5xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            {storeView !== 'hub' && (
              <button
                onClick={() => { setStoreView('hub'); setSelectedCountry(null); setDestsSearch(''); }}
                className="p-2 -ml-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/5 transition-all"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
            )}
            <div>
              <h1 className="text-xl font-bold text-white">
                {storeView === 'hub' && 'Магазин'}
                {storeView === 'esim' && 'eSIM / Сим-карты'}
                {storeView === 'digital' && 'Цифровые товары'}
              </h1>
              <p className="text-xs text-slate-400">
                {storeView === 'hub' && 'Выберите категорию'}
                {storeView === 'esim' && 'Интернет для путешествий'}
                {storeView === 'digital' && 'Ключи, игры и софт'}
              </p>
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

        {/* ═══════ Hub — Category Cards ═══════ */}
        {storeView === 'hub' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8 md:gap-12 pt-4">
            {/* eSIM Category */}
            <button
              onClick={() => setStoreView('esim')}
              className="group relative rounded-2xl border border-white/[0.08] hover:border-cyan-500/30 bg-transparent backdrop-blur-sm p-8 md:p-10 flex flex-col items-center text-center transition-all duration-300 hover:shadow-2xl hover:shadow-cyan-500/10 hover:-translate-y-1"
            >
              <div className="w-52 h-32 md:w-64 md:h-40 rounded-[50%] overflow-hidden border-2 border-cyan-500/20 group-hover:border-cyan-400/50 transition-colors shadow-lg shadow-cyan-500/5 mb-6">
                <img src="/store-esim.png" alt="eSIM" className="w-full h-full object-cover scale-110 group-hover:scale-125 transition-transform duration-500" />
              </div>
              <h2 className="text-lg font-bold text-white mb-1 group-hover:text-cyan-300 transition-colors">eSIM / Сим-карты</h2>
              <p className="text-xs text-slate-500">Интернет для путешествий</p>
              <ChevronRight className="absolute right-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-700 group-hover:text-cyan-400 transition-colors hidden md:block" />
            </button>

            {/* Digital Category */}
            <button
              onClick={() => setStoreView('digital')}
              className="group relative rounded-2xl border border-white/[0.08] hover:border-purple-500/30 bg-transparent backdrop-blur-sm p-8 md:p-10 flex flex-col items-center text-center transition-all duration-300 hover:shadow-2xl hover:shadow-purple-500/10 hover:-translate-y-1"
            >
              <div className="w-52 h-32 md:w-64 md:h-40 rounded-[50%] overflow-hidden border-2 border-purple-500/20 group-hover:border-purple-400/50 transition-colors shadow-lg shadow-purple-500/5 mb-6">
                <img src="/store-digital.png" alt="Digital" className="w-full h-full object-cover scale-110 group-hover:scale-125 transition-transform duration-500" />
              </div>
              <h2 className="text-lg font-bold text-white mb-1 group-hover:text-purple-300 transition-colors">Цифровые товары</h2>
              <p className="text-xs text-slate-500">Ключи, игры и софт</p>
              <ChevronRight className="absolute right-5 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-700 group-hover:text-purple-400 transition-colors hidden md:block" />
            </button>
          </div>
        )}

        {/* ═══════ eSIM — Country List ═══════ */}
        {storeView === 'esim' && !selectedCountry && (
          <>
            {/* Country search */}
            <div className="relative">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
              <input
                type="text"
                placeholder="Поиск по стране..."
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

            {/* Country list */}
            {destsLoading ? (
              <div className="rounded-2xl border border-white/[0.06] overflow-hidden bg-white/[0.02]">
                {[...Array(8)].map((_, i) => (
                  <div key={i} className="flex items-center gap-4 px-5 py-4 border-b border-white/[0.04] last:border-0 animate-pulse">
                    <div className="w-8 h-6 rounded bg-white/5" />
                    <div className="flex-1"><div className="h-4 bg-white/[0.06] rounded w-32" /></div>
                    <div className="h-3 bg-white/[0.04] rounded w-16" />
                  </div>
                ))}
              </div>
            ) : destinations.length === 0 ? (
              <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-12 text-center">
                <Globe className="w-10 h-10 text-slate-600 mx-auto mb-3" />
                <p className="text-slate-400 text-sm">{destsSearch ? 'Страна не найдена' : 'Направления загружаются...'}</p>
                {destsSearch && (
                  <button onClick={() => setDestsSearch('')} className="mt-3 text-xs text-blue-400 hover:text-blue-300">Сбросить поиск</button>
                )}
              </div>
            ) : (
              <div className="rounded-2xl border border-white/[0.06] overflow-hidden bg-white/[0.02]">
                {destinations.map((dest, idx) => (
                  <button
                    key={dest.country_code}
                    onClick={() => setSelectedCountry(dest)}
                    className={`w-full flex items-center gap-4 px-5 py-3.5 hover:bg-white/[0.04] transition-colors group ${
                      idx < destinations.length - 1 ? 'border-b border-white/[0.04]' : ''
                    }`}
                  >
                    <span className="text-2xl leading-none select-none w-8 text-center">{dest.flag_emoji || countryFlag(dest.country_code)}</span>
                    <span className="flex-1 text-left text-sm font-medium text-white">{dest.country_name}</span>
                    <span className="text-[11px] text-slate-500 font-medium">{dest.plan_count} {dest.plan_count === 1 ? 'тариф' : 'тарифов'}</span>
                    <ChevronRight className="w-4 h-4 text-slate-700 group-hover:text-blue-400 transition-colors" />
                  </button>
                ))}
              </div>
            )}
          </>
        )}

        {/* ═══════ eSIM Plans (selected country) ═══════ */}
        {storeView === 'esim' && selectedCountry && (
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
              <div className="rounded-2xl border border-white/[0.06] overflow-hidden bg-white/[0.02]">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="flex items-center gap-4 px-5 py-4 border-b border-white/[0.04] last:border-0 animate-pulse">
                    <div className="w-8 h-6 rounded bg-white/5" />
                    <div className="flex-1 space-y-1.5">
                      <div className="h-4 bg-white/[0.06] rounded w-40" />
                      <div className="h-3 bg-white/[0.04] rounded w-24" />
                    </div>
                    <div className="h-5 bg-white/[0.06] rounded w-16" />
                    <div className="h-8 bg-white/[0.04] rounded-lg w-20" />
                  </div>
                ))}
              </div>
            ) : plans.length === 0 ? (
              <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-12 text-center">
                <Wifi className="w-10 h-10 text-slate-600 mx-auto mb-3" />
                <p className="text-slate-400 text-sm">Нет доступных тарифов</p>
              </div>
            ) : (
              <div className="rounded-2xl border border-white/[0.06] overflow-hidden bg-white/[0.02]">
                {plans.map((plan, idx) => (
                  <div
                    key={plan.plan_id}
                    className={`flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-4 transition-colors ${
                      plan.in_stock ? 'hover:bg-white/[0.04]' : 'opacity-50'
                    } ${idx < plans.length - 1 ? 'border-b border-white/[0.04]' : ''}`}
                  >
                    <span className="text-xl leading-none select-none w-7 text-center shrink-0">{countryFlag(selectedCountry.country_code)}</span>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-white truncate">{plan.name}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <span className="text-[11px] text-blue-400 font-medium">{plan.data_gb} ГБ</span>
                        <span className="text-slate-600 text-[10px]">•</span>
                        <span className="text-[11px] text-slate-500">{plan.validity_days} дн.</span>
                      </div>
                    </div>
                    <div className="text-right shrink-0">
                      <span className="text-sm font-bold text-white">${plan.price_usd.toFixed(2)}</span>
                      {plan.old_price > 0 && plan.old_price > plan.price_usd && (
                        <span className="block text-[10px] text-slate-600 line-through">${plan.old_price.toFixed(2)}</span>
                      )}
                    </div>
                    {plan.in_stock ? (
                      <button
                        onClick={() => setConfirmPlan(plan)}
                        className="shrink-0 px-4 py-2 rounded-xl bg-blue-500/10 border border-blue-500/25 text-blue-400 text-xs font-semibold hover:bg-blue-500/20 hover:border-blue-500/40 transition-all"
                      >
                        Купить
                      </button>
                    ) : (
                      <span className="shrink-0 px-3 py-2 rounded-xl bg-white/5 text-slate-600 text-[11px] font-medium">
                        Нет в наличии
                      </span>
                    )}
                  </div>
                ))}
              </div>
            )}
          </>
        )}

        {/* ═══════ Digital — Products List ═══════ */}
        {storeView === 'digital' && (
          <>
            {digitalLoading ? (
              <div className="rounded-2xl border border-white/[0.06] overflow-hidden bg-white/[0.02]">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="flex items-center gap-4 px-5 py-4 border-b border-white/[0.04] last:border-0 animate-pulse">
                    <div className="w-8 h-8 rounded-lg bg-white/5" />
                    <div className="flex-1 space-y-1.5">
                      <div className="h-4 bg-white/[0.06] rounded w-36" />
                      <div className="h-3 bg-white/[0.04] rounded w-24" />
                    </div>
                    <div className="h-5 bg-white/[0.06] rounded w-14" />
                    <div className="h-8 bg-white/[0.04] rounded-lg w-20" />
                  </div>
                ))}
              </div>
            ) : digitalProducts.length === 0 ? (
              <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-12 text-center">
                <Gamepad2 className="w-10 h-10 text-slate-600 mx-auto mb-3" />
                <p className="text-slate-400 text-sm">Товаров пока нет</p>
              </div>
            ) : (
              <div className="rounded-2xl border border-white/[0.06] overflow-hidden bg-white/[0.02]">
                {digitalProducts.map((product, idx) => (
                  <div
                    key={product.id}
                    className={`flex items-center gap-3 sm:gap-4 px-4 sm:px-5 py-4 transition-colors ${
                      product.in_stock ? 'hover:bg-white/[0.04]' : 'opacity-50'
                    } ${idx < digitalProducts.length - 1 ? 'border-b border-white/[0.04]' : ''}`}
                  >
                    {product.image_url ? (
                      <img src={product.image_url} alt="" className="w-8 h-8 rounded-lg object-contain shrink-0 bg-white/[0.03] p-0.5" />
                    ) : (
                      <div className="w-8 h-8 rounded-lg bg-white/[0.03] flex items-center justify-center shrink-0">
                        <Gamepad2 className="w-4 h-4 text-slate-600" />
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-white truncate">{product.name}</p>
                      {product.description && (
                        <p className="text-[11px] text-slate-500 truncate mt-0.5">{product.description}</p>
                      )}
                    </div>
                    <div className="text-right shrink-0">
                      <span className="text-sm font-bold text-white">${product.price_usd}</span>
                      {product.old_price && parseFloat(product.old_price) > parseFloat(product.price_usd) && (
                        <span className="block text-[10px] text-slate-600 line-through">${parseFloat(product.old_price).toFixed(2)}</span>
                      )}
                    </div>
                    {product.in_stock ? (
                      <button
                        onClick={() => setConfirmDigital(product)}
                        className="shrink-0 px-4 py-2 rounded-xl bg-purple-500/10 border border-purple-500/25 text-purple-400 text-xs font-semibold hover:bg-purple-500/20 hover:border-purple-500/40 transition-all"
                      >
                        Купить
                      </button>
                    ) : (
                      <span className="shrink-0 px-3 py-2 rounded-xl bg-white/5 text-slate-600 text-[11px] font-medium">
                        Нет в наличии
                      </span>
                    )}
                  </div>
                ))}
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
