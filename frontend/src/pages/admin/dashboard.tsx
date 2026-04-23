import { useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import {
  Shield,
  Users,
  Wallet,
  CreditCard,
  MessageSquare,
  UserPlus,
  ListOrdered,
  Lock,
  RefreshCw,
  Pencil,
  Trash2,
  X,
  Copy,
  Check,
  AlertTriangle,
  ToggleLeft,
  ToggleRight,
} from 'lucide-react';

type DashboardStats = {
  totalUsers: number;
  totalBalance: string;
  activeCards: number;
  openTickets: number;
  todaySignups: number;
  totalCards: number;
};

type AezaBalance = { balance: number; currency: string; updated_at: string; status?: string };

type AezaServerInfo = {
  id: number; name: string; status: string; ip: string;
  cost_eur: number; expires_at: string;
  cpu: number; ram_mb: number; disk_gb: number; disk_type: string;
  os: string; location: string; api_status: string;
};

type VPNServerStatus = {
  active_clients: number;
  total_upload: number;
  total_download: number;
  total_traffic: number;
  server_limit_bytes: number;
  server_limit_gb: number;
  used_percent: number;
  traffic_alert: boolean;
  monthly_revenue: number;
  server_cost: number;
  margin: number;
  unique_vpn_clients: number;
  current_month_clients: number;
  prev_month_clients: number;
};

type VPNClientRow = {
  email: string;
  full_name: string;
  product_name: string;
  price_usd: number;
  created_at: string;
  provider_ref: string;
  activation_key: string;
  traffic_bytes: number;
  expire_ms: number;
  duration_days: number;
  used_bytes: number;
  used_percent: number;
  active: boolean;
};

const StatCard = ({
  icon,
  label,
  value,
  accent,
}: {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  accent: string;
}) => (
  <div className="glass-card p-5 flex items-center gap-4">
    <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${accent}`}>
      {icon}
    </div>
    <div>
      <p className="text-xs text-slate-400 uppercase tracking-wider">{label}</p>
      <p className="text-xl font-bold text-white">{value}</p>
    </div>
  </div>
);

export const AdminDashboardPage = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // ── Aeza infra balance ──
  const [aezaBalance, setAezaBalance] = useState<AezaBalance | null>(null);
  const [aezaLoading, setAezaLoading] = useState(false);
  const [aezaError, setAezaError] = useState('');
  const [activeKeys, setActiveKeys] = useState<number | null>(null);

  // ── Aeza Server Info ──
  const [aezaServer, setAezaServer] = useState<AezaServerInfo | null>(null);

  // ── VPN Server Status ──
  const [vpnServerStatus, setVpnServerStatus] = useState<VPNServerStatus | null>(null);
  const [vpnServerLoading, setVpnServerLoading] = useState(false);

  // ── VPN Active Clients ──
  const [vpnClients, setVpnClients] = useState<VPNClientRow[]>([]);

  // ── VPN Client Management Modals ──
  const [vpnDeleteConfirm, setVpnDeleteConfirm] = useState<{ email: string; providerRef: string } | null>(null);
  const [vpnDeleting, setVpnDeleting] = useState(false);
  const [vpnEditModal, setVpnEditModal] = useState<{
    providerRef: string;
    email: string;
    trafficBytes: number;
    expireMs: number;
  } | null>(null);
  const [vpnEditForm, setVpnEditForm] = useState({ trafficGB: '', expiryDate: '' });
  const [vpnEditing, setVpnEditing] = useState(false);
  const [copiedRef, setCopiedRef] = useState<string | null>(null);

  const [pin, setPin] = useState('');
  const [pinSaving, setPinSaving] = useState(false);
  const [pinOk, setPinOk] = useState(false);
  const [pinError, setPinError] = useState('');

  const load = async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<DashboardStats>('/admin/dashboard-stats');
      setStats(res.data || null);
    } catch {
      setStats(null);
      setError('Не удалось загрузить данные Dashboard');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    fetchAezaBalance();
    fetchAezaServerInfo();
    fetchActiveKeys();
    fetchVPNServerStatus();
    fetchVPNActiveClients();
  }, []);

  const totalBalanceLabel = useMemo(() => {
    if (!stats) return '—';
    const n = Number(stats.totalBalance);
    if (Number.isFinite(n)) return `€${n.toLocaleString('ru-RU')}`;
    return `€${stats.totalBalance}`;
  }, [stats]);

  const savePIN = async () => {
    setPinSaving(true);
    setPinOk(false);
    setPinError('');
    try {
      await apiClient.patch('/admin/staff-pin', { pin });
      setPin('');
      setPinOk(true);
      window.setTimeout(() => setPinOk(false), 2500);
    } catch {
      setPinError('Не удалось изменить PIN');
    } finally {
      setPinSaving(false);
    }
  };

  const fetchAezaBalance = async () => {
    setAezaLoading(true);
    setAezaError('');
    try {
      const res = await apiClient.get<AezaBalance>('/admin/infra/balance');
      setAezaBalance(res.data || null);
    } catch (err: any) {
      setAezaError(err?.response?.data?.error || 'Не удалось получить баланс');
    } finally {
      setAezaLoading(false);
    }
  };

  const fetchAezaServerInfo = async () => {
    try {
      const res = await apiClient.get<AezaServerInfo>('/admin/infra/server-info');
      setAezaServer(res.data || null);
    } catch {
      // ignore
    }
  };

  const fetchActiveKeys = async () => {
    try {
      const res = await apiClient.get<{ active_keys: number }>('/admin/infra/active-keys');
      setActiveKeys(res.data?.active_keys ?? null);
    } catch {
      // ignore
    }
  };

  const fetchVPNServerStatus = async () => {
    setVpnServerLoading(true);
    try {
      const res = await apiClient.get<VPNServerStatus>('/admin/infra/vpn-server-status');
      setVpnServerStatus(res.data || null);
    } catch {
      // ignore
    } finally {
      setVpnServerLoading(false);
    }
  };

  const fetchVPNActiveClients = async () => {
    try {
      const res = await apiClient.get<{ clients: VPNClientRow[] }>('/admin/infra/vpn-active-clients');
      setVpnClients(res.data?.clients || []);
    } catch {
      // ignore
    }
  };

  const copyText = async (text: string, key: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedRef(key);
      window.setTimeout(() => setCopiedRef(null), 1500);
    } catch {
      // ignore
    }
  };

  const handleDeleteVPNClient = async () => {
    if (!vpnDeleteConfirm) return;
    setVpnDeleting(true);
    try {
      await apiClient.delete(`/admin/vpn/client/${encodeURIComponent(vpnDeleteConfirm.providerRef)}`);
      setVpnDeleteConfirm(null);
      fetchVPNActiveClients();
      fetchVPNServerStatus();
      fetchActiveKeys();
    } catch {
      // ignore
    } finally {
      setVpnDeleting(false);
    }
  };

  const handleEditVPNClient = async () => {
    if (!vpnEditModal) return;
    setVpnEditing(true);
    try {
      const body: { total_bytes?: number; expiry_ms?: number } = {};
      if (vpnEditForm.trafficGB) {
        body.total_bytes = Math.round(parseFloat(vpnEditForm.trafficGB) * 1024 * 1024 * 1024);
      }
      if (vpnEditForm.expiryDate) {
        body.expiry_ms = new Date(vpnEditForm.expiryDate).getTime();
      }
      if (!body.total_bytes && !body.expiry_ms) {
        setVpnEditing(false);
        return;
      }
      await apiClient.patch(`/admin/vpn/client/${encodeURIComponent(vpnEditModal.providerRef)}`, body);
      setVpnEditModal(null);
      fetchVPNActiveClients();
    } catch {
      // ignore
    } finally {
      setVpnEditing(false);
    }
  };

  return (
    <div className="stagger-fade-in space-y-6">
      <div className="flex items-center gap-4">
        <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center">
          <Shield className="w-7 h-7 text-blue-300" />
        </div>
        <div className="min-w-0">
          <h1 className="text-2xl md:text-3xl font-bold text-white">Admin Panel</h1>
          <p className="text-slate-400 text-sm">Закрытая зона управления</p>
        </div>
      </div>

      {error ? <p className="text-sm text-red-400">{error}</p> : null}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard
          icon={<Users className="w-6 h-6 text-white" />}
          label="Всего юзеров"
          value={stats ? stats.totalUsers : loading ? '…' : '—'}
          accent="bg-blue-500/20 border border-blue-500/30"
        />
        <StatCard
          icon={<Wallet className="w-6 h-6 text-white" />}
          label="Сумма в кошельках"
          value={stats ? totalBalanceLabel : loading ? '…' : '—'}
          accent="bg-emerald-500/20 border border-emerald-500/30"
        />
        <StatCard
          icon={<CreditCard className="w-6 h-6 text-white" />}
          label="Активные карты"
          value={stats ? stats.activeCards : loading ? '…' : '—'}
          accent="bg-purple-500/20 border border-purple-500/30"
        />
        <StatCard
          icon={<MessageSquare className="w-6 h-6 text-white" />}
          label="Открытые тикеты"
          value={stats ? stats.openTickets : loading ? '…' : '—'}
          accent="bg-orange-500/20 border border-orange-500/30"
        />
        <StatCard
          icon={<UserPlus className="w-6 h-6 text-white" />}
          label="Регистрации сегодня"
          value={stats ? stats.todaySignups : loading ? '…' : '—'}
          accent="bg-cyan-500/20 border border-cyan-500/30"
        />
        <StatCard
          icon={<ListOrdered className="w-6 h-6 text-white" />}
          label="Всего карт"
          value={stats ? stats.totalCards : loading ? '…' : '—'}
          accent="bg-slate-500/20 border border-white/10"
        />
      </div>

      <div className="glass-card p-6">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500 to-purple-600 flex items-center justify-center">
            <Lock className="w-5 h-5 text-white" />
          </div>
          <div>
            <p className="text-white font-semibold">PIN админки</p>
            <p className="text-sm text-slate-400">Смена PIN для входа (4 цифры)</p>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
          <input
            value={pin}
            onChange={(e) => { setPin(e.target.value.replace(/[^\d]/g, '').slice(0, 4)); setPinError(''); }}
            inputMode="numeric"
            pattern="[0-9]*"
            maxLength={4}
            placeholder="0000"
            className="w-full sm:max-w-[240px] bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-white text-center text-lg tracking-[0.3em] font-mono outline-none focus:border-blue-500/40"
          />
          <button
            onClick={savePIN}
            disabled={pinSaving || pin.length !== 4}
            className="px-5 py-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-semibold transition-colors disabled:opacity-50"
          >
            {pinSaving ? '...' : 'Сохранить'}
          </button>
          {pinOk ? <span className="text-sm text-emerald-400">Сохранено</span> : null}
        </div>
        {pinError ? <p className="text-sm text-red-400 mt-3">{pinError}</p> : null}
      </div>

      {/* ── Финансы VPN ── */}
      {vpnServerStatus && (
        <div className="glass-card p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Финансы VPN (текущий месяц)</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/20">
              <p className="text-xs text-slate-400 mb-1">Выручка</p>
              <p className="text-2xl font-bold text-emerald-400">€{vpnServerStatus.monthly_revenue?.toFixed(2) ?? '0.00'}</p>
            </div>
            <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/20">
              <p className="text-xs text-slate-400 mb-1">Затраты</p>
              <p className="text-2xl font-bold text-red-400">€{vpnServerStatus.server_cost?.toFixed(2) ?? '4.94'}</p>
            </div>
            <div className="p-4 rounded-xl bg-blue-500/10 border border-blue-500/20">
              <p className="text-xs text-slate-400 mb-1">Прибыль</p>
              <p className="text-2xl font-bold text-blue-300">
                {(vpnServerStatus.margin ?? 0) >= 0 ? '+' : ''}€{vpnServerStatus.margin?.toFixed(2) ?? '0.00'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* ── Aeza Server Info ── */}
      {aezaServer && aezaServer.api_status !== 'error' && (
        <div className="glass-card p-6">
          <div className="flex items-center justify-between gap-4 flex-wrap">
            <div>
              <h3 className="text-lg font-semibold text-white">VPN-сервер (Aeza #{aezaServer.id})</h3>
              <p className="text-sm text-slate-400 mt-1">
                <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-500/20 text-emerald-400">
                  {aezaServer.status === 'active' ? 'Активен' : aezaServer.status}
                </span>
              </p>
            </div>
            <div className="text-right">
              <p className="text-xs text-slate-500">Стоимость</p>
              <p className="text-lg font-bold text-white">€{aezaServer.cost_eur.toFixed(2)}</p>
              <p className="text-xs text-slate-500 mt-1">
                Оплачен до{' '}
                <span className="text-slate-300">
                  {aezaServer.expires_at ? new Date(aezaServer.expires_at).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' }) : '—'}
                </span>
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mt-5">
            <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
              <p className="text-xs text-slate-500">CPU / RAM</p>
              <p className="text-sm text-slate-200 font-semibold mt-1">{aezaServer.cpu} vCPU / {(aezaServer.ram_mb / 1024).toFixed(0)} GB</p>
            </div>
            <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
              <p className="text-xs text-slate-500">Диск</p>
              <p className="text-sm text-slate-200 font-semibold mt-1">{aezaServer.disk_gb} GB {aezaServer.disk_type}</p>
            </div>
            <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
              <p className="text-xs text-slate-500">IP</p>
              <p className="text-sm text-slate-200 font-semibold mt-1">{aezaServer.ip || '—'}</p>
            </div>
            <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
              <p className="text-xs text-slate-500">Локация</p>
              <p className="text-sm text-slate-200 font-semibold mt-1">{aezaServer.location || '—'}</p>
            </div>
          </div>
        </div>
      )}

      {/* ── VPN Server Status ── */}
      <div className="glass-card p-6">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <h3 className="text-lg font-semibold text-white">Выданные доступы</h3>
            <p className="text-sm text-slate-400 mt-1">Статус и трафик сервера</p>
          </div>
          <button
            onClick={() => { fetchVPNServerStatus(); fetchActiveKeys(); fetchVPNActiveClients(); fetchAezaServerInfo(); }}
            disabled={vpnServerLoading}
            className="px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-xs text-white/60 hover:text-white hover:bg-white/10 transition-all disabled:opacity-50 inline-flex items-center gap-2"
          >
            <RefreshCw className="w-3.5 h-3.5" />
            {vpnServerLoading ? '...' : 'Обновить'}
          </button>
        </div>

        {vpnServerStatus ? (
          <div className="mt-5">
            {vpnServerStatus.traffic_alert && (
              <div className="mb-4 p-3 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 text-sm flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 mt-0.5 shrink-0" />
                <div>
                  <p className="font-semibold">Высокая загрузка трафика</p>
                  <p className="text-xs text-amber-200/70">Проверьте лимит/остаток трафика и при необходимости пополните.</p>
                </div>
              </div>
            )}

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
              <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
                <p className="text-xs text-slate-500">Активных ключей</p>
                <p className="text-xl font-bold text-white mt-1">{activeKeys !== null ? activeKeys : vpnServerStatus.active_clients}</p>
              </div>
              <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
                <p className="text-xs text-slate-500">Клиентов</p>
                <p className="text-xl font-bold text-white mt-1">{vpnServerStatus.unique_vpn_clients ?? 0}</p>
              </div>
              <div className="p-4 rounded-xl bg-white/[0.03] border border-white/10">
                <p className="text-xs text-slate-500">Лимит</p>
                <p className="text-xl font-bold text-white mt-1">{vpnServerStatus.server_limit_gb} ГБ</p>
              </div>
            </div>

            <div className="mb-3">
              <div className="flex items-center justify-between text-xs text-slate-500 mb-2">
                <span>Трафик сервера</span>
                <span>{(vpnServerStatus.total_traffic / (1024 * 1024 * 1024)).toFixed(1)} / {vpnServerStatus.server_limit_gb} ГБ</span>
              </div>
              <div className="h-2.5 bg-white/5 rounded-full overflow-hidden border border-white/10">
                <div
                  className={`h-full ${vpnServerStatus.used_percent > 70 ? 'bg-amber-500' : 'bg-emerald-500'}`}
                  style={{ width: `${Math.min(vpnServerStatus.used_percent, 100)}%` }}
                />
              </div>
              <div className="flex items-center justify-between text-xs text-slate-500 mt-2">
                <span>{vpnServerStatus.used_percent.toFixed(1)}% использовано</span>
                <span>↓ {(vpnServerStatus.total_download / (1024 * 1024 * 1024)).toFixed(2)} ГБ</span>
              </div>
            </div>
          </div>
        ) : (
          <div className="mt-6 text-sm text-slate-500">{vpnServerLoading ? 'Загрузка...' : 'Нет данных'}</div>
        )}
      </div>

      {/* ── Активные VPN-пользователи ── */}
      {vpnClients.length > 0 && (
        <div className="glass-card p-6">
          <div className="flex items-center justify-between gap-4 flex-wrap">
            <div>
              <h3 className="text-lg font-semibold text-white">Активные VPN-пользователи</h3>
              <p className="text-sm text-slate-400 mt-1">{vpnClients.length} записей</p>
            </div>
          </div>

          <div className="mt-5 overflow-x-auto">
            <table className="min-w-[980px] w-full text-left">
              <thead>
                <tr className="text-xs text-slate-500">
                  <th className="py-3 px-2 font-semibold">Email</th>
                  <th className="py-3 px-2 font-semibold">Тариф</th>
                  <th className="py-3 px-2 font-semibold">Ключ</th>
                  <th className="py-3 px-2 font-semibold">Использовано</th>
                  <th className="py-3 px-2 font-semibold">Срок</th>
                  <th className="py-3 px-2 font-semibold">Статус</th>
                  <th className="py-3 px-2 font-semibold text-right">Действия</th>
                </tr>
              </thead>
              <tbody>
                {vpnClients.map((c) => {
                  const totalGB = (c.traffic_bytes / (1024 * 1024 * 1024)).toFixed(0);
                  const usedGB = (c.used_bytes / (1024 * 1024 * 1024)).toFixed(1);
                  const daysLeft = c.expire_ms ? Math.ceil((c.expire_ms - Date.now()) / 86400000) : 0;
                  const keyShort = c.activation_key ? `vless://…${c.activation_key.slice(-8)}` : c.provider_ref;
                  return (
                    <tr key={c.provider_ref} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                      <td className="py-3 px-2 text-sm text-slate-200">{c.email}</td>
                      <td className="py-3 px-2 text-sm text-blue-300">{c.product_name}</td>
                      <td className="py-3 px-2">
                        <button
                          onClick={() => copyText(c.activation_key || c.provider_ref, c.provider_ref)}
                          className="inline-flex items-center gap-2 px-2.5 py-1.5 rounded-lg bg-white/5 border border-white/10 text-xs text-slate-300 hover:bg-white/10 transition-colors font-mono"
                        >
                          {copiedRef === c.provider_ref ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                          {keyShort}
                        </button>
                      </td>
                      <td className="py-3 px-2 text-sm text-slate-300">
                        {usedGB}/{totalGB} ГБ ({c.used_percent.toFixed(0)}%)
                      </td>
                      <td className="py-3 px-2 text-sm text-slate-300">
                        {daysLeft > 0 ? `${daysLeft} дн.` : '—'}
                      </td>
                      <td className="py-3 px-2">
                        <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${c.active ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-500/20 text-slate-400'}`}>
                          {c.active ? 'Активен' : 'Неактивен'}
                        </span>
                      </td>
                      <td className="py-3 px-2">
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => {
                              setVpnEditModal({ providerRef: c.provider_ref, email: c.email, trafficBytes: c.traffic_bytes, expireMs: c.expire_ms });
                              setVpnEditForm({ trafficGB: (c.traffic_bytes / (1024 * 1024 * 1024)).toFixed(0), expiryDate: c.expire_ms ? new Date(c.expire_ms).toISOString().slice(0, 10) : '' });
                            }}
                            className="p-2 rounded-lg bg-white/5 border border-white/10 text-slate-300 hover:bg-white/10 transition-colors"
                            title="Редактировать"
                          >
                            <Pencil className="w-4 h-4" />
                          </button>
                          <button
                            onClick={() => setVpnDeleteConfirm({ email: c.email, providerRef: c.provider_ref })}
                            className="p-2 rounded-lg bg-red-500/10 border border-red-500/30 text-red-300 hover:bg-red-500/20 transition-colors"
                            title="Удалить"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ── VPN Delete Confirmation Modal ── */}
      {vpnDeleteConfirm && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={() => !vpnDeleting && setVpnDeleteConfirm(null)}>
          <div className="glass-card p-6 w-full max-w-md mx-4" onClick={(e) => e.stopPropagation()}>
            <button onClick={() => !vpnDeleting && setVpnDeleteConfirm(null)} className="absolute top-3 right-3 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all">
              <X className="w-4 h-4" />
            </button>
            <h3 className="text-lg font-bold text-white mb-2">Удалить VPN-клиента?</h3>
            <p className="text-sm text-slate-400 mb-5">Вы уверены, что хотите удалить пользователя <span className="text-slate-200">{vpnDeleteConfirm.email || vpnDeleteConfirm.providerRef}</span> и заблокировать его доступ?</p>
            <div className="flex gap-3 justify-end">
              <button onClick={() => setVpnDeleteConfirm(null)} disabled={vpnDeleting} className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50">
                Отмена
              </button>
              <button onClick={handleDeleteVPNClient} disabled={vpnDeleting} className="px-4 py-2 rounded-xl bg-red-500/20 hover:bg-red-500/25 border border-red-500/30 text-red-200 text-sm font-semibold transition-colors disabled:opacity-50">
                {vpnDeleting ? 'Удаление...' : 'Удалить'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── VPN Edit Modal ── */}
      {vpnEditModal && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={() => !vpnEditing && setVpnEditModal(null)}>
          <div className="glass-card p-6 w-full max-w-md mx-4" onClick={(e) => e.stopPropagation()}>
            <button onClick={() => !vpnEditing && setVpnEditModal(null)} className="absolute top-3 right-3 p-2 rounded-xl bg-black/50 border border-white/10 text-slate-400 hover:text-white transition-all">
              <X className="w-4 h-4" />
            </button>
            <h3 className="text-lg font-bold text-white mb-2">Редактировать VPN</h3>
            <p className="text-sm text-slate-400 mb-5">{vpnEditModal.email || vpnEditModal.providerRef}</p>

            <div className="space-y-4">
              <div>
                <label className="text-xs text-slate-500">Лимит трафика (ГБ)</label>
                <input
                  value={vpnEditForm.trafficGB}
                  onChange={(e) => setVpnEditForm(prev => ({ ...prev, trafficGB: e.target.value }))}
                  placeholder={`Текущий: ${(vpnEditModal.trafficBytes / (1024 * 1024 * 1024)).toFixed(0)} ГБ`}
                  className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40 font-mono"
                />
              </div>
              <div>
                <label className="text-xs text-slate-500">Срок (дата)</label>
                <input
                  type="date"
                  value={vpnEditForm.expiryDate}
                  onChange={(e) => setVpnEditForm(prev => ({ ...prev, expiryDate: e.target.value }))}
                  className="mt-2 w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
                />
              </div>
            </div>

            <div className="flex gap-3 justify-end mt-6">
              <button onClick={() => setVpnEditModal(null)} disabled={vpnEditing} className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50">
                Отмена
              </button>
              <button onClick={handleEditVPNClient} disabled={vpnEditing} className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-sm font-semibold transition-colors disabled:opacity-50">
                {vpnEditing ? 'Сохранение...' : 'Сохранить'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

