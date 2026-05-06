import { useEffect, useMemo, useState } from 'react';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { listSubscriptions, setBlockedByCard, setSubscriptionBlocked, type CardSubscription } from '../services/subscriptions';
import { getUserCards, type Card as BackendCard } from '../services/cards';
import { Ban, Shield, ToggleLeft, ToggleRight, Search, Loader2 } from 'lucide-react';

type CardLite = { id: string; nickname?: string; last4?: string };

function fmtDate(iso?: string) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short', year: 'numeric' });
}

export const SubscriptionsPage = () => {
  const [items, setItems] = useState<CardSubscription[]>([]);
  const [cards, setCards] = useState<CardLite[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState<string | null>(null);
  const [query, setQuery] = useState('');
  const [selectedCardId, setSelectedCardId] = useState<string>('all');

  const fetchAll = async () => {
    setLoading(true);
    try {
      const [subs, backendCards] = await Promise.all([
        listSubscriptions(),
        getUserCards(),
      ]);
      setItems(subs || []);
      const mapped: CardLite[] = (backendCards || []).map((c: BackendCard) => ({
        id: String(c.id),
        nickname: c.nickname,
        last4: (c.last_4_digits || '').toString(),
      }));
      setCards(mapped);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAll();
  }, []);

  const cardMap = useMemo(() => {
    const m = new Map<string, CardLite>();
    for (const c of cards) m.set(String(c.id), c);
    return m;
  }, [cards]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return items.filter((s) => {
      if (selectedCardId !== 'all' && String(s.cardId) !== String(selectedCardId)) return false;
      if (q && !(s.merchantName || '').toLowerCase().includes(q)) return false;
      return true;
    });
  }, [items, query, selectedCardId]);

  const toggleOne = async (id: string, next: boolean) => {
    setBusy(id);
    try {
      await setSubscriptionBlocked(id, next);
      await fetchAll();
    } finally {
      setBusy(null);
    }
  };

  const blockAllByCard = async (cardId: string, next: boolean) => {
    setBusy(`card:${cardId}`);
    try {
      await setBlockedByCard(cardId, next);
      await fetchAll();
    } finally {
      setBusy(null);
    }
  };

  const uniqueCardIds = useMemo(() => {
    const set = new Set<string>();
    for (const s of items) set.add(String(s.cardId));
    return Array.from(set.values());
  }, [items]);

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />

        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-2xl md:text-3xl font-bold text-white mb-1">Защита от автоподписок</h1>
            <p className="text-slate-400 text-sm">Блокируйте списания конкретных сервисов, не блокируя карту целиком</p>
          </div>
          <div className="flex items-center gap-2">
            <div className="relative">
              <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
              <input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Поиск по сервису"
                className="h-10 pl-9 pr-3 bg-white/[0.04] border border-white/[0.10] rounded-xl text-white text-sm outline-none focus:border-blue-500/40"
              />
            </div>
            <select
              value={selectedCardId}
              onChange={(e) => setSelectedCardId(e.target.value)}
              className="h-10 px-3 bg-white/[0.04] border border-white/[0.10] rounded-xl text-white text-sm outline-none focus:border-blue-500/40"
            >
              <option value="all">Все карты</option>
              {uniqueCardIds.map((id) => {
                const c = cardMap.get(String(id));
                const label = c?.nickname ? `${c.nickname}${c.last4 ? ` •••• ${c.last4}` : ''}` : `•••• ${c?.last4 || id.slice(-4)}`;
                return <option key={id} value={id}>{label}</option>;
              })}
            </select>
          </div>
        </div>

        <div className="glass-card p-0 overflow-hidden">
          <div className="flex items-center gap-3 p-5 border-b border-white/10">
            <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center">
              <Shield className="w-5 h-5 text-amber-400" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-white font-semibold">Мои подписки</p>
              <p className="text-slate-500 text-xs">Подписки появляются автоматически после первого списания</p>
            </div>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="w-6 h-6 text-blue-400 animate-spin" />
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-8 text-center">
              <div className="w-12 h-12 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mx-auto mb-3">
                <Ban className="w-6 h-6 text-slate-500" />
              </div>
              <p className="text-white font-semibold mb-1">Подписки пока не обнаружены</p>
              <p className="text-slate-500 text-sm">Они появятся автоматически после первого рекуррентного списания</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead className="text-xs text-slate-500">
                  <tr className="border-b border-white/5">
                    <th className="py-3 px-5 font-semibold">Сервис</th>
                    <th className="py-3 px-5 font-semibold">Карта</th>
                    <th className="py-3 px-5 font-semibold">Последний платёж</th>
                    <th className="py-3 px-5 font-semibold">Списаний</th>
                    <th className="py-3 px-5 font-semibold">Статус</th>
                    <th className="py-3 px-5 font-semibold text-right">Действие</th>
                  </tr>
                </thead>
                <tbody className="text-sm">
                  {filtered.map((s) => {
                    const c = cardMap.get(String(s.cardId));
                    const cardLabel = c?.nickname ? `${c.nickname}${c.last4 ? ` •••• ${c.last4}` : ''}` : `•••• ${c?.last4 || String(s.cardId).slice(-4)}`;
                    const lastPay = `${s.lastCurrency === 'EUR' ? '€' : '$'}${parseFloat(s.lastAmount || '0').toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} • ${fmtDate(s.lastSeenAt)}`;
                    const isRowBusy = busy === s.id;
                    return (
                      <tr key={s.id} className="border-b border-white/5 hover:bg-white/[0.02]">
                        <td className="py-4 px-5">
                          <div className="text-white font-medium">{s.merchantName}</div>
                          <div className="text-[11px] text-slate-500 truncate max-w-[320px]">{s.merchantKey}</div>
                        </td>
                        <td className="py-4 px-5 text-slate-300">{cardLabel}</td>
                        <td className="py-4 px-5 text-slate-300">{lastPay}</td>
                        <td className="py-4 px-5 text-slate-300">{s.chargeCount}</td>
                        <td className="py-4 px-5">
                          {s.isBlocked ? (
                            <span className="inline-flex items-center gap-2 text-red-400">
                              <Ban className="w-4 h-4" />
                              Заблокировано
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-2 text-emerald-400">
                              <span className="w-2 h-2 rounded-full bg-emerald-400" />
                              Активно
                            </span>
                          )}
                        </td>
                        <td className="py-4 px-5">
                          <div className="flex justify-end items-center gap-2">
                            <button
                              disabled={isRowBusy}
                              onClick={() => toggleOne(s.id, !s.isBlocked)}
                              className={`px-3 py-2 rounded-xl text-xs font-semibold border transition-colors ${
                                s.isBlocked
                                  ? 'bg-white/5 border-white/10 text-slate-200 hover:bg-white/10'
                                  : 'bg-red-500/10 border-red-500/20 text-red-400 hover:bg-red-500/20'
                              } ${isRowBusy ? 'opacity-60 cursor-not-allowed' : ''}`}
                            >
                              {s.isBlocked ? 'Разблокировать' : 'Заблокировать'}
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {!loading && items.length > 0 && selectedCardId !== 'all' && (
            <div className="p-5 border-t border-white/10 flex items-center justify-between gap-3">
              <p className="text-xs text-slate-500">Массовое действие по выбранной карте</p>
              <div className="flex items-center gap-2">
                <button
                  disabled={busy === `card:${selectedCardId}`}
                  onClick={() => blockAllByCard(selectedCardId, true)}
                  className={`px-3 py-2 rounded-xl text-xs font-semibold bg-red-500/10 border border-red-500/20 text-red-400 hover:bg-red-500/20 transition-colors ${
                    busy === `card:${selectedCardId}` ? 'opacity-60 cursor-not-allowed' : ''
                  }`}
                >
                  <span className="inline-flex items-center gap-2">
                    <ToggleLeft className="w-4 h-4" />
                    Заблокировать все
                  </span>
                </button>
                <button
                  disabled={busy === `card:${selectedCardId}`}
                  onClick={() => blockAllByCard(selectedCardId, false)}
                  className={`px-3 py-2 rounded-xl text-xs font-semibold bg-white/5 border border-white/10 text-slate-200 hover:bg-white/10 transition-colors ${
                    busy === `card:${selectedCardId}` ? 'opacity-60 cursor-not-allowed' : ''
                  }`}
                >
                  <span className="inline-flex items-center gap-2">
                    <ToggleRight className="w-4 h-4" />
                    Разблокировать все
                  </span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
};

