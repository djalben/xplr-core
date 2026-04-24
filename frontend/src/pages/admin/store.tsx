import { useCallback, useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw, Shield } from 'lucide-react';

type StoreProduct = {
  id: string;
  categoryId: string;
  categorySlug: string;
  provider: string;
  externalId: string;
  name: string;
  description: string;
  country: string;
  countryCode: string;
  productType: 'esim' | 'digital' | 'vpn' | string;
  priceUsd: string;
  costPrice: string;
  markupPercent: string;
  oldPrice: string;
  dataGb: string;
  validityDays: number;
  imageUrl: string;
  inStock: boolean;
  meta: string;
  sortOrder: number;
};

type ProductTypeTab = 'vpn' | 'esim' | 'digital';

const TabBtn = ({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) => (
  <button
    onClick={onClick}
    className={`px-3 py-2 rounded-xl text-xs font-semibold border transition-colors ${
      active ? 'bg-blue-500/20 border-blue-500/30 text-blue-200' : 'bg-white/5 border-white/10 text-slate-300 hover:bg-white/10'
    }`}
  >
    {label}
  </button>
);

export const AdminStorePage = () => {
  const [tab, setTab] = useState<ProductTypeTab>('vpn');
  const [rows, setRows] = useState<StoreProduct[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState<StoreProduct | null>(null);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<StoreProduct[]>('/admin/store/products');
      setRows(res.data || []);
    } catch {
      setError('Не удалось загрузить товары');
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const sorted = useMemo(() => {
    const copy = rows.filter((p) => p.productType === tab);
    copy.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
    return copy;
  }, [rows, tab]);

  const openEdit = (p: StoreProduct) => {
    setEditing({ ...p });
  };

  const closeEdit = () => setEditing(null);

  const save = async () => {
    if (!editing) return;
    setSaving(true);
    setError('');
    try {
      const payload: Record<string, unknown> = {
        retail_price: parseFloat(editing.priceUsd),
        cost_price: parseFloat(editing.costPrice),
        markup_percent: parseFloat(editing.markupPercent),
      };
      await apiClient.patch(`/admin/store/products/${editing.id}`, payload);
      setEditing(null);
      await load();
    } catch {
      setError('Не удалось сохранить товар');
    } finally {
      setSaving(false);
    }
  };

  const fmt = (v: string) => {
    const n = Number(v);
    return Number.isFinite(n) ? n.toFixed(2) : v;
  };

  const productCount = (productType: ProductTypeTab) => rows.filter((p) => p.productType === productType).length;

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Магазин</h1>
          <p className="text-sm text-slate-400 mt-1">Товары: eSIM / VPN / Digital — цены, наценка, наличие</p>
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
        >
          <RefreshCw className="w-4 h-4" />
          Обновить
        </button>
      </div>

      <div className="flex flex-wrap gap-2">
        <TabBtn active={tab === 'esim'} onClick={() => { setTab('esim'); setEditing(null); }} label={`eSIM (${productCount('esim')})`} />
        <TabBtn active={tab === 'digital'} onClick={() => { setTab('digital'); setEditing(null); }} label={`Цифровые (${productCount('digital')})`} />
        <TabBtn active={tab === 'vpn'} onClick={() => { setTab('vpn'); setEditing(null); }} label={`VPN (${productCount('vpn')})`} />
      </div>

      {error ? <p className="text-sm text-red-400">{error}</p> : null}

      <div className="glass-card overflow-hidden">
        <div className="overflow-x-auto">
        <table className="w-full text-sm min-w-[760px]">
          <thead>
            <tr className="border-b border-white/10">
              <th className="text-left px-4 py-3 text-slate-400 font-medium">ID</th>
              <th className="text-left px-4 py-3 text-slate-400 font-medium">Товар</th>
              <th className="text-left px-4 py-3 text-slate-400 font-medium">Себестоимость</th>
              <th className="text-left px-4 py-3 text-slate-400 font-medium">Наценка %</th>
              <th className="text-left px-4 py-3 text-slate-400 font-medium">Розница</th>
              <th className="text-left px-4 py-3 text-slate-400 font-medium">Старая цена</th>
              <th className="text-left px-4 py-3 text-slate-400 font-medium">Статус</th>
              <th className="text-left px-4 py-3 text-slate-400 font-medium"></th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-slate-400">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin" /> Загрузка...
                  </span>
                </td>
              </tr>
            ) : sorted.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-slate-500">
                  Пусто
                </td>
              </tr>
            ) : (
              sorted.map((p) => {
                const isEditing = editing?.id === p.id;
                return (
                  <tr key={p.id} className={`border-b border-white/5 hover:bg-white/5 transition-colors ${isEditing ? 'bg-blue-500/5' : ''}`}>
                    <td className="px-4 py-3 text-slate-500 font-mono text-xs">{p.sortOrder || p.id.slice(0, 6)}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2.5">
                        {p.productType === 'vpn' ? (
                          <div className="w-7 h-7 rounded bg-gradient-to-br from-indigo-600/30 to-purple-600/30 border border-indigo-500/20 flex items-center justify-center shrink-0">
                            <Shield className="w-3.5 h-3.5 text-indigo-400" />
                          </div>
                        ) : p.imageUrl ? (
                          <img src={p.imageUrl} alt="" className="w-7 h-7 rounded object-contain shrink-0 bg-white/5 p-0.5" />
                        ) : (
                          <div className="w-7 h-7 rounded bg-white/5 shrink-0" />
                        )}
                        <div className="min-w-0">
                          <span className="text-white text-xs font-medium truncate block max-w-[180px]" title={p.name}>{p.name}</span>
                          <span className="text-[10px] text-slate-500 font-mono">{p.externalId}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      {isEditing ? (
                        <input
                          type="number"
                          step="0.01"
                          value={editing.costPrice}
                          onChange={(e) => setEditing({ ...editing, costPrice: e.target.value })}
                          className="w-20 px-2 py-1 bg-white/10 border border-blue-500/50 rounded text-white text-xs outline-none"
                        />
                      ) : (
                        <span className="text-emerald-400 font-medium text-xs">${fmt(p.costPrice)}</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {isEditing ? (
                        <input
                          type="number"
                          step="1"
                          value={editing.markupPercent}
                          onChange={(e) => setEditing({ ...editing, markupPercent: e.target.value })}
                          className="w-16 px-2 py-1 bg-white/10 border border-blue-500/50 rounded text-white text-xs outline-none"
                        />
                      ) : (
                        <span className="text-orange-400 font-medium text-xs">{Number(p.markupPercent).toFixed(0)}%</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {isEditing ? (
                        <input
                          type="number"
                          step="0.01"
                          value={editing.priceUsd}
                          onChange={(e) => setEditing({ ...editing, priceUsd: e.target.value })}
                          className="w-20 px-2 py-1 bg-white/10 border border-blue-500/50 rounded text-white text-xs outline-none"
                          autoFocus={tab === 'vpn'}
                        />
                      ) : (
                        <span className="text-white font-bold text-xs">${fmt(p.priceUsd)}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-slate-500 text-xs line-through">${fmt(p.oldPrice)}</td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-0.5 rounded-full text-[10px] font-medium ${p.inStock ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-400'}`}>
                        {p.inStock ? 'В наличии' : 'Нет'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      {isEditing ? (
                        <div className="flex gap-1">
                          <button onClick={save} disabled={saving} className="px-2.5 py-1 bg-emerald-500 hover:bg-emerald-600 text-white rounded text-xs transition-colors disabled:opacity-50">
                            {saving ? '...' : 'OK'}
                          </button>
                          <button onClick={closeEdit} disabled={saving} className="px-2.5 py-1 bg-white/10 hover:bg-white/20 text-slate-300 rounded text-xs transition-colors disabled:opacity-50">
                            ×
                          </button>
                        </div>
                      ) : (
                        <button onClick={() => openEdit(p)} className="text-blue-400 hover:text-blue-300 text-xs transition-colors">
                          Изменить
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
        </div>
      </div>
    </div>
  );
};

