import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  Filter,
  Search,
  ArrowUpRight,
  ArrowDownRight,
  Calendar,
  ChevronDown,
  FileSpreadsheet,
  TrendingUp,
  TrendingDown,
  Download
} from 'lucide-react';

interface Transaction {
  id: string;
  date: string;
  description: string;
  amount: number;
  currency: string;
  type: 'income' | 'expense';
  card: string;
  wallet: string;
  merchant: string;
  status: 'completed' | 'pending' | 'failed';
}

const StatusBadge = ({ status }: { status: string }) => {
  const { t } = useTranslation();
  const colors: Record<string, string> = {
    completed: 'badge-success',
    pending: 'badge-warning',
    failed: 'badge-error'
  };

  const labels: Record<string, string> = {
    completed: t('finance.completed'),
    pending: t('finance.pending'),
    failed: t('finance.failed')
  };

  return (
    <span className={`badge ${colors[status]}`}>
      {labels[status]}
    </span>
  );
};

const FilterDropdown = ({ 
  label, 
  options, 
  value, 
  onChange 
}: { 
  label: string; 
  options: string[]; 
  value: string; 
  onChange: (v: string) => void;
}) => (
  <div className="relative">
    <select 
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="xplr-select pr-10"
    >
      <option value="">{label}</option>
      {(options ?? []).map(opt => (
        <option key={opt} value={opt}>{opt}</option>
      ))}
    </select>
  </div>
);

export const FinancePage = () => {
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [cardFilter, setCardFilter] = useState('');
  const [walletFilter, setWalletFilter] = useState('');
  const [showFilters, setShowFilters] = useState(false);

  const transactions: Transaction[] = [
    { id: '1', date: '2024-12-20', description: 'Подписка Netflix', amount: 15.99, currency: '$', type: 'expense', card: '•••• 4521', wallet: 'Личный', merchant: 'Netflix Inc.', status: 'completed' },
    { id: '2', date: '2024-12-19', description: 'Зачисление зарплаты', amount: 4500.00, currency: '$', type: 'income', card: '•••• 4521', wallet: 'Личный', merchant: 'Employer LLC', status: 'completed' },
    { id: '3', date: '2024-12-19', description: 'Покупка на Amazon', amount: 89.50, currency: '$', type: 'expense', card: '•••• 7832', wallet: 'Покупки', merchant: 'Amazon.com', status: 'completed' },
    { id: '4', date: '2024-12-18', description: 'Поездка Uber', amount: 24.30, currency: '$', type: 'expense', card: '•••• 4521', wallet: 'Личный', merchant: 'Uber BV', status: 'completed' },
    { id: '5', date: '2024-12-18', description: 'Оплата за фриланс', amount: 850.00, currency: '$', type: 'income', card: '•••• 0923', wallet: 'Бизнес', merchant: 'Upwork Global', status: 'pending' },
    { id: '6', date: '2024-12-17', description: 'Spotify Premium', amount: 9.99, currency: '$', type: 'expense', card: '•••• 4521', wallet: 'Личный', merchant: 'Spotify AB', status: 'completed' },
    { id: '7', date: '2024-12-17', description: 'Продуктовый магазин', amount: 156.80, currency: '$', type: 'expense', card: '•••• 7832', wallet: 'Покупки', merchant: 'Whole Foods', status: 'completed' },
    { id: '8', date: '2024-12-16', description: 'Неудачная транзакция', amount: 200.00, currency: '$', type: 'expense', card: '•••• 0923', wallet: 'Бизнес', merchant: 'Unknown', status: 'failed' },
    { id: '9', date: '2024-12-16', description: 'Возврат - Amazon', amount: 45.00, currency: '$', type: 'income', card: '•••• 7832', wallet: 'Покупки', merchant: 'Amazon.com', status: 'completed' },
    { id: '10', date: '2024-12-15', description: 'Покупка в Steam', amount: 59.99, currency: '$', type: 'expense', card: '•••• 4521', wallet: 'Личный', merchant: 'Valve Corp.', status: 'completed' },
  ];

  const filteredTransactions = (transactions ?? []).filter(t => {
    if (searchQuery && !t.description.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    if (cardFilter && t.card !== cardFilter) return false;
    if (walletFilter && t.wallet !== walletFilter) return false;
    return true;
  });

  const cards = [...new Set((transactions ?? []).map(t => t.card))];
  const wallets = [...new Set((transactions ?? []).map(t => t.wallet))];

  const totalIncome = filteredTransactions
    .filter(t => t.type === 'income')
    .reduce((sum, t) => sum + t.amount, 0);
  
  const totalExpense = filteredTransactions
    .filter(t => t.type === 'expense')
    .reduce((sum, t) => sum + t.amount, 0);

  const netAmount = totalIncome - totalExpense;

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />
        
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">{t('finance.title')}</h1>
            <p className="text-slate-400">{t('finance.search')}</p>
          </div>
          <div className="flex items-center gap-2 flex-wrap">
            {[
              { label: 'Вчера', days: 1 },
              { label: 'Неделя', days: 7 },
              { label: 'Месяц', days: 30 },
              { label: 'Всё', days: 0 },
            ].map(period => (
              <button
                key={period.label}
                onClick={() => {
                  const now = new Date();
                  const cutoff = period.days > 0
                    ? new Date(now.getTime() - period.days * 86400000).toISOString().split('T')[0]
                    : '';
                  const data = cutoff
                    ? filteredTransactions.filter(t => t.date >= cutoff)
                    : filteredTransactions;
                  const header = 'Дата,Описание,Сумма,Валюта,Тип,Карта,Кошелёк,Мерчант/Получатель,Статус';
                  const rows = data.map(t =>
                    [t.date, `"${t.description}"`, t.amount.toFixed(2), t.currency, t.type === 'income' ? 'Поступление' : 'Списание', t.card, t.wallet, `"${t.merchant}"`, t.status === 'completed' ? 'Выполнено' : t.status === 'pending' ? 'В обработке' : 'Ошибка'].join(',')
                  );
                  const csv = '\uFEFF' + [header, ...rows].join('\n');
                  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
                  const url = URL.createObjectURL(blob);
                  const a = document.createElement('a');
                  a.href = url;
                  a.download = `xplr_report_${period.label.toLowerCase()}.csv`;
                  a.click();
                  URL.revokeObjectURL(url);
                }}
                className="flex items-center gap-2 px-4 py-2.5 glass-card hover:bg-white/10 text-white font-medium rounded-xl transition-all min-h-[44px] text-sm"
              >
                <Download className="w-4 h-4" />
                {period.label}
              </button>
            ))}
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div className="glass-card p-5">
            <div className="flex items-center justify-between mb-2">
              <p className="text-slate-400 text-sm">{t('finance.income')}</p>
              <TrendingUp className="w-4 h-4 text-emerald-400" />
            </div>
            <p className="text-2xl font-bold text-emerald-400 balance-display">+${totalIncome.toLocaleString()}</p>
          </div>
          <div className="glass-card p-5">
            <div className="flex items-center justify-between mb-2">
              <p className="text-slate-400 text-sm">{t('finance.expenses')}</p>
              <TrendingDown className="w-4 h-4 text-red-400" />
            </div>
            <p className="text-2xl font-bold text-white balance-display">-${totalExpense.toLocaleString()}</p>
          </div>
          <div className="glass-card p-5 border-blue-500/30">
            <div className="flex items-center justify-between mb-2">
              <p className="text-slate-400 text-sm">{t('finance.net')}</p>
              {netAmount >= 0 ? <TrendingUp className="w-4 h-4 text-blue-400" /> : <TrendingDown className="w-4 h-4 text-red-400" />}
            </div>
            <p className={`text-2xl font-bold balance-display ${netAmount >= 0 ? 'text-blue-400' : 'text-red-400'}`}>
              {netAmount >= 0 ? '+' : '-'}${Math.abs(netAmount).toLocaleString()}
            </p>
          </div>
        </div>

        {/* Search and Filters */}
        <div className="glass-card p-4 mb-6">
          <div className="flex flex-col md:flex-row gap-4">
            {/* Search */}
            <div className="relative flex-1">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
              <input
                type="text"
                placeholder={t('common.search')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="xplr-input w-full !pl-12"
              />
            </div>
            
            {/* Filter Toggle */}
            <button 
              onClick={() => setShowFilters(!showFilters)}
              className={`flex items-center gap-2 px-4 py-2.5 rounded-xl transition-colors font-medium ${
                showFilters ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30' : 'bg-white/5 text-slate-400 hover:text-white border border-white/10'
              }`}
            >
              <Filter className="w-4 h-4" />
              {t('finance.allCards')}
              <ChevronDown className={`w-4 h-4 transition-transform ${showFilters ? 'rotate-180' : ''}`} />
            </button>
          </div>

          {/* Expanded Filters */}
          {showFilters && (
            <div className="flex flex-wrap gap-4 mt-4 pt-4 border-t border-white/10">
              <div className="flex items-center gap-2">
                <Calendar className="w-4 h-4 text-slate-400" />
                <input type="date" className="xplr-input" />
                <span className="text-slate-500">до</span>
                <input type="date" className="xplr-input" />
              </div>
              <FilterDropdown
                label={t('finance.allCards')}
                options={cards}
                value={cardFilter}
                onChange={setCardFilter}
              />
              <FilterDropdown
                label={t('finance.allWallets')}
                options={wallets}
                value={walletFilter}
                onChange={setWalletFilter}
              />
              <button 
                onClick={() => { setSearchQuery(''); setCardFilter(''); setWalletFilter(''); }}
                className="text-sm text-blue-400 hover:text-blue-300 font-medium"
              >
                Сбросить фильтры
              </button>
            </div>
          )}
        </div>

        {/* Transaction Table */}
        <div className="glass-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="xplr-table min-w-[700px]">
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Описание</th>
                  <th>Сумма</th>
                  <th>Карта</th>
                  <th>Кошелёк</th>
                  <th>Статус</th>
                </tr>
              </thead>
              <tbody>
                {(filteredTransactions ?? []).map(transaction => (
                  <tr key={transaction.id}>
                    <td className="py-4 px-4">
                      <span className="text-slate-400">{transaction.date}</span>
                    </td>
                    <td className="py-4 px-4">
                      <div className="flex items-center gap-3">
                        <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${
                          transaction.type === 'income' ? 'bg-emerald-500/20' : 'bg-white/5'
                        }`}>
                          {transaction.type === 'income' ? (
                            <ArrowDownRight className="w-4 h-4 text-emerald-400" />
                          ) : (
                            <ArrowUpRight className="w-4 h-4 text-slate-400" />
                          )}
                        </div>
                        <span className="text-white font-medium">{transaction.description}</span>
                      </div>
                    </td>
                    <td className="py-4 px-4">
                      <span className={`font-semibold balance-display ${
                        transaction.type === 'income' ? 'text-emerald-400' : 'text-white'
                      }`}>
                        {transaction.type === 'income' ? '+' : '-'}{transaction.currency}{transaction.amount.toLocaleString()}
                      </span>
                    </td>
                    <td className="py-4 px-4">
                      <span className="text-slate-400 font-mono text-sm">{transaction.card}</span>
                    </td>
                    <td className="py-4 px-4">
                      <span className="text-slate-400">{transaction.wallet}</span>
                    </td>
                    <td className="py-4 px-4">
                      <StatusBadge status={transaction.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {(filteredTransactions ?? []).length === 0 && (
            <div className="py-12 text-center">
              <Search className="w-12 h-12 text-slate-600 mx-auto mb-4" />
              <p className="text-slate-400">{t('finance.noTransactions')}</p>
              <p className="text-sm text-slate-500">Попробуйте изменить параметры поиска</p>
            </div>
          )}
        </div>

        {/* Pagination */}
        <div className="flex flex-col md:flex-row items-center justify-between mt-4 gap-4">
          <p className="text-sm text-slate-500">
            Показано {(filteredTransactions ?? []).length} из {(transactions ?? []).length} операций
          </p>
          <div className="flex items-center gap-2 flex-wrap">
            <button className="px-4 py-2.5 glass-card hover:bg-white/10 text-slate-400 rounded-lg transition-colors min-h-[44px]">
              ← Назад
            </button>
            <button className="px-4 py-2.5 bg-blue-500 text-white rounded-lg font-medium min-h-[44px]">1</button>
            <button className="px-4 py-2.5 glass-card hover:bg-white/10 text-slate-400 rounded-lg transition-colors min-h-[44px]">2</button>
            <button className="px-4 py-2.5 glass-card hover:bg-white/10 text-slate-400 rounded-lg transition-colors min-h-[44px]">3</button>
            <button className="px-4 py-2.5 glass-card hover:bg-white/10 text-slate-400 rounded-lg transition-colors min-h-[44px]">
              Далее →
            </button>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
};
