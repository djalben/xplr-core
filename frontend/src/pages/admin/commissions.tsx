import { useCallback, useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw, Save } from 'lucide-react';

type CommissionRow = {
  id: string;
  key: string;
  value: string;
  description: string;
  updatedAt?: string;
};

type ExchangeRateRow = {
  id: string;
  currency_from: string;
  currency_to: string;
  base_rate: string;
  markup_percent: string;
  final_rate: string;
  updated_at?: string;
};

const keyLabel = (key: string) => {
  switch (key) {
    case 'fee_standard':
      return 'Комиссия STANDARD (%)';
    case 'fee_gold':
      return 'Комиссия GOLD (%)';
    case 'referral_percent':
      return 'Реферальный процент (%)';
    case 'card_issue_fee':
      return 'Выпуск карты ($)';
    default:
      return key;
  }
};

export const AdminCommissionsPage = () => {
  const [rows, setRows] = useState<CommissionRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [valueDraft, setValueDraft] = useState('');
  const [saving, setSaving] = useState(false);

  const [rates, setRates] = useState<ExchangeRateRow[]>([]);
  const [ratesLoading, setRatesLoading] = useState(false);
  const [ratesError, setRatesError] = useState('');
  const [editingPair, setEditingPair] = useState<string | null>(null);
  const [rateDraft, setRateDraft] = useState({ base: '', markup: '' });
  const [rateSaving, setRateSaving] = useState(false);
  const [rateSaved, setRateSaved] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<CommissionRow[]>('/admin/commissions');
      setRows(res.data || []);
    } catch {
      setError('Не удалось загрузить комиссии');
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const loadRates = useCallback(async () => {
    setRatesLoading(true);
    setRatesError('');
    try {
      const res = await apiClient.get<ExchangeRateRow[]>('/admin/exchange-rates');
      setRates(res.data || []);
    } catch {
      setRatesError('Не удалось загрузить курсы валют');
      setRates([]);
    } finally {
      setRatesLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
    loadRates();
  }, [load]);

  const startEdit = (r: CommissionRow) => {
    setEditingId(r.id);
    setValueDraft(String(r.value ?? ''));
  };

  const cancelEdit = () => {
    setEditingId(null);
    setValueDraft('');
  };

  const saveEdit = async () => {
    const id = editingId;
    if (!id) return;
    const v = parseFloat(valueDraft);
    if (Number.isNaN(v)) {
      setError('Введите корректное число');
      return;
    }
    setSaving(true);
    setError('');
    try {
      await apiClient.patch(`/admin/commissions/${id}`, { value: v });
      setEditingId(null);
      setValueDraft('');
      await load();
    } catch {
      setError('Не удалось сохранить комиссию');
    } finally {
      setSaving(false);
    }
  };

  const startEditRate = (r: ExchangeRateRow) => {
    const key = `${r.currency_from}/${r.currency_to}`;
    setEditingPair(key);
    setRateDraft({ base: String(r.base_rate ?? ''), markup: String(r.markup_percent ?? '0') });
    setRatesError('');
  };

  const cancelEditRate = () => {
    setEditingPair(null);
    setRateDraft({ base: '', markup: '' });
  };

  const saveRate = async () => {
    if (!editingPair) return;
    const [currency_from, currency_to] = editingPair.split('/');
    const base = parseFloat(rateDraft.base);
    const markup = parseFloat(rateDraft.markup || '0');

    if (!Number.isFinite(base) || base <= 0) {
      setRatesError('Введите корректный base rate');
      return;
    }
    if (!Number.isFinite(markup)) {
      setRatesError('Введите корректный markup');
      return;
    }

    setRateSaving(true);
    setRatesError('');
    try {
      await apiClient.patch('/admin/exchange-rates', { currency_from, currency_to, base_rate: base, markup_percent: markup });
      setEditingPair(null);
      setRateDraft({ base: '', markup: '' });
      await loadRates();
      setRateSaved(true);
      window.setTimeout(() => setRateSaved(false), 2500);
    } catch {
      setRatesError('Не удалось сохранить курс');
    } finally {
      setRateSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Комиссии</h1>
          <p className="text-sm text-slate-400 mt-1">Комиссии и внутренние курсы валют</p>
        </div>
        <button
          onClick={() => { load(); loadRates(); }}
          disabled={loading || ratesLoading}
          className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
        >
          <RefreshCw className="w-4 h-4" />
          Обновить
        </button>
      </div>

      <div className="glass-card p-6">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <p className="text-white font-semibold">Курсы валют</p>
            <p className="text-sm text-slate-400 mt-1">Внутренний курс = наша комиссия</p>
          </div>
          {rateSaved ? <span className="text-sm text-emerald-400">Сохранено</span> : null}
        </div>

        {ratesError ? <p className="text-sm text-red-400 mt-3">{ratesError}</p> : null}

        <div className="mt-4 overflow-x-auto">
          <table className="min-w-[920px] w-full text-left">
            <thead>
              <tr className="text-xs text-slate-500">
                <th className="py-3 px-2 font-semibold">Пара</th>
                <th className="py-3 px-2 font-semibold">Base rate</th>
                <th className="py-3 px-2 font-semibold">Markup (%)</th>
                <th className="py-3 px-2 font-semibold">Final rate</th>
                <th className="py-3 px-2 font-semibold text-right">Действия</th>
              </tr>
            </thead>
            <tbody>
              {ratesLoading ? (
                <tr>
                  <td colSpan={5} className="py-8 text-center text-slate-400">
                    <span className="inline-flex items-center gap-2">
                      <Loader2 className="w-4 h-4 animate-spin" /> Загрузка...
                    </span>
                  </td>
                </tr>
              ) : rates.length === 0 ? (
                <tr>
                  <td colSpan={5} className="py-10 text-center text-slate-500">Пусто</td>
                </tr>
              ) : (
                rates.map((r) => {
                  const key = `${r.currency_from}/${r.currency_to}`;
                  const isEditing = editingPair === key;
                  return (
                    <tr key={r.id || key} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                      <td className="py-3 px-2 text-sm text-slate-200 font-semibold">
                        <div className="font-mono text-xs text-slate-500">{r.currency_from}→{r.currency_to}</div>
                        {key}
                      </td>
                      <td className="py-3 px-2">
                        {isEditing ? (
                          <input
                            value={rateDraft.base}
                            onChange={(e) => setRateDraft((x) => ({ ...x, base: e.target.value }))}
                            className="w-40 bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                          />
                        ) : (
                          <span className="text-sm text-slate-200 font-mono">{r.base_rate}</span>
                        )}
                      </td>
                      <td className="py-3 px-2">
                        {isEditing ? (
                          <input
                            value={rateDraft.markup}
                            onChange={(e) => setRateDraft((x) => ({ ...x, markup: e.target.value }))}
                            className="w-40 bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                          />
                        ) : (
                          <span className="text-sm text-slate-200 font-mono">{r.markup_percent}</span>
                        )}
                      </td>
                      <td className="py-3 px-2">
                        <span className="text-sm text-slate-200 font-mono">{r.final_rate}</span>
                      </td>
                      <td className="py-3 px-2">
                        <div className="flex justify-end gap-2">
                          {isEditing ? (
                            <>
                              <button
                                onClick={saveRate}
                                disabled={rateSaving}
                                className="px-3 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                              >
                                <Save className="w-4 h-4" />
                                {rateSaving ? '...' : 'Сохранить'}
                              </button>
                              <button
                                onClick={cancelEditRate}
                                disabled={rateSaving}
                                className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors disabled:opacity-50"
                              >
                                Отмена
                              </button>
                            </>
                          ) : (
                            <button
                              onClick={() => startEditRate(r)}
                              className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors inline-flex items-center gap-2"
                            >
                              Редактировать
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      <div className="glass-card p-4 sm:p-6 overflow-x-auto">
        {error ? <p className="text-sm text-red-400 mb-4">{error}</p> : null}

        <table className="min-w-[920px] w-full text-left">
          <thead>
            <tr className="text-xs text-slate-500">
              <th className="py-3 px-2 font-semibold">Ключ</th>
              <th className="py-3 px-2 font-semibold">Описание</th>
              <th className="py-3 px-2 font-semibold">Значение</th>
              <th className="py-3 px-2 font-semibold text-right">Действия</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={4} className="py-8 text-center text-slate-400">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin" /> Загрузка...
                  </span>
                </td>
              </tr>
            ) : rows.length === 0 ? (
              <tr>
                <td colSpan={4} className="py-10 text-center text-slate-500">
                  Пусто
                </td>
              </tr>
            ) : (
              rows.map((r) => {
                const isEditing = editingId === r.id;
                return (
                  <tr key={r.id} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                    <td className="py-3 px-2 text-sm text-slate-200">
                      <div className="font-mono text-xs text-slate-500">{r.key}</div>
                      <div className="font-semibold">{keyLabel(r.key)}</div>
                    </td>
                    <td className="py-3 px-2 text-sm text-slate-400">{r.description || '—'}</td>
                    <td className="py-3 px-2">
                      {isEditing ? (
                        <input
                          value={valueDraft}
                          onChange={(e) => setValueDraft(e.target.value)}
                          className="w-40 bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
                        />
                      ) : (
                        <span className="text-sm text-slate-200 font-mono">{r.value}</span>
                      )}
                    </td>
                    <td className="py-3 px-2">
                      <div className="flex justify-end gap-2">
                        {isEditing ? (
                          <>
                            <button
                              onClick={saveEdit}
                              disabled={saving}
                              className="px-3 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                            >
                              <Save className="w-4 h-4" />
                              {saving ? '...' : 'Сохранить'}
                            </button>
                            <button
                              onClick={cancelEdit}
                              disabled={saving}
                              className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors disabled:opacity-50"
                            >
                              Отмена
                            </button>
                          </>
                        ) : (
                          <button
                            onClick={() => startEdit(r)}
                            className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors inline-flex items-center gap-2"
                          >
                            Редактировать
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

