import { useState, useEffect, useCallback } from 'react';
import { ShoppingBag, Search, Globe, Gamepad2, X, Copy, Check, Loader2, AlertTriangle, QrCode, Key, ChevronRight } from 'lucide-react';
import { DashboardLayout } from '../components/dashboard-layout';
import { getStoreCatalog, purchaseProduct, type StoreProduct, type StoreCategory } from '../api/store';

// ── Skeleton Loader ──
const ProductSkeleton = () => (
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

// ── Country flag emoji from country code ──
const countryFlag = (code: string) => {
  if (!code || code.length < 2 || code === 'GLOBAL') return '\u{1F30D}';
  const codePoints = code
    .toUpperCase()
    .split('')
    .map(c => 0x1f1e6 + c.charCodeAt(0) - 65);
  return String.fromCodePoint(...codePoints);
};

// ── Purchase Confirmation Modal ──
const ConfirmModal = ({
  product,
  loading,
  onConfirm,
  onClose,
}: {
  product: StoreProduct;
  loading: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) => {
  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape' && !loading) onClose(); };
    document.addEventListener('keydown', h);
    document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose, loading]);

  const isEsim = product.product_type === 'esim';

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={() => !loading && onClose()}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div className="relative w-full max-w-md rounded-2xl bg-[#111118] border border-white/10 shadow-2xl" onClick={e => e.stopPropagation()}>
        <button onClick={onClose} disabled={loading} className="absolute top-3 right-3 z-10 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all disabled:opacity-30">
          <X className="w-4 h-4" />
        </button>

        <div className="p-6">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center">
              <ShoppingBag className="w-6 h-6 text-blue-400" />
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
            {isEsim && product.data_gb && (
              <div className="flex justify-between items-center py-2.5 border-b border-white/5">
                <span className="text-sm text-slate-400">Объём</span>
                <span className="text-sm text-white font-medium">{product.data_gb} ГБ</span>
              </div>
            )}
            {isEsim && product.validity_days > 0 && (
              <div className="flex justify-between items-center py-2.5 border-b border-white/5">
                <span className="text-sm text-slate-400">Срок</span>
                <span className="text-sm text-white font-medium">{product.validity_days} дней</span>
              </div>
            )}
            {isEsim && product.country && (
              <div className="flex justify-between items-center py-2.5 border-b border-white/5">
                <span className="text-sm text-slate-400">Страна</span>
                <span className="text-sm text-white font-medium">{countryFlag(product.country_code)} {product.country}</span>
              </div>
            )}
            <div className="flex justify-between items-center py-3 bg-blue-500/5 rounded-xl px-3 border border-blue-500/10">
              <span className="text-sm text-slate-300 font-medium">Итого</span>
              <span className="text-xl text-blue-400 font-bold">${product.price_usd}</span>
            </div>
          </div>

          <div className="flex gap-3">
            <button
              onClick={onClose}
              disabled={loading}
              className="flex-1 py-3 rounded-xl bg-white/5 border border-white/10 text-slate-400 text-sm font-medium hover:bg-white/10 transition-all disabled:opacity-30"
            >
              Отмена
            </button>
            <button
              onClick={onConfirm}
              disabled={loading}
              className="flex-1 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-600 text-white text-sm font-medium hover:opacity-90 transition-all disabled:opacity-60 flex items-center justify-center gap-2"
            >
              {loading ? <><Loader2 className="w-4 h-4 animate-spin" /> Покупка...</> : 'Купить'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

// ── Purchase Result Modal (QR / Key) ──
const ResultModal = ({
  productName,
  priceUsd,
  activationKey,
  qrData,
  onClose,
}: {
  productName: string;
  priceUsd: string;
  activationKey: string;
  qrData: string;
  onClose: () => void;
}) => {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const h = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', h);
    document.body.style.overflow = 'hidden';
    return () => { document.removeEventListener('keydown', h); document.body.style.overflow = ''; };
  }, [onClose]);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  const displayValue = activationKey || qrData || '';

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div className="relative w-full max-w-md rounded-2xl bg-[#111118] border border-white/10 shadow-2xl" onClick={e => e.stopPropagation()}>
        <button onClick={onClose} className="absolute top-3 right-3 z-10 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all">
          <X className="w-4 h-4" />
        </button>

        <div className="p-6 text-center">
          {/* Success animation */}
          <div className="w-16 h-16 rounded-full bg-gradient-to-br from-green-500/20 to-emerald-500/20 border border-green-500/30 flex items-center justify-center mx-auto mb-4">
            <Check className="w-8 h-8 text-green-400" />
          </div>

          <h2 className="text-lg font-bold text-white mb-1">Покупка успешна!</h2>
          <p className="text-sm text-slate-400 mb-6">{productName} — ${priceUsd}</p>

          {/* QR Code area */}
          {qrData && (
            <div className="mb-6">
              <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium mb-4">
                <QrCode className="w-3.5 h-3.5" />
                eSIM QR-код
              </div>
              <div className="bg-white rounded-2xl p-6 inline-block mx-auto mb-3">
                {/* QR rendered via Google Charts API */}
                <img
                  src={`https://chart.googleapis.com/chart?cht=qr&chs=200x200&chl=${encodeURIComponent(qrData)}&choe=UTF-8`}
                  alt="QR Code"
                  className="w-[200px] h-[200px]"
                />
              </div>
              <p className="text-[11px] text-slate-500 mb-2">Отсканируйте камерой для установки eSIM</p>
            </div>
          )}

          {/* Activation Key area */}
          {activationKey && (
            <div className="mb-6">
              <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400 text-xs font-medium mb-4">
                <Key className="w-3.5 h-3.5" />
                Ключ активации
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-4 flex items-center justify-between gap-3">
                <code className="text-white font-mono text-sm break-all flex-1 text-left">{activationKey}</code>
                <button
                  onClick={() => copyToClipboard(activationKey)}
                  className="shrink-0 p-2.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 hover:bg-blue-500/20 transition-all"
                >
                  {copied ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
                </button>
              </div>
            </div>
          )}

          {/* Copy LPA for eSIM */}
          {qrData && (
            <div className="mb-6">
              <p className="text-[11px] text-slate-500 mb-2">Или скопируйте код активации вручную:</p>
              <div className="bg-white/5 border border-white/10 rounded-xl p-3 flex items-center justify-between gap-3">
                <code className="text-slate-300 font-mono text-[11px] break-all flex-1 text-left">{qrData}</code>
                <button
                  onClick={() => copyToClipboard(qrData)}
                  className="shrink-0 p-2 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 hover:bg-blue-500/20 transition-all"
                >
                  {copied ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
                </button>
              </div>
            </div>
          )}

          {!displayValue && (
            <p className="text-sm text-slate-400 mb-6">Данные активации будут отправлены в уведомлении.</p>
          )}

          <button
            onClick={onClose}
            className="w-full py-3 rounded-xl bg-white/5 border border-white/10 text-slate-300 text-sm font-medium hover:bg-white/10 transition-all"
          >
            Закрыть
          </button>
        </div>
      </div>
    </div>
  );
};

// ── Error toast ──
const ErrorToast = ({ message, onClose }: { message: string; onClose: () => void }) => (
  <div className="fixed bottom-24 left-1/2 -translate-x-1/2 z-[110] max-w-sm w-full mx-4 animate-in fade-in slide-in-from-bottom-4">
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
// Store Page
// ══════════════════════════════════════════════════════════════

export const StorePage = () => {
  const [categories, setCategories] = useState<StoreCategory[]>([]);
  const [products, setProducts] = useState<StoreProduct[]>([]);
  const [activeTab, setActiveTab] = useState('esim');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);

  // Purchase flow
  const [confirmProduct, setConfirmProduct] = useState<StoreProduct | null>(null);
  const [purchasing, setPurchasing] = useState(false);
  const [result, setResult] = useState<{ productName: string; priceUsd: string; activationKey: string; qrData: string } | null>(null);
  const [error, setError] = useState('');

  const loadCatalog = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getStoreCatalog({ category: activeTab, search: search || undefined });
      setCategories(res.categories);
      setProducts(res.products);
    } catch {
      setError('Не удалось загрузить каталог');
    } finally {
      setLoading(false);
    }
  }, [activeTab, search]);

  useEffect(() => {
    const debounce = setTimeout(() => loadCatalog(), search ? 300 : 0);
    return () => clearTimeout(debounce);
  }, [loadCatalog, search]);

  const handlePurchase = async () => {
    if (!confirmProduct) return;
    setPurchasing(true);
    setError('');
    try {
      const res = await purchaseProduct(confirmProduct.id);
      setConfirmProduct(null);
      setResult({
        productName: res.product_name,
        priceUsd: res.price_usd,
        activationKey: res.activation_key,
        qrData: res.qr_data,
      });
    } catch (err: any) {
      setConfirmProduct(null);
      const msg = err?.response?.data?.error || err?.response?.data || 'Ошибка при покупке';
      setError(typeof msg === 'string' ? msg : 'Ошибка при покупке');
    } finally {
      setPurchasing(false);
    }
  };

  const tabs = [
    { slug: 'esim', label: 'eSIM — Весь мир', icon: <Globe className="w-4 h-4" /> },
    { slug: 'digital', label: 'Цифровые товары', icon: <Gamepad2 className="w-4 h-4" /> },
  ];

  const isEsimTab = activeTab === 'esim';

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
        </div>

        {/* Tabs */}
        <div className="flex gap-2">
          {tabs.map(tab => (
            <button
              key={tab.slug}
              onClick={() => { setActiveTab(tab.slug); setSearch(''); }}
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

        {/* Search (eSIM tab) */}
        {isEsimTab && (
          <div className="relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
            <input
              type="text"
              placeholder="Поиск по стране (Турция, USA, Japan...)"
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="w-full pl-11 pr-4 py-3 bg-white/5 border border-white/10 rounded-xl text-sm text-white placeholder-slate-500 outline-none focus:border-blue-500/50 transition-colors"
            />
            {search && (
              <button
                onClick={() => setSearch('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 p-1 rounded-lg hover:bg-white/10 text-slate-500 hover:text-white transition-all"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
        )}

        {/* Products grid */}
        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(6)].map((_, i) => <ProductSkeleton key={i} />)}
          </div>
        ) : products.length === 0 ? (
          <div className="glass-card p-12 text-center">
            <ShoppingBag className="w-12 h-12 text-slate-600 mx-auto mb-3" />
            <p className="text-slate-400 text-sm">{search ? 'Ничего не найдено' : 'Товаров пока нет'}</p>
            {search && (
              <button onClick={() => setSearch('')} className="mt-3 text-xs text-blue-400 hover:text-blue-300 transition-colors">
                Сбросить поиск
              </button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {products.map(product => (
              <article
                key={product.id}
                onClick={() => product.in_stock && setConfirmProduct(product)}
                className={`glass-card p-5 transition-all group ${
                  product.in_stock
                    ? 'cursor-pointer hover:border-blue-500/20 hover:shadow-lg hover:shadow-blue-500/5'
                    : 'opacity-50 cursor-not-allowed'
                }`}
              >
                {/* Product name + country flag */}
                <div className="flex items-start gap-3 mb-3">
                  {isEsimTab && (
                    <span className="text-2xl shrink-0 mt-0.5">{countryFlag(product.country_code)}</span>
                  )}
                  <div className="flex-1 min-w-0">
                    <h3 className="text-sm font-bold text-white mb-1 truncate">{product.name}</h3>
                    {isEsimTab && product.country && (
                      <p className="text-xs text-slate-500">{product.country}</p>
                    )}
                    {!isEsimTab && product.description && (
                      <p className="text-xs text-slate-500 line-clamp-2">{product.description}</p>
                    )}
                  </div>
                </div>

                {/* eSIM specs */}
                {isEsimTab && (
                  <div className="flex items-center gap-3 mb-4">
                    {product.data_gb && (
                      <span className="px-2.5 py-1 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-[11px] font-medium">
                        {product.data_gb} ГБ
                      </span>
                    )}
                    {product.validity_days > 0 && (
                      <span className="px-2.5 py-1 rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400 text-[11px] font-medium">
                        {product.validity_days} дн.
                      </span>
                    )}
                  </div>
                )}

                {/* Price + Buy CTA */}
                <div className="flex items-center justify-between mt-auto">
                  <span className="text-lg font-bold text-white">${product.price_usd}</span>
                  {product.in_stock ? (
                    <span className="flex items-center gap-1 text-xs text-blue-400/60 group-hover:text-blue-400 transition-colors font-medium">
                      Купить <ChevronRight className="w-3.5 h-3.5" />
                    </span>
                  ) : (
                    <span className="flex items-center gap-1 text-xs text-red-400/60 font-medium">
                      <AlertTriangle className="w-3 h-3" /> Нет в наличии
                    </span>
                  )}
                </div>
              </article>
            ))}
          </div>
        )}
      </div>

      {/* Modals */}
      {confirmProduct && (
        <ConfirmModal
          product={confirmProduct}
          loading={purchasing}
          onConfirm={handlePurchase}
          onClose={() => !purchasing && setConfirmProduct(null)}
        />
      )}
      {result && (
        <ResultModal
          productName={result.productName}
          priceUsd={result.priceUsd}
          activationKey={result.activationKey}
          qrData={result.qrData}
          onClose={() => setResult(null)}
        />
      )}
      {error && <ErrorToast message={error} onClose={() => setError('')} />}
    </DashboardLayout>
  );
};
