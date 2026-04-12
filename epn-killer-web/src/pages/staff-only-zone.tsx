import React, { useState, useEffect, useCallback, useRef } from 'react';
import apiClient from '../api/axios';
import { useAuth } from '../store/auth-context';
import { SBPToggle } from '../components/sbp-toggle';
import { compressImage } from '../utils/compress-image';
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
  Eye,
  Ban,
  ShieldCheck,
  Send,
  Mail,
  FileSearch,
  X,
  Filter,
  Newspaper,
  Trash2,
  Plus,
  Upload,
  ImageIcon,
  Tag,
  ShoppingBag,
} from 'lucide-react';

type Tab = 'dashboard' | 'users' | 'commissions' | 'tickets' | 'news' | 'logs' | 'store';

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
  is_blocked: boolean;
  card_count: number;
  wallet_balance: string;
  is_telegram_linked: boolean;
  notification_pref: string;
  created_at: string;
  tier: string;
  tier_expires_at: string;
}

interface UserFullDetails {
  id: number;
  email: string;
  status: string;
  wallet_balance: string;
  balance_rub: string;
  is_admin: boolean;
  is_verified: boolean;
  is_blocked: boolean;
  telegram_chat_id: number;
  is_telegram_linked: boolean;
  notification_pref: string;
  created_at: string;
  cards: CardSummary[];
  transactions: TxSummary[];
}

interface CardSummary {
  id: number;
  last_4: string;
  card_status: string;
  card_balance: string;
  card_type: string;
  category: string;
  nickname: string;
  created_at: string;
}

interface TxSummary {
  id: number;
  amount: string;
  fee: string;
  transaction_type: string;
  status: string;
  details: string;
  source_type: string;
  currency: string;
  card_last_4: string;
  executed_at: string;
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
  claimed_by: number;
  created_at: string;
}

interface LiveChat {
  id: number;
  user_id: number;
  user_email: string;
  topic: string;
  status: string;
  claimed_by: number;
  claimer_email: string;
  message_count: number;
  created_at: string;
  updated_at: string;
}

interface ChatMsg {
  id: number;
  conversation_id: number;
  sender_type: string;
  sender_name: string;
  body: string;
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

// ── Grade Select (only STANDARD and GOLD tiers) ──
const GRADES = ['STANDARD', 'GOLD'] as const;
const gradeColors: Record<string, string> = {
  STANDARD: 'text-slate-400',
  GOLD: 'text-yellow-400',
};

// ── Russian labels for system_settings keys ──
const SETTING_LABELS: Record<string, string> = {
  sbp_enabled: 'СБП (пополнение)',
  gold_tier_price: 'Цена Gold-пакета (USD)',
  gold_tier_duration_days: 'Длительность Gold (дней)',
  fee_standard: 'Комиссия — тир Стандарт ($)',
  fee_gold: 'Комиссия — тир Gold ($)',
};

// ══════════════════════════════════════
// MAIN COMPONENT
// ══════════════════════════════════════
const SUPER_ADMIN_EMAIL = 'aalabin5@gmail.com';

export const StaffOnlyZone = () => {
  const { user: currentUser } = useAuth();
  const isSuperAdmin = currentUser?.email === SUPER_ADMIN_EMAIL;
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
  const [sysSettings, setSysSettings] = useState<{ key: string; value: string; description: string }[]>([]);
  const [editingSetting, setEditingSetting] = useState<{ key: string; value: string } | null>(null);
  const [logs, setLogs] = useState<AdminLog[]>([]);
  const [tickets, setTickets] = useState<SupportTicket[]>([]);
  const [ticketFilter, setTicketFilter] = useState('');
  const [liveChats, setLiveChats] = useState<LiveChat[]>([]);
  const [chatMessages, setChatMessages] = useState<ChatMsg[]>([]);
  const [viewingChatId, setViewingChatId] = useState<number | null>(null);
  const [saving, setSaving] = useState(false);
  const [freezeConfirm, setFreezeConfirm] = useState(false);
  const [inspectUser, setInspectUser] = useState<UserFullDetails | null>(null);
  const [inspectLoading, setInspectLoading] = useState(false);
  const [txFilter, setTxFilter] = useState('');
  const [blockConfirmUser, setBlockConfirmUser] = useState<AdminUser | null>(null);
  const [toast, setToast] = useState<{ msg: string; type: 'ok' | 'err' } | null>(null);
  const [newsList, setNewsList] = useState<{ id: number; title: string; content: string; image_url: string; status: string; created_at: string }[]>([]);
  const [newNewsTitle, setNewNewsTitle] = useState('');
  const [newNewsContent, setNewNewsContent] = useState('');
  const [newsImageFile, setNewsImageFile] = useState<File | null>(null);
  const [newsImagePreview, setNewsImagePreview] = useState<string | null>(null);
  const [newsPublishing, setNewsPublishing] = useState(false);
  const [newsUploadProgress, setNewsUploadProgress] = useState('');
  const [editingNews, setEditingNews] = useState<{ id: number; title: string; content: string; image_url: string; status: string } | null>(null);
  const newsFileRef = useRef<HTMLInputElement>(null);
  const editNewsFileRef = useRef<HTMLInputElement>(null);

  // ── Store pricing state ──
  const [storeProducts, setStoreProducts] = useState<{ id: number; name: string; product_type: string; provider: string; cost_price: string; markup_percent: string; retail_price: string; old_price: string; in_stock: boolean; external_id: string; image_url: string }[]>([]);
  const [editingProduct, setEditingProduct] = useState<{ id: number; cost_price: string; markup_percent: string; image_url: string } | null>(null);
  const [bulkDelta, setBulkDelta] = useState('10');
  const [bulkType, setBulkType] = useState('');
  const [storeSubTab, setStoreSubTab] = useState<'esim' | 'digital'>('esim');

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

  // ── System settings (tier fees, gold price, etc.) ──
  const loadSysSettings = useCallback(async () => {
    try {
      const res = await apiClient.get('/admin/system-settings');
      setSysSettings(res.data || []);
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
      const res = await apiClient.get('/admin/tickets');
      setTickets(res.data || []);
    } catch { /* ignore */ }
  }, []);

  // ── Live chats ──
  const loadChats = useCallback(async () => {
    try {
      const res = await apiClient.get('/admin/chats');
      setLiveChats(res.data || []);
    } catch { /* ignore */ }
  }, []);

  const viewChatMessages = async (chatId: number) => {
    try {
      const res = await apiClient.get(`/admin/chats/${chatId}/messages`);
      setChatMessages(res.data || []);
      setViewingChatId(chatId);
    } catch {
      showToast('Ошибка загрузки сообщений', 'err');
    }
  };


  const handleToggleRole = async (user: AdminUser) => {
    try {
      setSaving(true);
      const res = await apiClient.patch(`/admin/users/${user.id}/role`);
      const newAdmin = res.data.is_admin;
      showToast(`${user.email} — ${newAdmin ? 'назначен админом' : 'снят с админа'}`);
      setUsers(prev => prev.map(u => u.id === user.id ? { ...u, is_admin: newAdmin, role: newAdmin ? 'admin' : 'user' } : u));
      setSelectedUser(prev => prev && prev.id === user.id ? { ...prev, is_admin: newAdmin, role: newAdmin ? 'admin' : 'user' } : prev);
    } catch (err: any) {
      const msg = typeof err.response?.data === 'string' ? err.response.data : 'Ошибка изменения роли';
      showToast(msg, 'err');
    } finally { setSaving(false); }
  };

  const toggleBlock = async (user: AdminUser) => {
    try {
      setSaving(true);
      const res = await apiClient.post(`/admin/users/${user.id}/toggle-block`);
      const newBlocked = res.data.is_blocked;
      showToast(`${user.email} — ${newBlocked ? 'заблокирован' : 'разблокирован'}`);
      setBlockConfirmUser(null);
      setUsers(prev => prev.map(u => u.id === user.id ? { ...u, is_blocked: newBlocked } : u));
    } catch { showToast('Ошибка блокировки', 'err'); }
    finally { setSaving(false); }
  };

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

  const loadNews = useCallback(async () => {
    try {
      const res = await apiClient.get('/user/news', { params: { limit: 50, offset: 0 } });
      setNewsList(res.data?.items || []);
    } catch { /* ignore */ }
  }, []);

  const handleNewsImageSelect = async (file: File | null) => {
    if (!file) { setNewsImageFile(null); setNewsImagePreview(null); return; }
    try {
      const compressed = await compressImage(file);
      setNewsImageFile(compressed);
      setNewsImagePreview(URL.createObjectURL(compressed));
    } catch {
      showToast('Ошибка сжатия изображения', 'err');
    }
  };

  const publishNews = async () => {
    if (!newNewsTitle.trim() || !newNewsContent.trim()) return;
    setNewsPublishing(true);
    try {
      let imageUrl = '';
      // Step 1: Upload image if selected
      if (newsImageFile) {
        setNewsUploadProgress('Загрузка изображения...');
        const formData = new FormData();
        formData.append('image', newsImageFile);
        const uploadRes = await apiClient.post('/admin/upload-image', formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
        imageUrl = uploadRes.data?.url || '';
      }
      // Step 2: Create news as draft
      setNewsUploadProgress('Сохранение...');
      await apiClient.post('/admin/news', { title: newNewsTitle, content: newNewsContent, image_url: imageUrl, status: 'draft' });
      showToast('Новость сохранена как черновик');
      setNewNewsTitle(''); setNewNewsContent('');
      setNewsImageFile(null); setNewsImagePreview(null);
      setNewsUploadProgress('');
      loadNews();
    } catch { showToast('Ошибка публикации', 'err'); }
    finally { setNewsPublishing(false); setNewsUploadProgress(''); }
  };

  const toggleNewsStatus = async (id: number, currentStatus: string) => {
    const newStatus = currentStatus === 'published' ? 'draft' : 'published';
    setSaving(true);
    try {
      await apiClient.patch(`/admin/news/${id}`, { status: newStatus });
      showToast(newStatus === 'published' ? 'Новость опубликована' : 'Новость снята с публикации');
      loadNews();
    } catch { showToast('Ошибка обновления статуса', 'err'); }
    finally { setSaving(false); }
  };

  const deleteNews = async (id: number) => {
    try {
      await apiClient.delete(`/admin/news/${id}`);
      showToast('Новость удалена');
      loadNews();
    } catch { showToast('Ошибка удаления', 'err'); }
  };

  const saveEditedNews = async () => {
    if (!editingNews) return;
    setSaving(true);
    try {
      await apiClient.put(`/admin/news/${editingNews.id}`, {
        title: editingNews.title,
        content: editingNews.content,
        image_url: editingNews.image_url,
        status: editingNews.status,
      });
      showToast('Новость обновлена');
      setEditingNews(null);
      loadNews();
    } catch { showToast('Ошибка обновления', 'err'); }
    finally { setSaving(false); }
  };

  const handleEditNewsImageUpload = async (file: File | null) => {
    if (!file || !editingNews) return;
    try {
      const compressed = await compressImage(file);
      const formData = new FormData();
      formData.append('image', compressed);
      const res = await apiClient.post('/admin/upload-image', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
      setEditingNews({ ...editingNews, image_url: res.data.url });
      showToast('Изображение загружено');
    } catch { showToast('Ошибка загрузки', 'err'); }
  };

  // ── Load store products ──
  const loadStoreProducts = useCallback(async () => {
    try {
      const res = await apiClient.get('/admin/store/products');
      setStoreProducts(res.data || []);
    } catch { /* ignore */ }
  }, []);

  const handleSaveProduct = async () => {
    if (!editingProduct) return;
    setSaving(true);
    try {
      await apiClient.patch(`/admin/store/products/${editingProduct.id}`, {
        cost_price: parseFloat(editingProduct.cost_price),
        markup_percent: parseFloat(editingProduct.markup_percent),
        image_url: editingProduct.image_url,
      });
      showToast('Цена обновлена');
      setEditingProduct(null);
      loadStoreProducts();
    } catch { showToast('Ошибка сохранения', 'err'); }
    finally { setSaving(false); }
  };

  const handleBulkMarkup = async () => {
    const delta = parseFloat(bulkDelta);
    if (!delta || delta === 0) return;
    setSaving(true);
    try {
      const res = await apiClient.post('/admin/store/bulk-markup', { delta, product_type: storeSubTab });
      showToast(`Наценка +${delta}% применена к ${res.data.affected} товарам`);
      loadStoreProducts();
    } catch { showToast('Ошибка массового обновления', 'err'); }
    finally { setSaving(false); }
  };

  useEffect(() => {
    if (tab === 'users') loadAllUsers();
    if (tab === 'commissions') { loadCommissions(); loadSysSettings(); }
    if (tab === 'tickets') { loadTickets(); loadChats(); }
    if (tab === 'news') loadNews();
    if (tab === 'logs') loadLogs();
    if (tab === 'store') loadStoreProducts();
  }, [tab, loadAllUsers, loadCommissions, loadSysSettings, loadTickets, loadChats, loadNews, loadLogs, loadStoreProducts]);

  // ── Inspect User (Financial Passport) ──
  const inspectUserDetails = async (userId: number) => {
    setInspectLoading(true);
    setTxFilter('');
    try {
      const res = await apiClient.get(`/admin/users/${userId}/full-details`);
      setInspectUser(res.data);
    } catch {
      showToast('Ошибка загрузки данных пользователя', 'err');
    } finally {
      setInspectLoading(false);
    }
  };

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

  // ── Save system setting ──
  const handleSaveSetting = async () => {
    if (!editingSetting) return;
    setSaving(true);
    try {
      await apiClient.patch(`/admin/system-settings/${editingSetting.key}`, { value: editingSetting.value });
      showToast('Настройка обновлена');
      setEditingSetting(null);
      loadSysSettings();
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
      className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-medium transition-all whitespace-nowrap shrink-0 ${
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
          <TabBtn id="commissions" icon={Settings} label="Комиссии" />
          <TabBtn id="tickets" icon={MessageSquare} label="Тикеты" />
          <TabBtn id="news" icon={Newspaper} label="Новости" />
          <TabBtn id="store" icon={ShoppingBag} label="Магазин" />
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

            {/* System Settings */}
            <div className="glass-card p-6">
              <h3 className="text-lg font-bold text-white mb-4">Системные настройки</h3>
              <SBPToggle />
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
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Кошелёк</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Статус</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Карты</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Связь</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Срок Gold</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Дата</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map(u => (
                      <tr key={u.id} className={`border-b border-white/5 transition-colors ${u.is_blocked ? 'bg-red-500/[0.06] hover:bg-red-500/[0.1]' : 'hover:bg-white/5'}`}>
                        <td className="px-4 py-3 text-slate-300 font-mono text-xs">{u.id}</td>
                        <td className="px-4 py-3 text-white">{u.email}</td>
                        <td className="px-4 py-3 text-emerald-400 font-medium">${u.wallet_balance || '0.00'}</td>
                        <td className="px-4 py-3">
                          {u.is_blocked
                            ? <span className="px-2 py-0.5 rounded-full text-[10px] font-medium bg-red-500/20 text-red-400">Blocked</span>
                            : <span className="px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-500/20 text-emerald-400">Active</span>
                          }
                        </td>
                        <td className="px-4 py-3 text-slate-300">{u.card_count} шт.</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-1.5 relative group">
                            <Send className={`w-3.5 h-3.5 ${u.is_telegram_linked ? 'text-blue-400' : 'text-slate-600'}`} />
                            <Mail className={`w-3.5 h-3.5 ${u.notification_pref === 'email' || u.notification_pref === 'both' ? 'text-amber-400' : 'text-slate-600'}`} />
                            <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 bg-slate-800 border border-white/10 text-[10px] text-slate-300 rounded-lg whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                              Канал: {u.notification_pref === 'both' ? 'TG + Email' : u.notification_pref === 'telegram' ? 'Telegram' : 'Email'}
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-xs">
                          {u.tier === 'gold' && u.tier_expires_at ? (() => {
                            const exp = new Date(u.tier_expires_at);
                            const days = Math.ceil((exp.getTime() - Date.now()) / (1000 * 60 * 60 * 24));
                            const expired = days <= 0;
                            const color = expired ? 'text-red-400' : days <= 5 ? 'text-red-400' : days <= 30 ? 'text-orange-400' : 'text-emerald-400';
                            return <span className={`font-medium ${color}`}>{expired ? 'Истёк' : exp.toLocaleDateString('ru-RU')}</span>;
                          })() : <span className="text-slate-600">—</span>}
                        </td>
                        <td className="px-4 py-3 text-slate-500 text-xs">{u.created_at?.slice(0, 10)}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-1">
                            <button
                              onClick={() => inspectUserDetails(u.id)}
                              title="Инспектировать"
                              className="w-7 h-7 flex items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-400 hover:bg-indigo-500/20 border border-indigo-500/30 transition-all"
                            >
                              <FileSearch className="w-3.5 h-3.5" />
                            </button>
                            <button
                              onClick={() => setBlockConfirmUser(u)}
                              title={u.is_blocked ? 'Разблокировать' : 'Заблокировать'}
                              className={`w-7 h-7 flex items-center justify-center rounded-lg transition-all ${
                                u.is_blocked
                                  ? 'bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 border border-emerald-500/30'
                                  : 'bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/30'
                              }`}
                            >
                              <Ban className="w-3.5 h-3.5" />
                            </button>
                            <button
                              onClick={() => { setSelectedUser(u); setSelectedGrade(''); setSelectedStatus(''); }}
                              className="text-blue-400 hover:text-blue-300 transition-colors"
                            >
                              <ChevronRight className="w-4 h-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                    {users.length === 0 && (
                      <tr><td colSpan={9} className="px-4 py-8 text-center text-slate-500">Нет данных</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Block confirmation modal */}
            {blockConfirmUser && (
              <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setBlockConfirmUser(null)}>
                <div className="bg-slate-900 border border-white/10 rounded-2xl w-full max-w-sm p-6 shadow-2xl space-y-4" onClick={e => e.stopPropagation()}>
                  <div className="flex items-center gap-3">
                    <div className={`p-2.5 rounded-xl ${blockConfirmUser.is_blocked ? 'bg-emerald-500/20' : 'bg-red-500/20'}`}>
                      <Ban className={`w-5 h-5 ${blockConfirmUser.is_blocked ? 'text-emerald-400' : 'text-red-400'}`} />
                    </div>
                    <h4 className="text-white font-semibold text-sm">
                      {blockConfirmUser.is_blocked ? 'Разблокировать' : 'Заблокировать'} пользователя?
                    </h4>
                  </div>
                  <p className="text-sm text-slate-400">
                    Вы уверены, что хотите {blockConfirmUser.is_blocked ? 'разблокировать' : 'заблокировать'} пользователя <strong className="text-white">{blockConfirmUser.email}</strong>?
                    {!blockConfirmUser.is_blocked && <span className="block mt-1 text-red-400/80 text-xs">Пользователь немедленно потеряет доступ ко всем API.</span>}
                  </p>
                  <div className="flex gap-2">
                    <button
                      onClick={() => toggleBlock(blockConfirmUser)}
                      disabled={saving}
                      className={`flex-1 py-2.5 rounded-xl text-sm font-medium transition-colors disabled:opacity-50 ${
                        blockConfirmUser.is_blocked
                          ? 'bg-emerald-500 hover:bg-emerald-600 text-white'
                          : 'bg-red-500 hover:bg-red-600 text-white'
                      }`}
                    >
                      {saving ? '...' : blockConfirmUser.is_blocked ? 'Разблокировать' : 'Заблокировать'}
                    </button>
                    <button
                      onClick={() => setBlockConfirmUser(null)}
                      className="px-5 py-2.5 bg-white/10 hover:bg-white/20 text-slate-300 rounded-xl text-sm transition-colors"
                    >
                      Отмена
                    </button>
                  </div>
                </div>
              </div>
            )}

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
                  {/* Admin Role (super-admin only) */}
                  {isSuperAdmin && selectedUser && (
                    <div className="pt-2 border-t border-white/10">
                      <button
                        onClick={() => handleToggleRole(selectedUser)}
                        disabled={saving}
                        className={`w-full flex items-center justify-center gap-2 py-2.5 rounded-xl text-xs font-medium transition-all ${
                          selectedUser.is_admin
                            ? 'bg-orange-500/10 hover:bg-orange-500/20 border border-orange-500/30 text-orange-400'
                            : 'bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/30 text-blue-400'
                        }`}
                      >
                        <ShieldCheck className="w-3.5 h-3.5" />
                        {saving ? '...' : selectedUser.is_admin ? 'Снять роль админа' : 'Назначить админом'}
                      </button>
                    </div>
                  )}
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

            {/* ══ Financial Passport Modal ══ */}
            {(inspectUser || inspectLoading) && (
              <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={() => { setInspectUser(null); setInspectLoading(false); }}>
                <div className="bg-[#0c0c18]/95 border border-white/10 rounded-2xl w-full max-w-2xl mx-4 max-h-[90vh] overflow-hidden flex flex-col shadow-2xl" onClick={e => e.stopPropagation()}>
                  {/* Header */}
                  <div className="flex items-center justify-between px-6 py-4 border-b border-white/10">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded-xl bg-indigo-500/20">
                        <FileSearch className="w-5 h-5 text-indigo-400" />
                      </div>
                      <div>
                        <h3 className="text-base font-bold text-white">Финансовый паспорт</h3>
                        {inspectUser && <p className="text-xs text-slate-400">{inspectUser.email} · #{inspectUser.id}</p>}
                      </div>
                    </div>
                    <button onClick={() => { setInspectUser(null); setInspectLoading(false); }} className="text-slate-400 hover:text-white transition-colors"><X className="w-5 h-5" /></button>
                  </div>

                  {inspectLoading ? (
                    <div className="flex items-center justify-center py-16"><Loader2 className="w-6 h-6 text-indigo-400 animate-spin" /></div>
                  ) : inspectUser ? (
                    <div className="overflow-y-auto flex-1 px-6 py-4 space-y-5">
                      {/* Section 1: Communication Summary */}
                      <div className="space-y-2">
                        <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-2"><Send className="w-3.5 h-3.5" /> Сводка связи</h4>
                        <div className="grid grid-cols-2 gap-3">
                          <div className="bg-white/[0.04] border border-white/[0.08] rounded-xl p-3">
                            <p className="text-[10px] text-slate-500 mb-1">Telegram</p>
                            {inspectUser.is_telegram_linked ? (
                              <p className="text-sm text-blue-400 font-medium">ID: {inspectUser.telegram_chat_id}</p>
                            ) : (
                              <p className="text-sm text-slate-600">Не привязан</p>
                            )}
                          </div>
                          <div className="bg-white/[0.04] border border-white/[0.08] rounded-xl p-3">
                            <p className="text-[10px] text-slate-500 mb-1">Канал уведомлений</p>
                            <p className="text-sm text-white font-medium">
                              {inspectUser.notification_pref === 'both' ? '📩 TG + Email' : inspectUser.notification_pref === 'telegram' ? '📩 Telegram' : '📧 Email'}
                            </p>
                          </div>
                          <div className="bg-white/[0.04] border border-white/[0.08] rounded-xl p-3">
                            <p className="text-[10px] text-slate-500 mb-1">Кошелёк (USD)</p>
                            <p className="text-sm text-emerald-400 font-bold">${inspectUser.wallet_balance}</p>
                          </div>
                          <div className="bg-white/[0.04] border border-white/[0.08] rounded-xl p-3">
                            <p className="text-[10px] text-slate-500 mb-1">Статус</p>
                            <p className={`text-sm font-medium ${inspectUser.is_blocked ? 'text-red-400' : 'text-emerald-400'}`}>
                              {inspectUser.is_blocked ? 'Заблокирован' : inspectUser.status}
                            </p>
                          </div>
                        </div>
                      </div>

                      {/* Section 2: Card Balances */}
                      <div className="space-y-2">
                        <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-2"><CreditCard className="w-3.5 h-3.5" /> Балансы карт ({inspectUser.cards.length})</h4>
                        {inspectUser.cards.length === 0 ? (
                          <p className="text-sm text-slate-600 py-2">Нет карт</p>
                        ) : (
                          <div className="bg-white/[0.03] border border-white/[0.08] rounded-xl overflow-hidden">
                            <table className="w-full text-xs">
                              <thead>
                                <tr className="border-b border-white/10">
                                  <th className="text-left px-3 py-2 text-slate-500">Карта</th>
                                  <th className="text-left px-3 py-2 text-slate-500">Баланс</th>
                                  <th className="text-left px-3 py-2 text-slate-500">Статус</th>
                                  <th className="text-left px-3 py-2 text-slate-500">Тип</th>
                                </tr>
                              </thead>
                              <tbody>
                                {inspectUser.cards.map(c => (
                                  <tr key={c.id} className="border-b border-white/5">
                                    <td className="px-3 py-2 text-white font-mono">•••• {c.last_4}</td>
                                    <td className="px-3 py-2 text-emerald-400 font-medium">${c.card_balance}</td>
                                    <td className="px-3 py-2">
                                      <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${
                                        c.card_status === 'ACTIVE' ? 'bg-emerald-500/20 text-emerald-400' :
                                        c.card_status === 'FROZEN' ? 'bg-blue-500/20 text-blue-400' :
                                        c.card_status === 'BLOCKED' ? 'bg-red-500/20 text-red-400' :
                                        'bg-slate-500/20 text-slate-400'
                                      }`}>{c.card_status}</span>
                                    </td>
                                    <td className="px-3 py-2 text-slate-400">{c.card_type} · {c.category}</td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          </div>
                        )}
                      </div>

                      {/* Section 3: Transaction History */}
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-2"><Activity className="w-3.5 h-3.5" /> История операций ({inspectUser.transactions.length})</h4>
                          <div className="flex items-center gap-1">
                            <Filter className="w-3 h-3 text-slate-500" />
                            <select
                              value={txFilter}
                              onChange={e => setTxFilter(e.target.value)}
                              className="bg-white/[0.06] border border-white/10 rounded-lg px-2 py-1 text-[10px] text-slate-300 outline-none"
                            >
                              <option value="">Все</option>
                              <option value="FUND">Пополнение</option>
                              <option value="CAPTURE">Списание</option>
                              <option value="FEE">Комиссия</option>
                              <option value="CARD_REFUND">Возврат</option>
                            </select>
                          </div>
                        </div>
                        {(() => {
                          const filtered = txFilter
                            ? inspectUser.transactions.filter(t => t.transaction_type === txFilter)
                            : inspectUser.transactions;
                          return filtered.length === 0 ? (
                            <p className="text-sm text-slate-600 py-2">Нет операций{txFilter ? ' по фильтру' : ''}</p>
                          ) : (
                            <div className="bg-white/[0.03] border border-white/[0.08] rounded-xl overflow-hidden max-h-[300px] overflow-y-auto">
                              <table className="w-full text-xs">
                                <thead className="sticky top-0 bg-[#0c0c18]">
                                  <tr className="border-b border-white/10">
                                    <th className="text-left px-3 py-2 text-slate-500">Дата</th>
                                    <th className="text-left px-3 py-2 text-slate-500">Тип</th>
                                    <th className="text-left px-3 py-2 text-slate-500">Сумма</th>
                                    <th className="text-left px-3 py-2 text-slate-500">Карта</th>
                                    <th className="text-left px-3 py-2 text-slate-500">Детали</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {filtered.slice(0, 100).map(tx => (
                                    <tr key={tx.id} className="border-b border-white/5">
                                      <td className="px-3 py-2 text-slate-500 whitespace-nowrap">{tx.executed_at?.slice(0, 10)}</td>
                                      <td className="px-3 py-2">
                                        <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${
                                          tx.transaction_type === 'FUND' ? 'bg-emerald-500/20 text-emerald-400' :
                                          tx.transaction_type === 'CAPTURE' ? 'bg-orange-500/20 text-orange-400' :
                                          tx.transaction_type === 'FEE' ? 'bg-purple-500/20 text-purple-400' :
                                          'bg-slate-500/20 text-slate-400'
                                        }`}>{tx.transaction_type}</span>
                                      </td>
                                      <td className="px-3 py-2 text-white font-medium">{tx.currency === 'USD' ? '$' : tx.currency}{tx.amount}</td>
                                      <td className="px-3 py-2 text-slate-400 font-mono">{tx.card_last_4 ? `•••• ${tx.card_last_4}` : '—'}</td>
                                      <td className="px-3 py-2 text-slate-500 max-w-[150px] truncate" title={tx.details}>{tx.details || '—'}</td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                            </div>
                          );
                        })()}
                      </div>
                    </div>
                  ) : null}
                </div>
              </div>
            )}
          </div>
        )}

        {/* ════════════ COMMISSIONS TAB ════════════ */}
        {tab === 'commissions' && (
          <div className="space-y-6">
            {/* ── Section 1: Настройки тиров (system_settings) ── */}
            <div>
              <h3 className="text-sm font-semibold text-slate-300 mb-3 flex items-center gap-2">
                <Shield className="w-4 h-4 text-yellow-400" />
                Настройки тиров и комиссий
              </h3>
              <div className="glass-card overflow-hidden">
                <div className="overflow-x-auto">
                <table className="w-full text-sm min-w-[500px]">
                  <thead>
                    <tr className="border-b border-white/10">
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Параметр</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Описание</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Значение</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {sysSettings
                      .filter(s => ['fee_standard', 'fee_gold', 'gold_tier_price', 'gold_tier_duration_days'].includes(s.key))
                      .map(s => (
                        <tr key={s.key} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                          <td className="px-4 py-3 text-white text-xs font-medium break-words">{SETTING_LABELS[s.key] || s.key}</td>
                          <td className="px-4 py-3 text-slate-400 text-xs break-words max-w-[200px]">{s.description}</td>
                          <td className="px-4 py-3">
                            {editingSetting?.key === s.key ? (
                              <input
                                type="text"
                                value={editingSetting.value}
                                onChange={e => setEditingSetting({ ...editingSetting, value: e.target.value })}
                                className="w-28 px-2 py-1 bg-white/10 border border-blue-500/50 rounded text-white text-sm outline-none"
                              />
                            ) : (
                              <span className="text-emerald-400 font-medium">{s.value}</span>
                            )}
                          </td>
                          <td className="px-4 py-3">
                            {editingSetting?.key === s.key ? (
                              <div className="flex gap-1">
                                <button onClick={handleSaveSetting} disabled={saving} className="px-3 py-1 bg-emerald-500 hover:bg-emerald-600 text-white rounded text-xs transition-colors disabled:opacity-50">
                                  {saving ? '...' : 'Сохранить'}
                                </button>
                                <button onClick={() => setEditingSetting(null)} className="px-3 py-1 bg-white/10 hover:bg-white/20 text-slate-300 rounded text-xs transition-colors">
                                  ✕
                                </button>
                              </div>
                            ) : (
                              <button
                                onClick={() => setEditingSetting({ key: s.key, value: s.value })}
                                className="text-blue-400 hover:text-blue-300 text-xs transition-colors"
                              >
                                Изменить
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    {sysSettings.filter(s => ['fee_standard', 'fee_gold', 'gold_tier_price', 'gold_tier_duration_days'].includes(s.key)).length === 0 && (
                      <tr><td colSpan={4} className="px-4 py-8 text-center text-slate-500">Нет данных</td></tr>
                    )}
                  </tbody>
                </table>
                </div>
              </div>
            </div>

            {/* ── Section 2: Прочие комиссии (commission_config) ── */}
            <div>
              <h3 className="text-sm font-semibold text-slate-300 mb-3 flex items-center gap-2">
                <Settings className="w-4 h-4 text-blue-400" />
                Прочие комиссии
              </h3>
              <div className="glass-card overflow-hidden">
                <div className="overflow-x-auto">
                <table className="w-full text-sm min-w-[500px]">
                  <thead>
                    <tr className="border-b border-white/10">
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Параметр</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Описание</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium">Значение</th>
                      <th className="text-left px-4 py-3 text-slate-400 font-medium"></th>
                    </tr>
                  </thead>
                  <tbody>
                    {commissions
                      .filter(c => !['fee_standard', 'fee_gold', 'gold_tier_price', 'gold_tier_duration_days'].includes(c.key))
                      .filter((c, i, arr) => arr.findIndex(x => x.key === c.key) === i)
                      .map(c => (
                      <tr key={c.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                        <td className="px-4 py-3 text-white text-xs font-medium break-words max-w-[200px]">{c.description || c.key}</td>
                        <td className="px-4 py-3 text-slate-400 font-mono text-xs break-words">{c.key}</td>
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
                                {saving ? '...' : 'Сохранить'}
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
            </div>
          </div>
        )}

        {/* ════════════ TICKETS TAB ════════════ */}
        {tab === 'tickets' && (() => {
          // Normalize both sources into one unified list
          const normalizeStatus = (s: string) => {
            if (s === 'open') return 'open';
            if (s === 'in_progress' || s === 'claimed') return 'in_progress';
            if (s === 'resolved') return 'closed';
            if (s === 'closed') return 'closed';
            return s;
          };
          const statusLabel = (s: string) => s === 'open' ? 'Открыт' : s === 'in_progress' ? 'В работе' : 'Закрыт';
          const statusBadge = (s: string) =>
            s === 'open' ? 'bg-emerald-500/20 text-emerald-400'
            : s === 'in_progress' ? 'bg-orange-500/20 text-orange-400'
            : 'bg-slate-500/20 text-slate-400';

          type UnifiedItem = {
            uid: string; id: number; source: 'ticket' | 'chat';
            status: string; topic: string; client: string;
            claimer: string; claimedById: number; date: string; msgCount?: number;
          };

          const unified: UnifiedItem[] = [
            ...tickets.map(t => ({
              uid: `t-${t.id}`, id: t.id, source: 'ticket' as const,
              status: normalizeStatus(t.status),
              topic: t.subject || t.message || '—',
              client: t.email, claimer: '', claimedById: t.claimed_by || 0,
              date: t.created_at || '',
            })),
            ...liveChats.map(c => ({
              uid: `c-${c.id}`, id: c.id, source: 'chat' as const,
              status: normalizeStatus(c.claimed_by > 0 && c.status === 'open' ? 'in_progress' : c.status),
              topic: c.topic || 'Живой чат',
              client: c.user_email, claimer: c.claimer_email || '', claimedById: c.claimed_by || 0,
              date: c.updated_at || c.created_at || '',
              msgCount: c.message_count,
            })),
          ].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());

          const filtered = ticketFilter ? unified.filter(u => u.status === ticketFilter) : unified;

          return (
          <div className="space-y-4">
            {/* Single unified filter bar */}
            <div className="flex gap-2 flex-wrap">
              {[
                { value: '', label: 'Все' },
                { value: 'open', label: 'Открытые' },
                { value: 'in_progress', label: 'В работе' },
                { value: 'closed', label: 'Закрытые' },
              ].map(f => (
                <button
                  key={f.value}
                  onClick={() => setTicketFilter(f.value)}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-all ${
                    ticketFilter === f.value
                      ? 'border-blue-500 bg-blue-500/20 text-blue-400'
                      : 'border-white/10 bg-white/5 text-slate-400 hover:bg-white/10'
                  }`}
                >
                  {f.label}
                </button>
              ))}
            </div>

            {/* Unified ticket list */}
            <div className="space-y-3">
              {filtered.map(item => (
                <div key={item.uid} className="glass-card p-5">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      {/* Row 1: ID | Status | Topic | Msg count */}
                      <div className="flex items-center gap-2 mb-1.5 flex-wrap">
                        <span className="text-xs text-slate-500 font-mono">#{item.source === 'chat' ? `C${item.id}` : item.id}</span>
                        <span className={`px-2 py-0.5 rounded-full text-[10px] font-medium ${statusBadge(item.status)}`}>
                          {statusLabel(item.status)}
                        </span>
                        <span className="text-xs text-slate-500 truncate max-w-[220px]">{item.topic}</span>
                        {item.source === 'chat' && item.msgCount !== undefined && (
                          <span className="text-[10px] text-slate-600 bg-white/5 px-1.5 py-0.5 rounded">{item.msgCount} сообщ.</span>
                        )}
                        {item.source === 'chat' && (
                          <span className="text-[10px] text-blue-500/60 bg-blue-500/10 px-1.5 py-0.5 rounded">чат</span>
                        )}
                      </div>
                      {/* Row 2: Client | Claimer | Date */}
                      <div className="flex items-center gap-3 flex-wrap">
                        <p className="text-sm text-white font-medium">{item.client}</p>
                        {item.claimer ? (
                          <span className="text-xs text-orange-400">← {item.claimer}</span>
                        ) : item.status === 'open' ? (
                          <span className="text-xs text-slate-600">Не назначен</span>
                        ) : null}
                        <span className="text-[10px] text-slate-600 ml-auto shrink-0">{item.date ? new Date(item.date).toLocaleString('ru-RU') : ''}</span>
                      </div>
                    </div>

                    {/* Action buttons */}
                    <div className="flex items-center gap-1.5 shrink-0 self-center">
                      {item.source === 'chat' && (
                        <button
                          onClick={() => viewChatMessages(item.id)}
                          title="История"
                          className="w-8 h-8 flex items-center justify-center bg-white/5 hover:bg-white/10 border border-white/10 text-slate-400 hover:text-white rounded-lg transition-all"
                        >
                          <Eye className="w-3.5 h-3.5" />
                        </button>
                      )}
                      {item.source === 'ticket' && item.status !== 'in_progress' && item.status !== 'closed' && (
                        <button
                          onClick={() => updateTicketStatus(item.id, 'in_progress')}
                          className="px-2.5 py-1.5 bg-orange-500/10 hover:bg-orange-500/20 border border-orange-500/30 text-orange-400 rounded-lg text-[11px] font-medium transition-all"
                        >
                          В работу
                        </button>
                      )}
                      {item.source === 'ticket' && item.status !== 'closed' && (() => {
                        const ownedByOther = item.claimedById > 0 && item.claimedById !== Number(currentUser?.id) && !isSuperAdmin;
                        return (
                          <>
                            <div className="relative group">
                              <button
                                onClick={() => !ownedByOther && updateTicketStatus(item.id, 'resolved')}
                                disabled={ownedByOther}
                                className={`px-2.5 py-1.5 border rounded-lg text-[11px] font-medium transition-all ${
                                  ownedByOther
                                    ? 'bg-emerald-500/5 border-emerald-500/10 text-emerald-400/40 cursor-not-allowed'
                                    : 'bg-emerald-500/10 hover:bg-emerald-500/20 border-emerald-500/30 text-emerald-400'
                                }`}
                              >
                                Решено
                              </button>
                              {ownedByOther && (
                                <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 bg-slate-800 border border-white/10 text-[10px] text-slate-300 rounded-lg whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                  Этот тикет в работе у другого администратора
                                </div>
                              )}
                            </div>
                            <div className="relative group">
                              <button
                                onClick={() => !ownedByOther && updateTicketStatus(item.id, 'closed')}
                                disabled={ownedByOther}
                                className={`px-2.5 py-1.5 border rounded-lg text-[11px] font-medium transition-all ${
                                  ownedByOther
                                    ? 'bg-slate-500/5 border-slate-500/10 text-slate-400/40 cursor-not-allowed'
                                    : 'bg-slate-500/10 hover:bg-slate-500/20 border-slate-500/30 text-slate-400'
                                }`}
                              >
                                Закрыть
                              </button>
                              {ownedByOther && (
                                <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 bg-slate-800 border border-white/10 text-[10px] text-slate-300 rounded-lg whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none z-10">
                                  Этот тикет в работе у другого администратора
                                </div>
                              )}
                            </div>
                          </>
                        );
                      })()}
                    </div>
                  </div>
                </div>
              ))}
              {filtered.length === 0 && (
                <div className="glass-card p-8 text-center text-slate-500 text-sm">
                  Нет тикетов{ticketFilter ? ` со статусом «${ticketFilter === 'open' ? 'Открытые' : ticketFilter === 'in_progress' ? 'В работе' : 'Закрытые'}»` : ''}
                </div>
              )}
            </div>

            {/* ── Chat Messages Modal ── */}
            {viewingChatId && (
              <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setViewingChatId(null)}>
                <div className="bg-slate-900 border border-white/10 rounded-2xl w-full max-w-lg max-h-[80vh] flex flex-col shadow-2xl" onClick={e => e.stopPropagation()}>
                  <div className="flex items-center justify-between p-4 border-b border-white/10">
                    <h4 className="text-white font-semibold text-sm">Чат #{viewingChatId}</h4>
                    <button onClick={() => setViewingChatId(null)} className="text-slate-400 hover:text-white transition-colors text-lg">&times;</button>
                  </div>
                  <div className="flex-1 overflow-y-auto p-4 space-y-3">
                    {chatMessages.map(m => (
                      <div key={m.id} className={`flex ${m.sender_type === 'user' ? 'justify-end' : 'justify-start'}`}>
                        <div className={`max-w-[80%] rounded-2xl px-4 py-2.5 ${
                          m.sender_type === 'user'
                            ? 'bg-blue-500 text-white rounded-br-md'
                            : 'bg-white/[0.06] text-white border border-white/10 rounded-bl-md'
                        }`}>
                          {m.sender_type === 'admin' && (
                            <p className="text-xs text-blue-400 font-medium mb-1">{m.sender_name}</p>
                          )}
                          <p className="text-sm whitespace-pre-wrap break-words">{m.body}</p>
                          <p className="text-[10px] mt-1 text-slate-400">{m.created_at ? new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}</p>
                        </div>
                      </div>
                    ))}
                    {chatMessages.length === 0 && (
                      <p className="text-center text-slate-500 text-sm py-8">Нет сообщений</p>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
          );
        })()}


        {/* ════════════ NEWS TAB ════════════ */}
        {tab === 'news' && (
          <div className="space-y-4">
            {/* Create news form */}
            <div className="glass-card p-5">
              <h4 className="text-white text-sm font-semibold mb-3 flex items-center gap-2"><Plus className="w-4 h-4 text-blue-400" /> Новая публикация</h4>
              <div className="space-y-3">
                <div>
                  <label className="text-xs text-slate-500 mb-1 block">Заголовок</label>
                  <input value={newNewsTitle} onChange={e => setNewNewsTitle(e.target.value)} placeholder="Заголовок новости" className="w-full h-9 px-3 bg-white/5 border border-white/10 rounded-lg text-white text-sm focus:outline-none focus:border-blue-400" />
                </div>
                <div>
                  <label className="text-xs text-slate-500 mb-1 block">Текст (поддерживает переносы строк)</label>
                  <textarea value={newNewsContent} onChange={e => setNewNewsContent(e.target.value)} placeholder="Текст новости..." rows={5} className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white text-sm focus:outline-none focus:border-blue-400 resize-y" />
                </div>
                <div>
                  <label className="text-xs text-slate-500 mb-1 block">Изображение (необязательно, авто-сжатие до 300КБ)</label>
                  <input
                    ref={newsFileRef}
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    className="hidden"
                    onChange={e => handleNewsImageSelect(e.target.files?.[0] || null)}
                  />
                  {newsImagePreview ? (
                    <div className="relative group">
                      <img src={newsImagePreview} alt="Preview" className="w-full max-h-48 object-cover rounded-lg border border-white/10" />
                      <div className="absolute top-2 right-2 flex gap-1">
                        <button onClick={() => newsFileRef.current?.click()} className="p-1.5 bg-black/60 text-white rounded-lg hover:bg-black/80 transition-colors" title="Заменить">
                          <Upload className="w-3.5 h-3.5" />
                        </button>
                        <button onClick={() => { setNewsImageFile(null); setNewsImagePreview(null); }} className="p-1.5 bg-black/60 text-red-400 rounded-lg hover:bg-black/80 transition-colors" title="Удалить">
                          <X className="w-3.5 h-3.5" />
                        </button>
                      </div>
                      {newsImageFile && (
                        <p className="text-[10px] text-slate-500 mt-1">{newsImageFile.name} — {(newsImageFile.size / 1024).toFixed(0)} КБ</p>
                      )}
                    </div>
                  ) : (
                    <div
                      onClick={() => newsFileRef.current?.click()}
                      onDragOver={e => { e.preventDefault(); e.currentTarget.classList.add('border-blue-400'); }}
                      onDragLeave={e => { e.preventDefault(); e.currentTarget.classList.remove('border-blue-400'); }}
                      onDrop={e => { e.preventDefault(); e.currentTarget.classList.remove('border-blue-400'); handleNewsImageSelect(e.dataTransfer.files?.[0] || null); }}
                      className="w-full py-8 border-2 border-dashed border-white/10 rounded-lg flex flex-col items-center justify-center gap-2 cursor-pointer hover:border-blue-400/50 hover:bg-white/[0.02] transition-all"
                    >
                      <ImageIcon className="w-8 h-8 text-slate-600" />
                      <p className="text-xs text-slate-500">Перетащите изображение или нажмите для выбора</p>
                      <p className="text-[10px] text-slate-600">JPG, PNG, WebP · макс. 1200px · авто WebP</p>
                    </div>
                  )}
                </div>
                <button
                  onClick={publishNews}
                  disabled={!newNewsTitle.trim() || !newNewsContent.trim() || newsPublishing}
                  className="px-5 py-2.5 bg-blue-500 hover:bg-blue-600 disabled:opacity-40 text-white rounded-lg text-sm font-medium transition-colors flex items-center gap-2"
                >
                  {newsPublishing ? <><Loader2 className="w-4 h-4 animate-spin" /> {newsUploadProgress || 'Обработка...'}</> : <><Save className="w-4 h-4" /> Сохранить черновик</>}
                </button>
              </div>
            </div>

            {/* Edit news form (shown when editing) */}
            {editingNews && (
              <div className="glass-card p-5 border-l-2 border-blue-500">
                <div className="flex items-center justify-between mb-3">
                  <h4 className="text-white text-sm font-semibold flex items-center gap-2"><Tag className="w-4 h-4 text-blue-400" /> Редактирование #{editingNews.id}</h4>
                  <button onClick={() => setEditingNews(null)} className="text-slate-500 hover:text-white transition-colors"><X className="w-4 h-4" /></button>
                </div>
                <div className="space-y-3">
                  <div>
                    <label className="text-xs text-slate-500 mb-1 block">Заголовок</label>
                    <input value={editingNews.title} onChange={e => setEditingNews({ ...editingNews, title: e.target.value })} className="w-full h-9 px-3 bg-white/5 border border-white/10 rounded-lg text-white text-sm focus:outline-none focus:border-blue-400" />
                  </div>
                  <div>
                    <label className="text-xs text-slate-500 mb-1 block">Текст</label>
                    <textarea value={editingNews.content} onChange={e => setEditingNews({ ...editingNews, content: e.target.value })} rows={4} className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white text-sm focus:outline-none focus:border-blue-400 resize-y" />
                  </div>
                  <div>
                    <label className="text-xs text-slate-500 mb-1 block">Изображение</label>
                    <input ref={editNewsFileRef} type="file" accept="image/jpeg,image/png,image/webp" className="hidden" onChange={e => handleEditNewsImageUpload(e.target.files?.[0] || null)} />
                    {editingNews.image_url ? (
                      <div className="relative group">
                        <img src={editingNews.image_url} alt="" className="w-full max-h-36 object-cover rounded-lg border border-white/10" />
                        <div className="absolute top-2 right-2 flex gap-1">
                          <button onClick={() => editNewsFileRef.current?.click()} className="p-1.5 bg-black/60 text-white rounded-lg hover:bg-black/80 transition-colors" title="Заменить"><Upload className="w-3.5 h-3.5" /></button>
                          <button onClick={() => setEditingNews({ ...editingNews, image_url: '' })} className="p-1.5 bg-black/60 text-red-400 rounded-lg hover:bg-black/80 transition-colors" title="Удалить"><X className="w-3.5 h-3.5" /></button>
                        </div>
                      </div>
                    ) : (
                      <button onClick={() => editNewsFileRef.current?.click()} className="px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-xs text-slate-400 hover:bg-white/10 transition-colors flex items-center gap-1.5">
                        <Upload className="w-3.5 h-3.5" /> Загрузить изображение
                      </button>
                    )}
                  </div>
                  <div className="flex gap-2">
                    <button onClick={saveEditedNews} disabled={saving || !editingNews.title.trim() || !editingNews.content.trim()} className="px-4 py-2 bg-emerald-500 hover:bg-emerald-600 disabled:opacity-40 text-white rounded-lg text-xs font-medium transition-colors flex items-center gap-1.5">
                      {saving ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Save className="w-3.5 h-3.5" />} Сохранить
                    </button>
                    <button
                      onClick={() => toggleNewsStatus(editingNews.id, editingNews.status)}
                      disabled={saving}
                      className={`px-4 py-2 rounded-lg text-xs font-medium transition-colors flex items-center gap-1.5 disabled:opacity-40 ${
                        editingNews.status === 'published'
                          ? 'bg-orange-500/10 border border-orange-500/30 text-orange-400 hover:bg-orange-500/20'
                          : 'bg-blue-500/10 border border-blue-500/30 text-blue-400 hover:bg-blue-500/20'
                      }`}
                    >
                      <Send className="w-3.5 h-3.5" />
                      {editingNews.status === 'published' ? 'Убрать из публикации' : 'Опубликовать'}
                    </button>
                    <button onClick={() => setEditingNews(null)} className="px-4 py-2 bg-white/5 border border-white/10 text-slate-400 rounded-lg text-xs font-medium hover:bg-white/10 transition-colors">Отмена</button>
                  </div>
                </div>
              </div>
            )}

            {/* News list */}
            <div className="glass-card overflow-hidden">
              <div className="overflow-x-auto">
              <table className="w-full text-sm min-w-[500px]">
                <thead>
                  <tr className="border-b border-white/10">
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">ID</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Заголовок</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Статус</th>
                    <th className="text-left px-4 py-3 text-slate-400 font-medium">Дата</th>
                    <th className="text-right px-4 py-3 text-slate-400 font-medium w-32"></th>
                  </tr>
                </thead>
                <tbody>
                  {newsList.map(n => (
                    <tr key={n.id} className={`border-b border-white/5 hover:bg-white/5 transition-colors ${editingNews?.id === n.id ? 'bg-blue-500/5' : ''}`}>
                      <td className="px-4 py-3 text-slate-500 font-mono text-xs">{n.id}</td>
                      <td className="px-4 py-3 text-white text-xs break-words max-w-[200px]">{n.title}</td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded-full text-[10px] font-medium ${
                          n.status === 'published' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-500/20 text-slate-400'
                        }`}>
                          {n.status === 'published' ? 'Опубликовано' : 'Черновик'}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-slate-500 text-xs">{n.created_at ? new Date(n.created_at).toLocaleString('ru-RU') : ''}</td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex gap-1 justify-end">
                          <button
                            onClick={() => toggleNewsStatus(n.id, n.status)}
                            disabled={saving}
                            className={`p-1.5 rounded-lg transition-colors ${
                              n.status === 'published'
                                ? 'bg-orange-500/10 text-orange-400 hover:bg-orange-500/20'
                                : 'bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20'
                            }`}
                            title={n.status === 'published' ? 'Убрать из публикации' : 'Опубликовать'}
                          >
                            <Send className="w-3.5 h-3.5" />
                          </button>
                          <button onClick={() => setEditingNews({ id: n.id, title: n.title, content: n.content, image_url: n.image_url, status: n.status || 'draft' })} className="p-1.5 bg-blue-500/10 text-blue-400 rounded-lg hover:bg-blue-500/20 transition-colors" title="Редактировать">
                            <Tag className="w-3.5 h-3.5" />
                          </button>
                          <button onClick={() => deleteNews(n.id)} className="p-1.5 bg-red-500/10 text-red-400 rounded-lg hover:bg-red-500/20 transition-colors" title="Удалить">
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {newsList.length === 0 && (
                    <tr><td colSpan={5} className="px-4 py-8 text-center text-slate-500 text-sm">Нет новостей</td></tr>
                  )}
                </tbody>
              </table>
              </div>
            </div>
          </div>
        )}

        {/* ════════════ STORE TAB ════════════ */}
        {tab === 'store' && (() => {
          const filtered = storeProducts.filter(p => p.product_type === storeSubTab);
          return (
          <div className="space-y-5">
            {/* Header: sub-tabs + bulk markup */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              <div className="flex gap-2">
                {([['esim', 'eSIM'], ['digital', 'Цифровые']] as const).map(([key, label]) => (
                  <button key={key} onClick={() => { setStoreSubTab(key); setEditingProduct(null); }}
                    className={`px-4 py-2 rounded-xl text-xs font-medium border transition-all ${
                      storeSubTab === key
                        ? 'bg-blue-500/10 border-blue-500/30 text-blue-400'
                        : 'bg-white/5 border-white/10 text-slate-400 hover:bg-white/10'
                    }`}>
                    {label} ({storeProducts.filter(p => p.product_type === key).length})
                  </button>
                ))}
              </div>
              <div className="flex items-center gap-2">
                <input type="number" step="1" value={bulkDelta} onChange={e => setBulkDelta(e.target.value)} className="w-20 px-2 py-2 bg-white/5 border border-white/10 rounded-lg text-white text-xs outline-none focus:border-blue-500/50" placeholder="+%" />
                <button onClick={handleBulkMarkup} disabled={saving} className="px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white rounded-lg text-xs font-medium transition-colors disabled:opacity-50 flex items-center gap-1.5 whitespace-nowrap">
                  <TrendingUp className="w-3.5 h-3.5" />{saving ? '...' : `+${bulkDelta}% ${storeSubTab}`}
                </button>
              </div>
            </div>

            {/* Products table */}
            <div className="glass-card overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
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
                    {filtered.map(p => (
                      <tr key={p.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                        <td className="px-4 py-3 text-slate-500 font-mono text-xs">{p.id}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            {p.image_url ? (
                              <img src={p.image_url} alt="" className="w-6 h-6 rounded object-contain shrink-0 bg-white/5" />
                            ) : (
                              <div className="w-6 h-6 rounded bg-white/5 shrink-0" />
                            )}
                            <span className="text-white text-xs font-medium truncate max-w-[160px]" title={p.name}>{p.name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          {editingProduct?.id === p.id ? (
                            <input type="number" step="0.01" value={editingProduct.cost_price} onChange={e => setEditingProduct({ ...editingProduct, cost_price: e.target.value })} className="w-20 px-2 py-1 bg-white/10 border border-blue-500/50 rounded text-white text-xs outline-none" />
                          ) : (
                            <span className="text-emerald-400 font-medium text-xs">${parseFloat(p.cost_price).toFixed(2)}</span>
                          )}
                        </td>
                        <td className="px-4 py-3">
                          {editingProduct?.id === p.id ? (
                            <input type="number" step="1" value={editingProduct.markup_percent} onChange={e => setEditingProduct({ ...editingProduct, markup_percent: e.target.value })} className="w-16 px-2 py-1 bg-white/10 border border-blue-500/50 rounded text-white text-xs outline-none" />
                          ) : (
                            <span className="text-orange-400 font-medium text-xs">{parseFloat(p.markup_percent).toFixed(0)}%</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-white font-bold text-xs">${parseFloat(p.retail_price).toFixed(2)}</td>
                        <td className="px-4 py-3 text-slate-500 text-xs line-through">${parseFloat(p.old_price).toFixed(2)}</td>
                        <td className="px-4 py-3">
                          <span className={`px-2 py-0.5 rounded-full text-[10px] font-medium ${p.in_stock ? 'bg-emerald-500/20 text-emerald-400' : 'bg-red-500/20 text-red-400'}`}>
                            {p.in_stock ? 'В наличии' : 'Нет'}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          {editingProduct?.id === p.id ? (
                            <div className="flex gap-1">
                              <button onClick={handleSaveProduct} disabled={saving} className="px-2.5 py-1 bg-emerald-500 hover:bg-emerald-600 text-white rounded text-xs transition-colors disabled:opacity-50">
                                {saving ? '...' : 'OK'}
                              </button>
                              <button onClick={() => setEditingProduct(null)} className="px-2.5 py-1 bg-white/10 hover:bg-white/20 text-slate-300 rounded text-xs transition-colors">
                                ✕
                              </button>
                            </div>
                          ) : (
                            <button onClick={() => setEditingProduct({ id: p.id, cost_price: p.cost_price, markup_percent: p.markup_percent, image_url: p.image_url })} className="text-blue-400 hover:text-blue-300 text-xs transition-colors">
                              Изменить
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                    {filtered.length === 0 && (
                      <tr><td colSpan={8} className="px-4 py-8 text-center text-slate-500">Нет товаров</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Image URL editor (shown when editing a product) */}
            {editingProduct && (
              <div className="glass-card p-4">
                <label className="text-xs text-slate-500 mb-1.5 block">Изображение товара (URL логотипа / флага)</label>
                <div className="flex items-center gap-3">
                  {editingProduct.image_url && (
                    <img src={editingProduct.image_url} alt="" className="w-8 h-8 rounded object-contain bg-white/5 border border-white/10 shrink-0" />
                  )}
                  <input
                    type="text"
                    value={editingProduct.image_url}
                    onChange={e => setEditingProduct({ ...editingProduct, image_url: e.target.value })}
                    placeholder="https://cdn.simpleicons.org/steam/white"
                    className="flex-1 px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white text-xs outline-none focus:border-blue-500/50 placeholder-slate-600"
                  />
                </div>
              </div>
            )}
          </div>
          );
        })()}

        {/* ════════════ LOGS TAB ════════════ */}
        {tab === 'logs' && (
          <div className="space-y-4">
            <div className="glass-card overflow-hidden">
              <div className="overflow-x-auto">
              <table className="w-full text-sm min-w-[500px]">
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
                      <td className="px-4 py-3 text-slate-300 text-xs break-all">{l.admin_email}</td>
                      <td className="px-4 py-3 text-white text-xs break-words max-w-[250px]">{l.action}</td>
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
          </div>
        )}
      </div>
    </div>
  );
};
