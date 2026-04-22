import { useCallback, useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw, Inbox, CheckCircle2, User, MessageSquare, X, Send } from 'lucide-react';

type Ticket = {
  id: string;
  userId: string;
  adminId?: string | null;
  subject: string;
  status: 'NEW' | 'IN_PROGRESS' | 'DONE' | string;
  userMessage: string;
  adminReply?: string;
  createdAt: string;
  closedAt?: string | null;
};

const StatusPill = ({ s }: { s: string }) => {
  const cls =
    s === 'NEW'
      ? 'bg-blue-500/10 border-blue-500/20 text-blue-200'
      : s === 'IN_PROGRESS'
      ? 'bg-orange-500/10 border-orange-500/20 text-orange-200'
      : s === 'DONE'
      ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-200'
      : 'bg-white/5 border-white/10 text-slate-300';
  return <span className={`px-2 py-1 rounded-lg text-[10px] font-semibold border ${cls}`}>{s}</span>;
};

export const AdminTicketsPage = () => {
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [busyId, setBusyId] = useState<string | null>(null);
  const [closing, setClosing] = useState<Ticket | null>(null);
  const [reply, setReply] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<Ticket[]>('/admin/tickets');
      setTickets(res.data || []);
    } catch {
      setError('Не удалось загрузить тикеты');
      setTickets([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const take = async (t: Ticket) => {
    setBusyId(t.id);
    setError('');
    try {
      await apiClient.put(`/admin/tickets/${t.id}/take`);
      await load();
    } catch {
      setError('Не удалось взять тикет');
    } finally {
      setBusyId(null);
    }
  };

  const openClose = (t: Ticket) => {
    setClosing(t);
    setReply('');
  };

  const close = async () => {
    if (!closing) return;
    setBusyId(closing.id);
    setError('');
    try {
      await apiClient.put(`/admin/tickets/${closing.id}/close`, { reply });
      setClosing(null);
      setReply('');
      await load();
    } catch {
      setError('Не удалось закрыть тикет');
    } finally {
      setBusyId(null);
    }
  };

  const sorted = useMemo(() => {
    const copy = [...tickets];
    copy.sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
    return copy;
  }, [tickets]);

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Тикеты</h1>
          <p className="text-sm text-slate-400 mt-1">Список тикетов и обработка</p>
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
              <th className="py-3 px-2 font-semibold">Статус</th>
              <th className="py-3 px-2 font-semibold">Тема</th>
              <th className="py-3 px-2 font-semibold">User</th>
              <th className="py-3 px-2 font-semibold">Admin</th>
              <th className="py-3 px-2 font-semibold">Создан</th>
              <th className="py-3 px-2 font-semibold text-right">Действия</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-slate-400">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin" /> Загрузка...
                  </span>
                </td>
              </tr>
            ) : sorted.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-10 text-center text-slate-500">
                  <span className="inline-flex items-center gap-2">
                    <Inbox className="w-4 h-4" /> Тикетов нет
                  </span>
                </td>
              </tr>
            ) : (
              sorted.map((t) => {
                const isBusy = busyId === t.id;
                const created = t.createdAt ? new Date(t.createdAt).toLocaleString('ru-RU') : '';
                return (
                  <tr key={t.id} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                    <td className="py-3 px-2">
                      <StatusPill s={t.status} />
                    </td>
                    <td className="py-3 px-2 text-sm text-slate-200">
                      <div className="font-semibold">{t.subject || '—'}</div>
                      <div className="text-xs text-slate-500 truncate max-w-[520px]">{t.userMessage}</div>
                    </td>
                    <td className="py-3 px-2 text-xs text-slate-500 font-mono">{t.userId}</td>
                    <td className="py-3 px-2 text-xs text-slate-500 font-mono">{t.adminId || '—'}</td>
                    <td className="py-3 px-2 text-xs text-slate-500">{created}</td>
                    <td className="py-3 px-2">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => take(t)}
                          disabled={isBusy || t.status !== 'NEW'}
                          className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                          title="Взять в работу"
                        >
                          <User className="w-4 h-4" />
                          {isBusy ? '...' : 'Взять'}
                        </button>
                        <button
                          onClick={() => openClose(t)}
                          disabled={isBusy || t.status === 'DONE'}
                          className="px-3 py-2 rounded-xl bg-emerald-500/15 hover:bg-emerald-500/20 border border-emerald-500/20 text-emerald-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                          title="Закрыть тикет"
                        >
                          <CheckCircle2 className="w-4 h-4" />
                          {isBusy ? '...' : 'Закрыть'}
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {closing ? (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/70 backdrop-blur-sm p-4" onClick={() => setClosing(null)}>
          <div className="glass-card w-full max-w-xl p-6 space-y-4" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <h3 className="text-white font-bold text-lg flex items-center gap-2">
                  <MessageSquare className="w-5 h-5 text-emerald-300" /> Закрыть тикет
                </h3>
                <p className="text-sm text-slate-400 mt-1 truncate">#{closing.id}</p>
              </div>
              <button onClick={() => setClosing(null)} className="p-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200">
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="bg-white/5 border border-white/10 rounded-xl p-4">
              <p className="text-xs text-slate-500 mb-1">Сообщение пользователя</p>
              <p className="text-sm text-slate-200 whitespace-pre-wrap">{closing.userMessage}</p>
            </div>

            <div>
              <label className="block text-xs text-slate-500 mb-2">Ответ админа (необязательно)</label>
              <textarea
                value={reply}
                onChange={(e) => setReply(e.target.value)}
                rows={4}
                className="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm text-slate-200 outline-none focus:border-emerald-500/40"
                placeholder="Напиши ответ пользователю..."
              />
            </div>

            <div className="flex items-center justify-end gap-2">
              <button
                onClick={() => setClosing(null)}
                className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors"
              >
                Отмена
              </button>
              <button
                onClick={close}
                disabled={busyId === closing.id}
                className="px-4 py-2 rounded-xl bg-emerald-500/20 hover:bg-emerald-500/25 border border-emerald-500/25 text-emerald-100 text-sm font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
              >
                <Send className="w-4 h-4" />
                {busyId === closing.id ? '...' : 'Закрыть'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

