import React, { useState, useEffect, useCallback } from 'react';
import apiClient from '../api/axios';
import {
  Users,
  DollarSign,
  CreditCard,
  MessageSquare,
  Search,
  Shield,
  Settings,
  Activity,
  TrendingUp,
  ChevronRight,
  Save,
  AlertCircle,
  CheckCircle,
  Clock,
  UserCheck,
  Loader2,
  Zap,
} from 'lucide-react';

type Tab = 'dashboard' | 'users' | 'finances' | 'commissions' | 'tickets' | 'logs';

interface DashboardStats {
  total_users: number;
  total_balance: string;
  active_cards: number;
  open_tickets: number;
  today_signups: number;
  total_cards: number;
}

interface AdminUser {
  id: number;
  email: string;
  balance_rub: string;
  status: string;
  is_admin: boolean;
  role: string;
  is_verified: boolean;
  card_count: number;
  created_at: string;
}

interface CommissionRow {
  id: number;
  key: string;
  value: string;
  description: string;
  updated_at: string;
}

interface AdminLog {
  id: number;
  admin_id: number;
  admin_email: string;
  action: string;
  created_at: string;
}

interface SupportTicket {
  id: number;
  user_id: number;
  email: string;
  subject: string;
  message: string;
  status: string;
  created_at: string;
}

// ── Stat Card ──
const StatCard = ({ icon: Icon, label, value, accent }: { icon: any; label: string; value: string | number; accent: string }) => (
  <div className="glass-card p-5 flex items-center gap-4">
    <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${accent}`}>
      <Icon className="w-6 h-6 text-white" />
    </div>
    <div>
      <p className="text-xs text-slate-400 uppercase tracking-wider">{label}</p>
      <p className="text-xl font-bold text-white">{value}</p>
    </div>
  </div>
);

// ── Status Badge ──
const StatusBadge = ({ status }: { status: string }) => {
  const colors: Record<string, string> = {
    ACTIVE: 'bg-emerald-500/20 text-emerald-400',
    BANNED: 'bg-red-500/20 text-red-400',
    admin: 'bg-purple-500/20 text-purple-400',
    user: 'bg-slate-500/20 text-slate-400',
  };
  return (
    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${colors[status] || 'bg-slate-500/20 text-slate-400'}`}>
      {status}
    </span>
  );
};

// ── Grade Select ──
const GRADES = ['STANDARD', 'SILVER', 'GOLD', 'PLATINUM', 'BLACK'] as const;
const gradeColors: Record<string, string> = {
  STANDARD: 'text-slate-400',
  SILVER: 'text-slate-300',
  GOLD: 'text-yellow-400',
  PLATINUM: 'text-blue-400',
  BLACK: 'text-purple-400',
};

// ══════════════════════════════════════
// MAIN COMPONENT
// ══════════════════════════════════════
export const StaffOnlyZone = () => {
  const [tab, setTab] = useState<Tab>('dashboard');
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [searchQ, setSearchQ] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null);
  const [selectedGrade, setSelectedGrade] = useState('');
  const [selectedStatus, setSelectedStatus] = useState('');
  const [commissions, setCommissions] = useState<CommissionRow[]>([]);
  const [editingCommission, setEditingCommission] = useState<{ id: number; value: string } | null>(null);
  const [logs, setLogs] = useState<AdminLog[]>([]);
  const [tickets, setTickets] = useState<SupportTicket[]>([]);
  const [ticketFilter, setTicketFilter] = useState('');
  const [saving, setSaving] = useState(false);
  const [freezeConfirm, setFreezeConfirm] = useState(false);
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null);

  const showToast = (msg: string, type: 'ok' | 'err' = 'ok') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  // ── Fetch dashboard stats ──
  const fetchStats = useCallback(async () => {
    try {
      const res = await apiClient.get('/admin/dashboard');
      setStats(res.data);
    } catch { /* ignore */ }
  }, []);

  // ── Search users ──
  const searchUsers = useCallback(async () => {
    if (!searchQ.trim()) return;
    setIsSearching(true);
    try {
      const res = await apiClient.get('/admin/users/search', { params: { q: searchQ, limit: 50 } });
      setUsers(res.data || []);
    } catch {
      showToast('Ошибка поиска', 'err');
    } finally {
      setIsSearching(false);
    }
  }, [searchQ]);

  // ── Load all users ──
  const loadAllUsers = useCallback(async () => {
    try {
      const res = await apiClient.get('/admin/users');
      setUsers(res.data || []);
    } catch { /* ignore */ }
  }, []);

  // ── Commission config ──
  const loadCommissions = useCallback(async () => {
    try {
      const res = await apiClient.get('/admin/commissions');
      setCommissions(res.data || []);
    } catch { /* ignore */ }
  }, []);

  // ── Admin logs ──
  const loadLogs = useCallback(async () => {
    try {
      const res = await apiClient.get('/admin/logs', { params: { limit: 50 } });
      setLogs(res.data || []);
    } catch { /* ignore */ }
  }, []);

  // ── Support tickets ──
  const loadTickets = useCallback(async () => {
    try {
      const params: Record<string, string> = {};
      if (ticketFilter) params.status = ticketFilter;
      const res = await apiClient.get('/admin/tickets', { params });
      setTickets(res.data || []);
    } catch { /* ignore */ }
  }, [ticketFilter]);

  const updateTicketStatus = async (id: number, status: string) => {
    try {
      await apiClient.patch(`/admin/tickets/${id}`, { status });
      showToast(`Тикет #${id} → ${status}`);
      loadTickets();
      fetchStats();
    } catch {
      showToast('Ошибка обновления тикета', 'err');
    }
  };

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  useEffect(() => {
    if (tab === 'users') loadAllUsers();
    if (tab === 'commissions') loadCommissions();
    if (tab === 'tickets') loadTickets();
    if (tab === 'logs') loadLogs();
  }, [tab, loadAllUsers, loadCommissions, loadTickets, loadLogs]);

  // ── Emergency Freeze ──
  const handleEmergencyFreeze = async () => {
    if (!selectedUser) return;
    setSaving(true);
    try {
      const res = await apiClient.post(`/admin/users/${selectedUser.id}/emergency-freeze`);
      showToast(`FREEZE: ${selectedUser.email} — ${res.data.frozen_cards} cards frozen, BANNED`);
      setSelectedUser(null);
      setFreezeConfirm(false);
      loadAllUsers();
      fetchStats();
    } catch {
      showToast('Emergency Freeze failed', 'err');
    } finally {
      setSaving(false);
    }
  };

  // ── Update grade ──
  const handleUpdateGrade = async () => {
    if (!selectedUser || !selectedGrade) return;
    setSaving(true);
    try {
      await apiClient.patch(`/admin/users/${selectedUser.id}/grade`, { grade: selectedGrade });
      showToast(`Грейд ${selectedUser.email} → ${selectedGrade}`);
      setSelectedUser(null);
      loadAllUsers();
    } catch {
      showToast('Ошибка обновления грейда', 'err');
    } finally {
      setSaving(false);
    }
  };

  // ── Update status ──
  const handleUpdateStatus = async () => {
    if (!selectedUser || !selectedStatus) return;
    setSaving(true);
    try {
      await apiClient.patch(`/admin/users/${selectedUser.id}/status`, { status: selectedStatus });
      showToast(`Статус ${selectedUser.email} → ${selectedStatus}`);
      setSelectedUser(null);
      loadAllUsers();
    } catch {
      showToast('Ошибка обновления статуса', 'err');
    } finally {
      setSaving(false);
    }
  };

  // ── Save commission ──
  const handleSaveCommission = async () => {
    if (!editingCommission) return;
    setSaving(true);
    try {
      await apiClient.patch(`/admin/commissions/${editingCommission.id}`, { value: parseFloat(editingCommission.value) });
      showToast('Комиссия обновлена');
      setEditingCommission(null);
      loadCommissions();
    } catch {
      showToast('Ошибка сохранения', 'err');
    } finally {
      setSaving(false);
    }
  };

  // ── Tab button ──
  const TabBtn = ({ id, icon: Icon, label }: { id: Tab; icon: any; label: string }) => (
    <button
      onClick={() => setTab(id)}
      className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all whitespace-nowrap ${
        tab === id
          ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/10 text-blue-400 border border-blue-500/30'
          : 'text-slate-400 hover:text-white hover:bg-white/5'
      }`}
    >
      <Icon className="w-4 h-4" />
      {label}
    </button>
  );

  return (
    <div className="min-h-[100dvh] bg-transparent relative z-2">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Header */}
        <div className="flex items-center gap-3 mb-8">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-red-500 to-purple-600 flex items-center justify-center">
            <Shield className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">Admin Panel</h1>
            <p className="text-sm text-slate-500">Закрытая зона управления</p>
          </div>
        </div>

        {/* Toast */}
        {toast && (
          <div className={`fixed top-6 right-6 z-50 flex items-center gap-2 px-4 py-3 rounded-xl text-sm font-medium shadow-xl backdrop-blur ${
            toast.type === 'ok' ? 'bg-emerald-500/90 text-white' : 'bg-red-500/90 text-white'
          }`}>
            {toast.type === 'ok' ? <CheckCircle className="w-4 h-4" /> : <AlertCircle className="w-4 h-4" />}
            {toast.msg}
          </div>
        )}

        {/* Tabs */}
        <div className="flex gap-2 mb-6 overflow-x-auto pb-2 scrollbar-hide">
          <TabBtn id="dashboard" icon={Activity} label="Dashboard" />
          <TabBtn id="users" icon={Users} label="Юзеры" />
          <TabBtn id="finances" icon={TrendingUp} label="Финансы" />
          <TabBtn id="commissions" icon={Settings} label="Комиссии" />
          <TabBtn id="tickets" icon={MessageSquare} label="Тикеты" />
          <TabBtn id="logs" icon={Clock} label="Логи" />
        </div>

        {/* ════════════ DASHBOARD TAB ════════════ */}
        {tab === 'dashboard' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <StatCard icon={Users} label="Всего юзеров" value={stats?.total_users ?? '—'} accent="bg-blue-500" />
              <StatCard icon={DollarSign} label="Сумма в кошельках" value={stats ? `$${parseFloat(stats.total_balance).toLocaleString()}` : '—'} accent="bg-emerald-500" />
              <StatCard icon={CreditCard} label="Активные карты" value={stats?.active_cards ?? '—'} accent="bg-purple-500" />
              <StatCard icon={MessageSquare} label="Открытые тикеты" value={stats?.open_tickets ?? '—'} accent="bg-orange-500" />
              <StatCard icon={UserCheck} label="Регистрации сегодня" value={stats?.today_signups ?? '—'} accent="bg-cyan-500" />
              <StatCard icon={CreditCard} label="Всего карт" value={stats?.total_cards ?? '—'} accent="bg-slate-500" />
            </div>
          </div>
        )}

        {/* ════════════ USERS TAB ════════════ */}
        {tab === 'users' && (
          <div className="space-y-4">
            {/* Search bar */}
            <div className="flex gap-2">
              <div className="flex-1 relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                <input
                  type="text"
                  value={searchQ}
                  onChange={e => setSearchQ(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && searchUsers()}
                  placeholder="Поиск по email..."
                  className="w-full pl-10 pr-4 py-2.5 bg-white/5 border border-white/10 rounded-xl text-white text-sm placeholder-slate-500 outline-none focus:border-blue-500/50"
                />
              </div>
              <button
                onClick={searchUsers}
                disabled={isSearching}
                className="px-5 py-2.5 bg-blue-500 hover:bg-blue-600 text-white rounded-xl text-sm font-medium transition-colors disabled:opacity-50"
              >
                {isSearching ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Найти'}
              </button>
            </div>

            {/* Users table */}
            <div className="glass-card overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-white/10">
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">ID</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Email</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Баланс</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Статус</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Роль</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Карты</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Дата</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map(u => (
                      <tr key={u.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                        <td className="px-4 py-3 text-slate-300 font-mono text-xs">{u.id}</td>
                        <td className="px-4 py-3 text-white">{u.email}</td>
                        <td className="px-4 py-3 text-emerald-400 font-medium">${parseFloat(u.balance_rub).toFixed(2)}</td>
                        <td className="px-4 py-3"><StatusBadge status={u.status} /></td>
                        <td className="px-4 py-3"><StatusBadge status={u.role} /></td>
                        <td className="px-4 py-3 text-slate-300">{u.card_count}</td>
                        <td className="px-4 py-3 text-slate-500 text-xs">{u.created_at?.slice(0, 10)}</td>
                        <td className="px-4 py-3">
                          <button
                            onClick={() => { setSelectedUser(u); setSelectedGrade(''); setSelectedStatus(''); }}
                            className="text-blue-400 hover:text-blue-300 transition-colors"
                          >
                            <ChevronRight className="w-4 h-4" />
                          </button>
                        </td>
                      </tr>
                    ))}
                    {users.length === 0 && (
                      <tr><td colSpan={8} className="px-4 py-8 text-center text-slate-500">Нет данных</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>

            {/* User edit modal */}
            {selectedUser && (
              <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setSelectedUser(null)}>
                <div className="glass-card p-6 w-full max-w-md mx-4 space-y-5" onClick={e => e.stopPropagation()}>
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-bold text-white">Юзер #{selectedUser.id}</h3>
                    <button onClick={() => setSelectedUser(null)} className="text-slate-400 hover:text-white">✕</button>
                  </div>
                  <div className="space-y-1 text-sm text-slate-300">
                    <p><strong className="text-slate-400">Email:</strong> {selectedUser.email}</p>
                    <p><strong className="text-slate-400">Баланс:</strong> ${parseFloat(selectedUser.balance_rub).toFixed(2)}</p>
                    <p><strong className="text-slate-400">Статус:</strong> {selectedUser.status}</p>
                    <p><strong className="text-slate-400">Роль:</strong> {selectedUser.role}</p>
                    <p><strong className="text-slate-400">Verified:</strong> {selectedUser.is_verified ? 'Да' : 'Нет'}</p>
                  </div>
                  {/* Grade */}
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">Грейд</label>
                    <div className="flex gap-2 flex-wrap">
                      {GRADES.map(g => (
                        <button
                          key={g}
                          onClick={() => setSelectedGrade(g)}
                          className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-all ${
                            selectedGrade === g
                              ? 'border-blue-500 bg-blue-500/20 text-blue-400'
                              : 'border-white/10 bg-white/5 hover:bg-white/10 ' + gradeColors[g]
                          }`}
                        >
                          {g}
                        </button>
                      ))}
                    </div>
                    {selectedGrade && (
                      <button onClick={handleUpdateGrade} disabled={saving} className="mt-2 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-xs font-medium transition-colors disabled:opacity-50 flex items-center gap-1.5">
                        <Save className="w-3 h-3" />{saving ? 'Saving...' : 'Сохранить грейд'}
                      </button>
                    )}
                  </div>
                  {/* Emergency Freeze */}
                  <div className="pt-2 border-t border-white/10">
                    {!freezeConfirm ? (
                      <button
                        onClick={() => setFreezeConfirm(true)}
                        className="w-full flex items-center justify-center gap-2 py-2.5 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30 text-red-400 rounded-xl text-xs font-medium transition-all"
                      >
                        <Zap className="w-3.5 h-3.5" /> Emergency Freeze
                      </button>
                    ) : (
                      <div className="space-y-2">
                        <p className="text-xs text-red-400 text-center">Все карты будут заморожены, аккаунт заблокирован, баланс обнулён. Продолжить?</p>
                        <div className="flex gap-2">
                          <button
                            onClick={handleEmergencyFreeze}
                            disabled={saving}
                            className="flex-1 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-xs font-bold transition-colors disabled:opacity-50"
                          >
                            {saving ? 'Freezing...' : 'CONFIRM FREEZE'}
                          </button>
                          <button
                            onClick={() => setFreezeConfirm(false)}
                            className="px-4 py-2 bg-white/10 hover:bg-white/20 text-slate-300 rounded-lg text-xs transition-colors"
                          >
                            Отмена
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                  {/* Status */}
                  <div>
                    <label className="block text-xs text-slate-400 mb-1">Статус</label>
                    <div className="flex gap-2">
                      {['ACTIVE', 'BANNED'].map(s => (
                        <button
                          key={s}
                          onClick={() => setSelectedStatus(s)}
                          className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-all ${
                            selectedStatus === s
                              ? (s === 'ACTIVE' ? 'border-emerald-500 bg-emerald-500/20 text-emerald-400' : 'border-red-500 bg-red-500/20 text-red-400')
                              : 'border-white/10 bg-white/5 text-slate-400 hover:bg-white/10'
                          }`}
                        >
                          {s}
                        </button>
                      ))}
                    </div>
                    {selectedStatus && (
                      <button onClick={handleUpdateStatus} disabled={saving} className={`mt-2 px-4 py-2 rounded-lg text-xs font-medium transition-colors disabled:opacity-50 flex items-center gap-1.5 ${
                        selectedStatus === 'BANNED' ? 'bg-red-500 hover:bg-red-600 text-white' : 'bg-emerald-500 hover:bg-emerald-600 text-white'
                      }`}>
                        <Save className="w-3 h-3" />{saving ? 'Saving...' : 'Сохранить статус'}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {/* ════════════ FINANCES TAB ════════════ */}
        {tab === 'finances' && (
          <div className="space-y-4">
            <div className="glass-card p-6">
              <h3 className="text-lg font-bold text-white mb-4">Заявки на вывод средств</h3>
              <p className="text-sm text-slate-400">
                Функционал заявок на вывод (Withdrawal Requests) будет подключён после интеграции платёжного шлюза.
                Текущая архитектура готова — добавьте таблицу <code className="text-blue-400">withdrawal_requests</code> и обработчик.
              </p>
              <div className="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
                <StatCard icon={DollarSign} label="Сумма в кошельках" value={stats ? `$${parseFloat(stats.total_balance).toLocaleString()}` : '—'} accent="bg-emerald-500" />
                <StatCard icon={CreditCard} label="Активные карты" value={stats?.active_cards ?? '—'} accent="bg-purple-500" />
                <StatCard icon={Users} label="Всего юзеров" value={stats?.total_users ?? '—'} accent="bg-blue-500" />
              </div>
            </div>
          </div>
        )}

        {/* ════════════ COMMISSIONS TAB ════════════ */}
        {tab === 'commissions' && (
          <div className="space-y-4">
            <div className="glass-card overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/10">
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Параметр</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Описание</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Значение</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium"></th>
                  </tr>
                </thead>
                <tbody>
                  {commissions.map(c => (
                    <tr key={c.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="px-4 py-3 text-white font-mono text-xs">{c.key}</td>
                      <td className="px-4 py-3 text-slate-400 text-xs">{c.description}</td>
                      <td className="px-4 py-3">
                        {editingCommission?.id === c.id ? (
                          <input
                            type="number"
                            step="0.01"
                            value={editingCommission.value}
                            onChange={e => setEditingCommission({ ...editingCommission, value: e.target.value })}
                            className="w-24 px-2 py-1 bg-white/10 border border-blue-500/50 rounded text-white text-sm outline-none"
                          />
                        ) : (
                          <span className="text-emerald-400 font-medium">{c.value}</span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        {editingCommission?.id === c.id ? (
                          <div className="flex gap-1">
                            <button onClick={handleSaveCommission} disabled={saving} className="px-3 py-1 bg-emerald-500 hover:bg-emerald-600 text-white rounded text-xs transition-colors disabled:opacity-50">
                              {saving ? '...' : 'Save'}
                            </button>
                            <button onClick={() => setEditingCommission(null)} className="px-3 py-1 bg-white/10 hover:bg-white/20 text-slate-300 rounded text-xs transition-colors">
                              ✕
                            </button>
                          </div>
                        ) : (
                          <button
                            onClick={() => setEditingCommission({ id: c.id, value: c.value })}
                            className="text-blue-400 hover:text-blue-300 text-xs transition-colors"
                          >
                            Изменить
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                  {commissions.length === 0 && (
                    <tr><td colSpan={4} className="px-4 py-8 text-center text-slate-500">Нет данных</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* ════════════ TICKETS TAB ════════════ */}
        {tab === 'tickets' && (
          <div className="space-y-4">
            {/* Filter */}
            <div className="flex gap-2 flex-wrap">
              {['', 'open', 'in_progress', 'resolved', 'closed'].map(f => (
                <button
                  key={f}
                  onClick={() => setTicketFilter(f)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-all ${
                    ticketFilter === f
                      ? 'border-blue-500 bg-blue-500/20 text-blue-400'
                      : 'border-white/10 bg-white/5 text-slate-400 hover:bg-white/10'
                  }`}
                >
                  {f === '' ? 'Все' : f === 'open' ? 'Открытые' : f === 'in_progress' ? 'В работе' : f === 'resolved' ? 'Решено' : 'Закрыто'}
                </button>
              ))}
            </div>

            {/* Tickets list */}
            <div className="space-y-3">
              {tickets.map(t => (
                <div key={t.id} className="glass-card p-5">
                  <div className="flex items-start justify-between gap-4 mb-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-xs text-slate-500 font-mono">#{t.id}</span>
                        <span className={`px-2 py-0.5 rounded-full text-[10px] font-medium ${
                          t.status === 'open' ? 'bg-orange-500/20 text-orange-400'
                          : t.status === 'in_progress' ? 'bg-blue-500/20 text-blue-400'
                          : t.status === 'resolved' ? 'bg-emerald-500/20 text-emerald-400'
                          : 'bg-slate-500/20 text-slate-400'
                        }`}>
                          {t.status === 'open' ? 'Открыт' : t.status === 'in_progress' ? 'В работе' : t.status === 'resolved' ? 'Решено' : 'Закрыт'}
                        </span>
                        <span className="text-xs text-slate-500">{t.created_at ? new Date(t.created_at).toLocaleString('ru-RU') : ''}</span>
                      </div>
                      <p className="text-sm text-slate-300 mb-1">
                        <strong className="text-white">{t.email}</strong>
                      </p>
                      <p className="text-sm text-slate-400 whitespace-pre-wrap break-words">{t.message || t.subject}</p>
                    </div>
                  </div>
                  <div className="flex gap-2 flex-wrap">
                    {t.status !== 'in_progress' && (
                      <button onClick={() => updateTicketStatus(t.id, 'in_progress')} className="px-3 py-1.5 bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/30 text-blue-400 rounded-lg text-xs font-medium transition-all">
                        В работу
                      </button>
                    )}
                    {t.status !== 'resolved' && (
                      <button onClick={() => updateTicketStatus(t.id, 'resolved')} className="px-3 py-1.5 bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/30 text-emerald-400 rounded-lg text-xs font-medium transition-all">
                        Решено
                      </button>
                    )}
                    {t.status !== 'closed' && (
                      <button onClick={() => updateTicketStatus(t.id, 'closed')} className="px-3 py-1.5 bg-slate-500/10 hover:bg-slate-500/20 border border-slate-500/30 text-slate-400 rounded-lg text-xs font-medium transition-all">
                        Закрыть
                      </button>
                    )}
                  </div>
                </div>
              ))}
              {tickets.length === 0 && (
                <div className="glass-card p-8 text-center text-slate-500 text-sm">
                  Нет тикетов{ticketFilter ? ` со статусом "${ticketFilter}"` : ''}
                </div>
              )}
            </div>
          </div>
        )}

        {/* ════════════ LOGS TAB ════════════ */}
        {tab === 'logs' && (
          <div className="space-y-4">
            <div className="glass-card overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/10">
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">ID</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Админ</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Действие</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Время</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map(l => (
                    <tr key={l.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                      <td className="px-4 py-3 text-slate-500 font-mono text-xs">{l.id}</td>
                      <td className="px-4 py-3 text-slate-300 text-xs">{l.admin_email}</td>
                      <td className="px-4 py-3 text-white text-xs">{l.action}</td>
                      <td className="px-4 py-3 text-slate-500 text-xs">{l.created_at ? new Date(l.created_at).toLocaleString('ru-RU') : ''}</td>
                    </tr>
                  ))}
                  {logs.length === 0 && (
                    <tr><td colSpan={4} className="px-4 py-8 text-center text-slate-500">Нет логов</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
