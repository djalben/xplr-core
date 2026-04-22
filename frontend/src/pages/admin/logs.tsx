import { useCallback, useEffect, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw } from 'lucide-react';

type AdminLog = {
  id: string;
  adminId?: string | null;
  action: string;
  createdAt: string;
};

export const AdminLogsPage = () => {
  const [rows, setRows] = useState<AdminLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [limit, setLimit] = useState(50);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<AdminLog[]>('/admin/logs', { params: { limit } });
      setRows(res.data || []);
    } catch {
      setError('Не удалось загрузить логи');
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, [limit]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Логи</h1>
          <p className="text-sm text-slate-400 mt-1">Последние действия админов</p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
            className="px-3 py-2 rounded-xl bg-white/5 border border-white/10 text-slate-200 text-sm outline-none"
          >
            <option value={20}>20</option>
            <option value={50}>50</option>
            <option value={100}>100</option>
            <option value={200}>200</option>
          </select>
          <button
            onClick={load}
            disabled={loading}
            className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            <RefreshCw className="w-4 h-4" />
            Обновить
          </button>
        </div>
      </div>

      {error ? <p className="text-sm text-red-400">{error}</p> : null}

      <div className="glass-card p-4 sm:p-6 overflow-x-auto">
        <table className="min-w-[920px] w-full text-left">
          <thead>
            <tr className="text-xs text-slate-500">
              <th className="py-3 px-2 font-semibold">Time</th>
              <th className="py-3 px-2 font-semibold">Admin ID</th>
              <th className="py-3 px-2 font-semibold">Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={3} className="py-8 text-center text-slate-400">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin" /> Загрузка...
                  </span>
                </td>
              </tr>
            ) : rows.length === 0 ? (
              <tr>
                <td colSpan={3} className="py-10 text-center text-slate-500">Пусто</td>
              </tr>
            ) : (
              rows.map((r) => {
                const ts = r.createdAt ? new Date(r.createdAt).toLocaleString('ru-RU') : '';
                return (
                  <tr key={r.id} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                    <td className="py-3 px-2 text-xs text-slate-500">{ts}</td>
                    <td className="py-3 px-2 text-xs text-slate-500 font-mono">{r.adminId || '—'}</td>
                    <td className="py-3 px-2 text-sm text-slate-200">{r.action}</td>
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

