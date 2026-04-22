import { useCallback, useEffect, useMemo, useState } from 'react';
import apiClient from '../../services/axios';
import { Loader2, RefreshCw, Save, X, Plus, Trash2, UploadCloud, Eye, EyeOff } from 'lucide-react';

type NewsItem = {
  id: string;
  title: string;
  content: string;
  imageUrl: string;
  status: 'draft' | 'published' | 'archived' | string;
  createdAt: string;
  updatedAt: string;
};

const StatusPill = ({ s }: { s: string }) => {
  const cls =
    s === 'published'
      ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-200'
      : s === 'draft'
      ? 'bg-blue-500/10 border-blue-500/20 text-blue-200'
      : s === 'archived'
      ? 'bg-orange-500/10 border-orange-500/20 text-orange-200'
      : 'bg-white/5 border-white/10 text-slate-300';
  return <span className={`px-2 py-1 rounded-lg text-[10px] font-semibold border ${cls}`}>{s}</span>;
};

export const AdminNewsPage = () => {
  const [rows, setRows] = useState<NewsItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [creating, setCreating] = useState(false);

  const [newTitle, setNewTitle] = useState('');
  const [newContent, setNewContent] = useState('');
  const [newImageUrl, setNewImageUrl] = useState('');

  const [editing, setEditing] = useState<NewsItem | null>(null);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await apiClient.get<NewsItem[]>('/admin/news', { params: { limit: 200 } });
      setRows(res.data || []);
    } catch {
      setError('Не удалось загрузить новости');
      setRows([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const sorted = useMemo(() => {
    const copy = [...rows];
    copy.sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
    return copy;
  }, [rows]);

  const create = async () => {
    if (!newTitle.trim() || !newContent.trim()) return;
    setCreating(true);
    setError('');
    try {
      await apiClient.post('/admin/news', {
        title: newTitle.trim(),
        content: newContent.trim(),
        image_url: newImageUrl || '',
        status: 'draft',
      });
      setNewTitle('');
      setNewContent('');
      setNewImageUrl('');
      await load();
    } catch {
      setError('Не удалось создать новость');
    } finally {
      setCreating(false);
    }
  };

  const openEdit = (n: NewsItem) => setEditing({ ...n });
  const closeEdit = () => setEditing(null);

  const saveEdit = async () => {
    if (!editing) return;
    if (!editing.title.trim() || !editing.content.trim()) return;
    setSaving(true);
    setError('');
    try {
      await apiClient.put(`/admin/news/${editing.id}`, {
        title: editing.title.trim(),
        content: editing.content.trim(),
        image_url: editing.imageUrl || '',
        status: editing.status,
      });
      setEditing(null);
      await load();
    } catch {
      setError('Не удалось сохранить новость');
    } finally {
      setSaving(false);
    }
  };

  const patchStatus = async (n: NewsItem, next: 'draft' | 'published' | 'archived') => {
    setSaving(true);
    setError('');
    try {
      await apiClient.patch(`/admin/news/${n.id}`, { status: next });
      await load();
    } catch {
      setError('Не удалось изменить статус');
    } finally {
      setSaving(false);
    }
  };

  const del = async (n: NewsItem) => {
    if (!confirm('Удалить новость?')) return;
    setSaving(true);
    setError('');
    try {
      await apiClient.delete(`/admin/news/${n.id}`);
      await load();
    } catch {
      setError('Не удалось удалить новость');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-2xl font-bold text-white">Новости</h1>
          <p className="text-sm text-slate-400 mt-1">Draft/Publish/Archive + редактирование</p>
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
        >
          <RefreshCw className="w-4 h-4" />
          Обновить
        </button>
      </div>

      <div className="glass-card p-6 space-y-4">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          <div>
            <p className="text-white font-semibold">Создать новость</p>
            <p className="text-sm text-slate-400 mt-1">Создаётся как черновик</p>
          </div>
          <button
            onClick={create}
            disabled={creating || !newTitle.trim() || !newContent.trim()}
            className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-sm font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            {creating ? '...' : 'Создать'}
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="md:col-span-2">
            <label className="block text-xs text-slate-500 mb-2">Заголовок</label>
            <input
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
            />
          </div>
          <div className="md:col-span-2">
            <label className="block text-xs text-slate-500 mb-2">Текст</label>
            <textarea
              value={newContent}
              onChange={(e) => setNewContent(e.target.value)}
              rows={5}
              className="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm text-slate-200 outline-none focus:border-blue-500/40"
            />
          </div>
          <div className="md:col-span-2">
            <label className="block text-xs text-slate-500 mb-2">Image URL (опционально)</label>
            <div className="flex items-center gap-2 bg-white/5 border border-white/10 rounded-xl px-3 py-2">
              <UploadCloud className="w-4 h-4 text-slate-500" />
              <input
                value={newImageUrl}
                onChange={(e) => setNewImageUrl(e.target.value)}
                className="w-full bg-transparent outline-none text-sm text-slate-200"
              />
            </div>
          </div>
        </div>

        {error ? <p className="text-sm text-red-400">{error}</p> : null}
      </div>

      <div className="glass-card p-4 sm:p-6 overflow-x-auto">
        <table className="min-w-[1100px] w-full text-left">
          <thead>
            <tr className="text-xs text-slate-500">
              <th className="py-3 px-2 font-semibold">Статус</th>
              <th className="py-3 px-2 font-semibold">Заголовок</th>
              <th className="py-3 px-2 font-semibold">Создано</th>
              <th className="py-3 px-2 font-semibold">Обновлено</th>
              <th className="py-3 px-2 font-semibold text-right">Действия</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={5} className="py-8 text-center text-slate-400">
                  <span className="inline-flex items-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin" /> Загрузка...
                  </span>
                </td>
              </tr>
            ) : sorted.length === 0 ? (
              <tr>
                <td colSpan={5} className="py-10 text-center text-slate-500">
                  Пусто
                </td>
              </tr>
            ) : (
              sorted.map((n) => {
                const created = n.createdAt ? new Date(n.createdAt).toLocaleString('ru-RU') : '';
                const updated = n.updatedAt ? new Date(n.updatedAt).toLocaleString('ru-RU') : '';
                return (
                  <tr key={n.id} className="border-t border-white/5 hover:bg-white/[0.03] transition-colors">
                    <td className="py-3 px-2">
                      <StatusPill s={n.status} />
                    </td>
                    <td className="py-3 px-2 text-sm text-slate-200">
                      <div className="font-semibold">{n.title}</div>
                      <div className="text-xs text-slate-500 font-mono truncate max-w-[620px]">{n.id}</div>
                    </td>
                    <td className="py-3 px-2 text-xs text-slate-500">{created}</td>
                    <td className="py-3 px-2 text-xs text-slate-500">{updated}</td>
                    <td className="py-3 px-2">
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => openEdit(n)}
                          className="px-3 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-xs font-semibold transition-colors"
                        >
                          Редактировать
                        </button>
                        {n.status === 'published' ? (
                          <button
                            onClick={() => patchStatus(n, 'draft')}
                            disabled={saving}
                            className="px-3 py-2 rounded-xl bg-orange-500/15 hover:bg-orange-500/20 border border-orange-500/20 text-orange-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                            title="Снять с публикации"
                          >
                            <EyeOff className="w-4 h-4" />
                            Draft
                          </button>
                        ) : (
                          <button
                            onClick={() => patchStatus(n, 'published')}
                            disabled={saving}
                            className="px-3 py-2 rounded-xl bg-emerald-500/15 hover:bg-emerald-500/20 border border-emerald-500/20 text-emerald-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                            title="Опубликовать"
                          >
                            <Eye className="w-4 h-4" />
                            Publish
                          </button>
                        )}
                        <button
                          onClick={() => del(n)}
                          disabled={saving}
                          className="px-3 py-2 rounded-xl bg-red-500/15 hover:bg-red-500/20 border border-red-500/20 text-red-200 text-xs font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
                          title="Удалить"
                        >
                          <Trash2 className="w-4 h-4" />
                          Del
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {editing ? (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/70 backdrop-blur-sm p-4" onClick={closeEdit}>
          <div className="glass-card w-full max-w-2xl p-6 space-y-4" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <h3 className="text-white font-bold text-lg">Редактирование</h3>
                <p className="text-sm text-slate-400 mt-1 truncate">{editing.id}</p>
              </div>
              <button onClick={closeEdit} className="p-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200">
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-xs text-slate-500 mb-2">Заголовок</label>
                <input
                  value={editing.title}
                  onChange={(e) => setEditing({ ...editing, title: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
                />
              </div>
              <div>
                <label className="block text-xs text-slate-500 mb-2">Текст</label>
                <textarea
                  value={editing.content}
                  onChange={(e) => setEditing({ ...editing, content: e.target.value })}
                  rows={6}
                  className="w-full bg-white/5 border border-white/10 rounded-xl p-3 text-sm text-slate-200 outline-none focus:border-blue-500/40"
                />
              </div>
              <div>
                <label className="block text-xs text-slate-500 mb-2">Image URL</label>
                <input
                  value={editing.imageUrl || ''}
                  onChange={(e) => setEditing({ ...editing, imageUrl: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
                />
              </div>
              <div>
                <label className="block text-xs text-slate-500 mb-2">Status</label>
                <select
                  value={editing.status}
                  onChange={(e) => setEditing({ ...editing, status: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-3 py-2 text-sm text-slate-200 outline-none focus:border-blue-500/40"
                >
                  <option value="draft">draft</option>
                  <option value="published">published</option>
                  <option value="archived">archived</option>
                </select>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2">
              <button
                onClick={closeEdit}
                disabled={saving}
                className="px-4 py-2 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 text-slate-200 text-sm font-medium transition-colors disabled:opacity-50"
              >
                Отмена
              </button>
              <button
                onClick={saveEdit}
                disabled={saving}
                className="px-4 py-2 rounded-xl bg-blue-500/20 hover:bg-blue-500/25 border border-blue-500/30 text-blue-200 text-sm font-semibold transition-colors disabled:opacity-50 inline-flex items-center gap-2"
              >
                <Save className="w-4 h-4" />
                {saving ? '...' : 'Сохранить'}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
};

