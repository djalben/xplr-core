import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { DashboardLayout } from '../components/dashboard-layout';
import {
  TrendingUp,
  TrendingDown,
  CreditCard,
  DollarSign,
  ArrowUpRight,
  ArrowDownRight,
  Wallet,
  Sparkles,
  Target,
  Zap,
  Shield,
  Lock,
  ToggleLeft,
  ToggleRight,
  Info
} from 'lucide-react';
import apiClient from '../api/axios';
import { getUserGrade, type GradeInfo } from '../api/grade';
import { getWallet, type InternalBalance } from '../api/wallet';
import { WorldClocks } from '../components/world-clocks';
import { WalletTopUpModal } from '../components/wallet-topup-modal';

interface StatCardProps {
  title: string;
  value: string;
  subValue?: React.ReactNode;
  change?: number;
  icon: React.ReactNode;
  iconClass?: string;
  accent?: boolean;
  onClick?: () => void;
}

const StatCard = ({ title, value, subValue, change, icon, iconClass = 'stat-icon-blue', accent, onClick }: StatCardProps) => (
  <div className={`glass-card p-6 card-hover ${accent ? 'border-blue-500/30' : ''} ${onClick ? 'cursor-pointer' : ''}`} onClick={onClick}>
    <div className="flex items-start justify-between mb-4">
      <div className={`p-3 rounded-xl ${iconClass}`}>{icon}</div>
      {change !== undefined && (
        <div className={`flex items-center gap-1 text-sm font-medium ${change >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
          {change >= 0 ? <TrendingUp className="w-4 h-4" /> : <TrendingDown className="w-4 h-4" />}
          <span>{Math.abs(change)}%</span>
        </div>
      )}
    </div>
    <p className="text-slate-400 text-sm mb-1">{title}</p>
    <p className="text-2xl font-bold text-white balance-display">{value}</p>
    {subValue && <div className="text-sm text-slate-500 mt-1">{subValue}</div>}
  </div>
);

interface Transaction {
  id: string;
  description: string;
  amount: number;
  currency: string;
  type: 'income' | 'expense';
  date: string;
  card: string;
}

const TransactionRow = ({ transaction }: { transaction: Transaction }) => (
  <div className="flex items-center justify-between py-4 border-b border-white/5 last:border-0 hover:bg-white/[0.02] px-4 -mx-4 transition-colors rounded-lg">
    <div className="flex items-center gap-4">
      <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${
        transaction.type === 'income' ? 'bg-emerald-500/20' : 'bg-white/5'
      }`}>
        {transaction.type === 'income' ? (
          <ArrowDownRight className="w-5 h-5 text-emerald-400" />
        ) : (
          <ArrowUpRight className="w-5 h-5 text-slate-400" />
        )}
      </div>
      <div>
        <p className="text-white font-medium">{transaction.description}</p>
        <p className="text-sm text-slate-500">{transaction.card} • {transaction.date}</p>
      </div>
    </div>
    <p className={`font-semibold balance-display ${transaction.type === 'income' ? 'text-emerald-400' : 'text-white'}`}>
      {transaction.type === 'income' ? '+' : '-'}{transaction.currency}{Math.abs(transaction.amount).toLocaleString()}
    </p>
  </div>
);

const DAY_LABELS = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

const SpendingChart = ({ weeklyData }: { weeklyData: { label: string; amount: number }[] }) => {
  const { t } = useTranslation();
  const maxValue = Math.max(...weeklyData.map(d => d.amount), 1);
  const total = weeklyData.reduce((s, d) => s + d.amount, 0);

  return (
    <div className="glass-card p-6">
      <div className="flex items-center justify-between mb-1">
        <h3 className="block-title">{t('dashboard.weekExpenses')}</h3>
        <span className="text-sm text-slate-400">Итого: <span className="text-white font-semibold">${total.toLocaleString()}</span></span>
      </div>
      <div className="flex items-end justify-between gap-2 h-48">
        {weeklyData.map((day) => (
          <div key={day.label} className="flex-1 flex flex-col items-center gap-2">
            <div className="w-full flex flex-col items-center justify-end h-36">
              <div
                className={`w-full max-w-[40px] rounded-lg transition-all duration-150 shadow-lg ${
                  day.amount > 0
                    ? 'bg-gradient-to-t from-blue-600 to-blue-400 hover:from-blue-500 hover:to-blue-300 shadow-blue-500/20'
                    : 'bg-white/[0.06]'
                }`}
                style={{ height: day.amount > 0 ? `${Math.max((day.amount / maxValue) * 100, 4)}%` : '4%' }}
              />
            </div>
            <span className="text-xs text-slate-500">{day.label}</span>
            <span className="text-xs text-slate-300 font-medium">{day.amount > 0 ? `$${day.amount.toLocaleString()}` : '0'}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

const GradeIndicator = ({ grade }: { grade: string }) => {
  const gradeConfig: Record<string, { color: string; label: string }> = {
    'Standard': { color: 'grade-standard', label: 'Стандарт' },
    'Silver': { color: 'grade-silver', label: 'Серебро' },
    'Gold': { color: 'grade-gold', label: 'Золото' },
    'Platinum': { color: 'grade-platinum', label: 'Платина' },
    'Black': { color: 'grade-black', label: 'Блэк' }
  };
  const config = gradeConfig[grade] || gradeConfig['Standard'];
  return <span className={`grade-badge ${config.color}`}>{config.label}</span>;
};

const GradeProgressCard = ({ gradeInfo }: { gradeInfo: GradeInfo | null }) => {
  const grades = [
    { name: 'Стандарт', commission: '6.7%', color: 'bg-slate-500' },
    { name: 'Серебро', commission: '6.0%', color: 'bg-slate-400' },
    { name: 'Золото', commission: '5.0%', color: 'bg-amber-500' },
    { name: 'Платина', commission: '4.0%', color: 'bg-blue-500' },
    { name: 'Блэк', commission: '3.0%', color: 'bg-slate-900' },
  ];

  const gradeIndex = grades.findIndex(g => g.name === (gradeInfo?.grade || 'Стандарт'));
  const currentIdx = gradeIndex >= 0 ? gradeIndex : 0;
  const totalSpent = parseFloat(gradeInfo?.total_spent || '0');
  const nextSpend = parseFloat(gradeInfo?.next_spend || '1000000');
  const progress = nextSpend > 0 ? Math.min((totalSpent / nextSpend) * 100, 100) : 0;

  return (
    <div className="glass-card p-6">
      <div className="flex items-center justify-between mb-4">
        <h3 className="block-title mb-0">Прогресс грейда</h3>
        <div className="flex items-center gap-2">
          <Zap className="w-4 h-4 text-amber-400" />
          <span className="text-sm text-slate-400">Комиссия: <span className="text-blue-400 font-semibold">{gradeInfo?.fee_percent || '6.7'}%</span></span>
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex-1">
          <div className="flex justify-between text-sm mb-2">
            <GradeIndicator grade={gradeInfo?.grade || 'Standard'} />
            {gradeInfo?.next_grade && <span className="text-slate-500">→ {gradeInfo.next_grade}</span>}
          </div>
          <div className="progress-bar-container">
            <div className="progress-bar-fill progress-bar-blue" style={{ width: `${progress}%` }} />
          </div>
          <p className="text-xs text-slate-500 mt-2">
            Оборот: ${(totalSpent / 1000).toFixed(0)}K
          </p>
        </div>
      </div>
      <div className="mt-6 pt-4 border-t border-white/10">
        <p className="text-xs text-slate-500 mb-3">Шкала комиссий</p>
        <div className="flex items-center gap-1">
          {grades.map((grade, i) => (
            <div key={grade.name} className="flex-1 text-center">
              <div className={`h-1.5 ${grade.color} ${i === 0 ? 'rounded-l-full' : ''} ${i === grades.length - 1 ? 'rounded-r-full' : ''} ${i <= currentIdx ? 'opacity-100' : 'opacity-30'}`} />
              <p className={`text-[10px] mt-1 ${i <= currentIdx ? 'text-slate-300' : 'text-slate-600'}`}>{grade.commission}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export const DashboardPage = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [userData, setUserData] = useState<any>(null);
  const [gradeInfo, setGradeInfo] = useState<GradeInfo | null>(null);
  const [cardCount, setCardCount] = useState(0);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [wallet, setWallet] = useState<InternalBalance | null>(null);
  const [isWalletModalOpen, setIsWalletModalOpen] = useState(false);
  const [autoTopup, setAutoTopup] = useState(false);
  const [showAutoTopupTooltip, setShowAutoTopupTooltip] = useState(false);
  const [weeklySpending, setWeeklySpending] = useState<{ label: string; amount: number }[]>(
    Array.from({ length: 7 }, (_, i) => {
      const d = new Date(); d.setDate(d.getDate() - 6 + i);
      return { label: DAY_LABELS[d.getDay()], amount: 0 };
    })
  );

  useEffect(() => {
    fetchData();
    // Auto-refresh wallet balance every 30s (picks up webhook credits)
    const walletInterval = setInterval(() => {
      getWallet()
        .then((v) => setWallet(v))
        .catch(() => {});
    }, 30000);
    return () => clearInterval(walletInterval);
  }, []);

  const fetchData = async () => {
    try {
      const [userRes] = await Promise.all([
        apiClient.get('/user/me'),
      ]);
      setUserData(userRes.data);

      // Non-critical fetches
      try { const g = await getUserGrade(); setGradeInfo(g); } catch {}
      try { const v = await getWallet(); setWallet(v); } catch {}
      try { const c = await apiClient.get('/user/cards'); setCardCount(Array.isArray(c.data) ? c.data.length : 0); } catch {}
      try {
        const t = await apiClient.get('/user/report', { params: { limit: 3 } });
        const txs = t.data?.transactions ?? [];
        setTransactions((txs as any[]).slice(0, 3).map((tx: any, i: number) => ({
          id: String(i),
          description: tx.description || tx.type || 'Операция',
          amount: parseFloat(tx.amount_usd || tx.amount || '0'),
          currency: '$',
          type: parseFloat(tx.amount_usd || tx.amount || '0') >= 0 ? 'income' as const : 'expense' as const,
          date: tx.created_at ? new Date(tx.created_at).toLocaleDateString('ru-RU') : '',
          card: tx.card_last4 ? `•••• ${tx.card_last4}` : 'Кошелёк',
        })));
      } catch {}

      // Fetch 7-day spending data
      try {
        const now = new Date();
        const startDate = new Date(now);
        startDate.setDate(now.getDate() - 6);
        const fmt = (d: Date) => d.toISOString().slice(0, 10);
        const weekRes = await apiClient.get('/user/report', {
          params: { start_date: fmt(startDate), end_date: fmt(now), limit: 500 }
        });
        const weekTxs: any[] = weekRes.data?.transactions ?? [];
        // Build day-by-day map for last 7 days
        const dayMap: Record<string, number> = {};
        const labels: { key: string; label: string }[] = [];
        for (let i = 0; i < 7; i++) {
          const d = new Date(startDate);
          d.setDate(startDate.getDate() + i);
          const key = fmt(d);
          dayMap[key] = 0;
          labels.push({ key, label: DAY_LABELS[d.getDay()] });
        }
        weekTxs.forEach((tx: any) => {
          const amt = Math.abs(parseFloat(tx.amount_usd || tx.amount || '0'));
          const txDate = tx.created_at ? tx.created_at.slice(0, 10) : tx.executed_at?.slice(0, 10);
          if (txDate && dayMap[txDate] !== undefined) {
            dayMap[txDate] += amt;
          }
        });
        setWeeklySpending(labels.map(l => ({ label: l.label, amount: Math.round(dayMap[l.key]) })));
      } catch {}
    } catch (err) {
      console.error('Dashboard fetch error:', err);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <DashboardLayout>
        <div className="flex items-center justify-center h-64">
          <div className="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin" />
        </div>
      </DashboardLayout>
    );
  }

  const balancePers = Number(userData?.balance_personal ?? 0);
  const userName = userData?.email?.split('@')[0] || 'Пользователь';

  return (
    <DashboardLayout>
      <div>
        {/* Welcome Card */}
        <div className="glass-card p-6 mb-6 relative overflow-hidden min-h-[160px]">
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/10 to-purple-500/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
          <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <Sparkles className="w-5 h-5 text-amber-400" />
                <span className="text-sm text-slate-400">{t('dashboard.welcome')}</span>
              </div>
              <h2 className="text-2xl md:text-3xl font-bold welcome-gradient mb-2">
                {t('dashboard.hello', { name: userName })}
              </h2>
              <p className="text-slate-400">
                {t('dashboard.personalSubtitle')}
              </p>
            </div>
            {/* World Clocks — inline with greeting */}
            <div className="shrink-0">
              <WorldClocks />
            </div>
          </div>
        </div>

        {/* Stats Grid — enhanced widgets */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          {/* Баланс Кошелька + Пополнить */}
          <div className="glass-card p-6 border-blue-500/30">
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 rounded-xl stat-icon-blue"><Wallet className="w-6 h-6 text-blue-400" /></div>
            </div>
            <p className="text-slate-400 text-sm mb-1">Баланс Кошелька</p>
            <p className="text-2xl font-bold text-white balance-display">{wallet ? `$${Number(wallet.master_balance).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}` : `$${balancePers.toFixed(2)}`}</p>
            <button
              onClick={() => setIsWalletModalOpen(true)}
              className="mt-4 w-full py-2.5 bg-gradient-to-r from-amber-500 to-orange-500 text-white font-medium rounded-xl text-sm hover:shadow-lg hover:shadow-amber-500/25 transition-all"
            >
              Пополнить
            </button>
          </div>

          {/* Активные карты + Управление + Автопополнение */}
          <div className="glass-card p-6">
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 rounded-xl stat-icon-purple"><CreditCard className="w-6 h-6 text-purple-400" /></div>
            </div>
            <p className="text-slate-400 text-sm mb-1">{t('dashboard.activeCards')}</p>
            <p className="text-2xl font-bold text-white balance-display">{cardCount}</p>
            <button
              onClick={() => navigate('/cards')}
              className="mt-4 w-full py-2.5 bg-white/5 border border-white/10 text-white font-medium rounded-xl text-sm hover:bg-white/10 transition-all"
            >
              Управление картами
            </button>
            {/* Автопополнение toggle */}
            <div className="mt-3 flex items-center justify-between relative">
              <button
                onClick={() => {
                  const next = !autoTopup;
                  setAutoTopup(next);
                  apiClient.patch('/user/wallet/auto-topup', { enabled: next }).catch(() => setAutoTopup(!next));
                }}
                className="flex items-center gap-2 text-sm"
              >
                {autoTopup
                  ? <ToggleRight className="w-7 h-7 text-emerald-400" />
                  : <ToggleLeft className="w-7 h-7 text-slate-500" />
                }
                <span className={autoTopup ? 'text-emerald-400 font-medium' : 'text-slate-400'}>Автопополнение</span>
              </button>
              <button
                onMouseEnter={() => setShowAutoTopupTooltip(true)}
                onMouseLeave={() => setShowAutoTopupTooltip(false)}
                onClick={() => setShowAutoTopupTooltip(v => !v)}
                className="text-slate-500 hover:text-slate-300 transition-colors"
              >
                <Info className="w-4 h-4" />
              </button>
              {showAutoTopupTooltip && (
                <div className="absolute bottom-full right-0 mb-2 w-64 p-3 bg-[#1a1a24] border border-white/10 rounded-xl shadow-2xl z-50 text-xs text-slate-300 leading-relaxed">
                  При нехватке средств на карте, система автоматически переведёт нужную сумму из Кошелька.
                  <div className="absolute -bottom-1.5 right-4 w-3 h-3 bg-[#1a1a24] border-r border-b border-white/10 transform rotate-45" />
                </div>
              )}
            </div>
          </div>

          {/* Транзакции — clickable → /history, today spending */}
          <div
            className="glass-card p-6 cursor-pointer card-hover"
            onClick={() => navigate('/history')}
          >
            <div className="flex items-start justify-between mb-4">
              <div className="p-3 rounded-xl stat-icon-green"><DollarSign className="w-6 h-6 text-emerald-400" /></div>
            </div>
            <p className="text-slate-400 text-sm mb-1">{t('dashboard.transactions')}</p>
            <p className="text-2xl font-bold text-white balance-display">{transactions.length}</p>
            <p className="text-sm text-slate-400 mt-3">
              Расходы за сегодня: <span className="text-white font-semibold">${(() => {
                const today = new Date().toLocaleDateString('ru-RU');
                return transactions
                  .filter(tx => tx.type === 'expense' && tx.date === today)
                  .reduce((sum, tx) => sum + Math.abs(tx.amount), 0)
                  .toLocaleString();
              })()}</span>
            </p>
          </div>
        </div>

        {/* Charts and Transactions */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          <SpendingChart weeklyData={weeklySpending} />

          <div className="glass-card p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="block-title mb-0">{t('dashboard.recentOps')}</h3>
              <button onClick={() => navigate('/history')} className="text-sm text-blue-400 hover:text-blue-300 transition-colors font-medium">
                {t('dashboard.viewAll')}
              </button>
            </div>
            <div>
              {(transactions ?? []).length > 0 ? (
                transactions.map(tx => <TransactionRow key={tx.id} transaction={tx} />)
              ) : (
                <p className="text-slate-500 text-sm text-center py-8">{t('dashboard.noOps')}</p>
              )}
            </div>
          </div>
        </div>

      </div>
      {isWalletModalOpen && <WalletTopUpModal onClose={() => setIsWalletModalOpen(false)} onSuccess={() => { getWallet().then(v => setWallet(v)).catch(() => {}); }} />}
    </DashboardLayout>
  );
};
