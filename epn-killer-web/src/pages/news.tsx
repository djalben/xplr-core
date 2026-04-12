import { useState, useEffect, useCallback } from 'react';
import { Newspaper, Bell, BellOff, ChevronLeft, ChevronRight, ImageIcon, X } from 'lucide-react';
import { DashboardLayout } from '../components/dashboard-layout';
import { getNews, getNewsNotifications, updateNewsNotifications, markNewsAsRead, type NewsItem } from '../api/news';

const PAGE_SIZE = 6;

// ── News Detail Modal ──
const NewsModal = ({ item, onClose }: { item: NewsItem; onClose: () => void }) => {
  // Close on Escape key
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('keydown', handleKey);
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', handleKey);
      document.body.style.overflow = '';
    };
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 sm:p-6" onClick={onClose}>
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" />
      <div
        className="relative w-full max-w-2xl max-h-[85vh] flex flex-col rounded-2xl bg-[#111118] border border-white/10 shadow-2xl"
        onClick={e => e.stopPropagation()}
      >
        {/* Close button — always visible */}
        <button
          onClick={onClose}
          className="absolute top-3 right-3 z-50 p-2 rounded-xl bg-black/70 backdrop-blur-sm border border-white/10 text-slate-400 hover:text-white hover:bg-white/10 transition-all"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Scrollable content */}
        <div className="overflow-y-auto flex-1 rounded-2xl">
        {/* Image — full width, no crop */}
        {item.image_url ? (
          <img
            src={item.image_url}
            alt={item.title}
            className="w-full h-auto rounded-t-2xl"
            onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
          />
        ) : (
          <div className="h-32 bg-gradient-to-br from-blue-500/10 to-purple-500/10 rounded-t-2xl flex items-center justify-center">
            <ImageIcon className="w-10 h-10 text-slate-600" />
          </div>
        )}

        {/* Content */}
        <div className="p-6 sm:p-8">
          <p className="text-[11px] text-slate-500 mb-3">
            {new Date(item.created_at).toLocaleDateString('ru-RU', { day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })}
          </p>
          <h2 className="text-lg sm:text-xl font-bold text-white mb-4 leading-tight">{item.title}</h2>
          <div className="text-sm text-slate-300 leading-relaxed whitespace-pre-line break-words">
            {item.content}
          </div>
        </div>
        </div>{/* /overflow-y-auto */}
      </div>
    </div>
  );
};

export const NewsPage = () => {
  const [news, setNews] = useState<NewsItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [notifyEnabled, setNotifyEnabled] = useState(true);
  const [notifyLoading, setNotifyLoading] = useState(false);
  const [selectedNews, setSelectedNews] = useState<NewsItem | null>(null);

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

  // Mark news as read when page loads
  useEffect(() => {
    markNewsAsRead().catch(() => {});
  }, []);

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
                <div className="h-32 bg-white/5 rounded-lg mb-3" />
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
                onClick={() => setSelectedNews(item)}
                className="glass-card overflow-hidden hover:border-blue-500/20 transition-all group cursor-pointer"
              >
                {item.image_url ? (
                  <div className="relative h-40 bg-white/5 overflow-hidden">
                    <img
                      src={item.image_url}
                      alt={item.title}
                      className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                    />
                  </div>
                ) : (
                  <div className="h-24 bg-gradient-to-br from-blue-500/10 to-purple-500/10 flex items-center justify-center">
                    <ImageIcon className="w-8 h-8 text-slate-600" />
                  </div>
                )}
                <div className="p-5">
                  <h3 className="text-sm font-bold text-white mb-2 line-clamp-2">{item.title}</h3>
                  <p className="text-xs text-slate-400 whitespace-pre-line line-clamp-3 mb-3">{item.content}</p>
                  <div className="flex items-center justify-between">
                    <p className="text-[10px] text-slate-600">
                      {new Date(item.created_at).toLocaleDateString('ru-RU', { day: '2-digit', month: 'long', year: 'numeric' })}
                    </p>
                    <span className="text-[10px] text-blue-400/60 group-hover:text-blue-400 transition-colors">Читать →</span>
                  </div>
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

      {/* News Detail Modal */}
      {selectedNews && <NewsModal item={selectedNews} onClose={() => setSelectedNews(null)} />}
    </DashboardLayout>
  );
};
