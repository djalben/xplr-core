import { useCallback, useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw, Save, X, Tag, Percent, Image as ImageIcon, Layers, ToggleLeft, ToggleRight } from 'lucide-react';

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
  const [bulkDelta, setBulkDelta] = useState('');
  const [bulkSaving, setBulkSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<StoreProduct[]>('/admin/store/products', { params: { product_type: tab } });
      setRows(res.data || []);
    } catch {
      setError('Не удалось загрузить товары');
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, [tab]);

  useEffect(() => {
    load();
  }, [load]);

  const sorted = useMemo(() => {
    const copy = [...rows];
    copy.sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0));
    return copy;
  }, [rows]);

  const openEdit = (p: StoreProduct) => {
    setEditing({ ...p });
  };

  const closeEdit = () => setEditing(null);

  const save = async () => {
    if (!editing) return;
    setSaving(true);
    setError('');
    try {
      const payload: any = {
        cost_price: parseFloat(editing.costPrice),
        markup_percent: parseFloat(editing.markupPercent),
        image_url: editing.imageUrl,
        retail_price: parseFloat(editing.priceUsd),
        in_stock: !!editing.inStock,
        meta: editing.meta,
        sort_order: Number(editing.sortOrder ?? 0),
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

  const applyBulk = async () => {
    const delta = parseFloat(bulkDelta);
    if (!delta || Number.isNaN(delta)) return;
    setBulkSaving(true);
    setError('');
    try {
      const res = await apiClient.post('/admin/store/bulk-markup', { delta, product_type: tab });
      const affected = res.data?.affected ?? 0;
      setBulkDelta('');
      await load();
      if (affected === 0) {
        setError('Наценка применена, но affected=0 (проверь тип товаров)');
      }
    } catch {
      setError('Не удалось применить массовую наценку');
    } finally {
      setBulkSaving(false);
    }
  };

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
        <TabBtn active={tab === 'vpn'} onClick={() => setTab('vpn')} label="VPN" />
        <TabBtn active={tab === 'esim'} onClick={() => setTab('esim')} label="eSIM" />
        <TabBtn active={tab === 'digital'} onClick={() => setTab('digital')} label="Digital" />
      </div>

      <div className="glass-card p-4 sm:p-6">
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:justify-between">
          <div className="text-sm text-slate-400">
            Массовая наценка для <span className="text-slate-200 font-semibold">{tab}</span>
          </div>
          <div className="flex gap-2 items-center">
            <input
              value={bulkDelta}
              onChange={(e) => setBulkDelta(e.target.value)}
              placeholder="+5"
              className="w-24 bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
            />
            <button
              onClick={applyBulk}
              disabled={bulkSaving}
              className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-sm font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
            >
              <Percent className="w-4 h-4" />
              {bulkSaving ? '...' : 'Применить'}
            </button>
          </div>
        </div>
        {error ? <p className="text-sm text-red-400 mt-4">{error}</p> : null}
      </div>

      <div className="glass-card p-4 sm:p-6 overflow-x-auto">
        <table className="min-w-[1100px] w-full text-left">
          <thead>
            <tr className="text-xs text-slate-500">
              <th className="py-3 px-2 font-semibold">Тип</th>
              <th className="py-3 px-2 font-semibold">Название</th>
              <th className="py-3 px-2 font-semibold">Цена</th>
              <th className="py-3 px-2 font-semibold">Себестоимость</th>
              <th className="py-3 px-2 font-semibold">Наценка %</th>
              <th className="py-3 px-2 font-semibold">InStock</th>
              <th className="py-3 px-2 font-semibold text-right">Действия</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={7} className="py-8 text-center text-slate-400">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin" /> Загрузка...
                  </span>
                </td>
              </tr>
            ) : sorted.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-10 text-center text-slate-500">
                  Пусто
                </td>
              </tr>
            ) : (
              sorted.map((p) => (
                <tr key={p.id} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                  <td className="py-3 px-2 text-xs text-slate-500 font-mono">{p.productType}</td>
                  <td className="py-3 px-2 text-sm text-slate-200">
                    <div className="font-semibold">{p.name}</div>
                    <div className="text-xs text-slate-500">{p.provider} · {p.externalId}</div>
                  </td>
                  <td className="py-3 px-2 text-emerald-400 font-medium">{p.priceUsd}</td>
                  <td className="py-3 px-2 text-emerald-400 font-medium">{p.costPrice}</td>
                  <td className="py-3 px-2 text-emerald-400 font-medium">{p.markupPercent}</td>
                  <td className="py-3 px-2">
                    {p.inStock ? (
                      <span className="inline-flex items-center gap-1 text-xs text-emerald-300">
                        <ToggleRight className="w-4 h-4" /> yes
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-xs text-red-300">
                        <ToggleLeft className="w-4 h-4" /> no
                      </span>
                    )}
                  </td>
                  <td className="py-3 px-2">
                    <div className="flex justify-end">
                      <button
                        onClick={() => openEdit(p)}
                        className="text-blue-400 hover:text-blue-300 text-xs transition-colors"
                      >
                        Редактировать
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {editing ? (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/70 backdrop-blur-sm p-4" onClick={closeEdit}>
          <div className="glass-card w-full max-w-2xl p-6 space-y-4" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <h3 className="text-white font-bold text-lg flex items-center gap-2">
                  <Tag className="w-5 h-5 text-blue-300" /> Редактирование товара
                </h3>
                <p className="text-sm text-slate-400 mt-1 truncate">{editing.name}</p>
              </div>
              <button onClick={closeEdit} className="p-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200">
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs text-slate-500 mb-2">Retail price (USD)</label>
                <input
                  value={editing.priceUsd}
                  onChange={(e) => setEditing({ ...editing, priceUsd: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                />
              </div>
              <div>
                <label className="block text-xs text-slate-500 mb-2">Cost price</label>
                <input
                  value={editing.costPrice}
                  onChange={(e) => setEditing({ ...editing, costPrice: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                />
              </div>
              <div>
                <label className="block text-xs text-slate-500 mb-2">Markup %</label>
                <input
                  value={editing.markupPercent}
                  onChange={(e) => setEditing({ ...editing, markupPercent: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                />
              </div>
              <div>
                <label className="block text-xs text-slate-500 mb-2">Sort order</label>
                <input
                  value={String(editing.sortOrder ?? 0)}
                  onChange={(e) => setEditing({ ...editing, sortOrder: Number(e.target.value) })}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-xs text-slate-500 mb-2">Image URL</label>
                <div className="flex items-center gap-2 bg-white/5 border border-white/10 rounded-xl px-3 py-2">
                  <ImageIcon className="w-4 h-4 text-slate-500" />
                  <input
                    value={editing.imageUrl}
                    onChange={(e) => setEditing({ ...editing, imageUrl: e.target.value })}
                    className="w-full bg-transparent outline-none text-sm text-slate-200"
                  />
                </div>
              </div>
              <div className="md:col-span-2">
                <label className="block text-xs text-slate-500 mb-2">Meta (JSON text)</label>
                <textarea
                  value={editing.meta}
                  onChange={(e) => setEditing({ ...editing, meta: e.target.value })}
                  rows={4}
                  className="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                />
              </div>
              <div className="md:col-span-2 flex items-center justify-between gap-3">
                <div className="flex items-center gap-2 text-sm text-slate-300">
                  <Layers className="w-4 h-4 text-slate-500" />
                  <span>In stock</span>
                </div>
                <button
                  onClick={() => setEditing({ ...editing, inStock: !editing.inStock })}
                  className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors inline-flex items-center gap-2"
                >
                  {editing.inStock ? <ToggleRight className="w-5 h-5 text-emerald-400" /> : <ToggleLeft className="w-5 h-5 text-slate-400" />}
                  {editing.inStock ? 'В наличии' : 'Нет в наличии'}
                </button>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2">
              <button
                onClick={closeEdit}
                disabled={saving}
                className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50"
              >
                Отмена
              </button>
              <button
                onClick={save}
                disabled={saving}
                className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-sm font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
              >
                <Save className="w-4 h-4" />
                {saving ? '...' : 'Сохранить'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

