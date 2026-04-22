import { useCallback, useEffect, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw, Save } from 'lucide-react';

type SystemSetting = {
  key: string;
  value: string;
  description: string;
  updatedAt?: string;
};

export const AdminSystemSettingsPage = () => {
  const [rows, setRows] = useState<SystemSetting[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [valueDraft, setValueDraft] = useState('');
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<SystemSetting[]>('/admin/system-settings');
      setRows(res.data || []);
    } catch {
      setError('Не удалось загрузить system settings');
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const startEdit = (s: SystemSetting) => {
    setEditingKey(s.key);
    setValueDraft(String(s.value ?? ''));
  };

  const cancel = () => {
    setEditingKey(null);
    setValueDraft('');
  };

  const save = async () => {
    if (!editingKey) return;
    setSaving(true);
    setError('');
    try {
      await apiClient.patch(`/admin/system-settings/${editingKey}`, { value: valueDraft });
      setEditingKey(null);
      setValueDraft('');
      await load();
    } catch {
      setError('Не удалось сохранить значение');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">System settings</h1>
          <p className="text-sm text-slate-400 mt-1">Хранилище ключ-значение для админских параметров</p>
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

      {error ? <p className="text-sm text-red-400">{error}</p> : null}

      <div className="glass-card p-4 sm:p-6 overflow-x-auto">
        <table className="min-w-[920px] w-full text-left">
          <thead>
            <tr className="text-xs text-slate-500">
              <th className="py-3 px-2 font-semibold">Key</th>
              <th className="py-3 px-2 font-semibold">Description</th>
              <th className="py-3 px-2 font-semibold">Value</th>
              <th className="py-3 px-2 font-semibold text-right">Actions</th>
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
                <td colSpan={4} className="py-10 text-center text-slate-500">Пусто</td>
              </tr>
            ) : (
              rows.map((r) => {
                const isEditing = editingKey === r.key;
                return (
                  <tr key={r.key} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                    <td className="py-3 px-2 text-xs text-slate-500 font-mono">{r.key}</td>
                    <td className="py-3 px-2 text-sm text-slate-400">{r.description || '—'}</td>
                    <td className="py-3 px-2">
                      {isEditing ? (
                        <input
                          value={valueDraft}
                          onChange={(e) => setValueDraft(e.target.value)}
                          className="w-full max-w-[420px] bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
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
                              onClick={save}
                              disabled={saving}
                              className="px-3 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                            >
                              <Save className="w-4 h-4" />
                              {saving ? '...' : 'Сохранить'}
                            </button>
                            <button
                              onClick={cancel}
                              disabled={saving}
                              className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors disabled:opacity-50"
                            >
                              Отмена
                            </button>
                          </>
                        ) : (
                          <button
                            onClick={() => startEdit(r)}
                            className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors"
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

