import { useState, useEffect } from 'react';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  ArrowUpRight,
  ArrowDownRight,
  Search,
  Wallet,
  CreditCard,
  Clock
} from 'lucide-react';
import apiClient, { API_BASE_URL } from '../api/axios';

type Period = 'day' | 'week' | 'month';

interface HistoryTx {
  id: string;
  description: string;
  amount: number;
  currency: string;
  type: 'income' | 'expense';
  date: string;       // ISO or formatted
  time: string;
  cardLast4: string;  // '' = wallet op
}

const periodLabels: Record<Period, string> = {
  day: 'День',
  week: 'Неделя',
  month: 'Месяц',
};

export const HistoryPage = () => {
  const [period, setPeriod] = useState<Period>('month');
  const [searchQuery, setSearchQuery] = useState('');
  const [transactions, setTransactions] = useState<HistoryTx[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fmt = (d: Date) => d.toISOString().slice(0, 10);

  useEffect(() => {
    fetchTransactions();
  }, [period]);

  const fetchTransactions = async () => {
    setIsLoading(true);
    try {
      const now = new Date();
      const start = new Date(now);
      if (period === 'day') start.setDate(now.getDate() - 1);
      else if (period === 'week') start.setDate(now.getDate() - 7);
      else start.setDate(now.getDate() - 31);

      const res = await apiClient.get(`${API_BASE_URL}/user/transactions`, {
        params: { start_date: fmt(start), end_date: fmt(now), limit: 200 }
      });
      const txs: any[] = res.data?.transactions ?? [];
      const sourceLabels: Record<string, string> = {
        wallet_topup: 'Пополнение кошелька',
        card_transfer: 'Перевод на карту',
        card_charge: 'Списание',
        referral_bonus: 'Реферальный бонус',
        refund: 'Возврат',
        commission: 'Комиссия',
      };
      const mapped: HistoryTx[] = txs.map((tx: any, i: number) => {
        const amt = parseFloat(tx.amount || '0');
        const executedAt = tx.executed_at ? new Date(tx.executed_at) : new Date();
        const srcType = tx.source_type || 'card_charge';
        const desc = tx.details || sourceLabels[srcType] || tx.transaction_type || 'Операция';
        const cur = tx.currency === 'RUB' ? '₽' : tx.currency === 'EUR' ? '€' : '$';
        return {
          id: tx.transaction_id ? String(tx.transaction_id) : String(i),
          description: desc,
          amount: Math.abs(amt),
          currency: cur,
          type: (srcType === 'wallet_topup' || srcType === 'referral_bonus' || srcType === 'refund' || amt > 0) ? 'income' as const : 'expense' as const,
          date: executedAt.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' }),
          time: executedAt.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }),
          cardLast4: tx.card_last_4_digits || '',
        };
      });
      setTransactions(mapped);
    } catch {
      setTransactions([]);
    } finally {
      setIsLoading(false);
    }
  };

  const filtered = transactions.filter(tx => {
    if (searchQuery && !tx.description.toLowerCase().includes(searchQuery.toLowerCase())) return false;
    return true;
  });

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />

        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-white mb-1">История</h1>
            <p className="text-slate-400 text-sm">Все операции по кошельку и картам</p>
          </div>
        </div>

        {/* Period filters */}
        <div className="flex items-center gap-2 mb-4">
          {(Object.keys(periodLabels) as Period[]).map(p => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-4 py-2 rounded-xl text-sm font-medium transition-all ${
                period === p
                  ? 'bg-blue-500/20 text-blue-400 border border-blue-500/40'
                  : 'bg-white/[0.04] text-slate-400 border border-white/[0.08] hover:bg-white/[0.08]'
              }`}
            >
              {periodLabels[p]}
            </button>
          ))}
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400 pointer-events-none" />
          <input
            type="text"
            placeholder="Поиск по операциям..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full h-12 pl-12 pr-4 bg-white/[0.04] border border-white/[0.08] rounded-xl text-white text-sm focus:outline-none focus:border-blue-400 focus:ring-1 focus:ring-blue-400/50 transition-colors placeholder:text-slate-600"
          />
        </div>

        {/* Transaction list */}
        <div className="glass-card overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <div className="w-8 h-8 border-2 border-blue-500/30 border-t-blue-500 rounded-full animate-spin" />
            </div>
          ) : filtered.length > 0 ? (
            <div className="divide-y divide-white/[0.05]">
              {filtered.map(tx => (
                <div key={tx.id} className="flex items-center gap-4 px-5 py-4 hover:bg-white/[0.02] transition-colors">
                  {/* Icon */}
                  <div className={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
                    tx.type === 'income' ? 'bg-emerald-500/15' : 'bg-white/[0.05]'
                  }`}>
                    {tx.type === 'income' ? (
                      <ArrowDownRight className="w-5 h-5 text-emerald-400" />
                    ) : (
                      <ArrowUpRight className="w-5 h-5 text-slate-400" />
                    )}
                  </div>

                  {/* Description + meta */}
                  <div className="flex-1 min-w-0">
                    <p className="text-white font-medium text-sm truncate">{tx.description}</p>
                    <div className="flex items-center gap-2 mt-0.5">
                      {tx.cardLast4 ? (
                        <span className="flex items-center gap-1 text-xs text-slate-500">
                          <CreditCard className="w-3 h-3" />
                          •••• {tx.cardLast4}
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-xs text-slate-500">
                          <Wallet className="w-3 h-3" />
                          Кошелёк
                        </span>
                      )}
                      <span className="text-slate-600">·</span>
                      <span className="flex items-center gap-1 text-xs text-slate-500">
                        <Clock className="w-3 h-3" />
                        {tx.time}
                      </span>
                    </div>
                  </div>

                  {/* Amount + date */}
                  <div className="text-right shrink-0">
                    <p className={`font-semibold text-sm ${tx.type === 'income' ? 'text-emerald-400' : 'text-white'}`}>
                      {tx.type === 'income' ? '+' : '−'}{tx.currency}{tx.amount.toLocaleString()}
                    </p>
                    <p className="text-xs text-slate-500 mt-0.5">{tx.date}</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="py-16 text-center">
              <Search className="w-12 h-12 text-slate-600 mx-auto mb-4" />
              <p className="text-slate-400">Операции не найдены</p>
              <p className="text-sm text-slate-500 mt-1">Попробуйте изменить фильтр или запрос</p>
            </div>
          )}
        </div>

        <p className="text-xs text-slate-500 mt-3 text-center">
          Показано {filtered.length} операций
        </p>
      </div>
    </DashboardLayout>
  );
};
