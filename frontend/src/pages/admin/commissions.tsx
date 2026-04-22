import { useCallback, useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw, Save, ToggleLeft, ToggleRight, AlertTriangle } from 'lucide-react';

type CommissionRow = {
  id: string;
  key: string;
  value: string;
  description: string;
  updatedAt?: string;
};

const keyLabel = (key: string) => {
  switch (key) {
    case 'sbp_topup_enabled':
      return 'СБП пополнение (0/1)';
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
  const [sbpLoading, setSbpLoading] = useState(false);

  const sbp = useMemo(() => rows.find((r) => r.key === 'sbp_topup_enabled') || null, [rows]);
  const sbpEnabled = useMemo(() => {
    if (!sbp) return null;
    return !(parseFloat(sbp.value) < 0.5);
  }, [sbp]);

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

  useEffect(() => {
    load();
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

  const toggleSBP = async () => {
    if (sbpEnabled === null) return;
    const next = !sbpEnabled;
    setSbpLoading(true);
    setError('');
    try {
      await apiClient.patch('/admin/sbp-topup', { enabled: next });
      await load();
    } catch {
      setError('Не удалось переключить СБП');
    } finally {
      setSbpLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Комиссии</h1>
          <p className="text-sm text-slate-400 mt-1">Настройки комиссий и сервисных флагов</p>
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

      <div className="glass-card p-6">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <p className="text-white font-semibold">Пополнение через СБП</p>
            <p className="text-sm text-slate-400 mt-1">
              Быстрое глобальное включение/выключение. Сейчас:{' '}
              <span className="text-slate-200">{sbpEnabled === null ? '—' : sbpEnabled ? 'включено' : 'выключено'}</span>
            </p>
          </div>
          <button
            onClick={toggleSBP}
            disabled={sbpEnabled === null || sbpLoading}
            className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {sbpEnabled ? <ToggleRight className="w-5 h-5 text-emerald-400" /> : <ToggleLeft className="w-5 h-5 text-slate-400" />}
            {sbpLoading ? '...' : sbpEnabled ? 'Выключить' : 'Включить'}
          </button>
        </div>
        <div className="mt-4 flex items-start gap-2 text-xs text-slate-500">
          <AlertTriangle className="w-4 h-4 shrink-0 text-slate-600" />
          <p>
            Это меняет ключ <span className="font-mono">sbp_topup_enabled</span> в <span className="font-mono">commission_config</span> (1/0).
          </p>
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

