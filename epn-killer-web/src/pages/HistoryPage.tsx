import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../api/axios';
import { theme, XPLR_STORAGE_MODE } from '../theme/theme';
import SidebarLayout from '../components/SidebarLayout';

interface Transaction {
  transaction_id: number;
  amount: number;
  transaction_type: string;
  status: string;
  executed_at: string;
  merchant?: string;
  card_last_4?: string;
}

interface UserData {
  balance_arbitrage?: number;
  balance_personal?: number;
  balance?: number;
  balance_rub?: number;
}

const HistoryPage: React.FC = () => {
  const navigate = useNavigate();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [userData, setUserData] = useState<UserData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [walletFilter, setWalletFilter] = useState<'all' | 'arbitrage' | 'personal'>('all');

  const appMode = localStorage.getItem(XPLR_STORAGE_MODE) || 'professional';
  const isProfessional = appMode === 'professional';

  const [filters, setFilters] = useState({
    start_date: '',
    end_date: '',
    transaction_type: '',
    status: '',
  });

  const getConfig = () => ({ headers: { Authorization: `Bearer ${localStorage.getItem('token')}` } });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [txRes, userRes] = await Promise.all([
        axios.get(`${API_BASE_URL}/user/transactions`, { ...getConfig(), params: filters }),
        axios.get(`${API_BASE_URL}/user/me`, getConfig()),
      ]);
      setTransactions(Array.isArray(txRes.data) ? txRes.data : []);
      setUserData(userRes.data);
    } catch (e) { console.error('Error fetching history:', e); }
    finally { setIsLoading(false); }
  };

  if (isLoading) {
    return <SidebarLayout><div style={{ padding: 40, textAlign: 'center', color: theme.colors.textSecondary }}>Загрузка...</div></SidebarLayout>;
  }

  return (
    <SidebarLayout>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <button onClick={() => navigate('/dashboard')} style={{
            padding: '8px 16px', backgroundColor: 'rgba(255,255,255,0.04)',
            border: `1px solid ${theme.colors.border}`, borderRadius: 8,
            color: theme.colors.textSecondary, fontSize: 13, cursor: 'pointer', fontWeight: 600
          }}>← Назад</button>
          <h1 style={{ margin: 0, fontSize: 24, fontWeight: 700 }}>История транзакций</h1>
        </div>
        <button onClick={fetchData} style={{
          padding: '10px 20px', backgroundColor: theme.colors.accentMuted,
          border: `1px solid ${theme.colors.accentBorder}`, borderRadius: 8,
          color: theme.colors.accent, fontSize: 13, cursor: 'pointer', fontWeight: 600
        }}>Обновить</button>
      </div>

      {/* Wallet Toggle */}
      <div style={{ display: 'flex', gap: 10, marginBottom: 24 }}>
        {([
          { key: 'all' as const, label: 'Все', color: theme.colors.accent },
          { key: 'arbitrage' as const, label: `💳 Арбитраж · $${Number(userData?.balance_arbitrage ?? 0).toFixed(2)}`, color: '#3b82f6' },
          { key: 'personal' as const, label: `✈️ Личный · $${Number(userData?.balance_personal ?? 0).toFixed(2)}`, color: '#14b8a6' },
        ]).map(w => (
          <button key={w.key} onClick={() => setWalletFilter(w.key)} style={{
            padding: '10px 20px', borderRadius: 10, fontSize: 13, fontWeight: 600, cursor: 'pointer',
            backgroundColor: walletFilter === w.key ? `${w.color}20` : 'rgba(255,255,255,0.03)',
            border: `1px solid ${walletFilter === w.key ? w.color : theme.colors.border}`,
            color: walletFilter === w.key ? w.color : theme.colors.textSecondary,
            transition: '0.2s'
          }}>{w.label}</button>
        ))}
      </div>

      {/* Filters */}
      <div style={{
        display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))', gap: 12,
        marginBottom: 24, padding: 20,
        backgroundColor: theme.colors.backgroundCard, borderRadius: 12,
        border: `1px solid ${theme.colors.border}`
      }}>
        {[
          { label: 'С даты', type: 'date', value: filters.start_date, onChange: (v: string) => setFilters({ ...filters, start_date: v }) },
          { label: 'По дату', type: 'date', value: filters.end_date, onChange: (v: string) => setFilters({ ...filters, end_date: v }) },
        ].map((f, i) => (
          <div key={i}>
            <label style={{ display: 'block', fontSize: 11, color: theme.colors.textSecondary, marginBottom: 4, textTransform: 'uppercase' }}>{f.label}</label>
            <input type={f.type} value={f.value} onChange={e => f.onChange(e.target.value)}
              style={{
                width: '100%', padding: '8px 12px', backgroundColor: 'rgba(255,255,255,0.04)',
                border: `1px solid ${theme.colors.border}`, borderRadius: 8,
                color: '#fff', fontSize: 13, outline: 'none', boxSizing: 'border-box'
              }} />
          </div>
        ))}
        <div>
          <label style={{ display: 'block', fontSize: 11, color: theme.colors.textSecondary, marginBottom: 4, textTransform: 'uppercase' }}>Тип</label>
          <select value={filters.transaction_type} onChange={e => setFilters({ ...filters, transaction_type: e.target.value })}
            style={{
              width: '100%', padding: '8px 12px', backgroundColor: 'rgba(255,255,255,0.04)',
              border: `1px solid ${theme.colors.border}`, borderRadius: 8,
              color: '#fff', fontSize: 13, outline: 'none'
            }}>
            <option value="">Все</option>
            <option value="deposit">Deposit</option>
            <option value="withdrawal">Withdrawal</option>
            <option value="card_issue">Card Issue</option>
          </select>
        </div>
        <div style={{ display: 'flex', alignItems: 'flex-end' }}>
          <button onClick={fetchData} style={{
            width: '100%', padding: '9px 16px', backgroundColor: theme.colors.accent,
            border: 'none', borderRadius: 8, color: theme.colors.background,
            fontSize: 13, fontWeight: 600, cursor: 'pointer'
          }}>Применить</button>
        </div>
      </div>

      {/* Transactions Table */}
      <div style={{
        backgroundColor: theme.colors.backgroundCard, borderRadius: 12,
        border: `1px solid ${theme.colors.border}`, overflow: 'hidden'
      }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              {['Дата', 'Описание', 'Карта', 'Сумма', 'Статус'].map(h => (
                <th key={h} style={{
                  textAlign: 'left', padding: '14px 16px', fontSize: 11,
                  color: theme.colors.textSecondary, fontWeight: 600,
                  textTransform: 'uppercase', letterSpacing: 0.5,
                  borderBottom: `1px solid ${theme.colors.border}`
                }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {(transactions ?? []).length === 0 ? (
              <tr>
                <td colSpan={5} style={{ padding: 40, textAlign: 'center', color: theme.colors.textSecondary }}>
                  Нет транзакций
                </td>
              </tr>
            ) : (transactions ?? []).map(tx => {
              const isIncoming = tx.transaction_type === 'deposit' || tx.transaction_type === 'refund';
              const statusColor = tx.status === 'completed' ? theme.colors.success : tx.status === 'pending' ? theme.colors.warning : theme.colors.error;
              return (
                <tr key={tx.transaction_id} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                  <td style={{ padding: '14px 16px', fontSize: 13, color: theme.colors.textSecondary }}>
                    {tx.executed_at ? new Date(tx.executed_at).toLocaleDateString('ru-RU') : '—'}
                  </td>
                  <td style={{ padding: '14px 16px', fontSize: 13 }}>
                    <span style={{
                      display: 'inline-block', width: 22, height: 22, borderRadius: '50%',
                      backgroundColor: isIncoming ? 'rgba(0,224,150,0.15)' : 'rgba(255,59,59,0.15)',
                      textAlign: 'center', lineHeight: '22px', fontSize: 11, marginRight: 8
                    }}>{isIncoming ? '↑' : '↓'}</span>
                    {tx.merchant || tx.transaction_type || 'Transaction'}
                  </td>
                  <td style={{ padding: '14px 16px', fontSize: 13, color: theme.colors.textSecondary }}>
                    ..{tx.card_last_4 || '****'}
                  </td>
                  <td style={{
                    padding: '14px 16px', fontSize: 14, fontWeight: 600,
                    color: isIncoming ? '#00e096' : '#ff6b6b'
                  }}>
                    {isIncoming ? '+' : '-'}${Number(tx.amount || 0).toFixed(2)}
                  </td>
                  <td style={{ padding: '14px 16px', fontSize: 12, color: statusColor, fontWeight: 600 }}>
                    {tx.status}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </SidebarLayout>
  );
};

export default HistoryPage;
