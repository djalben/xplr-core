import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  FileText, QrCode, Key, Copy, Check, X, Download, Smartphone,
  ArrowLeft, ShoppingBag, Loader2, Clock, ChevronDown, ChevronUp, Wifi
} from 'lucide-react';
import { DashboardLayout } from '../components/dashboard-layout';
import { getStoreOrders, type StoreOrder } from '../services/store';

// ── Country flag emoji ──
const countryFlag = (code: string) => {
  if (!code || code.length < 2 || code === 'GLOBAL') return '\u{1F30D}';
  const codePoints = code.toUpperCase().split('').map(c => 0x1f1e6 + c.charCodeAt(0) - 65);
  return String.fromCodePoint(...codePoints);
};

// ── QR viewer modal (re-view purchased eSIM) ──
const QRViewerModal = ({ order, onClose }: { order: StoreOrder; onClose: () => void }) => {
  const [copied, setCopied] = useState('');

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

  const qrData = order.qr_data || '';
  const activationKey = order.activation_key || '';
  const isEsim = !!qrData;

  const lpaString = qrData;
  const smdp = lpaString.includes('$') ? lpaString.split('$')[1] : '';
  const matchingId = lpaString.includes('$') ? lpaString.split('$')[2] : '';
  const qrUrl = qrData ? `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(qrData)}` : '';

  const downloadInstruction = () => {
    const content = `
XPLR eSIM — Инструкция по активации
=====================================

Товар: ${order.product_name}
Цена: $${order.price_usd}
Дата покупки: ${new Date(order.created_at).toLocaleString('ru-RU')}

─────────────────────────────────────

СПОСОБ 1: QR-код
Откройте Настройки → Сотовая связь → Добавить eSIM → Сканировать QR-код.
Отсканируйте QR-код из приложения XPLR (Магазин → Мои покупки).

СПОСОБ 2: Ручная установка
Откройте Настройки → Сотовая связь → Добавить eSIM → Ввести данные вручную.

SM-DP+ адрес: ${smdp || 'N/A'}
Код активации: ${matchingId || 'N/A'}
LPA строка: ${lpaString || 'N/A'}

─────────────────────────────────────

Поддержка: https://xplr.pro/support
    `.trim();

    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `XPLR_eSIM_${order.id}.txt`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  };

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
          {/* Header */}
          <div className="text-center mb-6">
            <div className="w-14 h-14 rounded-full bg-gradient-to-br from-blue-500/20 to-cyan-500/20 border border-blue-500/30 flex items-center justify-center mx-auto mb-3">
              {isEsim ? <QrCode className="w-7 h-7 text-blue-400" /> : <Key className="w-7 h-7 text-purple-400" />}
            </div>
            <h2 className="text-lg font-bold text-white mb-1">{order.product_name}</h2>
            <p className="text-sm text-slate-400">${order.price_usd} — {new Date(order.created_at).toLocaleDateString('ru-RU')}</p>
          </div>

          {/* eSIM QR + instructions */}
          {isEsim && (
            <>
              <div className="text-center mb-6">
                <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium mb-4">
                  <QrCode className="w-3.5 h-3.5" />
                  Отсканируйте для установки
                </div>
                <div className="bg-white rounded-2xl p-5 inline-block mx-auto shadow-lg">
                  <img src={qrUrl} alt="eSIM QR" className="w-[220px] h-[220px]" />
                </div>
              </div>

              <div className="space-y-3 mb-6">
                <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                  <Smartphone className="w-4 h-4 text-blue-400" />
                  Как установить вручную
                </div>
                <div className="text-xs text-slate-500 mb-3">
                  <p>Настройки → Сотовая связь → Добавить eSIM → Ввести данные вручную</p>
                </div>

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

              <div className="flex gap-3">
                <button
                  onClick={downloadInstruction}
                  className="flex-1 py-3 rounded-xl bg-white/5 border border-white/10 text-slate-300 text-sm font-medium hover:bg-white/10 transition-all flex items-center justify-center gap-2"
                >
                  <Download className="w-4 h-4" />
                  Скачать инструкцию
                </button>
                <button onClick={onClose} className="flex-1 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-purple-600 text-white text-sm font-medium hover:opacity-90 transition-all">
                  Закрыть
                </button>
              </div>
            </>
          )}

          {/* Digital product key */}
          {!isEsim && activationKey && (
            <>
              <div className="mb-6">
                <div className="text-center mb-4">
                  <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-purple-500/10 border border-purple-500/20 text-purple-400 text-xs font-medium">
                    <Key className="w-3.5 h-3.5" />
                    Ключ активации
                  </div>
                </div>
                <div className="bg-white/5 border border-white/10 rounded-xl p-4 flex items-center justify-between gap-3">
                  <code className="text-white font-mono text-sm break-all flex-1 text-left">{activationKey}</code>
                  <button
                    onClick={() => copyText(activationKey, 'key')}
                    className="shrink-0 p-2.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 hover:bg-blue-500/20 transition-all"
                  >
                    {copied === 'key' ? <Check className="w-4 h-4 text-green-400" /> : <Copy className="w-4 h-4" />}
                  </button>
                </div>
              </div>
              <button onClick={onClose} className="w-full py-3 rounded-xl bg-white/5 border border-white/10 text-slate-300 text-sm font-medium hover:bg-white/10 transition-all">
                Закрыть
              </button>
            </>
          )}

          {/* No data */}
          {!isEsim && !activationKey && (
            <>
              <p className="text-sm text-slate-400 text-center mb-6">Данные активации были отправлены в уведомлении.</p>
              <button onClick={onClose} className="w-full py-3 rounded-xl bg-white/5 border border-white/10 text-slate-300 text-sm font-medium hover:bg-white/10 transition-all">
                Закрыть
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

// ── Order card skeleton ──
const OrderSkeleton = () => (
  <div className="glass-card p-5 animate-pulse">
    <div className="flex items-center justify-between mb-3">
      <div className="h-5 bg-white/10 rounded w-1/3" />
      <div className="h-4 bg-white/5 rounded w-16" />
    </div>
    <div className="h-3 bg-white/5 rounded w-2/3 mb-2" />
    <div className="h-3 bg-white/5 rounded w-1/2" />
  </div>
);

// ══════════════════════════════════════════════════════════════
// Purchases Page
// ══════════════════════════════════════════════════════════════

export const PurchasesPage = () => {
  const navigate = useNavigate();
  const [orders, setOrders] = useState<StoreOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [viewOrder, setViewOrder] = useState<StoreOrder | null>(null);

  const loadOrders = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getStoreOrders(50);
      setOrders(res.orders);
    } catch {
      // silent
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadOrders(); }, [loadOrders]);

  const statusLabel = (status: string) => {
    switch (status) {
      case 'completed': return { text: 'Выполнен', color: 'text-green-400 bg-green-500/10 border-green-500/20' };
      case 'pending': return { text: 'В обработке', color: 'text-yellow-400 bg-yellow-500/10 border-yellow-500/20' };
      case 'failed': return { text: 'Ошибка', color: 'text-red-400 bg-red-500/10 border-red-500/20' };
      default: return { text: status, color: 'text-slate-400 bg-white/5 border-white/10' };
    }
  };

  const hasActivationData = (order: StoreOrder) => !!(order.qr_data || order.activation_key);

  return (
    <DashboardLayout>
      <div className="max-w-3xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/store')}
            className="p-2 rounded-xl bg-white/5 border border-white/10 text-slate-400 hover:text-white hover:bg-white/10 transition-all"
          >
            <ArrowLeft className="w-4 h-4" />
          </button>
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center">
              <FileText className="w-5 h-5 text-blue-400" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">Мои покупки</h1>
              <p className="text-xs text-slate-400">История заказов и данные активации</p>
            </div>
          </div>
        </div>

        {/* Orders list */}
        {loading ? (
          <div className="space-y-3">
            {[...Array(4)].map((_, i) => <OrderSkeleton key={i} />)}
          </div>
        ) : orders.length === 0 ? (
          <div className="glass-card p-12 text-center">
            <ShoppingBag className="w-12 h-12 text-slate-600 mx-auto mb-3" />
            <p className="text-slate-400 text-sm mb-4">У вас пока нет покупок</p>
            <button
              onClick={() => navigate('/store')}
              className="px-6 py-2.5 rounded-xl bg-gradient-to-r from-blue-500 to-purple-600 text-white text-sm font-medium hover:opacity-90 transition-all"
            >
              Перейти в магазин
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {orders.map(order => {
              const status = statusLabel(order.status);
              const isEsim = !!order.qr_data;
              return (
                <div
                  key={order.id}
                  onClick={() => hasActivationData(order) && setViewOrder(order)}
                  className={`glass-card p-5 transition-all ${
                    hasActivationData(order)
                      ? 'cursor-pointer hover:border-blue-500/20 hover:shadow-lg hover:shadow-blue-500/5'
                      : ''
                  }`}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex items-center gap-3">
                      <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
                        isEsim
                          ? 'bg-gradient-to-br from-blue-500/20 to-cyan-500/20 border border-blue-500/30'
                          : 'bg-gradient-to-br from-purple-500/20 to-pink-500/20 border border-purple-500/30'
                      }`}>
                        {isEsim ? <Wifi className="w-5 h-5 text-blue-400" /> : <Key className="w-5 h-5 text-purple-400" />}
                      </div>
                      <div>
                        <h3 className="text-sm font-bold text-white">{order.product_name}</h3>
                        <p className="text-[11px] text-slate-500">
                          {new Date(order.created_at).toLocaleDateString('ru-RU', {
                            day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
                          })}
                        </p>
                      </div>
                    </div>
                    <div className="text-right">
                      <span className="text-sm font-bold text-white">${order.price_usd}</span>
                      <div className={`mt-1 inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-medium border ${status.color}`}>
                        {status.text}
                      </div>
                    </div>
                  </div>

                  {/* Quick info */}
                  {hasActivationData(order) && (
                    <div className="flex items-center justify-between mt-2 pt-2 border-t border-white/5">
                      <div className="flex items-center gap-2">
                        {isEsim ? (
                          <span className="inline-flex items-center gap-1 text-[11px] text-blue-400/70">
                            <QrCode className="w-3 h-3" /> QR-код доступен
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-[11px] text-purple-400/70">
                            <Key className="w-3 h-3" /> Ключ доступен
                          </span>
                        )}
                      </div>
                      <span className="text-[11px] text-slate-500">Нажмите для просмотра</span>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* QR/Key Viewer Modal */}
      {viewOrder && <QRViewerModal order={viewOrder} onClose={() => setViewOrder(null)} />}
    </DashboardLayout>
  );
};
