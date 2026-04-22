import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import apiClient from '../../services/axios';
import {
  Shield,
  Newspaper,
  ShoppingBag,
  Settings,
  Clock,
  MessageSquare,
  Users,
  ToggleLeft,
  ToggleRight,
  ArrowRight,
} from 'lucide-react';

type CommissionRow = {
  id: string;
  key: string;
  value: string;
  description?: string;
};

const QuickCard = ({
  title,
  description,
  to,
  icon,
}: {
  title: string;
  description: string;
  to: string;
  icon: React.ReactNode;
}) => (
  <Link to={to} className="block">
    <div className="glass-card p-6 card-hover cursor-pointer">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-white font-semibold truncate">{title}</p>
          <p className="text-sm text-slate-400 mt-1">{description}</p>
        </div>
        <div className="shrink-0 flex items-center gap-3">
          <div className="w-11 h-11 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center text-slate-300">
            {icon}
          </div>
          <ArrowRight className="w-4 h-4 text-slate-500" />
        </div>
      </div>
    </div>
  </Link>
);

export const AdminDashboardPage = () => {
  const [sbpEnabled, setSbpEnabled] = useState<boolean | null>(null);
  const [sbpLoading, setSbpLoading] = useState(false);
  const [sbpError, setSbpError] = useState('');

  const sbpLabel = useMemo(() => {
    if (sbpEnabled === null) return 'загрузка...';
    return sbpEnabled ? 'включено' : 'выключено';
  }, [sbpEnabled]);

  const loadSBP = async () => {
    setSbpError('');
    try {
      const res = await apiClient.get<CommissionRow[]>('/admin/commissions');
      const row = (res.data || []).find((x) => x.key === 'sbp_topup_enabled');
      const enabled = row ? !(parseFloat(row.value) < 0.5) : true;
      setSbpEnabled(enabled);
    } catch (e: any) {
      setSbpEnabled(null);
      setSbpError('Не удалось загрузить статус СБП');
    }
  };

  useEffect(() => {
    loadSBP();
  }, []);

  const toggleSBP = async () => {
    if (sbpEnabled === null) return;
    const next = !sbpEnabled;
    setSbpLoading(true);
    setSbpError('');
    setSbpEnabled(next);
    try {
      await apiClient.patch('/admin/sbp-topup', { enabled: next });
    } catch {
      setSbpEnabled(!next);
      setSbpError('Не удалось переключить СБП');
    } finally {
      setSbpLoading(false);
    }
  };

  return (
    <div className="stagger-fade-in space-y-6">
      <div className="flex items-center gap-4">
        <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center">
          <Shield className="w-7 h-7 text-blue-300" />
        </div>
        <div className="min-w-0">
          <h1 className="text-2xl md:text-3xl font-bold text-white">Админка</h1>
          <p className="text-slate-400 text-sm">Панель управления и быстрые действия</p>
        </div>
      </div>

      <div className="glass-card p-6">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <p className="text-white font-semibold">Пополнение через СБП</p>
            <p className="text-sm text-slate-400 mt-1">
              Глобальное включение/выключение пополнения (профилактика). Сейчас: <span className="text-slate-200">{sbpLabel}</span>
            </p>
            {sbpError ? <p className="text-sm text-red-400 mt-2">{sbpError}</p> : null}
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
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <QuickCard
          title="Новости"
          description="Создание, редактирование и публикация новостей"
          to="/admin/news"
          icon={<Newspaper className="w-5 h-5" />}
        />
        <QuickCard
          title="Магазин"
          description="Цифровые товары, eSIM, VPN: цены, наценка, наличие"
          to="/admin/store"
          icon={<ShoppingBag className="w-5 h-5" />}
        />
        <QuickCard
          title="Комиссии"
          description="Fee, referral и прочие параметры платформы"
          to="/admin/commissions"
          icon={<Settings className="w-5 h-5" />}
        />
        <QuickCard
          title="System settings"
          description="Ключи/значения настроек системы"
          to="/admin/system-settings"
          icon={<Settings className="w-5 h-5" />}
        />
        <QuickCard
          title="Тикеты"
          description="Заявки поддержки: взять в работу и закрыть"
          to="/admin/tickets"
          icon={<MessageSquare className="w-5 h-5" />}
        />
        <QuickCard
          title="Пользователи"
          description="Поиск, роль админа, бан/разбан, грейды"
          to="/admin/users"
          icon={<Users className="w-5 h-5" />}
        />
        <QuickCard
          title="Логи"
          description="Последние действия админов"
          to="/admin/logs"
          icon={<Clock className="w-5 h-5" />}
        />
      </div>
    </div>
  );
};

