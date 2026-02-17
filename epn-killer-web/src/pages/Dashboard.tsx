import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { Line, Pie } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler);

import { API_BASE_URL } from '../api/axios';
import { theme, XPLR_STORAGE_MODE } from '../theme/theme';
import SidebarLayout from '../components/SidebarLayout';

interface UserData {
  id: number;
  email: string;
  balance: number;
  balance_rub?: number;
  balance_arbitrage?: number;
  balance_personal?: number;
  status: string;
  grade?: string;
  fee_percent?: string;
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [userData, setUserData] = useState<UserData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const appMode = localStorage.getItem(XPLR_STORAGE_MODE) || 'professional';
  const isProfessional = appMode === 'professional';

  // Top-up
  const [isTopingUp, setIsTopingUp] = useState(false);
  const [topUpWallet, setTopUpWallet] = useState<'arbitrage' | 'personal'>('arbitrage');

  // Stats
  const [userGrade, setUserGrade] = useState<{ grade: string; fee_percent: string } | null>(null);
  const [spendStats, setSpendStats] = useState<{ category: string; total_spent: string; tx_count: number }[]>([]);
  const [exchangeRates, setExchangeRates] = useState<{ currency_from: string; currency_to: string; final_rate: string }[]>([]);
  const [cardCount, setCardCount] = useState(0);
  const [txCount, setTxCount] = useState(0);

  const getConfig = () => ({ headers: { Authorization: `Bearer ${localStorage.getItem('token')}` } });

  useEffect(() => { fetchData(); }, []);
  useEffect(() => { if (toast) { const t = setTimeout(() => setToast(null), 3500); return () => clearTimeout(t); } }, [toast]);

  const fetchData = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) { navigate('/login'); return; }
      const config = getConfig();

      const userRes = await axios.get(`${API_BASE_URL}/user/me`, config);
      setUserData(userRes.data);

      try { const g = await axios.get(`${API_BASE_URL}/user/grade`, config); setUserGrade({ grade: g.data.grade, fee_percent: g.data.fee_percent }); } catch {}
      try { const c = await axios.get(`${API_BASE_URL}/user/cards`, config); setCardCount(Array.isArray(c.data) ? c.data.length : 0); } catch {}
      try { const s = await axios.get(`${API_BASE_URL}/user/stats`, config); setSpendStats(Array.isArray(s.data?.categories) ? s.data.categories : []); } catch {}
      try { const r = await axios.get(`${API_BASE_URL}/exchange-rates`, config); setExchangeRates(Array.isArray(r.data) ? r.data : []); } catch {}
      try {
        const t = await axios.get(`${API_BASE_URL}/user/report`, config);
        setTxCount(Array.isArray(t.data?.transactions) ? t.data.transactions.length : 0);
      } catch {}
    } catch (err) { console.error('Dashboard fetch error:', err); }
    finally { setIsLoading(false); }
  };

  const handleTopUp = async () => {
    setIsTopingUp(true);
    try {
      const res = await axios.post(`${API_BASE_URL}/user/topup`, { wallet: topUpWallet }, getConfig());
      const balArb = res.data.balance_arbitrage;
      const balPers = res.data.balance_personal;
      if (userData) {
        setUserData({
          ...userData,
          balance_arbitrage: balArb ? parseFloat(balArb) : userData.balance_arbitrage,
          balance_personal: balPers ? parseFloat(balPers) : userData.balance_personal,
        });
      }
      const label = topUpWallet === 'arbitrage' ? 'Арбитраж' : 'Личный';
      setToast({ message: `${label}: +$${parseFloat(res.data.amount_usd || '100').toFixed(2)}`, type: 'success' });
    } catch { setToast({ message: 'Не удалось пополнить', type: 'error' }); }
    finally { setIsTopingUp(false); }
  };

  // Chart data
  const pieData = {
    labels: (spendStats ?? []).map(s => s.category),
    datasets: [{
      data: (spendStats ?? []).map(s => parseFloat(s.total_spent)),
      backgroundColor: ['#3b82f6', '#14b8a6', '#8b5cf6', '#f59e0b'],
      borderWidth: 0
    }]
  };

  const lineData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [{
      label: 'Spend',
      data: [120, 230, 180, 340, 280, 190, 310],
      borderColor: theme.colors.accent,
      backgroundColor: theme.colors.accentMuted,
      fill: true,
      tension: 0.4,
      pointRadius: 0
    }]
  };

  const chartOptions: any = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: { grid: { display: false }, ticks: { color: theme.colors.textSecondary, font: { size: 11 } }, border: { display: false } },
      y: { grid: { color: theme.colors.border }, ticks: { color: theme.colors.textSecondary, font: { size: 11 }, callback: (v: any) => `$${v}` }, border: { display: false } }
    }
  };

  if (isLoading) {
    return <SidebarLayout><div style={{ padding: 60, textAlign: 'center', color: theme.colors.textSecondary, fontSize: 16 }}>Загрузка...</div></SidebarLayout>;
  }

  return (
    <SidebarLayout>
      {/* Toast */}
      {toast && (
        <div style={{
          position: 'fixed', top: 20, right: 20, zIndex: 99999,
          padding: '14px 24px', borderRadius: 12,
          backgroundColor: toast.type === 'success' ? 'rgba(16, 185, 129, 0.9)' : 'rgba(239, 68, 68, 0.9)',
          color: '#fff', fontWeight: 600, fontSize: 14, backdropFilter: 'blur(12px)'
        }}>{toast.message}</div>
      )}

      {/* Header */}
      <div style={{ marginBottom: 32 }}>
        <h1 style={{ margin: 0, fontSize: 28, fontWeight: 800, letterSpacing: '-0.5px' }}>
          {isProfessional ? 'Account Overview' : 'Мой кабинет'}
        </h1>
        <p style={{ margin: '6px 0 0', color: theme.colors.textSecondary, fontSize: 14 }}>
          Добро пожаловать, {userData?.email?.split('@')[0] || 'User'}
        </p>
      </div>

      {/* Wallets + Quick Stats */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 16, marginBottom: 32 }}>
        {/* Arbitrage Wallet */}
        <div onClick={() => setTopUpWallet('arbitrage')} style={{
          padding: 20, borderRadius: 14, cursor: 'pointer', transition: '0.2s',
          backgroundColor: topUpWallet === 'arbitrage' ? 'rgba(59, 130, 246, 0.12)' : theme.colors.backgroundCard,
          border: topUpWallet === 'arbitrage' ? '1px solid rgba(59, 130, 246, 0.4)' : `1px solid ${theme.colors.border}`
        }}>
          <div style={{ fontSize: 11, color: '#3b82f6', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 6, fontWeight: 600 }}>💳 Арбитраж</div>
          <div style={{ fontSize: 32, fontWeight: 800, letterSpacing: '-1px' }}>${Number(userData?.balance_arbitrage ?? userData?.balance ?? 0).toFixed(2)}</div>
        </div>

        {/* Personal Wallet */}
        <div onClick={() => setTopUpWallet('personal')} style={{
          padding: 20, borderRadius: 14, cursor: 'pointer', transition: '0.2s',
          backgroundColor: topUpWallet === 'personal' ? 'rgba(20, 184, 166, 0.12)' : theme.colors.backgroundCard,
          border: topUpWallet === 'personal' ? '1px solid rgba(20, 184, 166, 0.4)' : `1px solid ${theme.colors.border}`
        }}>
          <div style={{ fontSize: 11, color: '#14b8a6', textTransform: 'uppercase', letterSpacing: 1, marginBottom: 6, fontWeight: 600 }}>✈️ Личный</div>
          <div style={{ fontSize: 32, fontWeight: 800, letterSpacing: '-1px' }}>${Number(userData?.balance_personal ?? 0).toFixed(2)}</div>
        </div>

        {/* Cards count */}
        <div onClick={() => navigate('/cards')} style={{
          padding: 20, borderRadius: 14, cursor: 'pointer',
          backgroundColor: theme.colors.backgroundCard, border: `1px solid ${theme.colors.border}`, transition: '0.2s'
        }}>
          <div style={{ fontSize: 11, color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 6, fontWeight: 600 }}>Активные карты</div>
          <div style={{ fontSize: 32, fontWeight: 800 }}>{cardCount}</div>
        </div>

        {/* Transactions count */}
        <div onClick={() => navigate('/history')} style={{
          padding: 20, borderRadius: 14, cursor: 'pointer',
          backgroundColor: theme.colors.backgroundCard, border: `1px solid ${theme.colors.border}`, transition: '0.2s'
        }}>
          <div style={{ fontSize: 11, color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: 1, marginBottom: 6, fontWeight: 600 }}>Транзакции</div>
          <div style={{ fontSize: 32, fontWeight: 800 }}>{txCount}</div>
        </div>
      </div>

      {/* Top Up */}
      <div style={{ marginBottom: 32, display: 'flex', gap: 12, alignItems: 'center' }}>
        <button onClick={handleTopUp} disabled={isTopingUp} style={{
          padding: '12px 28px', backgroundColor: theme.colors.accent,
          color: theme.colors.background, border: 'none', borderRadius: 10,
          fontWeight: 700, fontSize: 14, cursor: isTopingUp ? 'not-allowed' : 'pointer',
          opacity: isTopingUp ? 0.6 : 1, transition: '0.2s'
        }}>
          {isTopingUp ? '...' : `+ Top Up → ${topUpWallet === 'arbitrage' ? 'Арбитраж' : 'Личный'}`}
        </button>
        {isProfessional && userGrade && (
          <span style={{
            padding: '8px 14px', backgroundColor: theme.colors.accentMuted,
            border: `1px solid ${theme.colors.accentBorder}`, borderRadius: 8,
            fontSize: 12, color: theme.colors.accent, fontWeight: 600
          }}>
            {userGrade.grade} · {parseFloat(userGrade.fee_percent).toFixed(1)}%
          </span>
        )}
      </div>

      {/* Exchange Rates */}
      {(exchangeRates ?? []).length > 0 && (
        <div style={{ display: 'flex', gap: 12, marginBottom: 32, flexWrap: 'wrap' }}>
          {(exchangeRates ?? []).map((r, i) => (
            <div key={i} style={{
              backgroundColor: theme.colors.backgroundCard, border: `1px solid ${theme.colors.border}`,
              borderRadius: 10, padding: '10px 16px', display: 'flex', alignItems: 'center', gap: 10
            }}>
              <span style={{ fontWeight: 700, fontSize: 13 }}>{r.currency_from}/{r.currency_to}</span>
              <span style={{ color: theme.colors.accent, fontWeight: 800, fontSize: 15 }}>
                {parseFloat(r.final_rate).toFixed(2)}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Charts */}
      {isProfessional && (
        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 20, marginBottom: 32 }}>
          {/* Spend Line Chart */}
          <div style={{
            backgroundColor: theme.colors.backgroundCard, border: `1px solid ${theme.colors.border}`,
            borderRadius: 14, padding: 24
          }}>
            <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 700 }}>Расходы за неделю</h3>
            <div style={{ height: 240 }}>
              <Line data={lineData} options={chartOptions} />
            </div>
          </div>

          {/* Spend Pie Chart */}
          <div style={{
            backgroundColor: theme.colors.backgroundCard, border: `1px solid ${theme.colors.border}`,
            borderRadius: 14, padding: 24
          }}>
            <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 700 }}>По категориям</h3>
            <div style={{ height: 240, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              {(spendStats ?? []).length > 0 ? (
                <Pie data={pieData} options={{ responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom' as const, labels: { color: theme.colors.textSecondary, boxWidth: 12, padding: 12 } } } }} />
              ) : (
                <div style={{ color: theme.colors.textSecondary, fontSize: 13 }}>Нет данных</div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Quick Links */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 16 }}>
        {[
          { label: 'Мои карты', desc: 'Управление картами', icon: '💳', path: '/cards' },
          { label: 'Транзакции', desc: 'История операций', icon: '💸', path: '/history' },
          ...(isProfessional ? [
            { label: 'Команды', desc: 'Управление командой', icon: '👥', path: '/teams' },
            { label: 'API', desc: 'Документация и ключи', icon: '🔌', path: '/api' },
          ] : []),
          { label: 'Рефералы', desc: 'Пригласи друзей', icon: '🎁', path: '/referrals' },
        ].map(link => (
          <div key={link.path} onClick={() => navigate(link.path)} style={{
            padding: 20, borderRadius: 14, cursor: 'pointer',
            backgroundColor: theme.colors.backgroundCard, border: `1px solid ${theme.colors.border}`,
            transition: '0.2s'
          }}
          onMouseEnter={e => { e.currentTarget.style.borderColor = theme.colors.accent; e.currentTarget.style.transform = 'translateY(-2px)'; }}
          onMouseLeave={e => { e.currentTarget.style.borderColor = theme.colors.border; e.currentTarget.style.transform = 'translateY(0)'; }}
          >
            <div style={{ fontSize: 28, marginBottom: 10 }}>{link.icon}</div>
            <div style={{ fontWeight: 700, fontSize: 15, marginBottom: 4 }}>{link.label}</div>
            <div style={{ fontSize: 12, color: theme.colors.textSecondary }}>{link.desc}</div>
          </div>
        ))}
      </div>
    </SidebarLayout>
  );
};

export default Dashboard;
