import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMode } from '../store/mode-context';
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
  Zap
} from 'lucide-react';
import apiClient, { API_BASE_URL } from '../api/axios';
import { getUserGrade, type GradeInfo } from '../api/grade';

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

const SpendingChart = () => {
  const days = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
  const values = [340, 520, 280, 680, 420, 250, 580];
  const maxValue = Math.max(...values);

  return (
    <div className="glass-card p-6">
      <h3 className="block-title">Расходы за неделю</h3>
      <div className="flex items-end justify-between gap-2 h-48">
        {days.map((day, i) => (
          <div key={day} className="flex-1 flex flex-col items-center gap-2">
            <div className="w-full flex flex-col items-center justify-end h-36">
              <div
                className="w-full max-w-[40px] bg-gradient-to-t from-blue-600 to-blue-400 rounded-lg transition-all duration-500 hover:from-blue-500 hover:to-blue-300 cursor-pointer shadow-lg shadow-blue-500/20"
                style={{ height: `${(values[i] / maxValue) * 100}%` }}
              />
            </div>
            <span className="text-xs text-slate-500">{day}</span>
            <span className="text-xs text-slate-300 font-medium">${values[i]}</span>
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
  const { mode } = useMode();
  const navigate = useNavigate();
  const [userData, setUserData] = useState<any>(null);
  const [gradeInfo, setGradeInfo] = useState<GradeInfo | null>(null);
  const [cardCount, setCardCount] = useState(0);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [userRes] = await Promise.all([
        apiClient.get(`${API_BASE_URL}/user/me`),
      ]);
      setUserData(userRes.data);

      // Non-critical fetches
      try { const g = await getUserGrade(); setGradeInfo(g); } catch {}
      try { const c = await apiClient.get(`${API_BASE_URL}/user/cards`); setCardCount(Array.isArray(c.data) ? c.data.length : 0); } catch {}
      try {
        const t = await apiClient.get(`${API_BASE_URL}/user/report`);
        const txs = t.data?.transactions ?? [];
        setTransactions((txs as any[]).slice(0, 5).map((tx: any, i: number) => ({
          id: String(i),
          description: tx.description || tx.type || 'Операция',
          amount: parseFloat(tx.amount_usd || tx.amount || '0'),
          currency: '$',
          type: parseFloat(tx.amount_usd || tx.amount || '0') >= 0 ? 'income' as const : 'expense' as const,
          date: tx.created_at ? new Date(tx.created_at).toLocaleDateString('ru-RU') : '',
          card: tx.card_last4 ? `•••• ${tx.card_last4}` : 'Кошелёк',
        })));
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

  const balanceArb = Number(userData?.balance_arbitrage ?? userData?.balance ?? 0);
  const balancePers = Number(userData?.balance_personal ?? 0);
  const userName = userData?.email?.split('@')[0] || 'Пользователь';

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        {/* Welcome Card */}
        <div className="glass-card p-6 mb-6 relative overflow-hidden">
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/10 to-purple-500/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
          <div className="relative z-10 flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <Sparkles className="w-5 h-5 text-amber-400" />
                <span className="text-sm text-slate-400">Добро пожаловать!</span>
              </div>
              <h2 className="text-2xl md:text-3xl font-bold welcome-gradient mb-2">
                Привет, {userName}!
              </h2>
              <p className="text-slate-400">
                {mode === 'PERSONAL'
                  ? 'Отличный день для управления финансами'
                  : 'Ваш арбитражный кабинет готов к работе'}
              </p>
            </div>
            <div className="hidden md:block">
              <div className="w-16 h-16 rounded-2xl gradient-accent flex items-center justify-center shadow-lg shadow-blue-500/25">
                {mode === 'PERSONAL' ? <Wallet className="w-8 h-8 text-white" /> : <Target className="w-8 h-8 text-white" />}
              </div>
            </div>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          {mode === 'PERSONAL' ? (
            <>
              <StatCard title="Баланс" value={`$${balancePers.toFixed(2)}`} icon={<Wallet className="w-6 h-6 text-blue-400" />} iconClass="stat-icon-blue" accent />
              <StatCard title="Активные карты" value={String(cardCount)} icon={<CreditCard className="w-6 h-6 text-purple-400" />} iconClass="stat-icon-purple" onClick={() => navigate('/cards')} />
              <StatCard title="Транзакции" value={String(transactions.length)} icon={<DollarSign className="w-6 h-6 text-emerald-400" />} iconClass="stat-icon-green" onClick={() => navigate('/finance')} />
            </>
          ) : (
            <>
              <StatCard title="Баланс" value={`$${balanceArb.toFixed(2)}`} subValue={gradeInfo ? <GradeIndicator grade={gradeInfo.grade} /> : undefined} icon={<Wallet className="w-6 h-6 text-blue-400" />} iconClass="stat-icon-blue" accent />
              <StatCard title="Активные карты" value={String(cardCount)} icon={<CreditCard className="w-6 h-6 text-purple-400" />} iconClass="stat-icon-purple" onClick={() => navigate('/cards')} />
              <StatCard title="Транзакции" value={String(transactions.length)} icon={<DollarSign className="w-6 h-6 text-amber-400" />} iconClass="stat-icon-yellow" onClick={() => navigate('/finance')} />
            </>
          )}
        </div>

        {/* Charts and Transactions */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
          <SpendingChart />

          <div className="glass-card p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="block-title mb-0">Последние операции</h3>
              <button onClick={() => navigate('/finance')} className="text-sm text-blue-400 hover:text-blue-300 transition-colors font-medium">
                Смотреть все →
              </button>
            </div>
            <div>
              {(transactions ?? []).length > 0 ? (
                transactions.map(tx => <TransactionRow key={tx.id} transaction={tx} />)
              ) : (
                <p className="text-slate-500 text-sm text-center py-8">Нет операций</p>
              )}
            </div>
          </div>
        </div>

        {/* Grade Progress - Only for Arbitrage */}
        {mode === 'ARBITRAGE' && <GradeProgressCard gradeInfo={gradeInfo} />}
      </div>
    </DashboardLayout>
  );
};
