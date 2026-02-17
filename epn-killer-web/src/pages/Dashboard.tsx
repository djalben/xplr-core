import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
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

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

import { API_BASE_URL } from '../api/axios';
import { setCardAutoReplenishment, unsetCardAutoReplenishment } from '../api/cards';
import { theme, XPLR_STORAGE_MODE, type AppMode } from '../theme/theme';

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

interface Card {
  id: number;
  nickname: string;
  last_4_digits: string;
  bin: string;
  card_status: string;
  daily_spend_limit: number;
  auto_replenish_enabled?: boolean;
  auto_replenish_threshold?: string;
  auto_replenish_amount?: string;
  card_balance?: string;
  card_type?: string;
  category?: string;
  service_slug?: string;
  team_id?: number;
}

interface Transaction {
  transaction_id: number;
  amount: number;
  transaction_type: string;
  status: string;
  executed_at: string;
  merchant?: string;
  card_last_4?: string;
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [userData, setUserData] = useState<UserData | null>(null);
  const [cards, setCards] = useState<Card[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Derive activeMenu from URL path
  const activeMenu = (() => {
    const p = location.pathname;
    if (p === '/cards') return 'cards';
    if (p === '/finance') return 'finance';
    if (p === '/api') return 'api';
    return 'dashboard';
  })();
  const [appMode, setAppMode] = useState<AppMode>(() => {
    const stored = localStorage.getItem(XPLR_STORAGE_MODE);
    return (stored === 'personal' || stored === 'professional') ? stored : 'professional';
  });

  const isProfessional = appMode === 'professional';
  const setMode = (mode: AppMode) => {
    setAppMode(mode);
    localStorage.setItem(XPLR_STORAGE_MODE, mode);
  };

  // Grade and filters
  const [userGrade, setUserGrade] = useState<{ grade: string; fee_percent: string } | null>(null);
  const [transactionFilters, setTransactionFilters] = useState({
    start_date: '',
    end_date: '',
    transaction_type: '',
    status: '',
    search: ''
  });

  // Modal states
  const [showCreateCardModal, setShowCreateCardModal] = useState(false);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [showAutoReplenishModal, setShowAutoReplenishModal] = useState(false);
  const [selectedCardId, setSelectedCardId] = useState<number | null>(null);
  const [newCardType, setNewCardType] = useState<'VISA' | 'MasterCard'>('VISA');
  const [newCardCategory, setNewCardCategory] = useState<'arbitrage' | 'travel' | 'services'>('arbitrage');
  const [newCardNickname, setNewCardNickname] = useState('');
  const [isCreatingCard, setIsCreatingCard] = useState(false);
  const [newCardCount, setNewCardCount] = useState(1);

  // activeSection derived from appMode
  const activeSection: 'arbitrage' | 'personal' = isProfessional ? 'arbitrage' : 'personal';

  // Mobile sidebar
  const [sidebarOpen, setSidebarOpen] = useState(false);
  
  // Top-up state
  const [isTopingUp, setIsTopingUp] = useState(false);
  const [topUpWallet, setTopUpWallet] = useState<'arbitrage' | 'personal'>('arbitrage');

  // Card details reveal state
  const [revealedCardDetails, setRevealedCardDetails] = useState<Record<number, { full_number: string; cvv: string; expiry: string } | null>>({});
  const [loadingCardDetails, setLoadingCardDetails] = useState<Record<number, boolean>>({});

  // Auto-replenish states
  const [autoReplenishThreshold, setAutoReplenishThreshold] = useState('');
  const [autoReplenishAmount, setAutoReplenishAmount] = useState('');
  const [isSettingAutoReplenish, setIsSettingAutoReplenish] = useState(false);

  // Telegram connect state
  const [telegramChatId, setTelegramChatId] = useState('');
  const [isSavingTelegram, setIsSavingTelegram] = useState(false);
  const [showTelegramInput, setShowTelegramInput] = useState(false);

  // Spend stats for pie chart
  const [spendStats, setSpendStats] = useState<{ category: string; total_spent: string; tx_count: number }[]>([]);

  // Exchange rates
  const [exchangeRates, setExchangeRates] = useState<{ currency_from: string; currency_to: string; final_rate: string }[]>([]);

  // Limit edit modal
  const [showLimitModal, setShowLimitModal] = useState(false);
  const [limitCardId, setLimitCardId] = useState<number | null>(null);
  const [limitValue, setLimitValue] = useState('');
  const [isSavingLimit, setIsSavingLimit] = useState(false);

  // Toast notification (success/error for block/unblock)
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  useEffect(() => {
    fetchDashboardData();

    // Poll for new transactions every 30 seconds
    const transactionInterval = setInterval(() => {
      fetchDashboardData();
    }, 30000);

    return () => clearInterval(transactionInterval);
  }, []);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 3500);
    return () => clearTimeout(t);
  }, [toast]);

  const fetchDashboardData = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        navigate('/login');
        return;
      }

      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };

      // Fetch user data
      const userResponse = await axios.get(`${API_BASE_URL}/user/me`, config);
      setUserData(userResponse.data);
      
      // Fetch grade info
      try {
        const gradeResponse = await axios.get(`${API_BASE_URL}/user/grade`, config);
        setUserGrade({
          grade: gradeResponse.data.grade,
          fee_percent: gradeResponse.data.fee_percent
        });
      } catch (error) {
        console.error('Error fetching grade:', error);
      }

      // Fetch cards data
      try {
        const cardsResponse = await axios.get(`${API_BASE_URL}/user/cards`, config);
        setCards(Array.isArray(cardsResponse.data) ? cardsResponse.data : []);
      } catch (error) {
        console.error('Error fetching cards:', error);
        setCards([]);
      }

      // Fetch transactions data with filters
      try {
        const params = new URLSearchParams();
        if (transactionFilters.start_date) params.append('start_date', transactionFilters.start_date);
        if (transactionFilters.end_date) params.append('end_date', transactionFilters.end_date);
        if (transactionFilters.transaction_type) params.append('transaction_type', transactionFilters.transaction_type);
        if (transactionFilters.status) params.append('status', transactionFilters.status);
        if (transactionFilters.search) params.append('search', transactionFilters.search);
        
        const transactionsResponse = await axios.get(
          `${API_BASE_URL}/user/report${params.toString() ? '?' + params.toString() : ''}`,
          config
        );
        setTransactions(Array.isArray(transactionsResponse.data?.transactions) ? transactionsResponse.data.transactions : []);
      } catch (error) {
        console.error('Error fetching transactions:', error);
        setTransactions([]);
      }

      // Fetch spend stats for pie chart
      try {
        const statsResponse = await axios.get(`${API_BASE_URL}/user/stats`, config);
        setSpendStats(Array.isArray(statsResponse.data?.categories) ? statsResponse.data.categories : []);
      } catch (error) {
        console.error('Error fetching stats:', error);
      }

      // Fetch exchange rates (public, no auth needed)
      try {
        const ratesResponse = await axios.get(`${API_BASE_URL}/rates`);
        setExchangeRates(Array.isArray(ratesResponse.data) ? ratesResponse.data : []);
      } catch (error) {
        console.error('Error fetching rates:', error);
      }

      setIsLoading(false);
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
      setIsLoading(false);
      // If unauthorized, redirect to login
      if (axios.isAxiosError(error) && error.response?.status === 401) {
        localStorage.removeItem('token');
        navigate('/login');
      }
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  // Create new card
  const handleCreateCard = async () => {
    console.log('[CREATE CARD] Starting card creation...');
    console.log('[CREATE CARD] Nickname:', newCardNickname);
    console.log('[CREATE CARD] Type:', newCardType);

    if (!newCardNickname.trim()) {
      setToast({ message: 'Пожалуйста, введите название карты', type: 'error' });
      return;
    }

    setIsCreatingCard(true);
    try {
      const token = localStorage.getItem('token');
      console.log('[CREATE CARD] Token exists:', !!token);

      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };

      const requestData = {
        count: newCardCount,
        nickname: newCardNickname,
        daily_limit: 500,
        merchant_name: newCardNickname || 'Default Merchant',
        card_type: newCardType,
        category: newCardCategory
      };
      console.log('[CREATE CARD] Sending request:', requestData);

      // Make real API call to create card
      const response = await axios.post(`${API_BASE_URL}/user/cards/issue`, requestData, config);
      console.log('[CREATE CARD] Response received:', response.data);

      console.log('[CREATE CARD] Refreshing dashboard data...');
      // Refresh dashboard data to show new card
      await fetchDashboardData();
      console.log('[CREATE CARD] Dashboard data refreshed successfully');

      // Show success message
      setToast({ message: 'Карта успешно создана!', type: 'success' });

      // Close modal and reset form AFTER successful creation with a small delay
      // to avoid React DOM conflicts
      setTimeout(() => {
        setShowCreateCardModal(false);
        setNewCardNickname('');
        setNewCardType('VISA');
        setNewCardCategory('arbitrage');
        setNewCardCount(1);
      }, 100);
    } catch (error) {
      console.error('[CREATE CARD] Error:', error);
      if (axios.isAxiosError(error)) {
        const errorMsg = error.response?.data?.message || 'Не удалось создать карту';
        console.error('[CREATE CARD] API Error:', errorMsg);
        setToast({ message: errorMsg, type: 'error' });
      } else {
        setToast({ message: 'Не удалось создать карту', type: 'error' });
      }
    } finally {
      setIsCreatingCard(false);
      console.log('[CREATE CARD] Process completed');
    }
  };

  // Change card status (block/unblock/freeze/unfreeze/close)
  const handleToggleCardBlock = async (cardId: number, currentStatus: string) => {
    let newStatus: string;
    if (currentStatus === 'BLOCKED') {
      newStatus = 'ACTIVE';
    } else if (currentStatus === 'FROZEN') {
      newStatus = 'ACTIVE';
    } else if (currentStatus === 'ACTIVE_TO_FREEZE') {
      newStatus = 'FROZEN';
    } else if (currentStatus === 'CLOSE_CARD') {
      newStatus = 'CLOSED';
    } else {
      newStatus = currentStatus === 'ACTIVE' ? 'BLOCKED' : 'ACTIVE';
    }
    try {
      const token = localStorage.getItem('token');
      const config = {
        headers: { Authorization: `Bearer ${token}` }
      };

      await axios.patch(
        `${API_BASE_URL}/user/cards/${cardId}/status`,
        { status: newStatus },
        config
      );

      setShowConfirmDialog(false);
      setSelectedCardId(null);

      // Instant UI update
      setCards((prev) =>
        prev.map((c) =>
          c.id === cardId ? { ...c, card_status: newStatus } : c
        )
      );

      const msgs: Record<string, string> = {
        BLOCKED: 'Card blocked successfully.',
        ACTIVE: 'Card activated successfully.',
        FROZEN: 'Card frozen successfully.',
        CLOSED: 'Card closed permanently.',
      };
      setToast({
        message: msgs[newStatus] || `Card status changed to ${newStatus}`,
        type: 'success'
      });
    } catch (error) {
      console.error('Error toggling card status:', error);
      let msg = 'Failed to update card status';
      if (axios.isAxiosError(error) && error.response?.data != null) {
        const d = error.response.data;
        msg = typeof d === 'string' ? d : (d as { message?: string }).message || msg;
      }
      setToast({ message: msg, type: 'error' });
    }
  };

  // Open auto-replenish modal
  const openAutoReplenishModal = (cardId: number) => {
    setSelectedCardId(cardId);
    const card = (cards ?? []).find(c => c.id === cardId);
    if (card) {
      setAutoReplenishThreshold(card.auto_replenish_threshold || '');
      setAutoReplenishAmount(card.auto_replenish_amount || '');
    }
    setShowAutoReplenishModal(true);
  };

  // Handle auto-replenish setup
  const handleSetAutoReplenish = async () => {
    if (!selectedCardId) return;
    
    const threshold = parseFloat(autoReplenishThreshold);
    const amount = parseFloat(autoReplenishAmount);

    if (isNaN(threshold) || threshold <= 0) {
      setToast({ message: 'Порог должен быть больше 0', type: 'error' });
      return;
    }
    if (isNaN(amount) || amount <= 0) {
      setToast({ message: 'Сумма пополнения должна быть больше 0', type: 'error' });
      return;
    }

    setIsSettingAutoReplenish(true);
    try {
      await setCardAutoReplenishment(selectedCardId, {
        enabled: true,
        threshold,
        amount
      });
      setToast({ message: 'Автопополнение настроено успешно!', type: 'success' });
      setShowAutoReplenishModal(false);
      setAutoReplenishThreshold('');
      setAutoReplenishAmount('');
      await fetchDashboardData();
    } catch (error) {
      console.error('Error setting auto-replenish:', error);
      setToast({ message: 'Не удалось настроить автопополнение', type: 'error' });
    } finally {
      setIsSettingAutoReplenish(false);
    }
  };

  // Handle disable auto-replenish
  const handleDisableAutoReplenish = async (cardId: number) => {
    try {
      await unsetCardAutoReplenishment(cardId);
      setToast({ message: 'Автопополнение отключено', type: 'success' });
      await fetchDashboardData();
    } catch (error) {
      console.error('Error disabling auto-replenish:', error);
      setToast({ message: 'Не удалось отключить автопополнение', type: 'error' });
    }
  };

  // Save Telegram Chat ID
  const handleSaveTelegram = async () => {
    if (!telegramChatId.trim()) return;
    setIsSavingTelegram(true);
    try {
      const token = localStorage.getItem('token');
      const config = { headers: { Authorization: `Bearer ${token}` } };
      await axios.post(`${API_BASE_URL}/user/settings/telegram`, { chat_id: parseInt(telegramChatId) }, config);
      setToast({ message: 'Telegram connected successfully!', type: 'success' });
      setShowTelegramInput(false);
    } catch (error) {
      console.error('Error saving Telegram chat ID:', error);
      setToast({ message: 'Failed to connect Telegram', type: 'error' });
    } finally {
      setIsSavingTelegram(false);
    }
  };

  // Open limit edit modal
  const openLimitModal = (cardId: number, currentLimit: number) => {
    setLimitCardId(cardId);
    setLimitValue(String(currentLimit || 0));
    setShowLimitModal(true);
  };

  // Save card spend limit
  const handleSaveLimit = async () => {
    if (limitCardId === null) return;
    setIsSavingLimit(true);
    try {
      const token = localStorage.getItem('token');
      const config = { headers: { Authorization: `Bearer ${token}` } };
      await axios.patch(`${API_BASE_URL}/user/cards/${limitCardId}/limit`, { limit: parseFloat(limitValue) || 0 }, config);
      setCards(prev => prev.map(c => c.id === limitCardId ? { ...c, daily_spend_limit: parseFloat(limitValue) || 0 } : c));
      setToast({ message: 'Spend limit updated', type: 'success' });
      setShowLimitModal(false);
    } catch (error) {
      console.error('Error saving limit:', error);
      setToast({ message: 'Failed to update limit', type: 'error' });
    } finally {
      setIsSavingLimit(false);
    }
  };

  // Top up balance
  const handleTopUp = async () => {
    setIsTopingUp(true);
    try {
      const token = localStorage.getItem('token');
      const config = { headers: { Authorization: `Bearer ${token}` } };
      const response = await axios.post(`${API_BASE_URL}/user/topup`, { wallet: topUpWallet }, config);
      const amountUsd = response.data.amount_usd;
      const amountRub = response.data.amount_rub;
      const rate = response.data.rate;
      const balArb = response.data.balance_arbitrage;
      const balPers = response.data.balance_personal;
      if (userData) {
        setUserData({
          ...userData,
          balance_arbitrage: balArb ? parseFloat(balArb) : userData.balance_arbitrage,
          balance_personal: balPers ? parseFloat(balPers) : userData.balance_personal,
        });
      }
      const walletLabel = topUpWallet === 'arbitrage' ? 'Арбитраж' : 'Личный';
      const rateInfo = rate ? ` (${parseFloat(amountRub).toFixed(0)} ₽ × ${parseFloat(rate).toFixed(2)})` : '';
      setToast({ message: `${walletLabel}: +$${parseFloat(amountUsd || '100').toFixed(2)}${rateInfo}`, type: 'success' });
    } catch (error) {
      console.error('Error topping up:', error);
      setToast({ message: 'Не удалось пополнить баланс', type: 'error' });
    } finally {
      setIsTopingUp(false);
    }
  };

  // Reveal card details (PAN, CVV, expiry)
  const handleRevealCardDetails = async (cardId: number) => {
    // Toggle off if already revealed
    if (revealedCardDetails[cardId]) {
      setRevealedCardDetails(prev => ({ ...prev, [cardId]: null }));
      return;
    }
    setLoadingCardDetails(prev => ({ ...prev, [cardId]: true }));
    try {
      const token = localStorage.getItem('token');
      const config = { headers: { Authorization: `Bearer ${token}` } };
      const response = await axios.get(`${API_BASE_URL}/user/cards/${cardId}/mock-details`, config);
      setRevealedCardDetails(prev => ({ ...prev, [cardId]: response.data }));
    } catch (error) {
      console.error('Error fetching card details:', error);
      setToast({ message: 'Failed to load card details', type: 'error' });
    } finally {
      setLoadingCardDetails(prev => ({ ...prev, [cardId]: false }));
    }
  };

  // Copy card number to clipboard
  const handleCopyCardNumber = (cardId: number, last4: string) => {
    const details = revealedCardDetails[cardId];
    const number = details ? details.full_number : `•••• •••• •••• ${last4}`;
    navigator.clipboard.writeText(number.replace(/\s/g, ''));
    setToast({ message: 'Номер скопирован. Используйте его для оплаты в Facebook/Google.', type: 'success' });
  };

  // Open confirmation dialog
  const openBlockConfirmation = (cardId: number) => {
    setSelectedCardId(cardId);
    setShowConfirmDialog(true);
  };

  if (isLoading) {
    return (
      <div style={{
        width: '100vw',
        height: '100vh',
        backgroundColor: theme.colors.background,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        margin: 0,
        padding: 0
      }}>
        <div style={{
          color: theme.colors.textPrimary,
          fontSize: '24px',
          fontWeight: '600',
          letterSpacing: '-1px'
        }}>
          Loading...
        </div>
      </div>
    );
  }

  // Chart data for last 7 days spending
  const chartData = {
    labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    datasets: [
      {
        label: 'Daily Spend',
        data: [450, 680, 520, 890, 1240, 760, 950],
        borderColor: theme.colors.accent,
        backgroundColor: (context: any) => {
          const ctx = context.chart.ctx;
          const gradient = ctx.createLinearGradient(0, 0, 0, 300);
          gradient.addColorStop(0, 'rgba(6, 182, 212, 0.35)');
          gradient.addColorStop(1, 'rgba(6, 182, 212, 0)');
          return gradient;
        },
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 6,
        pointHoverBackgroundColor: theme.colors.accent,
        borderWidth: 2
      }
    ]
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: theme.colors.backgroundElevated,
        titleColor: theme.colors.textPrimary,
        bodyColor: theme.colors.accent,
        borderColor: theme.colors.border,
        borderWidth: 1,
        padding: 12,
        displayColors: false,
        callbacks: {
          label: (context: any) => `₽${context.parsed.y}`
        }
      }
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: { color: theme.colors.textSecondary, font: { size: 11 } },
        border: { display: false }
      },
      y: {
        grid: { color: theme.colors.border, drawBorder: false },
        ticks: {
          color: theme.colors.textSecondary,
          font: { size: 11 },
          callback: (value: any) => `₽${value}`
        },
        border: { display: false }
      }
    }
  };

  const sidebarMenuItems: Array<'dashboard' | 'cards' | 'finance' | 'team' | 'api'> = isProfessional
    ? ['dashboard', 'cards', 'finance', 'team', 'api']
    : ['dashboard', 'cards', 'finance'];

  // BIN options for arbitrage mass issue
  const arbBins = [
    { bin: '487013', label: 'VISA 487013', fee: 5.00, topup: 2.5 },
    { bin: '539453', label: 'MC 539453', fee: 4.50, topup: 2.0 },
    { bin: '414720', label: 'VISA 414720', fee: 6.00, topup: 3.0 },
  ];
  const [selectedBin, setSelectedBin] = useState(arbBins[0].bin);

  return (
    <div style={{
      width: '100vw',
      height: '100vh',
      backgroundColor: 'rgba(255, 255, 255, 0.03)',
      color: theme.colors.textPrimary,
      display: 'flex',
      overflow: 'hidden',
      fontFamily: theme.fonts.regular,
      margin: 0,
      padding: 0,
      position: 'fixed',
      top: 0,
      left: 0,
      zIndex: 10
    }}>
      {/* Mobile burger button */}
      <button
        onClick={() => setSidebarOpen(!sidebarOpen)}
        style={{
          display: 'none',
          position: 'fixed',
          top: '12px',
          left: '12px',
          zIndex: 10002,
          padding: '10px 12px',
          backgroundColor: 'rgba(255,255,255,0.08)',
          backdropFilter: 'blur(12px)',
          border: '1px solid rgba(255,255,255,0.1)',
          borderRadius: '10px',
          color: '#fff',
          fontSize: '18px',
          cursor: 'pointer',
          ...(typeof window !== 'undefined' && window.innerWidth < 768 ? { display: 'block' } : {})
        }}
      >
        {sidebarOpen ? '✕' : '☰'}
      </button>
      {/* Sidebar */}
      <aside style={{
        width: '260px',
        minWidth: '260px',
        backgroundColor: 'rgba(255, 255, 255, 0.04)',
        backdropFilter: 'blur(16px)',
        borderRight: `1px solid ${theme.colors.border}`,
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        ...(typeof window !== 'undefined' && window.innerWidth < 768 ? {
          position: 'fixed' as const,
          top: 0,
          left: sidebarOpen ? 0 : -280,
          height: '100vh',
          zIndex: 10001,
          transition: 'left 0.3s ease',
          boxShadow: sidebarOpen ? '4px 0 20px rgba(0,0,0,0.5)' : 'none'
        } : {})
      }}>
        <div style={{
          fontSize: '20px',
          fontWeight: '800',
          color: theme.colors.accent,
          marginBottom: '16px',
          letterSpacing: '2px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px'
        }}>
          ✦ XPLR
        </div>

        {/* Mode Toggle: Professional / Personal */}
        <div style={{
          marginBottom: '28px',
          padding: '6px',
          backgroundColor: theme.colors.backgroundCard,
          borderRadius: theme.borderRadius.md,
          border: `1px solid ${theme.colors.border}`,
          display: 'flex',
          gap: '4px'
        }}>
          <button
            type="button"
            onClick={() => setMode('professional')}
            style={{
              flex: 1,
              padding: '10px 12px',
              border: 'none',
              borderRadius: '8px',
              fontSize: '12px',
              fontWeight: 600,
              cursor: 'pointer',
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              transition: 'all 0.2s',
              backgroundColor: isProfessional ? theme.colors.accent : 'transparent',
              color: isProfessional ? theme.colors.background : theme.colors.textSecondary
            }}
            onMouseEnter={(e) => {
              if (!isProfessional) {
                e.currentTarget.style.color = theme.colors.textPrimary;
                e.currentTarget.style.backgroundColor = theme.colors.accentMuted;
              }
            }}
            onMouseLeave={(e) => {
              if (!isProfessional) {
                e.currentTarget.style.color = theme.colors.textSecondary;
                e.currentTarget.style.backgroundColor = 'transparent';
              }
            }}
          >
            ARBITRAGE
          </button>
          <button
            type="button"
            onClick={() => setMode('personal')}
            style={{
              flex: 1,
              padding: '10px 12px',
              border: 'none',
              borderRadius: '8px',
              fontSize: '12px',
              fontWeight: 600,
              cursor: 'pointer',
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              transition: 'all 0.2s',
              backgroundColor: !isProfessional ? theme.colors.accent : 'transparent',
              color: !isProfessional ? theme.colors.background : theme.colors.textSecondary
            }}
            onMouseEnter={(e) => {
              if (isProfessional) {
                e.currentTarget.style.color = theme.colors.textPrimary;
                e.currentTarget.style.backgroundColor = theme.colors.accentMuted;
              }
            }}
            onMouseLeave={(e) => {
              if (isProfessional) {
                e.currentTarget.style.color = theme.colors.textSecondary;
                e.currentTarget.style.backgroundColor = 'transparent';
              }
            }}
          >
            PERSONAL
          </button>
        </div>

        {sidebarMenuItems.map((item) => (
          <div
            key={item}
            onClick={() => {
              const routes: Record<string, string> = { dashboard: '/dashboard', cards: '/cards', finance: '/finance', team: '/teams', api: '/api' };
              navigate(routes[item] || '/dashboard');
            }}
            style={{
              padding: '12px 15px',
              color: activeMenu === item ? theme.colors.accent : theme.colors.textSecondary,
              cursor: 'pointer',
              borderRadius: theme.borderRadius.sm,
              marginBottom: '4px',
              transition: '0.2s',
              fontSize: '14px',
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              backgroundColor: activeMenu === item ? theme.colors.accentMuted : 'transparent',
              borderLeft: activeMenu === item ? `3px solid ${theme.colors.accent}` : '3px solid transparent'
            }}
            onMouseEnter={(e) => {
              if (activeMenu !== item) {
                e.currentTarget.style.backgroundColor = theme.colors.backgroundCard;
                e.currentTarget.style.color = theme.colors.textPrimary;
              }
            }}
            onMouseLeave={(e) => {
              if (activeMenu !== item) {
                e.currentTarget.style.backgroundColor = 'transparent';
                e.currentTarget.style.color = theme.colors.textSecondary;
              }
            }}
          >
            {item === 'dashboard' && '📊 Dashboard'}
            {item === 'cards' && '💳 Cards'}
            {item === 'finance' && '💸 Finance'}
            {item === 'team' && '👥 Teams'}
            {item === 'api' && '🔌 API & Trackers'}
          </div>
        ))}

        {/* Referrals link */}
        <div
          onClick={() => navigate('/referrals')}
          style={{
            padding: '12px 15px',
            color: theme.colors.textSecondary,
            cursor: 'pointer',
            borderRadius: theme.borderRadius.sm,
            marginBottom: '4px',
            transition: '0.2s',
            fontSize: '14px',
            display: 'flex',
            alignItems: 'center',
            gap: '10px',
            backgroundColor: 'transparent',
            borderLeft: '3px solid transparent'
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.backgroundColor = theme.colors.backgroundCard;
            e.currentTarget.style.color = theme.colors.textPrimary;
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.backgroundColor = 'transparent';
            e.currentTarget.style.color = theme.colors.textSecondary;
          }}
        >
          ✈️ Referrals
        </div>

        {/* Telegram Connect */}
        <div style={{
          marginTop: 'auto',
          marginBottom: '8px',
          padding: '12px',
          backgroundColor: theme.colors.backgroundCard,
          borderRadius: theme.borderRadius.md,
          border: `1px solid ${theme.colors.border}`
        }}>
          {!showTelegramInput ? (
            <button
              onClick={() => setShowTelegramInput(true)}
              style={{
                width: '100%',
                padding: '10px',
                backgroundColor: 'rgba(0, 136, 204, 0.15)',
                border: '1px solid rgba(0, 136, 204, 0.3)',
                borderRadius: '8px',
                color: '#0088cc',
                fontSize: '12px',
                fontWeight: '600',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '6px'
              }}
            >
              ✈️ Connect Telegram
            </button>
          ) : (
            <div>
              <div style={{ fontSize: '11px', color: theme.colors.textSecondary, marginBottom: '8px' }}>
                Enter your Telegram Chat ID:
              </div>
              <input
                type="text"
                placeholder="e.g. 123456789"
                value={telegramChatId}
                onChange={(e) => setTelegramChatId(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 10px',
                  backgroundColor: 'rgba(255,255,255,0.05)',
                  border: '1px solid rgba(255,255,255,0.1)',
                  borderRadius: '6px',
                  color: '#fff',
                  fontSize: '13px',
                  outline: 'none',
                  marginBottom: '8px',
                  boxSizing: 'border-box'
                }}
              />
              <div style={{ display: 'flex', gap: '6px' }}>
                <button
                  onClick={() => setShowTelegramInput(false)}
                  style={{
                    flex: 1, padding: '7px', fontSize: '11px',
                    backgroundColor: 'transparent', border: `1px solid ${theme.colors.border}`,
                    borderRadius: '6px', color: theme.colors.textSecondary, cursor: 'pointer'
                  }}
                >Cancel</button>
                <button
                  onClick={handleSaveTelegram}
                  disabled={isSavingTelegram || !telegramChatId.trim()}
                  style={{
                    flex: 1, padding: '7px', fontSize: '11px',
                    backgroundColor: '#0088cc', border: 'none',
                    borderRadius: '6px', color: '#fff', fontWeight: '600',
                    cursor: isSavingTelegram ? 'not-allowed' : 'pointer',
                    opacity: isSavingTelegram || !telegramChatId.trim() ? 0.5 : 1
                  }}
                >{isSavingTelegram ? '...' : 'Save'}</button>
              </div>
            </div>
          )}
        </div>

        <div
          onClick={handleLogout}
          style={{
            padding: '12px 15px',
            color: theme.colors.textSecondary,
            cursor: 'pointer',
            borderRadius: theme.borderRadius.sm,
            fontSize: '14px',
            display: 'flex',
            alignItems: 'center',
            gap: '10px'
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.backgroundColor = theme.colors.error + '20';
            e.currentTarget.style.color = theme.colors.error;
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.backgroundColor = 'transparent';
            e.currentTarget.style.color = theme.colors.textSecondary;
          }}
        >
          ⚙️ Logout
        </div>
      </aside>

      {/* Main Content */}
      <div style={{
        flex: 1,
        padding: '30px',
        overflowY: 'auto',
        backgroundColor: 'rgba(255, 255, 255, 0.02)',
        backdropFilter: 'blur(12px)'
      }}>
        {/* Header */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '30px'
        }}>
          <div>
            <h1 style={{ margin: 0, fontSize: '24px', fontWeight: '700', color: theme.colors.textPrimary }}>
              {isProfessional ? 'Account Overview' : 'My Cards & Balance'}
            </h1>
            <p style={{ margin: '5px 0 0', color: theme.colors.textSecondary, fontSize: '14px' }}>
              {isProfessional ? `Welcome back, ${userData?.email?.split('@')[0] || 'User'}` : `Hello, ${userData?.email?.split('@')[0] || 'User'}`}
            </p>
          </div>
          <div style={{ display: 'flex', gap: '16px', alignItems: 'stretch' }}>
            {/* Arbitrage Wallet */}
            <div style={{
              background: 'rgba(30, 58, 95, 0.5)',
              backdropFilter: 'blur(20px)',
              padding: '16px 20px',
              borderRadius: theme.borderRadius.lg,
              border: topUpWallet === 'arbitrage' ? `1px solid ${theme.colors.accent}` : `1px solid ${theme.colors.border}`,
              minWidth: '170px',
              cursor: 'pointer',
              transition: '0.2s'
            }} onClick={() => setTopUpWallet('arbitrage')}>
              <div style={{ color: '#3b82f6', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '4px' }}>
                💳 Арбитраж
              </div>
              <div style={{ fontSize: '28px', fontWeight: '700', color: theme.colors.textPrimary, letterSpacing: '-1px' }}>
                ${Number(userData?.balance_arbitrage ?? userData?.balance ?? 0).toFixed(2)}
              </div>
            </div>
            {/* Personal Wallet */}
            <div style={{
              background: 'rgba(19, 78, 74, 0.5)',
              backdropFilter: 'blur(20px)',
              padding: '16px 20px',
              borderRadius: theme.borderRadius.lg,
              border: topUpWallet === 'personal' ? `1px solid #14b8a6` : `1px solid ${theme.colors.border}`,
              minWidth: '170px',
              cursor: 'pointer',
              transition: '0.2s'
            }} onClick={() => setTopUpWallet('personal')}>
              <div style={{ color: '#14b8a6', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.5px', marginBottom: '4px' }}>
                ✈️ Личный
              </div>
              <div style={{ fontSize: '28px', fontWeight: '700', color: theme.colors.textPrimary, letterSpacing: '-1px' }}>
                ${Number(userData?.balance_personal ?? 0).toFixed(2)}
              </div>
            </div>
            {/* Top Up */}
            <div style={{
              display: 'flex', flexDirection: 'column', justifyContent: 'center', gap: '8px'
            }}>
              {isProfessional && userGrade && (
                <div style={{
                  padding: '6px 10px',
                  backgroundColor: theme.colors.accentMuted,
                  borderRadius: theme.borderRadius.sm,
                  border: `1px solid ${theme.colors.accentBorder}`,
                  fontSize: '11px',
                  color: theme.colors.accent,
                  whiteSpace: 'nowrap'
                }}>
                  {userGrade.grade} • {parseFloat(userGrade.fee_percent).toFixed(1)}%
                </div>
              )}
              <button
                onClick={handleTopUp}
                disabled={isTopingUp}
                style={{
                  padding: '10px 20px',
                  backgroundColor: theme.colors.accent,
                  color: theme.colors.background,
                  border: 'none',
                  borderRadius: theme.borderRadius.sm,
                  fontWeight: '600',
                  fontSize: '13px',
                  cursor: isTopingUp ? 'not-allowed' : 'pointer',
                  opacity: isTopingUp ? 0.6 : 1,
                  transition: '0.2s',
                  whiteSpace: 'nowrap'
                }}
              >
                {isTopingUp ? '...' : `+ Top Up → ${topUpWallet === 'arbitrage' ? 'Арбитраж' : 'Личный'}`}
              </button>
            </div>
          </div>
        </div>

        {/* ============ DASHBOARD VIEW ============ */}
        {activeMenu === 'dashboard' && <>
        {/* Exchange Rates */}
        {(exchangeRates?.length ?? 0) > 0 && (
          <div style={{
            display: 'flex', gap: '12px', marginBottom: '20px', flexWrap: 'wrap'
          }}>
            {(exchangeRates ?? []).map((r, i) => (
              <div key={i} style={{
                backgroundColor: theme.colors.backgroundCard,
                border: `1px solid ${theme.colors.border}`,
                borderRadius: '12px',
                padding: '16px 20px',
                backdropFilter: 'blur(20px)',
                flex: '1 1 200px',
                minWidth: '180px'
              }}>
                <div style={{ fontSize: '11px', color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: '1px', marginBottom: '6px' }}>
                  💱 {r.currency_from} → {r.currency_to}
                </div>
                <div style={{ fontSize: '24px', fontWeight: '800', color: '#00e096' }}>
                  {parseFloat(r.final_rate).toFixed(2)}
                </div>
                <div style={{ fontSize: '11px', color: theme.colors.textSecondary, marginTop: '4px' }}>
                  1 {r.currency_to} = {parseFloat(r.final_rate).toFixed(2)} {r.currency_from}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Dashboard Grid — только в Professional (арбитраж + статистика по BIN/картам) */}
        {isProfessional && (
        <div style={{
          display: 'grid',
          gridTemplateColumns: '2fr 1fr',
          gap: '20px',
          marginBottom: '20px'
        }}>
          <div style={{
            backgroundColor: theme.colors.backgroundCard,
            backdropFilter: 'blur(20px)',
            padding: '20px',
            borderRadius: theme.borderRadius.lg,
            border: `1px solid ${theme.colors.border}`,
            height: '300px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            <Line data={chartData} options={chartOptions} />
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
            <div style={{
              background: theme.colors.backgroundCard,
              backdropFilter: 'blur(20px)',
              padding: '20px',
              borderRadius: theme.borderRadius.lg,
              border: `1px solid ${theme.colors.border}`,
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center'
            }}>
              <div style={{
                color: theme.colors.textSecondary,
                fontSize: '12px',
                textTransform: 'uppercase',
                letterSpacing: '0.5px',
                marginBottom: '10px',
                alignSelf: 'flex-start'
              }}>
                Spend by Category (30d)
              </div>
              {(spendStats?.length ?? 0) > 0 ? (
                <div style={{ width: '140px', height: '140px' }}>
                  <Pie
                    data={{
                      labels: (spendStats ?? []).map(s => {
                        const labels: Record<string, string> = { arbitrage: 'Реклама', travel: 'Путешествия', services: 'Зарубежные сервисы' };
                        return labels[s.category] || s.category;
                      }),
                      datasets: [{
                        data: (spendStats ?? []).map(s => parseFloat(s.total_spent)),
                        backgroundColor: (spendStats ?? []).map(s => {
                          const colors: Record<string, string> = { arbitrage: '#3b82f6', travel: '#14b8a6', services: '#8b5cf6' };
                          return colors[s.category] || '#6b7280';
                        }),
                        borderWidth: 0
                      }]
                    }}
                    options={{
                      responsive: true,
                      maintainAspectRatio: true,
                      plugins: {
                        legend: { display: false },
                        tooltip: {
                          backgroundColor: '#1a1a2e',
                          titleColor: '#fff',
                          bodyColor: '#00e096',
                          callbacks: { label: (ctx: any) => `$${ctx.parsed.toFixed(2)}` }
                        }
                      }
                    }}
                  />
                </div>
              ) : (
                <div style={{ color: '#666', fontSize: '12px', marginTop: '20px' }}>No spend data yet</div>
              )}
              <div style={{ display: 'flex', gap: '12px', marginTop: '10px', flexWrap: 'wrap', justifyContent: 'center' }}>
                {(spendStats ?? []).map(s => (
                  <div key={s.category} style={{ fontSize: '10px', color: '#aaa', display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <span style={{
                      width: '8px', height: '8px', borderRadius: '50%', display: 'inline-block',
                      backgroundColor: s.category === 'arbitrage' ? '#3b82f6' : s.category === 'travel' ? '#14b8a6' : '#8b5cf6'
                    }}></span>
                    {s.category === 'arbitrage' ? 'Реклама' : s.category === 'travel' ? 'Путешествия' : 'Зарубежные сервисы'}
                  </div>
                ))}
              </div>
            </div>

            <div style={{
              background: theme.colors.backgroundCard,
              backdropFilter: 'blur(20px)',
              padding: '20px',
              borderRadius: theme.borderRadius.lg,
              border: `1px solid ${theme.colors.border}`,
              flex: 0
            }}>
              <div style={{
                color: theme.colors.textSecondary,
                fontSize: '12px',
                textTransform: 'uppercase',
                letterSpacing: '0.5px'
              }}>
                Active Cards
              </div>
              <div style={{ fontSize: '24px', fontWeight: 'bold', marginTop: '10px', color: theme.colors.textPrimary }}>
                {(cards ?? []).filter(c => c.card_status === 'ACTIVE').length} <span style={{ fontSize: '14px', color: theme.colors.textSecondary }}>/ {cards?.length ?? 0}</span>
              </div>
              <div style={{ fontSize: '12px', color: theme.colors.textSecondary, marginTop: '5px' }}>
                {(cards ?? []).filter(c => c.card_status === 'BLOCKED').length} blocked · {(cards ?? []).filter(c => c.card_status === 'FROZEN').length} frozen · {(cards ?? []).filter(c => c.card_status === 'CLOSED').length} closed
              </div>
            </div>
          </div>
        </div>
        )}

        </>}

        {/* ============ CARDS VIEW ============ */}
        {(activeMenu === 'dashboard' || activeMenu === 'cards') && <>
        {/* Cards Section */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '15px'
        }}>
          <h3 style={{ margin: 0, fontSize: '18px', fontWeight: '700', color: theme.colors.textPrimary }}>My Cards</h3>
          <button
          onClick={() => { setNewCardCategory(activeSection === 'arbitrage' ? 'arbitrage' : 'travel'); setNewCardCount(activeSection === 'arbitrage' ? 1 : 1); setShowCreateCardModal(true); }}
          style={{
            backgroundColor: theme.colors.accent,
            color: theme.colors.background,
            border: 'none',
            padding: '10px 20px',
            borderRadius: theme.borderRadius.sm,
            fontWeight: '600',
            cursor: 'pointer',
            transition: '0.2s',
            fontSize: '14px'
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.opacity = '0.9';
            e.currentTarget.style.backgroundColor = theme.colors.accentHover;
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.opacity = '1';
            e.currentTarget.style.backgroundColor = theme.colors.accent;
          }}>
            {isProfessional ? '+ Массовый выпуск' : '+ Выпустить карту'}
          </button>
        </div>

        {/* Mode indicator badge */}
        <div style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: '8px',
          padding: '8px 16px',
          backgroundColor: isProfessional ? 'rgba(59, 130, 246, 0.12)' : 'rgba(20, 184, 166, 0.12)',
          border: `1px solid ${isProfessional ? 'rgba(59, 130, 246, 0.3)' : 'rgba(20, 184, 166, 0.3)'}`,
          borderRadius: '8px',
          marginBottom: '16px',
          fontSize: '13px',
          color: isProfessional ? '#3b82f6' : '#14b8a6',
          fontWeight: '600'
        }}>
          {isProfessional ? '💳 Арбитраж' : '✈️ Личные карты'}
          <span style={{ opacity: 0.6, fontWeight: '400' }}>
            {(activeSection === 'arbitrage'
              ? (cards ?? []).filter(c => (c.category || 'arbitrage') === 'arbitrage')
              : (cards ?? []).filter(c => { const cat = c.category || 'arbitrage'; return cat === 'travel' || cat === 'services'; })
            ).length} карт
          </span>
        </div>

        {/* Fee info for arbitrage */}
        {isProfessional && (
          <div style={{
            display: 'flex', gap: '12px', marginBottom: '16px', flexWrap: 'wrap'
          }}>
            {arbBins.map(b => (
              <div key={b.bin} style={{
                padding: '8px 14px',
                backgroundColor: 'rgba(255,255,255,0.04)',
                border: '1px solid rgba(255,255,255,0.08)',
                borderRadius: '8px',
                fontSize: '11px',
                color: theme.colors.textSecondary
              }}>
                <span style={{ color: '#fff', fontWeight: '600' }}>{b.label}</span>
                <span style={{ margin: '0 6px', opacity: 0.4 }}>|</span>
                Выпуск: ${b.fee.toFixed(2)}
                <span style={{ margin: '0 6px', opacity: 0.4 }}>|</span>
                Пополнение: {b.topup}%
              </div>
            ))}
          </div>
        )}

        {/* Cards Grid */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
          gap: '20px',
          marginBottom: '30px'
        }}>
          {(activeSection === 'arbitrage'
            ? (cards ?? []).filter(c => (c.category || 'arbitrage') === 'arbitrage')
            : (cards ?? []).filter(c => { const cat = c.category || 'arbitrage'; return cat === 'travel' || cat === 'services'; })
          ).length === 0 ? (
            <div style={{
              gridColumn: '1 / -1',
              textAlign: 'center',
              padding: '60px 20px',
              color: '#888c95',
              fontSize: '16px'
            }}>
              {(cards?.length ?? 0) === 0 ? 'Нет активных карт. Нажмите кнопку выпуска.' : activeSection === 'arbitrage' ? 'Нет карт для рекламы.' : 'Нет личных карт.'}
            </div>
          ) : (
            (activeSection === 'arbitrage'
              ? (cards ?? []).filter(c => (c.category || 'arbitrage') === 'arbitrage')
              : (cards ?? []).filter(c => { const cat = c.category || 'arbitrage'; return cat === 'travel' || cat === 'services'; })
            ).map((card) => {
              const catColors: Record<string, { bg: string; border: string }> = {
                arbitrage: { bg: 'rgba(30, 58, 95, 0.7)', border: 'rgba(59, 130, 246, 0.3)' },
                travel:    { bg: 'rgba(19, 78, 74, 0.7)', border: 'rgba(20, 184, 166, 0.3)' },
                services:  { bg: 'rgba(40, 40, 50, 0.7)', border: 'rgba(120, 120, 140, 0.3)' },
              };
              const cc = catColors[(card.category || 'arbitrage')] || catColors.arbitrage;
              const details = revealedCardDetails[card.id];
              const isLoadingDetails = loadingCardDetails[card.id];
              return (
            <div
              key={card.id}
              style={{
                backgroundColor: card.card_status === 'CLOSED' ? 'rgba(30, 30, 30, 0.6)'
                  : card.card_status === 'FROZEN' ? 'rgba(30, 60, 90, 0.8)'
                  : card.card_status === 'BLOCKED' ? theme.colors.backgroundCard : cc.bg,
                backdropFilter: 'blur(20px)',
                border: card.card_status === 'CLOSED' ? '1px solid rgba(80, 80, 80, 0.4)'
                  : card.card_status === 'FROZEN' ? '1px solid rgba(59, 130, 246, 0.4)'
                  : card.card_status === 'BLOCKED' ? `1px solid ${theme.colors.error}` : `1px solid ${cc.border}`,
                borderRadius: theme.borderRadius.lg,
                padding: '20px',
                transition: '0.3s',
                opacity: card.card_status === 'CLOSED' ? 0.5 : card.card_status === 'BLOCKED' ? 0.7 : 1,
                filter: card.card_status === 'CLOSED' ? 'grayscale(100%)' : 'none'
              }}
              onMouseEnter={(e) => {
                if (card.card_status !== 'BLOCKED') {
                  e.currentTarget.style.borderColor = theme.colors.accent;
                  e.currentTarget.style.transform = 'translateY(-2px)';
                }
              }}
              onMouseLeave={(e) => {
                if (card.card_status !== 'BLOCKED') {
                  e.currentTarget.style.borderColor = cc.border;
                  e.currentTarget.style.transform = 'translateY(0)';
                }
              }}
            >
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                marginBottom: '20px'
              }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ fontWeight: '700', fontStyle: 'italic', color: '#fff' }}>
                    {card.bin?.startsWith('4') ? 'VISA' : 'MasterCard'}
                  </span>
                  {isProfessional && card.bin && (
                    <span style={{ fontSize: '10px', padding: '3px 6px', borderRadius: '4px', background: 'rgba(59,130,246,0.15)', color: '#3b82f6', fontFamily: theme.fonts.mono }}>
                      BIN {card.bin}
                    </span>
                  )}
                </div>
                <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
                  <span style={{
                    fontSize: '10px',
                    padding: '4px 8px',
                    borderRadius: '4px',
                    fontWeight: '600',
                    background: 'rgba(255,255,255,0.1)',
                    color: '#ccc',
                    textTransform: 'capitalize'
                  }}>
                    {(card.category || 'arbitrage') === 'arbitrage' ? 'Реклама' : (card.category || 'arbitrage') === 'travel' ? 'Путешествия' : 'Сервисы'}
                  </span>
                  <span style={{
                    fontSize: '10px',
                    padding: '4px 8px',
                    borderRadius: '4px',
                    fontWeight: '700',
                    background: card.card_status === 'ACTIVE' ? theme.colors.accentMuted
                      : card.card_status === 'FROZEN' ? 'rgba(59, 130, 246, 0.2)'
                      : card.card_status === 'CLOSED' ? 'rgba(100, 100, 100, 0.3)'
                      : `${theme.colors.error}30`,
                    color: card.card_status === 'ACTIVE' ? theme.colors.accent
                      : card.card_status === 'FROZEN' ? '#3b82f6'
                      : card.card_status === 'CLOSED' ? '#888'
                      : theme.colors.error
                  }}>
                    {card.card_status}
                  </span>
                </div>
              </div>

              <div style={{
                fontFamily: theme.fonts.mono,
                fontSize: '18px',
                letterSpacing: '2px',
                color: theme.colors.textSecondary,
                marginBottom: '10px'
              }}>
                <span
                onClick={() => handleCopyCardNumber(card.id, card.last_4_digits)}
                style={{ cursor: 'pointer' }}
                title="Нажмите, чтобы скопировать номер"
              >
                {details ? details.full_number.replace(/(.{4})/g, '$1 ').trim() : `•••• •••• •••• ${card.last_4_digits}`}
              </span>
              </div>

              {details && (
                <div style={{
                  display: 'flex',
                  gap: '20px',
                  fontSize: '13px',
                  color: '#fff',
                  marginBottom: '10px',
                  padding: '10px 12px',
                  backgroundColor: 'rgba(0,0,0,0.25)',
                  borderRadius: '8px',
                  fontFamily: theme.fonts.mono
                }}>
                  <div><span style={{ color: '#888', fontSize: '10px' }}>CVV</span><br />{details.cvv}</div>
                  <div><span style={{ color: '#888', fontSize: '10px' }}>EXPIRY</span><br />{details.expiry}</div>
                </div>
              )}

              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                fontSize: '12px',
                color: theme.colors.textSecondary,
                marginBottom: '5px'
              }}>
                <span>{card.nickname}</span>
                <span>₽{card.daily_spend_limit || 0} Daily Limit</span>
              </div>

              {card.auto_replenish_enabled && (card.category || 'arbitrage') === 'arbitrage' && (
                <div style={{
                  fontSize: '11px',
                  color: theme.colors.accent,
                  marginBottom: '10px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '5px'
                }}>
                  <span>🔄</span>
                  <span>Auto-top-up: ${card.auto_replenish_threshold || 0} → ${card.auto_replenish_amount || 0}</span>
                </div>
              )}
              {(card.category || 'arbitrage') !== 'arbitrage' && (
                <div style={{
                  fontSize: '11px',
                  color: '#14b8a6',
                  marginBottom: '10px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '5px'
                }}>
                  <span>📅</span>
                  <span>Срок действия: 1 год</span>
                </div>
              )}

              <div style={{ marginTop: '15px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
                <button
                onClick={() => handleRevealCardDetails(card.id)}
                disabled={isLoadingDetails}
                style={{
                  background: details ? theme.colors.accentMuted : 'transparent',
                  border: `1px solid ${details ? theme.colors.accent : theme.colors.border}`,
                  color: details ? theme.colors.accent : theme.colors.textSecondary,
                  padding: '5px 10px',
                  borderRadius: '8px',
                  fontSize: '11px',
                  cursor: isLoadingDetails ? 'wait' : 'pointer'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = theme.colors.accent;
                  e.currentTarget.style.color = theme.colors.accent;
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = details ? theme.colors.accent : theme.colors.border;
                  e.currentTarget.style.color = details ? theme.colors.accent : theme.colors.textSecondary;
                }}>
                  {isLoadingDetails ? '...' : details ? '🔒 Hide' : '👁 Show Details'}
                </button>
                {card.card_status !== 'CLOSED' && (
                <button
                onClick={() => openLimitModal(card.id, card.daily_spend_limit)}
                style={{
                  background: 'transparent',
                  border: `1px solid ${theme.colors.border}`,
                  color: theme.colors.textSecondary,
                  padding: '5px 10px',
                  borderRadius: '8px',
                  fontSize: '11px',
                  cursor: 'pointer'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = theme.colors.accent;
                  e.currentTarget.style.color = theme.colors.accent;
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = theme.colors.border;
                  e.currentTarget.style.color = theme.colors.textSecondary;
                }}>
                  ⚡ Limit: ${card.daily_spend_limit || 0}
                </button>
                )}
                {(card.category || 'arbitrage') === 'arbitrage' && (
                <>
                <button
                onClick={() => openAutoReplenishModal(card.id)}
                style={{
                  background: card.auto_replenish_enabled ? theme.colors.accentMuted : 'transparent',
                  border: `1px solid ${card.auto_replenish_enabled ? theme.colors.accent : theme.colors.border}`,
                  color: card.auto_replenish_enabled ? theme.colors.accent : theme.colors.textSecondary,
                  padding: '5px 10px',
                  borderRadius: '8px',
                  fontSize: '11px',
                  cursor: 'pointer'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = theme.colors.accent;
                  e.currentTarget.style.color = theme.colors.accent;
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = card.auto_replenish_enabled ? theme.colors.accent : theme.colors.border;
                  e.currentTarget.style.color = card.auto_replenish_enabled ? theme.colors.accent : theme.colors.textSecondary;
                }}>
                  {card.auto_replenish_enabled ? '⚙️ Auto-top-up' : 'Auto-top-up'}
                </button>
                {card.auto_replenish_enabled && (
                  <button
                  onClick={() => handleDisableAutoReplenish(card.id)}
                  style={{
                    background: 'transparent',
                    border: `1px solid ${theme.colors.error}50`,
                    color: theme.colors.error,
                    padding: '5px 10px',
                    borderRadius: '8px',
                    fontSize: '11px',
                    cursor: 'pointer'
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.borderColor = theme.colors.error;
                    e.currentTarget.style.backgroundColor = theme.colors.error + '20';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.borderColor = theme.colors.error + '50';
                    e.currentTarget.style.backgroundColor = 'transparent';
                  }}>
                    Disable
                  </button>
                )}
                </>)}
                {card.card_status !== 'CLOSED' && (
                <button
                onClick={() => handleToggleCardBlock(card.id, card.card_status === 'FROZEN' ? 'FROZEN' : 'ACTIVE_TO_FREEZE')}
                style={{
                  background: card.card_status === 'FROZEN' ? 'rgba(59, 130, 246, 0.15)' : 'transparent',
                  border: card.card_status === 'FROZEN' ? '1px solid rgba(59, 130, 246, 0.5)' : `1px solid ${theme.colors.border}`,
                  color: card.card_status === 'FROZEN' ? '#3b82f6' : theme.colors.textSecondary,
                  padding: '5px 10px',
                  borderRadius: '8px',
                  fontSize: '11px',
                  cursor: 'pointer'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = '#3b82f6';
                  e.currentTarget.style.color = '#3b82f6';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = card.card_status === 'FROZEN' ? 'rgba(59, 130, 246, 0.5)' : theme.colors.border;
                  e.currentTarget.style.color = card.card_status === 'FROZEN' ? '#3b82f6' : theme.colors.textSecondary;
                }}>
                  {card.card_status === 'FROZEN' ? '☀️ Unfreeze' : '❄️ Freeze'}
                </button>
                )}
                {card.card_status !== 'CLOSED' && (
                <button
                onClick={() => {
                  if (window.confirm('Are you sure you want to close this card? This action cannot be undone.')) {
                    handleToggleCardBlock(card.id, 'CLOSE_CARD');
                  }
                }}
                style={{
                  background: 'transparent',
                  border: `1px solid ${theme.colors.border}`,
                  color: theme.colors.textSecondary,
                  padding: '5px 10px',
                  borderRadius: '8px',
                  fontSize: '11px',
                  cursor: 'pointer'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = theme.colors.error;
                  e.currentTarget.style.color = theme.colors.error;
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = theme.colors.border;
                  e.currentTarget.style.color = theme.colors.textSecondary;
                }}>
                  ✕ Close
                </button>
                )}
                {card.card_status === 'CLOSED' && (
                  <span style={{ fontSize: '11px', color: '#666', fontStyle: 'italic' }}>Card closed</span>
                )}
              </div>
            </div>
              );
            })
          )}
        </div>

        </>}

        {/* ============ FINANCE VIEW ============ */}
        {(activeMenu === 'dashboard' || activeMenu === 'finance') && <>

        {/* Finance Wallet Toggle (only on dedicated finance page) */}
        {activeMenu === 'finance' && (
          <div style={{ display: 'flex', gap: '8px', marginBottom: '20px' }}>
            {(['arbitrage', 'personal'] as const).map(w => (
              <button key={w} onClick={() => setTopUpWallet(w)} style={{
                padding: '10px 20px',
                backgroundColor: topUpWallet === w ? (w === 'arbitrage' ? 'rgba(59, 130, 246, 0.2)' : 'rgba(20, 184, 166, 0.2)') : 'rgba(255,255,255,0.03)',
                border: `1px solid ${topUpWallet === w ? (w === 'arbitrage' ? '#3b82f6' : '#14b8a6') : theme.colors.border}`,
                borderRadius: '10px', color: topUpWallet === w ? (w === 'arbitrage' ? '#3b82f6' : '#14b8a6') : theme.colors.textSecondary,
                fontSize: '13px', fontWeight: '600', cursor: 'pointer', transition: '0.2s'
              }}>
                {w === 'arbitrage' ? '💳 Арбитраж' : '✈️ Личный'}
                <span style={{ marginLeft: '8px', fontWeight: '800' }}>
                  ${Number(w === 'arbitrage' ? (userData?.balance_arbitrage ?? 0) : (userData?.balance_personal ?? 0)).toFixed(2)}
                </span>
              </button>
            ))}
          </div>
        )}

        {/* Transactions Table */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '15px'
        }}>
          <h3 style={{ margin: 0, fontSize: '18px', fontWeight: '700' }}>
            Recent Transactions
          </h3>
          <button
            onClick={() => {
              // Применить фильтры
              fetchDashboardData();
            }}
            style={{
              padding: '8px 16px',
              backgroundColor: theme.colors.accentMuted,
              border: `1px solid ${theme.colors.accentBorder}`,
              borderRadius: theme.borderRadius.sm,
              color: theme.colors.accent,
              fontSize: '12px',
              cursor: 'pointer',
              fontWeight: '600'
            }}>
            🔍 Фильтры
          </button>
        </div>
        
        {/* Filters Panel */}
        <div style={{
          backgroundColor: theme.colors.backgroundCard,
          borderRadius: theme.borderRadius.lg,
          padding: '20px',
          marginBottom: '20px',
          border: `1px solid ${theme.colors.border}`,
          display: 'flex',
          flexWrap: 'wrap',
          gap: '15px',
          alignItems: 'flex-end'
        }}>
          <div style={{ flex: '1 1 200px' }}>
            <label style={{
              display: 'block',
              fontSize: '11px',
              color: '#888c95',
              marginBottom: '5px',
              textTransform: 'uppercase'
            }}>С даты</label>
            <input
              type="date"
              value={transactionFilters.start_date}
              onChange={(e) => setTransactionFilters({ ...transactionFilters, start_date: e.target.value })}
              style={{
                width: '100%',
                padding: '8px 12px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '14px',
                outline: 'none'
              }}
            />
          </div>
          <div style={{ flex: '1 1 200px' }}>
            <label style={{
              display: 'block',
              fontSize: '11px',
              color: '#888c95',
              marginBottom: '5px',
              textTransform: 'uppercase'
            }}>По дату</label>
            <input
              type="date"
              value={transactionFilters.end_date}
              onChange={(e) => setTransactionFilters({ ...transactionFilters, end_date: e.target.value })}
              style={{
                width: '100%',
                padding: '8px 12px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '14px',
                outline: 'none'
              }}
            />
          </div>
          <div style={{ flex: '1 1 150px' }}>
            <label style={{
              display: 'block',
              fontSize: '11px',
              color: '#888c95',
              marginBottom: '5px',
              textTransform: 'uppercase'
            }}>Тип</label>
            <select
              value={transactionFilters.transaction_type}
              onChange={(e) => setTransactionFilters({ ...transactionFilters, transaction_type: e.target.value })}
              style={{
                width: '100%',
                padding: '8px 12px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '14px',
                outline: 'none',
                cursor: 'pointer'
              }}>
              <option value="">Все</option>
              <option value="FUND">Пополнение</option>
              <option value="CAPTURE">Списание</option>
              <option value="DECLINE">Отказ</option>
            </select>
          </div>
          <div style={{ flex: '1 1 150px' }}>
            <label style={{
              display: 'block',
              fontSize: '11px',
              color: '#888c95',
              marginBottom: '5px',
              textTransform: 'uppercase'
            }}>Статус</label>
            <select
              value={transactionFilters.status}
              onChange={(e) => setTransactionFilters({ ...transactionFilters, status: e.target.value })}
              style={{
                width: '100%',
                padding: '8px 12px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '14px',
                outline: 'none',
                cursor: 'pointer'
              }}>
              <option value="">Все</option>
              <option value="APPROVED">Успешно</option>
              <option value="DECLINED">Отклонено</option>
            </select>
          </div>
          <div style={{ flex: '1 1 200px' }}>
            <label style={{
              display: 'block',
              fontSize: '11px',
              color: '#888c95',
              marginBottom: '5px',
              textTransform: 'uppercase'
            }}>Поиск</label>
            <input
              type="text"
              value={transactionFilters.search}
              onChange={(e) => setTransactionFilters({ ...transactionFilters, search: e.target.value })}
              placeholder="Поиск по описанию..."
              style={{
                width: '100%',
                padding: '8px 12px',
                backgroundColor: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '14px',
                outline: 'none'
              }}
            />
          </div>
          <button
            onClick={() => {
              setTransactionFilters({ start_date: '', end_date: '', transaction_type: '', status: '', search: '' });
              fetchDashboardData();
            }}
            style={{
              padding: '8px 16px',
              backgroundColor: 'transparent',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: '8px',
              color: '#888c95',
              fontSize: '12px',
              cursor: 'pointer',
              whiteSpace: 'nowrap'
            }}>
            Сбросить
          </button>
        </div>
        <div style={{
          backgroundColor: theme.colors.backgroundCard,
          backdropFilter: 'blur(20px)',
          borderRadius: theme.borderRadius.lg,
          border: `1px solid ${theme.colors.border}`,
          padding: '20px'
        }}>
          <table style={{
            width: '100%',
            borderCollapse: 'collapse'
          }}>
            <thead>
              <tr>
                <th style={{
                  textAlign: 'left',
                  color: '#888c95',
                  fontSize: '12px',
                  paddingBottom: '15px',
                  borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
                  fontWeight: '600'
                }}>Time</th>
                <th style={{
                  textAlign: 'left',
                  color: '#888c95',
                  fontSize: '12px',
                  paddingBottom: '15px',
                  borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
                  fontWeight: '600'
                }}>Merchant</th>
                <th style={{
                  textAlign: 'left',
                  color: '#888c95',
                  fontSize: '12px',
                  paddingBottom: '15px',
                  borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
                  fontWeight: '600'
                }}>Card</th>
                <th style={{
                  textAlign: 'left',
                  color: '#888c95',
                  fontSize: '12px',
                  paddingBottom: '15px',
                  borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
                  fontWeight: '600'
                }}>Amount</th>
                <th style={{
                  textAlign: 'left',
                  color: '#888c95',
                  fontSize: '12px',
                  paddingBottom: '15px',
                  borderBottom: '1px solid rgba(255, 255, 255, 0.1)',
                  fontWeight: '600'
                }}>Status</th>
              </tr>
            </thead>
            <tbody>
              {(transactions?.length ?? 0) === 0 ? (
                <tr>
                  <td colSpan={5} style={{
                    padding: '40px 0',
                    textAlign: 'center',
                    color: '#888c95',
                    fontSize: '14px'
                  }}>
                    No transactions yet
                  </td>
                </tr>
              ) : (
                (transactions ?? []).slice(0, 10).map((tx, index) => {
                  const txDate = new Date(tx.executed_at);
                  const timeStr = txDate.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false });
                  const statusColor = tx.status === 'APPROVED' || tx.status === 'SUCCESS' ? '#00e096' : '#ff3b3b';
                  const isIncoming = tx.transaction_type === 'FUND' || tx.transaction_type === 'ISSUE';

                  return (
                <tr key={tx.transaction_id || index}>
                  <td style={{
                    padding: '15px 0',
                    fontSize: '14px',
                    borderBottom: '1px solid rgba(255,255,255,0.05)'
                  }}>{timeStr}</td>
                  <td style={{
                    padding: '15px 0',
                    fontSize: '14px',
                    borderBottom: '1px solid rgba(255,255,255,0.05)'
                  }}>
                    <span style={{
                      width: '24px',
                      height: '24px',
                      background: isIncoming ? 'rgba(0, 224, 150, 0.15)' : 'rgba(255, 59, 59, 0.15)',
                      borderRadius: '50%',
                      display: 'inline-flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '12px',
                      marginRight: '10px'
                    }}>
                      {isIncoming ? '↑' : '↓'}
                    </span>
                    {tx.merchant || tx.transaction_type || 'Transaction'}
                  </td>
                  <td style={{
                    padding: '15px 0',
                    fontSize: '14px',
                    borderBottom: '1px solid rgba(255,255,255,0.05)'
                  }}>..{tx.card_last_4 || '****'}</td>
                  <td style={{
                    padding: '15px 0',
                    fontSize: '14px',
                    borderBottom: '1px solid rgba(255,255,255,0.05)',
                    color: isIncoming ? '#00e096' : '#ff6b6b',
                    fontWeight: '600'
                  }}>{isIncoming ? '+' : '-'}$ {Number(tx.amount || 0).toFixed(2)}</td>
                  <td style={{
                    padding: '15px 0',
                    fontSize: '14px',
                    borderBottom: '1px solid rgba(255,255,255,0.05)',
                    color: statusColor
                  }}>{tx.status}</td>
                </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
        </>}

        {/* ============ API VIEW ============ */}
        {activeMenu === 'api' && (
          <div style={{
            backgroundColor: theme.colors.backgroundCard,
            borderRadius: theme.borderRadius.lg,
            border: `1px solid ${theme.colors.border}`,
            padding: '40px',
            textAlign: 'center'
          }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>🔌</div>
            <h2 style={{ margin: '0 0 12px', fontSize: '22px', fontWeight: '700' }}>API & Trackers</h2>
            <p style={{ color: theme.colors.textSecondary, fontSize: '14px', maxWidth: '400px', margin: '0 auto 24px', lineHeight: '1.6' }}>
              Подключайте трекеры и используйте API для автоматизации управления картами и транзакциями.
            </p>
            <div style={{
              padding: '16px',
              backgroundColor: 'rgba(255,255,255,0.03)',
              borderRadius: '12px',
              border: '1px solid rgba(255,255,255,0.08)',
              fontFamily: theme.fonts.mono,
              fontSize: '13px',
              color: theme.colors.accent,
              textAlign: 'left'
            }}>
              <div style={{ color: theme.colors.textSecondary, fontSize: '11px', marginBottom: '8px' }}>API Base URL</div>
              <code>https://xplr-web.vercel.app/api/v1/</code>
            </div>
            <div style={{ marginTop: '20px', color: theme.colors.textMuted, fontSize: '12px' }}>
              Документация скоро будет доступна
            </div>
          </div>
        )}

      </div>

      {/* Limit Edit Modal */}
      {showLimitModal && (
        <div style={{
          position: 'fixed', top: 0, left: 0, width: '100vw', height: '100vh',
          backgroundColor: 'rgba(0, 0, 0, 0.8)', backdropFilter: 'blur(10px)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 10000
        }} onClick={() => setShowLimitModal(false)}>
          <div style={{
            backgroundColor: 'rgba(18, 18, 18, 0.95)', backdropFilter: 'blur(40px)',
            borderRadius: '24px', padding: '40px', width: '90%', maxWidth: '400px',
            border: '1px solid rgba(255, 255, 255, 0.1)', boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)'
          }} onClick={(e) => e.stopPropagation()}>
            <h2 style={{ margin: '0 0 20px 0', fontSize: '22px', fontWeight: '700', color: '#fff' }}>
              Set Daily Limit
            </h2>
            <label style={{ display: 'block', fontSize: '12px', fontWeight: '600', color: '#888c95', textTransform: 'uppercase', letterSpacing: '1px', marginBottom: '10px' }}>
              Limit Amount ($)
            </label>
            <input
              type="number"
              min="0"
              step="50"
              value={limitValue}
              onChange={(e) => setLimitValue(e.target.value)}
              style={{
                width: '100%', padding: '16px', backgroundColor: 'rgba(255,255,255,0.05)',
                border: '1px solid rgba(255,255,255,0.1)', borderRadius: '12px',
                color: '#fff', fontSize: '18px', fontWeight: '600', outline: 'none', marginBottom: '24px'
              }}
            />
            <div style={{ display: 'flex', gap: '12px' }}>
              <button onClick={() => setShowLimitModal(false)} style={{
                flex: 1, padding: '14px', backgroundColor: 'rgba(255,255,255,0.05)',
                border: '1px solid rgba(255,255,255,0.1)', borderRadius: '12px',
                color: '#888', fontWeight: '600', fontSize: '14px', cursor: 'pointer'
              }}>Cancel</button>
              <button onClick={handleSaveLimit} disabled={isSavingLimit} style={{
                flex: 1, padding: '14px', backgroundColor: '#00e096',
                border: 'none', borderRadius: '12px',
                color: '#0a0a0a', fontWeight: '700', fontSize: '14px',
                cursor: isSavingLimit ? 'not-allowed' : 'pointer', opacity: isSavingLimit ? 0.6 : 1
              }}>{isSavingLimit ? 'Saving...' : 'Save Limit'}</button>
            </div>
          </div>
        </div>
      )}

      {/* Toast notification */}
      {toast && (
        <div
          style={{
            position: 'fixed',
            bottom: '24px',
            left: '50%',
            transform: 'translateX(-50%)',
            padding: '12px 24px',
            borderRadius: theme.borderRadius.lg,
            backgroundColor: toast.type === 'success' ? theme.colors.accentMuted : theme.colors.error + '25',
            border: `1px solid ${toast.type === 'success' ? theme.colors.accent : theme.colors.error}`,
            color: toast.type === 'success' ? theme.colors.accent : theme.colors.error,
            fontSize: '14px',
            fontWeight: '600',
            zIndex: 10001,
            boxShadow: '0 8px 24px rgba(0,0,0,0.3)',
            backdropFilter: 'blur(12px)'
          }}
        >
          {toast.message}
        </div>
      )}

      {/* Confirmation Dialog (Block / Unblock) */}
      {showConfirmDialog && (() => {
        const card = selectedCardId ? (cards ?? []).find((c) => c.id === selectedCardId) : null;
        const isUnblock = card?.card_status === 'BLOCKED';
        return (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100vw',
          height: '100vh',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          backdropFilter: 'blur(10px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 10000
        }}
        onClick={() => setShowConfirmDialog(false)}>
          <div style={{
            backgroundColor: 'rgba(18, 18, 18, 0.95)',
            backdropFilter: 'blur(40px)',
            borderRadius: '24px',
            padding: '40px',
            width: '90%',
            maxWidth: '400px',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)',
            textAlign: 'center'
          }}
          onClick={(e) => e.stopPropagation()}>
            <div style={{
              width: '64px',
              height: '64px',
              borderRadius: '50%',
              backgroundColor: isUnblock ? 'rgba(0, 224, 150, 0.2)' : 'rgba(255, 59, 59, 0.2)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 24px',
              fontSize: '32px'
            }}>
              {isUnblock ? '🔓' : '⚠️'}
            </div>

            <h2 style={{
              margin: '0 0 12px 0',
              fontSize: '24px',
              fontWeight: '700',
              color: '#ffffff',
              letterSpacing: '-1px'
            }}>
              {isUnblock ? 'Unblock This Card?' : 'Block This Card?'}
            </h2>

            <p style={{
              margin: '0 0 32px 0',
              fontSize: '14px',
              color: '#888c95',
              lineHeight: '1.6'
            }}>
              {isUnblock
                ? 'This card will be active again and can be used for transactions.'
                : 'This card will be immediately blocked and cannot be used for transactions until you unblock it.'}
            </p>

            <div style={{ display: 'flex', gap: '12px' }}>
              <button
                onClick={() => setShowConfirmDialog(false)}
                style={{
                  flex: 1,
                  padding: '16px',
                  backgroundColor: 'rgba(255, 255, 255, 0.05)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  color: '#888c95',
                  fontWeight: '600',
                  fontSize: '16px',
                  cursor: 'pointer',
                  transition: '0.2s'
                }}>
                Cancel
              </button>
              <button
                onClick={() => {
                  if (selectedCardId && card) {
                    handleToggleCardBlock(selectedCardId, card.card_status);
                  }
                }}
                style={{
                  flex: 1,
                  padding: '16px',
                  backgroundColor: isUnblock ? '#00e096' : '#ff3b3b',
                  border: 'none',
                  borderRadius: '12px',
                  color: isUnblock ? '#000' : '#ffffff',
                  fontWeight: '700',
                  fontSize: '16px',
                  cursor: 'pointer',
                  transition: '0.2s'
                }}>
                {isUnblock ? 'Unblock Card' : 'Block Card'}
              </button>
            </div>
          </div>
        </div>
        );
      })()}

      {/* Auto-Replenish Modal */}
      {showAutoReplenishModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100vw',
          height: '100vh',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          backdropFilter: 'blur(10px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 10000
        }}
        onClick={() => setShowAutoReplenishModal(false)}>
          <div style={{
            backgroundColor: 'rgba(18, 18, 18, 0.95)',
            backdropFilter: 'blur(40px)',
            borderRadius: '24px',
            padding: '40px',
            width: '90%',
            maxWidth: '500px',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)'
          }}
          onClick={(e) => e.stopPropagation()}>
            <h2 style={{
              margin: '0 0 24px 0',
              fontSize: '24px',
              fontWeight: '700',
              color: '#ffffff',
              letterSpacing: '-1px'
            }}>
              Настройка автопополнения
            </h2>

            <div style={{ marginBottom: '20px' }}>
              <label style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: '600',
                color: '#888c95',
                marginBottom: '8px',
                textTransform: 'uppercase',
                letterSpacing: '1px'
              }}>
                Порог (минимальный баланс карты)
              </label>
              <input
                type="number"
                step="0.01"
                min="0"
                value={autoReplenishThreshold}
                onChange={(e) => setAutoReplenishThreshold(e.target.value)}
                placeholder="50.00"
                style={{
                  width: '100%',
                  padding: '12px 16px',
                  backgroundColor: 'rgba(255, 255, 255, 0.05)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  color: '#ffffff',
                  fontSize: '16px',
                  outline: 'none'
                }}
              />
            </div>

            <div style={{ marginBottom: '32px' }}>
              <label style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: '600',
                color: '#888c95',
                marginBottom: '8px',
                textTransform: 'uppercase',
                letterSpacing: '1px'
              }}>
                Сумма пополнения
              </label>
              <input
                type="number"
                step="0.01"
                min="0"
                value={autoReplenishAmount}
                onChange={(e) => setAutoReplenishAmount(e.target.value)}
                placeholder="100.00"
                style={{
                  width: '100%',
                  padding: '12px 16px',
                  backgroundColor: 'rgba(255, 255, 255, 0.05)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  color: '#ffffff',
                  fontSize: '16px',
                  outline: 'none'
                }}
              />
            </div>

            <div style={{ display: 'flex', gap: '12px' }}>
              <button
                onClick={() => {
                  setShowAutoReplenishModal(false);
                  setAutoReplenishThreshold('');
                  setAutoReplenishAmount('');
                }}
                style={{
                  flex: 1,
                  padding: '16px',
                  backgroundColor: 'transparent',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  color: '#888c95',
                  fontWeight: '600',
                  fontSize: '16px',
                  cursor: 'pointer',
                  transition: '0.2s'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = '#fff';
                  e.currentTarget.style.color = '#fff';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.1)';
                  e.currentTarget.style.color = '#888c95';
                }}>
                Отмена
              </button>
              <button
                onClick={handleSetAutoReplenish}
                disabled={isSettingAutoReplenish}
                style={{
                  flex: 1,
                  padding: '16px',
                  backgroundColor: '#00e096',
                  border: 'none',
                  borderRadius: '12px',
                  color: '#000',
                  fontWeight: '700',
                  fontSize: '16px',
                  cursor: isSettingAutoReplenish ? 'not-allowed' : 'pointer',
                  opacity: isSettingAutoReplenish ? 0.5 : 1,
                  transition: '0.2s'
                }}>
                {isSettingAutoReplenish ? 'Настройка...' : 'Включить'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Create Card Modal */}
      {showCreateCardModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100vw',
          height: '100vh',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          backdropFilter: 'blur(10px)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          zIndex: 10000
        }}
        onClick={() => setShowCreateCardModal(false)}>
          <div style={{
            backgroundColor: 'rgba(18, 18, 18, 0.95)',
            backdropFilter: 'blur(40px)',
            borderRadius: '24px',
            padding: '40px',
            width: '90%',
            maxWidth: '500px',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.5)'
          }}
          onClick={(e) => e.stopPropagation()}>
            <h2 style={{
              margin: '0 0 24px 0',
              fontSize: '28px',
              fontWeight: '700',
              color: '#ffffff',
              letterSpacing: '-1px'
            }}>
              Issue New Card
            </h2>

            <div style={{ marginBottom: '24px' }}>
              <label style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: '600',
                color: '#888c95',
                textTransform: 'uppercase',
                letterSpacing: '1px',
                marginBottom: '12px'
              }}>
                Card Type
              </label>
              <div style={{ display: 'flex', gap: '12px' }}>
                <button
                  onClick={() => setNewCardType('VISA')}
                  style={{
                    flex: 1,
                    padding: '16px',
                    backgroundColor: newCardType === 'VISA' ? 'rgba(0, 224, 150, 0.2)' : 'rgba(255, 255, 255, 0.05)',
                    border: newCardType === 'VISA' ? '2px solid #00e096' : '2px solid rgba(255, 255, 255, 0.1)',
                    borderRadius: '12px',
                    color: newCardType === 'VISA' ? '#00e096' : '#888c95',
                    fontWeight: '700',
                    fontSize: '16px',
                    cursor: 'pointer',
                    transition: '0.2s'
                  }}>
                  VISA
                </button>
                <button
                  onClick={() => setNewCardType('MasterCard')}
                  style={{
                    flex: 1,
                    padding: '16px',
                    backgroundColor: newCardType === 'MasterCard' ? 'rgba(0, 224, 150, 0.2)' : 'rgba(255, 255, 255, 0.05)',
                    border: newCardType === 'MasterCard' ? '2px solid #00e096' : '2px solid rgba(255, 255, 255, 0.1)',
                    borderRadius: '12px',
                    color: newCardType === 'MasterCard' ? '#00e096' : '#888c95',
                    fontWeight: '700',
                    fontSize: '16px',
                    cursor: 'pointer',
                    transition: '0.2s'
                  }}>
                  MasterCard
                </button>
              </div>
            </div>

            {/* Category selection — different per section */}
            <div style={{ marginBottom: '24px' }}>
              <label style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: '600',
                color: '#888c95',
                textTransform: 'uppercase',
                letterSpacing: '1px',
                marginBottom: '12px'
              }}>
                {activeSection === 'arbitrage' ? 'Категория' : 'Тип карты'}
              </label>
              {activeSection === 'arbitrage' ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {arbBins.map(b => (
                    <button key={b.bin} onClick={() => { setSelectedBin(b.bin); setNewCardType(b.bin.startsWith('4') ? 'VISA' : 'MasterCard'); }}
                      style={{
                        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                        padding: '14px 16px',
                        backgroundColor: selectedBin === b.bin ? 'rgba(59, 130, 246, 0.2)' : 'rgba(255,255,255,0.05)',
                        border: selectedBin === b.bin ? '2px solid #3b82f6' : '2px solid rgba(255,255,255,0.1)',
                        borderRadius: '12px', cursor: 'pointer', transition: '0.2s', textAlign: 'left'
                      }}>
                      <div>
                        <div style={{ color: selectedBin === b.bin ? '#3b82f6' : '#fff', fontWeight: '600', fontSize: '14px' }}>{b.label}</div>
                        <div style={{ color: '#888c95', fontSize: '11px', marginTop: '2px' }}>Пополнение: {b.topup}% • Facebook, Google, TikTok</div>
                      </div>
                      <div style={{ color: selectedBin === b.bin ? '#3b82f6' : '#888c95', fontSize: '15px', fontWeight: '700' }}>${b.fee.toFixed(2)}</div>
                    </button>
                  ))}
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                  {([
                    { key: 'travel' as const, label: 'Для путешествий', desc: 'Отели, авиабилеты, аренда • 1 год', feeRub: 990, feeUsd: '$3.00' },
                    { key: 'services' as const, label: 'Для зарубежных сервисов', desc: 'Подписки, онлайн-сервисы • 1 год', feeRub: 690, feeUsd: '$2.00' },
                  ]).map((cat) => {
                    const rubRate = (exchangeRates ?? []).find(r => r.currency_to === 'USD');
                    const rate = rubRate ? parseFloat(rubRate.final_rate) : 99;
                    const priceRub = Math.ceil(parseFloat(cat.feeUsd.replace('$', '')) * rate);
                    return (
                    <button
                      key={cat.key}
                      onClick={() => setNewCardCategory(cat.key)}
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        padding: '14px 16px',
                        backgroundColor: newCardCategory === cat.key ? 'rgba(0, 224, 150, 0.15)' : 'rgba(255, 255, 255, 0.05)',
                        border: newCardCategory === cat.key ? '2px solid #00e096' : '2px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '12px',
                        cursor: 'pointer',
                        transition: '0.2s',
                        textAlign: 'left'
                      }}
                    >
                      <div>
                        <div style={{
                          color: newCardCategory === cat.key ? '#00e096' : '#fff',
                          fontWeight: '600',
                          fontSize: '14px'
                        }}>{cat.label}</div>
                        <div style={{
                          color: '#888c95',
                          fontSize: '12px',
                          marginTop: '2px'
                        }}>{cat.desc}</div>
                      </div>
                      <div style={{
                        color: newCardCategory === cat.key ? '#00e096' : '#888c95',
                        fontSize: '15px',
                        fontWeight: '700',
                        whiteSpace: 'nowrap'
                      }}>{priceRub}₽</div>
                    </button>
                    );
                  })}
                </div>
              )}
            </div>

            <div style={{ marginBottom: '32px' }}>
              <label style={{
                display: 'block',
                fontSize: '12px',
                fontWeight: '600',
                color: '#888c95',
                textTransform: 'uppercase',
                letterSpacing: '1px',
                marginBottom: '12px'
              }}>
                Card Nickname
              </label>

              {/* Quantity selector — only for arbitrage (1–100) */}
              {activeSection === 'arbitrage' && (
                <div style={{ marginBottom: '20px' }}>
                  <label style={{
                    display: 'block',
                    fontSize: '12px',
                    fontWeight: '600',
                    color: '#888c95',
                    textTransform: 'uppercase',
                    letterSpacing: '1px',
                    marginBottom: '8px'
                  }}>
                    Количество карт (1–100)
                  </label>
                  <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
                    <input
                      type="number"
                      min="1"
                      max="100"
                      value={newCardCount}
                      onChange={(e) => {
                        const v = Math.max(1, Math.min(100, parseInt(e.target.value) || 1));
                        setNewCardCount(v);
                      }}
                      style={{
                        width: '80px',
                        padding: '12px',
                        backgroundColor: 'rgba(255, 255, 255, 0.05)',
                        border: '2px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '10px',
                        color: '#00e096',
                        fontWeight: '700',
                        fontSize: '18px',
                        textAlign: 'center',
                        outline: 'none'
                      }}
                    />
                    {[5, 10, 25, 50].map((qty) => (
                      <button
                        key={qty}
                        onClick={() => setNewCardCount(qty)}
                        style={{
                          padding: '10px 14px',
                          backgroundColor: newCardCount === qty ? 'rgba(0, 224, 150, 0.2)' : 'rgba(255, 255, 255, 0.05)',
                          border: newCardCount === qty ? '2px solid #00e096' : '2px solid rgba(255, 255, 255, 0.1)',
                          borderRadius: '10px',
                          color: newCardCount === qty ? '#00e096' : '#888c95',
                          fontWeight: '700',
                          fontSize: '14px',
                          cursor: 'pointer',
                          transition: '0.2s'
                        }}
                      >
                        {qty}
                      </button>
                    ))}
                  </div>
                  <div style={{ fontSize: '11px', color: '#888c95', marginTop: '8px' }}>
                    Итого: ${((arbBins.find(b => b.bin === selectedBin)?.fee ?? 5) * newCardCount).toFixed(2)} ({newCardCount} × ${(arbBins.find(b => b.bin === selectedBin)?.fee ?? 5).toFixed(2)})
                  </div>
                </div>
              )}

              <input
                type="text"
                placeholder="e.g., FB Ads Campaign"
                value={newCardNickname}
                onChange={(e) => setNewCardNickname(e.target.value)}
                style={{
                  width: '100%',
                  padding: '16px',
                  backgroundColor: 'rgba(255, 255, 255, 0.05)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  color: '#ffffff',
                  fontSize: '16px',
                  outline: 'none'
                }}
              />
            </div>

            <div style={{ display: 'flex', gap: '12px' }}>
              <button
                onClick={() => setShowCreateCardModal(false)}
                style={{
                  flex: 1,
                  padding: '16px',
                  backgroundColor: 'rgba(255, 255, 255, 0.05)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: '12px',
                  color: '#888c95',
                  fontWeight: '600',
                  fontSize: '16px',
                  cursor: 'pointer',
                  transition: '0.2s'
                }}>
                Cancel
              </button>
              <button
                onClick={handleCreateCard}
                disabled={isCreatingCard}
                style={{
                  flex: 1,
                  padding: '16px',
                  backgroundColor: '#00e096',
                  border: 'none',
                  borderRadius: '12px',
                  color: '#000000',
                  fontWeight: '700',
                  fontSize: '16px',
                  cursor: isCreatingCard ? 'not-allowed' : 'pointer',
                  opacity: isCreatingCard ? 0.5 : 1,
                  transition: '0.2s',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '8px'
                }}>
                {isCreatingCard && (
                  <div style={{
                    width: '16px',
                    height: '16px',
                    border: '2px solid #000000',
                    borderTopColor: 'transparent',
                    borderRadius: '50%',
                    animation: 'spin 0.6s linear infinite'
                  }} />
                )}
                {isCreatingCard ? 'Создание...' : activeSection === 'arbitrage'
                  ? (() => {
                      const binFee = arbBins.find(b => b.bin === selectedBin)?.fee ?? 5;
                      return newCardCount > 1 ? `Выпустить ${newCardCount} карт — $${(binFee * newCardCount).toFixed(2)}` : `Выпустить карту — $${binFee.toFixed(2)}`;
                    })()
                  : (() => {
                      const rubRate = (exchangeRates ?? []).find(r => r.currency_to === 'USD');
                      const rate = rubRate ? parseFloat(rubRate.final_rate) : 99;
                      const feeUsd = newCardCategory === 'travel' ? 3 : 2;
                      return `Выпустить — ${Math.ceil(feeUsd * rate)}₽`;
                    })()}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Dashboard;
