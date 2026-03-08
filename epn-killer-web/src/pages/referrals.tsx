import React, { useState, useEffect } from 'react';
import apiClient, { API_BASE_URL } from '../api/axios';
import { DashboardLayout } from '../components/dashboard-layout';
import { BackButton } from '../components/back-button';
import { ShareModal } from '../components/share-modal';
import { 
  Copy,
  Gift,
  Users,
  DollarSign,
  Check,
  ArrowRight,
  Trophy,
  Share2,
  QrCode,
  ExternalLink,
  Shield
} from 'lucide-react';

interface Referral {
  id: string;
  name: string;
  email: string;
  joinedDate: string;
  status: 'active' | 'pending' | 'expired';
  earnings: number;
}

const StatCard = ({ icon, label, value, accent, iconClass = 'stat-icon-blue' }: { 
  icon: React.ReactNode; 
  label: string; 
  value: string; 
  accent?: boolean;
  iconClass?: string;
}) => (
  <div className={`glass-card p-5 ${accent ? 'border-blue-300' : ''}`}>
    <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-3 ${iconClass}`}>
      {icon}
    </div>
    <p className="text-slate-500 text-sm mb-1">{label}</p>
    <p className="text-2xl font-bold text-white balance-display">{value}</p>
  </div>
);

const TierCard = ({ tier, reward, volume, current }: { tier: string; reward: string; volume: string; current: boolean }) => (
  <div className={`p-4 rounded-xl border transition-all ${
    current 
      ? 'bg-gradient-to-br from-blue-500/10 to-purple-500/10 border-blue-500/30' 
      : 'bg-white/[0.02] border-white/10 opacity-60'
  }`}>
    <div className="flex items-center justify-between mb-2">
      <span className={`text-sm font-semibold ${current ? 'text-blue-400' : 'text-gray-400'}`}>{tier}</span>
      {current && <Check className="w-4 h-4 text-blue-400" />}
    </div>
    <p className="text-white font-bold text-lg mb-1">{reward}</p>
    <p className="text-xs text-gray-500">{volume}</p>
  </div>
);

// Simple QR Code component using SVG
const QRCodeDisplay = ({ value, size = 120 }: { value: string; size?: number }) => {
  // This is a simplified QR-like pattern generator
  // In production, you'd use a library like qrcode.react
  const generatePattern = () => {
    const seed = value.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
    const pattern: boolean[][] = [];
    const gridSize = 21;
    
    for (let i = 0; i < gridSize; i++) {
      pattern[i] = [];
      for (let j = 0; j < gridSize; j++) {
        // Corner patterns (fixed for all QR codes)
        if ((i < 7 && j < 7) || (i < 7 && j >= gridSize - 7) || (i >= gridSize - 7 && j < 7)) {
          const cornerI = i < 7 ? i : i - (gridSize - 7);
          const cornerJ = j < 7 ? j : j - (gridSize - 7);
          if (cornerI === 0 || cornerI === 6 || cornerJ === 0 || cornerJ === 6 ||
              (cornerI >= 2 && cornerI <= 4 && cornerJ >= 2 && cornerJ <= 4)) {
            pattern[i][j] = true;
          } else {
            pattern[i][j] = false;
          }
        } else {
          // Random-ish pattern based on seed
          const hash = ((seed * (i + 1) * (j + 1)) % 17);
          pattern[i][j] = hash < 8;
        }
      }
    }
    return pattern;
  };

  const pattern = generatePattern();
  const cellSize = size / 21;

  return (
    <div className="qr-code-container">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {pattern.map((row, i) =>
          row.map((cell, j) =>
            cell && (
              <rect
                key={`${i}-${j}`}
                x={j * cellSize}
                y={i * cellSize}
                width={cellSize}
                height={cellSize}
                fill="#111114"
              />
            )
          )
        )}
      </svg>
    </div>
  );
};

export const ReferralsPage = () => {
  const [copied, setCopied] = useState(false);
  const [showQR, setShowQR] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const [referralCode, setReferralCode] = useState('');
  const [referralLink, setReferralLink] = useState('');
  const [referrals, setReferrals] = useState<Referral[]>([]);
  const [stats, setStats] = useState({ totalReferrals: 0, earnings: 0, pending: 0, rewardPerReferral: 10, bonusForNew: 5 });
  const personalStats = stats;

  useEffect(() => {
    const fetchReferralData = async () => {
      try {
        const res = await apiClient.get(`${API_BASE_URL}/user/referrals/info`);
        const data = res.data;
        setReferralCode(data.referral_code || '');
        setReferralLink(data.referral_link || '');
        setStats({
          totalReferrals: data.stats?.total_referrals || 0,
          earnings: parseFloat(data.stats?.total_earnings || '0'),
          pending: parseFloat(data.stats?.pending_amount || '0'),
          rewardPerReferral: data.reward_per_referral || 10,
          bonusForNew: data.bonus_for_new || 5,
        });
        const recent: Referral[] = (data.recent_referrals || []).map((r: any) => ({
          id: String(r.id),
          name: r.email?.split('@')[0] || 'User',
          email: r.email || '',
          joinedDate: r.joined_date ? new Date(r.joined_date).toLocaleDateString('ru-RU') : '',
          status: r.is_active ? 'active' as const : 'pending' as const,
          earnings: parseFloat(r.earnings || '0'),
        }));
        setReferrals(recent);
      } catch {
        // keep defaults
      } finally {
        setIsLoading(false);
      }
    };
    fetchReferralData();
  }, []);

  const handleCopy = () => {
    navigator.clipboard.writeText(referralLink);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleShare = () => {
    setShowShareModal(true);
  };

  const statusLabels: Record<string, string> = {
    active: 'Активен',
    pending: 'Ожидание',
    expired: 'Истёк'
  };

  return (
    <DashboardLayout>
      <div className="stagger-fade-in">
        <BackButton />
        
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">Реферальная программа</h1>
          <p className="text-slate-500">
            {`Приглашайте друзей и получайте $${personalStats.rewardPerReferral} за каждого. Ваш друг тоже получит $${personalStats.bonusForNew}!`}
          </p>
        </div>

        {/* Referral Link Card */}
        <div className="glass-card p-6 mb-8 relative overflow-hidden">
          {/* Background decoration */}
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/10 to-purple-500/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
          
          <div className="relative z-10">
            <div className="flex flex-col lg:flex-row lg:items-start gap-6">
              {/* Link section */}
              <div className="flex-1 min-w-0 overflow-hidden">
                <div className="flex items-center gap-3 mb-4">
                  <div className="w-12 h-12 rounded-xl gradient-accent flex items-center justify-center shadow-lg shadow-blue-500/30 shrink-0">
                    <Share2 className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-white">Ваша реферальная ссылка</h3>
                    <p className="text-sm text-gray-400">Поделитесь и зарабатывайте вместе</p>
                  </div>
                </div>
                
                <div className="flex flex-col gap-3 mb-4">
                  <div className="flex items-center bg-white/5 rounded-xl px-4 py-3 border border-white/10 overflow-hidden min-w-0">
                    <span className="text-white font-mono text-sm truncate">{referralLink}</span>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <button 
                      onClick={handleCopy}
                      className={`px-5 py-3 rounded-xl font-medium transition-all flex items-center gap-2 justify-center ${
                        copied 
                          ? 'bg-green-500 text-white' 
                          : 'gradient-accent text-white hover:opacity-90'
                      }`}
                    >
                      {copied ? <Check className="w-5 h-5" /> : <Copy className="w-5 h-5" />}
                      {copied ? 'Скопировано!' : 'Копировать'}
                    </button>
                    <button 
                      onClick={() => setShowQR(!showQR)}
                      className={`p-3 rounded-xl transition-colors shrink-0 ${
                        showQR ? 'bg-blue-500/20 text-blue-400' : 'glass-card hover:bg-white/10 text-gray-400'
                      }`}
                      title="QR-код"
                    >
                      <QrCode className="w-5 h-5" />
                    </button>
                    <button 
                      onClick={handleShare}
                      className="p-3 glass-card hover:bg-white/10 text-gray-400 rounded-xl transition-colors shrink-0"
                      title="Поделиться"
                    >
                      <ExternalLink className="w-5 h-5" />
                    </button>
                  </div>
                </div>

                <div className="flex flex-wrap items-center gap-4 text-sm">
                  <p className="text-gray-400 flex items-center">
                    Код: <span className="text-white font-mono font-semibold ml-1">{referralCode}</span>
                  </p>
                  <div className="flex items-center gap-2 text-green-400">
                    <Gift className="w-4 h-4 shrink-0" />
                    <span>+${personalStats.bonusForNew} новому пользователю</span>
                  </div>
                </div>
              </div>

              {/* QR Code section */}
              {showQR && (
                <div className="flex flex-col items-center p-4 glass-card animate-fade-in">
                  <QRCodeDisplay value={referralLink} size={140} />
                  <p className="text-xs text-gray-400 mt-3">Сканируйте для перехода</p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-2 gap-4 mb-8">
          <StatCard 
            icon={<Users className="w-5 h-5 text-blue-400" />}
            label="Всего рефералов"
            value={stats.totalReferrals.toString()}
            iconClass="stat-icon-blue"
            accent
          />
          <StatCard 
            icon={<DollarSign className="w-5 h-5 text-green-400" />}
            label="Заработано"
            value={`$${stats.earnings}`}
            iconClass="stat-icon-green"
          />
        </div>

        {/* How it works */}
        {
          <div className="glass-card p-6 mb-8">
            <h3 className="block-title flex items-center gap-2">
              <Gift className="w-5 h-5 text-blue-400" />
              Как это работает
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-center gap-3 p-4 rounded-xl bg-white/[0.02]">
                <div className="w-8 h-8 rounded-full gradient-accent flex items-center justify-center text-white font-bold text-sm shrink-0">1</div>
                <div>
                  <p className="text-white font-medium mb-1">Поделитесь ссылкой</p>
                  <p className="text-sm text-gray-400">Отправьте ссылку друзьям</p>
                </div>
              </div>
              <div className="flex items-center gap-3 p-4 rounded-xl bg-white/[0.02]">
                <div className="w-8 h-8 rounded-full gradient-accent flex items-center justify-center text-white font-bold text-sm shrink-0">2</div>
                <div>
                  <p className="text-white font-medium mb-1">Друг регистрируется</p>
                  <p className="text-sm text-gray-400">И получает бонус ${personalStats.bonusForNew}</p>
                </div>
              </div>
              <div className="flex items-center gap-3 p-4 rounded-xl bg-white/[0.02]">
                <div className="w-8 h-8 rounded-full gradient-accent flex items-center justify-center text-white font-bold text-sm shrink-0">3</div>
                <div>
                  <p className="text-white font-medium mb-1">Вы получаете ${personalStats.rewardPerReferral}</p>
                  <p className="text-sm text-gray-400">Сразу после активации</p>
                </div>
              </div>
            </div>
          </div>
        }

        {/* Referral List */}
        <div className="glass-card overflow-hidden mb-6">
          <div className="p-6 border-b border-white/10 flex items-center justify-between">
            <h3 className="block-title mb-0">Последние рефералы</h3>
            <span className="text-sm text-gray-400">{referrals.length} всего</span>
          </div>
          <div className="overflow-x-auto">
            <table className="xplr-table min-w-[500px]">
              <thead>
                <tr>
                  <th>Пользователь</th>
                  <th>Присоединился</th>
                  <th>Статус</th>
                  <th>Заработок</th>
                </tr>
              </thead>
              <tbody>
                {referrals.map(ref => (
                  <tr key={ref.id}>
                    <td className="py-4 px-4">
                      <div>
                        <p className="text-white font-medium">{ref.name}</p>
                        <p className="text-sm text-gray-500">{ref.email}</p>
                      </div>
                    </td>
                    <td className="py-4 px-4">
                      <span className="text-gray-400">{ref.joinedDate}</span>
                    </td>
                    <td className="py-4 px-4">
                      <span className={`badge ${
                        ref.status === 'active' ? 'badge-success' :
                        ref.status === 'pending' ? 'badge-warning' : 'badge-error'
                      }`}>
                        {statusLabels[ref.status]}
                      </span>
                    </td>
                    <td className="py-4 px-4">
                      <span className={`font-semibold ${ref.earnings > 0 ? 'text-green-400' : 'text-gray-500'}`}>
                        {ref.earnings > 0 ? `+$${ref.earnings}` : '-'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Transfer to Wallet */}
        <div className="glass-card p-6 flex flex-col sm:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500/20 to-orange-500/20 border border-amber-500/30 flex items-center justify-center">
              <Shield className="w-6 h-6 text-amber-400" />
            </div>
            <div>
              <p className="text-white font-semibold">Перевести в Кошелёк</p>
              <p className="text-xs text-slate-400">Минимум для вывода: $100</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <span className="text-lg font-bold text-white">${stats.earnings}</span>
            <button 
              disabled={stats.earnings < 100}
              className={`flex items-center gap-2 px-6 py-3 font-medium rounded-xl transition-all ${
                stats.earnings >= 100
                  ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white shadow-lg shadow-amber-500/25 hover:shadow-amber-500/40 hover:opacity-90'
                  : 'bg-white/5 text-slate-500 border border-white/10 cursor-not-allowed'
              }`}
            >
              <Shield className="w-5 h-5" />
              В Кошелёк
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {showShareModal && (
        <ShareModal
          url={referralLink}
          text="Присоединяйся к XPLR и получи бонус!"
          onClose={() => setShowShareModal(false)}
        />
      )}
    </DashboardLayout>
  );
};
