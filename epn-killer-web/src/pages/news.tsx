import { useState, useEffect, useCallback } from 'react';
import { Newspaper, Bell, BellOff, ChevronLeft, ChevronRight, ImageIcon } from 'lucide-react';
import { DashboardLayout } from '../components/dashboard-layout';
import { getNews, getNewsNotifications, updateNewsNotifications, type NewsItem } from '../api/news';

const PAGE_SIZE = 6;

export const NewsPage = () => {
  const [news, setNews] = useState<NewsItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [notifyEnabled, setNotifyEnabled] = useState(true);
  const [notifyLoading, setNotifyLoading] = useState(false);

  const totalPages = Math.ceil(total / PAGE_SIZE);

  const loadNews = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getNews(PAGE_SIZE, page * PAGE_SIZE);
      setNews(res.items);
      setTotal(res.total);
    } catch {
      /* ignore */
    } finally {
      setLoading(false);
    }
  }, [page]);

  const loadNotifyPref = useCallback(async () => {
    try {
      const res = await getNewsNotifications();
      setNotifyEnabled(res.enabled);
    } catch {
      /* ignore */
    }
  }, []);

  useEffect(() => {
    loadNews();
  }, [loadNews]);

  useEffect(() => {
    loadNotifyPref();
  }, [loadNotifyPref]);

  const toggleNotify = async () => {
    setNotifyLoading(true);
    try {
      await updateNewsNotifications(!notifyEnabled);
      setNotifyEnabled(!notifyEnabled);
    } catch {
      /* ignore */
    } finally {
      setNotifyLoading(false);
    }
  };

  return (
    <DashboardLayout>
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 flex items-center justify-center">
              <Newspaper className="w-5 h-5 text-blue-400" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">Новости</h1>
              <p className="text-xs text-slate-400">{total} {total === 1 ? 'публикация' : total >= 2 && total <= 4 ? 'публикации' : 'публикаций'}</p>
            </div>
          </div>

          {/* Notification toggle */}
          <button
            onClick={toggleNotify}
            disabled={notifyLoading}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-medium border transition-all ${
              notifyEnabled
                ? 'bg-blue-500/10 border-blue-500/30 text-blue-400 hover:bg-blue-500/20'
                : 'bg-white/5 border-white/10 text-slate-400 hover:bg-white/10'
            }`}
          >
            {notifyEnabled ? <Bell className="w-4 h-4" /> : <BellOff className="w-4 h-4" />}
            {notifyEnabled ? 'Уведомления включены' : 'Уведомления выключены'}
          </button>
        </div>

        {/* News cards */}
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="glass-card p-5 animate-pulse">
                <div className="h-4 bg-white/10 rounded w-3/4 mb-3" />
                <div className="h-3 bg-white/5 rounded w-full mb-2" />
                <div className="h-3 bg-white/5 rounded w-2/3" />
              </div>
            ))}
          </div>
        ) : news.length === 0 ? (
          <div className="glass-card p-12 text-center">
            <Newspaper className="w-12 h-12 text-slate-600 mx-auto mb-3" />
            <p className="text-slate-400 text-sm">Новостей пока нет</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {news.map(item => (
              <article
                key={item.id}
                className="glass-card overflow-hidden hover:border-blue-500/20 transition-all group"
              >
                {item.image_url && (
                  <div className="relative h-40 bg-white/5 overflow-hidden">
                    <img
                      src={item.image_url}
                      alt={item.title}
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                    />
                  </div>
                )}
                {!item.image_url && (
                  <div className="h-24 bg-gradient-to-br from-blue-500/10 to-purple-500/10 flex items-center justify-center">
                    <ImageIcon className="w-8 h-8 text-slate-600" />
                  </div>
                )}
                <div className="p-5">
                  <h3 className="text-sm font-bold text-white mb-2 line-clamp-2">{item.title}</h3>
                  <p className="text-xs text-slate-400 whitespace-pre-line line-clamp-4 mb-3">{item.content}</p>
                  <p className="text-[10px] text-slate-600">
                    {new Date(item.created_at).toLocaleDateString('ru-RU', { day: '2-digit', month: 'long', year: 'numeric' })}
                  </p>
                </div>
              </article>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-2">
            <button
              onClick={() => setPage(p => Math.max(0, p - 1))}
              disabled={page === 0}
              className="p-2 rounded-lg bg-white/5 border border-white/10 text-slate-400 hover:bg-white/10 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            {Array.from({ length: totalPages }, (_, i) => (
              <button
                key={i}
                onClick={() => setPage(i)}
                className={`w-8 h-8 rounded-lg text-xs font-medium transition-all ${
                  page === i
                    ? 'bg-blue-500/20 border border-blue-500/30 text-blue-400'
                    : 'bg-white/5 border border-white/10 text-slate-400 hover:bg-white/10'
                }`}
              >
                {i + 1}
              </button>
            ))}
            <button
              onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1}
              className="p-2 rounded-lg bg-white/5 border border-white/10 text-slate-400 hover:bg-white/10 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>
    </DashboardLayout>
  );
};
