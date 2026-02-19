import React, { useState, useRef, useEffect, useCallback } from 'react';
import { DashboardLayout } from '../components/dashboard-layout';
import { ModalPortal } from '../components/modal-portal';
import { BackButton } from '../components/back-button';
import { useAuth } from '../store/auth-context';
import { 
  UserPlus,
  Crown,
  Shield,
  User,
  Mail,
  Copy,
  Check,
  X,
  Link,
  MessageCircle,
  KanbanSquare,
  Send,
  Paperclip,
  CheckCheck,
  Clock,
  GripVertical,
  Plus,
  ChevronLeft,
  ChevronRight,
  Trash2
} from 'lucide-react';

const LS_TASKS_KEY = 'xplr_kanban_tasks';
const LS_CHAT_KEY = 'xplr_team_chat';

/* ──────────────────────────────── Types ──────────────────────────────── */

interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'member';
  avatar: string;
  status: 'online' | 'offline' | 'away';
  label?: string;
}

interface ChatMessage {
  id: string;
  senderId: string;
  text: string;
  time: string;
  read: boolean;
}

interface KanbanTask {
  id: string;
  title: string;
  description: string;
  assigneeId: string;
  column: 'backlog' | 'progress' | 'done';
}

/* ──────────────────────────────── Data ──────────────────────────────── */

const teamMembers: TeamMember[] = [
  { id: '1', name: 'Алексей Петров', email: 'alex@xplr.io', role: 'owner', avatar: 'АП', status: 'online', label: 'Lead' },
  { id: '2', name: 'Мария Иванова', email: 'maria@xplr.io', role: 'admin', avatar: 'МИ', status: 'online', label: 'Admin' },
  { id: '3', name: 'Дмитрий Козлов', email: 'dmitry@xplr.io', role: 'member', avatar: 'ДК', status: 'away', label: 'Buyer' },
  { id: '4', name: 'Елена Смирнова', email: 'elena@xplr.io', role: 'member', avatar: 'ЕС', status: 'offline', label: 'Buyer' },
];

const initialMessages: ChatMessage[] = [
  { id: 'm1', senderId: '1', text: 'Привет команда! Сегодня запускаем новый оффер — лимит 200 карт.', time: '10:15', read: true },
  { id: 'm2', senderId: '2', text: 'Принято, готовлю крео. Какие гео приоритетные?', time: '10:17', read: true },
  { id: 'm3', senderId: '1', text: 'DE, FR, ES — основные. UK пока на паузе.', time: '10:18', read: true },
  { id: 'm4', senderId: '3', text: 'Понял, начинаю лить на DE. Бюджет на день?', time: '10:22', read: true },
  { id: 'm5', senderId: '1', text: '$5K на тест, если ROI > 30% — масштабируем до $20K', time: '10:24', read: true },
  { id: 'm6', senderId: '4', text: 'Я возьму FR и ES, карты уже готовы.', time: '10:30', read: true },
  { id: 'm7', senderId: '2', text: 'Крео загружены в папку, 4 варианта. Проверьте.', time: '10:45', read: false },
];

const initialTasks: KanbanTask[] = [
  { id: 'k1', title: 'Протестировать оффер DE', description: 'Запустить тестовый трафик $2K', assigneeId: '3', column: 'progress' },
  { id: 'k2', title: 'Подготовить крео v2', description: 'Новые баннеры для FR гео', assigneeId: '2', column: 'progress' },
  { id: 'k3', title: 'Масштабировать ES', description: 'Увеличить бюджет до $10K/день', assigneeId: '4', column: 'backlog' },
  { id: 'k4', title: 'Анализ ROI за неделю', description: 'Сводная таблица по всем гео', assigneeId: '1', column: 'backlog' },
  { id: 'k5', title: 'Настроить API трекер', description: 'Интегрировать postback для конверсий', assigneeId: '2', column: 'backlog' },
  { id: 'k6', title: 'Выпустить 50 карт EUR', description: 'Для нового потока FR', assigneeId: '3', column: 'done' },
  { id: 'k7', title: 'Закрыть старый оффер UK', description: 'Стоп трафик, архивировать данные', assigneeId: '1', column: 'done' },
];

/* ──────────────────────────── Small Components ───────────────────────── */

const RoleBadge = ({ role, label }: { role: string; label?: string }) => {
  const config: Record<string, { icon: React.ReactNode; color: string; text: string }> = {
    owner: { icon: <Crown className="w-3 h-3" />, color: 'text-yellow-400', text: 'Owner' },
    admin: { icon: <Shield className="w-3 h-3" />, color: 'text-blue-400', text: 'Admin' },
    member: { icon: <User className="w-3 h-3" />, color: 'text-slate-400', text: 'Member' },
  };
  const c = config[role] || config.member;
  return (
    <span className={`inline-flex items-center gap-1 text-[10px] font-semibold ${c.color}`}>
      {c.icon}
      {label || c.text}
    </span>
  );
};

const StatusDot = ({ status }: { status: string }) => {
  const c: Record<string, string> = { online: 'bg-emerald-500', away: 'bg-amber-500', offline: 'bg-slate-500' };
  return <div className={`w-2.5 h-2.5 rounded-full ${c[status] || c.offline} ring-2 ring-[#0a0a0f]`} />;
};

const Avatar = ({ member, size = 'md' }: { member: TeamMember; size?: 'sm' | 'md' }) => {
  const s = size === 'sm' ? 'w-7 h-7 text-[10px]' : 'w-10 h-10 text-sm';
  const gradient = member.role === 'owner'
    ? 'bg-gradient-to-br from-red-500 to-orange-600'
    : 'bg-gradient-to-br from-blue-500 to-purple-600';
  return (
    <div className="relative shrink-0">
      <div className={`${s} rounded-xl ${gradient} flex items-center justify-center text-white font-semibold shadow-lg`}>
        {member.avatar}
      </div>
      <div className="absolute -bottom-0.5 -right-0.5"><StatusDot status={member.status} /></div>
    </div>
  );
};

/* ──────────────────────────── Invite Modal ────────────────────────────── */

const InviteModal = ({ onClose }: { onClose: () => void }) => {
  const [copied, setCopied] = useState(false);
  const inviteLink = 'https://xplr.io/team/join/abc123xyz';

  const handleCopy = () => {
    navigator.clipboard.writeText(inviteLink);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <ModalPortal>
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
      <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-[440px] max-h-[90dvh] overflow-y-auto shadow-2xl shadow-black/60 animate-scale-in">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-xl font-semibold text-white">Пригласить участника</h3>
          <button onClick={onClose} className="p-2 hover:bg-white/10 rounded-lg transition-colors">
            <X className="w-5 h-5 text-gray-400" />
          </button>
        </div>
        <div className="space-y-5">
          <div>
            <label className="block text-sm text-gray-400 mb-2">Email адрес</label>
            <div className="relative">
              <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500" />
              <input type="email" placeholder="colleague@company.com" className="xplr-input w-full !pl-12" />
            </div>
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-2">Роль</label>
            <select className="xplr-select w-full">
              <option value="member">Участник</option>
              <option value="admin">Админ</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-gray-400 mb-2">Лимит карт в месяц</label>
            <input type="number" placeholder="50" className="xplr-input w-full" />
          </div>
          <div className="section-divider" />
          <div>
            <label className="block text-sm text-gray-400 mb-2">Или отправьте ссылку-приглашение</label>
            <div className="flex gap-2">
              <div className="flex-1 flex items-center gap-2 bg-white/5 rounded-lg px-3 py-2 border border-white/10">
                <Link className="w-4 h-4 text-gray-500" />
                <span className="text-sm text-gray-400 truncate">{inviteLink}</span>
              </div>
              <button onClick={handleCopy} className={`px-4 py-2 rounded-lg transition-all flex items-center gap-2 ${copied ? 'bg-green-500/20 text-green-400' : 'glass-card hover:bg-white/10 text-gray-400'}`}>
                {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
              </button>
            </div>
          </div>
        </div>
        <div className="flex gap-3 mt-6">
          <button onClick={onClose} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-gray-300 font-medium rounded-xl transition-colors">Отмена</button>
          <button className="flex-1 px-4 py-3 gradient-accent text-white font-medium rounded-xl transition-colors">Отправить приглашение</button>
        </div>
      </div>
    </div>
    </ModalPortal>
  );
};

/* ──────────────────────────── Member List Panel ──────────────────────── */

const MemberListPanel = ({ members, selectedId, onSelect }: { members: TeamMember[]; selectedId: string | null; onSelect: (id: string) => void }) => (
  <div className="flex flex-col h-full">
    <div className="p-3 border-b border-white/[0.06]">
      <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">Участники · {members.length}</p>
    </div>
    <div className="flex-1 overflow-y-auto">
      {members.map(m => (
        <button
          key={m.id}
          onClick={() => onSelect(m.id)}
          className={`w-full flex items-center gap-3 px-3 py-3 transition-colors duration-150 ${
            selectedId === m.id ? 'bg-white/[0.06]' : 'hover:bg-white/[0.03]'
          }`}
        >
          <Avatar member={m} />
          <div className="flex-1 min-w-0 text-left">
            <p className="text-sm font-medium text-white truncate">{m.name}</p>
            <RoleBadge role={m.role} label={m.label} />
          </div>
        </button>
      ))}
    </div>
  </div>
);

/* ──────────────────────────── Chat Panel ─────────────────────────────── */

const ChatPanel = ({ messages, members, onSend }: {
  messages: ChatMessage[];
  members: TeamMember[];
  onSend: (text: string) => void;
}) => {
  const [input, setInput] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);
  const ownerId = '1';

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages.length]);

  const getMember = (id: string) => members.find(m => m.id === id) || members[0];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;
    onSend(input.trim());
    setInput('');
  };

  return (
    <div className="flex flex-col h-full">
      {/* Messages area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.map(msg => {
          const isOwner = msg.senderId === ownerId;
          const sender = getMember(msg.senderId);
          return (
            <div key={msg.id} className={`flex gap-2 ${isOwner ? 'justify-end' : 'justify-start'}`}>
              {!isOwner && <Avatar member={sender} size="sm" />}
              <div className={`max-w-[75%] ${isOwner ? 'order-first' : ''}`}>
                {!isOwner && (
                  <p className="text-[10px] text-slate-500 mb-1 ml-1">{sender.name}</p>
                )}
                <div
                  className={`px-3.5 py-2.5 rounded-2xl text-sm leading-relaxed ${
                    isOwner
                      ? 'bg-gradient-to-r from-red-600 to-orange-500 text-white rounded-br-md'
                      : 'bg-white/[0.06] text-slate-200 border border-white/[0.06] rounded-bl-md'
                  }`}
                >
                  {msg.text}
                </div>
                <div className={`flex items-center gap-1 mt-0.5 ${isOwner ? 'justify-end' : 'justify-start'} px-1`}>
                  <span className="text-[10px] text-slate-600">{msg.time}</span>
                  {isOwner && (
                    <CheckCheck className={`w-3 h-3 ${msg.read ? 'text-blue-400' : 'text-slate-600'}`} />
                  )}
                </div>
              </div>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>

      {/* Input area */}
      <form onSubmit={handleSubmit} className="shrink-0 p-3 border-t border-white/[0.06]">
        <div className="relative flex items-center gap-2">
          <button type="button" className="p-2 hover:bg-white/10 rounded-lg transition-colors shrink-0">
            <Paperclip className="w-5 h-5 text-slate-500" />
          </button>
          <input
            value={input}
            onChange={e => setInput(e.target.value)}
            placeholder="Написать сообщение..."
            className="xplr-input w-full !pl-4 !pr-12 !py-2.5 text-sm"
          />
          <button
            type="submit"
            className="absolute right-2 p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <Send className={`w-4 h-4 ${input.trim() ? 'text-blue-400' : 'text-slate-600'}`} />
          </button>
        </div>
      </form>
    </div>
  );
};

/* ──────────────────────────── Kanban Board ───────────────────────────── */

const columnOrder: KanbanTask['column'][] = ['backlog', 'progress', 'done'];

const KanbanColumn = ({ title, icon, color, columnKey, tasks, members, onAddTask, onMoveTask, onDeleteTask }: {
  title: string;
  icon: React.ReactNode;
  color: string;
  columnKey: KanbanTask['column'];
  tasks: KanbanTask[];
  members: TeamMember[];
  onAddTask: (column: KanbanTask['column']) => void;
  onMoveTask: (taskId: string, direction: 'left' | 'right') => void;
  onDeleteTask: (taskId: string) => void;
}) => {
  const getMember = (id: string) => members.find(m => m.id === id);
  const colIdx = columnOrder.indexOf(columnKey);
  const canLeft = colIdx > 0;
  const canRight = colIdx < columnOrder.length - 1;

  return (
    <div className="flex-1 min-w-[260px]">
      <div className="flex items-center gap-2 mb-3 px-1">
        {icon}
        <span className={`text-sm font-semibold ${color}`}>{title}</span>
        <span className="ml-auto text-xs text-slate-600 bg-white/[0.04] px-2 py-0.5 rounded-full">{tasks.length}</span>
      </div>
      <div className="space-y-2 pb-6">
        {tasks.map(task => {
          const assignee = getMember(task.assigneeId);
          return (
            <div key={task.id} className="glass-card p-3 group hover:border-white/20 transition-all duration-150">
              <div className="flex items-start gap-2">
                <GripVertical className="w-4 h-4 text-slate-700 mt-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white mb-1">{task.title}</p>
                  <p className="text-xs text-slate-500 mb-2">{task.description}</p>
                  <div className="flex items-center justify-between">
                    {assignee ? (
                      <div className="flex items-center gap-2">
                        <Avatar member={assignee} size="sm" />
                        <span className="text-[11px] text-slate-400">{assignee.name.split(' ')[0]}</span>
                      </div>
                    ) : <div />}
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      {canLeft && (
                        <button onClick={() => onMoveTask(task.id, 'left')} className="p-1 hover:bg-white/10 rounded transition-colors" title="Влево">
                          <ChevronLeft className="w-3.5 h-3.5 text-slate-400" />
                        </button>
                      )}
                      {canRight && (
                        <button onClick={() => onMoveTask(task.id, 'right')} className="p-1 hover:bg-white/10 rounded transition-colors" title="Вправо">
                          <ChevronRight className="w-3.5 h-3.5 text-slate-400" />
                        </button>
                      )}
                      <button onClick={() => onDeleteTask(task.id)} className="p-1 hover:bg-red-500/20 rounded transition-colors" title="Удалить">
                        <Trash2 className="w-3.5 h-3.5 text-slate-500 hover:text-red-400" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          );
        })}
        <button
          onClick={() => onAddTask(columnKey)}
          className="w-full py-2.5 flex items-center justify-center gap-1.5 text-xs font-medium text-slate-400 bg-white/[0.04] hover:bg-white/[0.08] hover:text-white rounded-lg border border-dashed border-white/10 hover:border-white/20 transition-all duration-150"
        >
          <Plus className="w-3.5 h-3.5" />
          Добавить задачу
        </button>
      </div>
    </div>
  );
};

const KanbanBoard = ({ tasks, members, onAddTask, onMoveTask, onDeleteTask }: {
  tasks: KanbanTask[];
  members: TeamMember[];
  onAddTask: (column: KanbanTask['column']) => void;
  onMoveTask: (taskId: string, direction: 'left' | 'right') => void;
  onDeleteTask: (taskId: string) => void;
}) => (
  <div className="flex-1 overflow-x-auto overflow-y-auto p-4">
    <div className="flex gap-4 min-w-[820px]">
      <KanbanColumn
        title="В очереди"
        icon={<Clock className="w-4 h-4 text-slate-400" />}
        color="text-slate-300"
        columnKey="backlog"
        tasks={tasks.filter(t => t.column === 'backlog')}
        members={members}
        onAddTask={onAddTask}
        onMoveTask={onMoveTask}
        onDeleteTask={onDeleteTask}
      />
      <KanbanColumn
        title="В процессе"
        icon={<div className="w-4 h-4 rounded-full bg-blue-500/20 flex items-center justify-center"><div className="w-2 h-2 rounded-full bg-blue-400 animate-pulse" /></div>}
        color="text-blue-400"
        columnKey="progress"
        tasks={tasks.filter(t => t.column === 'progress')}
        members={members}
        onAddTask={onAddTask}
        onMoveTask={onMoveTask}
        onDeleteTask={onDeleteTask}
      />
      <KanbanColumn
        title="Готово"
        icon={<Check className="w-4 h-4 text-emerald-400" />}
        color="text-emerald-400"
        columnKey="done"
        tasks={tasks.filter(t => t.column === 'done')}
        members={members}
        onAddTask={onAddTask}
        onMoveTask={onMoveTask}
        onDeleteTask={onDeleteTask}
      />
    </div>
  </div>
);

/* ──────────────────────────── Add Task Modal ──────────────────────────── */

const AddTaskModal = ({ column, members, onClose, onSave }: {
  column: KanbanTask['column'];
  members: TeamMember[];
  onClose: () => void;
  onSave: (task: Omit<KanbanTask, 'id'>) => void;
}) => {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [assigneeId, setAssigneeId] = useState(members[0]?.id || '');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;
    onSave({ title: title.trim(), description: description.trim(), assigneeId, column });
    onClose();
  };

  const columnLabels: Record<string, string> = { backlog: 'В очереди', progress: 'В процессе', done: 'Готово' };

  return (
    <ModalPortal>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
        <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-[440px] shadow-2xl shadow-black/60 animate-scale-in">
          <div className="flex items-center justify-between mb-5">
            <h3 className="text-lg font-semibold text-white">Новая задача</h3>
            <button onClick={onClose} className="p-2 hover:bg-white/10 rounded-lg transition-colors">
              <X className="w-5 h-5 text-slate-400" />
            </button>
          </div>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1.5">Название *</label>
              <input
                value={title}
                onChange={e => setTitle(e.target.value)}
                placeholder="Что нужно сделать?"
                className="xplr-input w-full"
                autoFocus
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1.5">Описание</label>
              <input
                value={description}
                onChange={e => setDescription(e.target.value)}
                placeholder="Детали задачи..."
                className="xplr-input w-full"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1.5">Исполнитель</label>
              <select
                value={assigneeId}
                onChange={e => setAssigneeId(e.target.value)}
                className="xplr-select w-full"
              >
                {members.map(m => (
                  <option key={m.id} value={m.id}>{m.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1.5">Колонка</label>
              <div className="px-3 py-2.5 bg-white/[0.04] border border-white/10 rounded-xl text-sm text-white">
                {columnLabels[column]}
              </div>
            </div>
            <div className="flex gap-3 pt-2">
              <button type="button" onClick={onClose} className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-slate-300 font-medium rounded-xl transition-colors">
                Отмена
              </button>
              <button
                type="submit"
                disabled={!title.trim()}
                className="flex-1 px-4 py-3 gradient-accent text-white font-medium rounded-xl transition-all disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Создать
              </button>
            </div>
          </form>
        </div>
      </div>
    </ModalPortal>
  );
};

/* ──────────────────────────── Main Page ──────────────────────────────── */

/* ── localStorage helpers ── */
const loadTasks = (): KanbanTask[] => {
  try {
    const raw = localStorage.getItem(LS_TASKS_KEY);
    if (raw) return JSON.parse(raw) as KanbanTask[];
  } catch { /* ignore */ }
  return initialTasks;
};

const loadMessages = (): ChatMessage[] => {
  try {
    const raw = localStorage.getItem(LS_CHAT_KEY);
    if (raw) return JSON.parse(raw) as ChatMessage[];
  } catch { /* ignore */ }
  return initialMessages;
};

export const TeamsPage = () => {
  const [showInvite, setShowInvite] = useState(false);
  const [activeTab, setActiveTab] = useState<'chat' | 'tasks'>('chat');
  const [selectedMember, setSelectedMember] = useState<string | null>(null);
  const [showMobileMembers, setShowMobileMembers] = useState(false);
  const [addTaskColumn, setAddTaskColumn] = useState<KanbanTask['column'] | null>(null);
  const { user } = useAuth();

  // ── Tasks state with localStorage persistence ──
  const [tasks, setTasks] = useState<KanbanTask[]>(loadTasks);

  useEffect(() => {
    localStorage.setItem(LS_TASKS_KEY, JSON.stringify(tasks));
  }, [tasks]);

  const handleAddTask = useCallback((data: Omit<KanbanTask, 'id'>) => {
    setTasks(prev => [...prev, { ...data, id: `k${Date.now()}` }]);
  }, []);

  const handleMoveTask = useCallback((taskId: string, direction: 'left' | 'right') => {
    setTasks(prev => prev.map(t => {
      if (t.id !== taskId) return t;
      const idx = columnOrder.indexOf(t.column);
      const newIdx = direction === 'right' ? Math.min(idx + 1, columnOrder.length - 1) : Math.max(idx - 1, 0);
      return { ...t, column: columnOrder[newIdx] };
    }));
  }, []);

  const handleDeleteTask = useCallback((taskId: string) => {
    setTasks(prev => prev.filter(t => t.id !== taskId));
  }, []);

  // ── Chat state with localStorage persistence ──
  const [messages, setMessages] = useState<ChatMessage[]>(loadMessages);

  useEffect(() => {
    localStorage.setItem(LS_CHAT_KEY, JSON.stringify(messages));
  }, [messages]);

  const handleSendMessage = useCallback((text: string) => {
    const newMsg: ChatMessage = {
      id: `m${Date.now()}`,
      senderId: user?.id || '1',
      text,
      time: new Date().toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }),
      read: false,
    };
    setMessages(prev => [...prev, newMsg]);
  }, [user?.id]);

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />

        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
          <div>
            <h1 className="text-3xl font-bold text-white mb-1">Командный центр</h1>
            <p className="text-slate-500">Чат и управление задачами</p>
          </div>
          <button
            onClick={() => setShowInvite(true)}
            className="flex items-center gap-2 px-5 py-3 gradient-accent text-white font-medium rounded-xl transition-all shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40"
          >
            <UserPlus className="w-5 h-5" />
            Пригласить
          </button>
        </div>

        {/* Tab Switcher */}
        <div className="flex items-center gap-1 p-1 glass-card mb-6 w-fit">
          <button
            onClick={() => setActiveTab('chat')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold transition-all duration-150 ${
              activeTab === 'chat'
                ? 'gradient-accent text-white shadow-lg shadow-blue-500/25'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            <MessageCircle className="w-4 h-4" />
            Обсуждение
          </button>
          <button
            onClick={() => setActiveTab('tasks')}
            className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold transition-all duration-150 ${
              activeTab === 'tasks'
                ? 'gradient-accent text-white shadow-lg shadow-blue-500/25'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            <KanbanSquare className="w-4 h-4" />
            Задачи
          </button>
        </div>

        {/* Content */}
        <div className="glass-card overflow-hidden" style={{ height: 'calc(100dvh - 300px)', minHeight: '400px' }}>
          {activeTab === 'chat' ? (
            <div className="flex h-full">
              {/* Member list — desktop */}
              <div className="hidden md:flex w-64 border-r border-white/[0.06] flex-col shrink-0">
                <MemberListPanel members={teamMembers} selectedId={selectedMember} onSelect={setSelectedMember} />
              </div>

              {/* Mobile member toggle */}
              <div className="md:hidden absolute top-2 right-2 z-10">
                <button onClick={() => setShowMobileMembers(!showMobileMembers)} className="p-2 glass-card">
                  <User className="w-4 h-4 text-slate-400" />
                </button>
              </div>

              {/* Mobile member sheet */}
              {showMobileMembers && (
                <div className="md:hidden fixed inset-0 z-50 flex">
                  <div className="absolute inset-0 bg-black/60" onClick={() => setShowMobileMembers(false)} />
                  <div className="relative w-72 bg-[#0a0a0f] border-r border-white/[0.08] h-full">
                    <MemberListPanel members={teamMembers} selectedId={selectedMember} onSelect={(id) => { setSelectedMember(id); setShowMobileMembers(false); }} />
                  </div>
                </div>
              )}

              {/* Chat area */}
              <div className="flex-1 flex flex-col min-w-0">
                <ChatPanel messages={messages} members={teamMembers} onSend={handleSendMessage} />
              </div>
            </div>
          ) : (
            <KanbanBoard
              tasks={tasks}
              members={teamMembers}
              onAddTask={(col) => setAddTaskColumn(col)}
              onMoveTask={handleMoveTask}
              onDeleteTask={handleDeleteTask}
            />
          )}
        </div>

        {/* Invite Modal */}
        {showInvite && <InviteModal onClose={() => setShowInvite(false)} />}
        {addTaskColumn && (
          <AddTaskModal
            column={addTaskColumn}
            members={teamMembers}
            onClose={() => setAddTaskColumn(null)}
            onSave={handleAddTask}
          />
        )}
      </div>
    </DashboardLayout>
  );
};
