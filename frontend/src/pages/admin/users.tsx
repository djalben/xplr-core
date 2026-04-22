import { useCallback, useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Search, ShieldCheck, ShieldOff, Ban, CheckCircle2, Loader2, RefreshCw } from 'lucide-react';

type AdminUser = {
  id: string;
  display_name: string;
  email: string;
  balance_rub: string;
  status: string;
  is_admin: boolean;
  role: string;
  is_verified: boolean;
  is_blocked: boolean;
  created_at: string;
};

type UsersListResponse = {
  users: AdminUser[];
  total?: number;
  limit?: number;
  offset?: number;
};

const Pill = ({ ok, label }: { ok: boolean; label: string }) => (
  <span
    className={`px-2 py-1 rounded-lg text-[10px] font-semibold border ${
      ok ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-300' : 'bg-red-500/10 border-red-500/20 text-red-300'
    }`}
  >
    {label}
  </span>
);

export const AdminUsersPage = () => {
  const [q, setQ] = useState('');
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [error, setError] = useState('');

  const canSearch = useMemo(() => q.trim().length >= 2, [q]);

  const loadUsers = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<UsersListResponse>('/admin/users', { params: { limit: 50, offset: 0 } });
      setUsers(res.data?.users ?? []);
    } catch (e: any) {
      setError('Не удалось загрузить пользователей');
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const searchUsers = useCallback(async () => {
    if (!canSearch) return;
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<{ users: AdminUser[] }>('/admin/users/search', { params: { q: q.trim(), limit: 50 } });
      setUsers(res.data?.users ?? []);
    } catch {
      setError('Не удалось выполнить поиск');
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, [q, canSearch]);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  const setStatus = async (user: AdminUser, next: 'ACTIVE' | 'BANNED') => {
    setBusyId(user.id);
    setError('');
    try {
      await apiClient.patch(`/admin/users/${user.id}/status`, { status: next });
      const isBlocked = next !== 'ACTIVE';
      setUsers((prev) => prev.map((u) => (u.id === user.id ? { ...u, status: isBlocked ? 'BLOCKED' : 'ACTIVE', is_blocked: isBlocked } : u)));
    } catch {
      setError('Не удалось изменить статус');
    } finally {
      setBusyId(null);
    }
  };

  const toggleAdmin = async (user: AdminUser) => {
    setBusyId(user.id);
    setError('');
    try {
      const res = await apiClient.patch(`/admin/users/${user.id}/role`);
      const next = !!res.data?.is_admin;
      setUsers((prev) => prev.map((u) => (u.id === user.id ? { ...u, is_admin: next, role: next ? 'admin' : 'user' } : u)));
    } catch {
      setError('Не удалось изменить роль');
    } finally {
      setBusyId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Пользователи</h1>
          <p className="text-sm text-slate-400 mt-1">Поиск, бан/разбан, роль админа</p>
        </div>
        <button
          onClick={loadUsers}
          disabled={loading}
          className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
        >
          <RefreshCw className="w-4 h-4" />
          Обновить
        </button>
      </div>

      <div className="glass-card p-4 sm:p-6">
        <div className="flex flex-col sm:flex-row gap-3 sm:items-center sm:justify-between">
          <div className="flex-1 flex items-center gap-2 bg-white/5 border border-white/10 rounded-xl px-3 py-2">
            <Search className="w-4 h-4 text-slate-500" />
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Поиск по email (минимум 2 символа)"
              className="w-full bg-transparent outline-none text-sm text-slate-200 placeholder:text-slate-600"
            />
          </div>
          <div className="flex gap-2">
            <button
              onClick={searchUsers}
              disabled={!canSearch || loading}
              className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-sm font-semibold transition-colors disabled:opacity-50"
            >
              Найти
            </button>
            <button
              onClick={() => {
                setQ('');
                loadUsers();
              }}
              disabled={loading}
              className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50"
            >
              Сброс
            </button>
          </div>
        </div>

        {error ? <p className="text-sm text-red-400 mt-4">{error}</p> : null}

        <div className="mt-4 overflow-x-auto">
          <table className="min-w-[920px] w-full text-left">
            <thead>
              <tr className="text-xs text-slate-500">
                <th className="py-3 px-2 font-semibold">Email</th>
                <th className="py-3 px-2 font-semibold">ID</th>
                <th className="py-3 px-2 font-semibold">Статус</th>
                <th className="py-3 px-2 font-semibold">Админ</th>
                <th className="py-3 px-2 font-semibold">Баланс</th>
                <th className="py-3 px-2 font-semibold">Создан</th>
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
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-8 text-center text-slate-500">
                    Ничего не найдено
                  </td>
                </tr>
              ) : (
                users.map((u) => {
                  const isBusy = busyId === u.id;
                  const created = u.created_at ? new Date(u.created_at).toLocaleString('ru-RU') : '';
                  return (
                    <tr key={u.id} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                      <td className="py-3 px-2 text-sm text-slate-200">{u.email}</td>
                      <td className="py-3 px-2 text-xs text-slate-500 font-mono">{u.id}</td>
                      <td className="py-3 px-2">
                        <Pill ok={!u.is_blocked} label={u.is_blocked ? 'BANNED' : 'ACTIVE'} />
                      </td>
                      <td className="py-3 px-2">
                        <Pill ok={u.is_admin} label={u.is_admin ? 'ADMIN' : 'USER'} />
                      </td>
                      <td className="py-3 px-2 text-sm text-slate-200">{u.balance_rub || '0'}</td>
                      <td className="py-3 px-2 text-xs text-slate-500">{created}</td>
                      <td className="py-3 px-2">
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => toggleAdmin(u)}
                            disabled={isBusy}
                            className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                            title="Переключить роль админа"
                          >
                            {u.is_admin ? <ShieldOff className="w-4 h-4 text-orange-300" /> : <ShieldCheck className="w-4 h-4 text-blue-300" />}
                            {isBusy ? '...' : u.is_admin ? 'Снять' : 'Дать'}
                          </button>

                          {u.is_blocked ? (
                            <button
                              onClick={() => setStatus(u, 'ACTIVE')}
                              disabled={isBusy}
                              className="px-3 py-2 rounded-xl bg-emerald-500/15 hover:bg-emerald-500/20 border border-emerald-500/20 text-emerald-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                              title="Разбанить"
                            >
                              <CheckCircle2 className="w-4 h-4" />
                              {isBusy ? '...' : 'Разбан'}
                            </button>
                          ) : (
                            <button
                              onClick={() => setStatus(u, 'BANNED')}
                              disabled={isBusy}
                              className="px-3 py-2 rounded-xl bg-red-500/15 hover:bg-red-500/20 border border-red-500/20 text-red-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                              title="Забанить"
                            >
                              <Ban className="w-4 h-4" />
                              {isBusy ? '...' : 'Бан'}
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
    </div>
  );
};

