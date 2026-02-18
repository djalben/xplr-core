import React, { useState } from 'react';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { 
  UserPlus,
  MoreVertical,
  Crown,
  Shield,
  User,
  Mail,
  Trash2,
  Edit,
  BarChart3,
  Copy,
  Check,
  X,
  Link
} from 'lucide-react';

interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: 'owner' | 'admin' | 'member';
  avatar: string;
  cardsIssued: number;
  volumeThisMonth: number;
  lastActive: string;
  status: 'online' | 'offline' | 'away';
}

const RoleBadge = ({ role }: { role: string }) => {
  const config: Record<string, { icon: React.ReactNode; color: string; label: string }> = {
    owner: { 
      icon: <Crown className="w-3 h-3" />, 
      color: 'bg-gradient-to-r from-yellow-500/20 to-orange-500/20 text-yellow-400 border-yellow-500/30',
      label: 'Владелец'
    },
    admin: { 
      icon: <Shield className="w-3 h-3" />, 
      color: 'bg-gradient-to-r from-blue-500/20 to-purple-500/20 text-blue-400 border-blue-500/30',
      label: 'Админ'
    },
    member: { 
      icon: <User className="w-3 h-3" />, 
      color: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
      label: 'Участник'
    },
  };

  const { icon, color, label } = config[role] || config.member;

  return (
    <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border ${color}`}>
      {icon}
      {label}
    </span>
  );
};

const StatusIndicator = ({ status }: { status: string }) => {
  const colors: Record<string, string> = {
    online: 'bg-emerald-500',
    offline: 'bg-slate-400',
    away: 'bg-amber-500'
  };

  return (
    <div className={`w-3 h-3 rounded-full ${colors[status]} ring-2 ring-white animate-pulse`} />
  );
};

const InviteModal = ({ onClose }: { onClose: () => void }) => {
  const [copied, setCopied] = useState(false);
  const inviteLink = 'https://xplr.io/team/join/abc123xyz';

  const handleCopy = () => {
    navigator.clipboard.writeText(inviteLink);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/80 backdrop-blur-md" onClick={onClose} />
      <div className="relative bg-[#050507]/95 backdrop-blur-3xl border border-white/10 p-6 rounded-2xl w-full max-w-md shadow-2xl shadow-black/60 animate-scale-in">
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
              <input
                type="email"
                placeholder="colleague@company.com"
                className="xplr-input w-full pl-10"
              />
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
            <input
              type="number"
              placeholder="50"
              className="xplr-input w-full"
            />
          </div>
          
          <div className="section-divider" />
          
          <div>
            <label className="block text-sm text-gray-400 mb-2">Или отправьте ссылку-приглашение</label>
            <div className="flex gap-2">
              <div className="flex-1 flex items-center gap-2 bg-white/5 rounded-lg px-3 py-2 border border-white/10">
                <Link className="w-4 h-4 text-gray-500" />
                <span className="text-sm text-gray-400 truncate">{inviteLink}</span>
              </div>
              <button 
                onClick={handleCopy}
                className={`px-4 py-2 rounded-lg transition-all flex items-center gap-2 ${
                  copied ? 'bg-green-500/20 text-green-400' : 'glass-card hover:bg-white/10 text-gray-400'
                }`}
              >
                {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
              </button>
            </div>
          </div>
        </div>
        
        <div className="flex gap-3 mt-6">
          <button 
            onClick={onClose}
            className="flex-1 px-4 py-3 glass-card hover:bg-white/10 text-gray-300 font-medium rounded-xl transition-colors"
          >
            Отмена
          </button>
          <button className="flex-1 px-4 py-3 gradient-accent text-white font-medium rounded-xl transition-colors">
            Отправить приглашение
          </button>
        </div>
      </div>
    </div>
  );
};

export const TeamsPage = () => {
  const [showInvite, setShowInvite] = useState(false);

  const teamMembers: TeamMember[] = [
    { 
      id: '1', 
      name: 'Алексей Петров', 
      email: 'alex@xplr.io', 
      role: 'owner', 
      avatar: 'АП',
      cardsIssued: 127,
      volumeThisMonth: 245000,
      lastActive: 'Сейчас',
      status: 'online'
    },
    { 
      id: '2', 
      name: 'Мария Иванова', 
      email: 'maria@xplr.io', 
      role: 'admin', 
      avatar: 'МИ',
      cardsIssued: 84,
      volumeThisMonth: 156000,
      lastActive: '5 мин назад',
      status: 'online'
    },
    { 
      id: '3', 
      name: 'Дмитрий Козлов', 
      email: 'dmitry@xplr.io', 
      role: 'member', 
      avatar: 'ДК',
      cardsIssued: 42,
      volumeThisMonth: 78000,
      lastActive: '2 часа назад',
      status: 'away'
    },
    { 
      id: '4', 
      name: 'Елена Смирнова', 
      email: 'elena@xplr.io', 
      role: 'member', 
      avatar: 'ЕС',
      cardsIssued: 31,
      volumeThisMonth: 52000,
      lastActive: 'Вчера',
      status: 'offline'
    },
  ];

  const totalVolume = teamMembers.reduce((sum, m) => sum + m.volumeThisMonth, 0);
  const totalCards = teamMembers.reduce((sum, m) => sum + m.cardsIssued, 0);

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />
        
        {/* Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">Ваша команда</h1>
            <p className="text-slate-500">Управление участниками и правами доступа</p>
          </div>
          <button 
            onClick={() => setShowInvite(true)}
            className="flex items-center gap-2 px-5 py-3 gradient-accent text-white font-medium rounded-xl transition-all shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40"
          >
            <UserPlus className="w-5 h-5" />
            Пригласить участника
          </button>
        </div>

        {/* Team Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          <div className="glass-card p-5">
            <p className="text-gray-400 text-sm mb-1">Участников</p>
            <p className="text-2xl font-bold text-white">{teamMembers.length}</p>
          </div>
          <div className="glass-card p-5">
            <p className="text-gray-400 text-sm mb-1">Онлайн сейчас</p>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
              <p className="text-2xl font-bold text-green-400">
                {teamMembers.filter(m => m.status === 'online').length}
              </p>
            </div>
          </div>
          <div className="glass-card p-5">
            <p className="text-gray-400 text-sm mb-1">Выпущено карт</p>
            <p className="text-2xl font-bold text-white">{totalCards}</p>
          </div>
          <div className="glass-card p-5 border-blue-500/30">
            <p className="text-gray-400 text-sm mb-1">Оборот команды</p>
            <p className="text-2xl font-bold text-blue-400 balance-display">${(totalVolume / 1000).toFixed(0)}K</p>
          </div>
        </div>

        {/* Shared Limits */}
        <div className="glass-card p-6 mb-8">
          <h3 className="block-title">Общие лимиты</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-gray-400">Выпуск карт в день</span>
                <span className="text-white font-semibold">284 / 500</span>
              </div>
              <div className="progress-bar-container">
                <div className="progress-bar-fill progress-bar-blue" style={{ width: '56.8%' }} />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-gray-400">Месячный оборот</span>
                <span className="text-white font-semibold">$531K / $1M</span>
              </div>
              <div className="progress-bar-container">
                <div className="progress-bar-fill progress-bar-green" style={{ width: '53.1%' }} />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-sm mb-2">
                <span className="text-gray-400">API запросы</span>
                <span className="text-white font-semibold">8.2K / 50K</span>
              </div>
              <div className="progress-bar-container">
                <div className="progress-bar-fill progress-bar-purple" style={{ width: '16.4%' }} />
              </div>
            </div>
          </div>
        </div>

        {/* Members List */}
        <div className="glass-card overflow-hidden">
          <div className="p-6 border-b border-white/10">
            <h3 className="block-title mb-0">Статистика команды</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="xplr-table min-w-[800px]">
              <thead>
                <tr>
                  <th>Участник</th>
                  <th>Роль</th>
                  <th>Карт выпущено</th>
                  <th>Оборот</th>
                  <th>Активность</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {teamMembers.map(member => (
                  <tr key={member.id}>
                    <td className="py-4 px-4">
                      <div className="flex items-center gap-3">
                        <div className="relative">
                          <div className="w-11 h-11 rounded-xl gradient-accent flex items-center justify-center text-white font-semibold text-sm shadow-lg shadow-blue-500/20">
                            {member.avatar}
                          </div>
                          <div className="absolute -bottom-0.5 -right-0.5">
                            <StatusIndicator status={member.status} />
                          </div>
                        </div>
                        <div>
                          <p className="text-white font-medium">{member.name}</p>
                          <p className="text-sm text-gray-500">{member.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="py-4 px-4">
                      <RoleBadge role={member.role} />
                    </td>
                    <td className="py-4 px-4">
                      <span className="text-white font-semibold">{member.cardsIssued}</span>
                    </td>
                    <td className="py-4 px-4">
                      <span className="text-white font-semibold balance-display">${member.volumeThisMonth.toLocaleString()}</span>
                    </td>
                    <td className="py-4 px-4">
                      <span className={`text-sm ${member.status === 'online' ? 'text-green-400' : 'text-gray-500'}`}>
                        {member.lastActive}
                      </span>
                    </td>
                    <td className="py-4 px-4">
                      <div className="flex items-center gap-1">
                        <button className="p-2.5 hover:bg-white/10 rounded-lg transition-colors" title="Статистика">
                          <BarChart3 className="w-4 h-4 text-gray-400" />
                        </button>
                        <button className="p-2.5 hover:bg-white/10 rounded-lg transition-colors" title="Редактировать">
                          <Edit className="w-4 h-4 text-gray-400" />
                        </button>
                        {member.role !== 'owner' && (
                          <button className="p-2.5 hover:bg-red-500/20 rounded-lg transition-colors" title="Удалить">
                            <Trash2 className="w-4 h-4 text-gray-400 hover:text-red-400" />
                          </button>
                        )}
                        <button className="p-2.5 hover:bg-white/10 rounded-lg transition-colors">
                          <MoreVertical className="w-4 h-4 text-gray-400" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Invite Modal */}
        {showInvite && <InviteModal onClose={() => setShowInvite(false)} />}
      </div>
      
      <style>{`
        @keyframes scale-in {
          from { opacity: 0; transform: scale(0.95); }
          to { opacity: 1; transform: scale(1); }
        }
        .animate-scale-in {
          animation: scale-in 0.2s ease-out forwards;
        }
      `}</style>
    </DashboardLayout>
  );
};
