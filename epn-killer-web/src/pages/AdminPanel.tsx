import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../api/axios';
import { theme } from '../theme/theme';

interface AdminStats {
  total_users: number;
  total_balance: string;
  total_cards: number;
  active_cards: number;
  frozen_cards: number;
  closed_cards: number;
  blocked_cards: number;
}

interface AdminUser {
  id: number;
  email: string;
  balance_rub: string;
  status: string;
  is_admin: boolean;
  card_count: number;
  created_at: string;
}

const AdminPanel: React.FC = () => {
  const navigate = useNavigate();
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [search, setSearch] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchAdminData();
  }, []);

  const fetchAdminData = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) { navigate('/login'); return; }
      const config = { headers: { Authorization: `Bearer ${token}` } };

      const [statsRes, usersRes] = await Promise.all([
        axios.get(`${API_BASE_URL}/admin/stats`, config),
        axios.get(`${API_BASE_URL}/admin/users`, config),
      ]);
      setStats(statsRes.data);
      setUsers(usersRes.data || []);
      setIsLoading(false);
    } catch (err: any) {
      setIsLoading(false);
      if (err?.response?.status === 403) {
        setError('Access denied. Admin privileges required.');
      } else if (err?.response?.status === 401) {
        navigate('/login');
      } else {
        setError('Failed to load admin data.');
      }
    }
  };

  const filteredUsers = users.filter(u =>
    u.email.toLowerCase().includes(search.toLowerCase())
  );

  if (isLoading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', backgroundColor: theme.colors.background, color: theme.colors.textPrimary }}>
        Loading admin panel...
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh', backgroundColor: theme.colors.background, color: theme.colors.error, gap: '20px' }}>
        <div style={{ fontSize: '48px' }}>🔒</div>
        <div style={{ fontSize: '18px', fontWeight: '700' }}>{error}</div>
        <button onClick={() => navigate('/dashboard')} style={{
          padding: '10px 24px', backgroundColor: theme.colors.accent, color: '#0a0a0a',
          border: 'none', borderRadius: '8px', fontWeight: '600', cursor: 'pointer', fontSize: '14px'
        }}>Back to Dashboard</button>
      </div>
    );
  }

  const statCards = [
    { label: 'Total Users', value: stats?.total_users || 0, icon: '👥', color: '#3b82f6' },
    { label: 'Total TVL', value: `$${parseFloat(stats?.total_balance || '0').toLocaleString()}`, icon: '💰', color: '#00e096' },
    { label: 'Total Cards', value: stats?.total_cards || 0, icon: '💳', color: '#8b5cf6' },
    { label: 'Active Cards', value: stats?.active_cards || 0, icon: '✅', color: '#14b8a6' },
  ];

  return (
    <div style={{ minHeight: '100vh', backgroundColor: theme.colors.background, color: theme.colors.textPrimary, padding: '30px' }}>
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '30px' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: '28px', fontWeight: '800', letterSpacing: '-0.5px' }}>
            Admin Panel
          </h1>
          <p style={{ margin: '5px 0 0', color: theme.colors.textSecondary, fontSize: '14px' }}>
            Platform overview & user management
          </p>
        </div>
        <button onClick={() => navigate('/dashboard')} style={{
          padding: '10px 20px', backgroundColor: 'rgba(255,255,255,0.05)',
          border: `1px solid ${theme.colors.border}`, borderRadius: '8px',
          color: theme.colors.textSecondary, fontSize: '13px', cursor: 'pointer', fontWeight: '600'
        }}>← Dashboard</button>
      </div>

      {/* Stats Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px', marginBottom: '30px' }}>
        {statCards.map((card) => (
          <div key={card.label} style={{
            backgroundColor: theme.colors.backgroundCard,
            border: `1px solid ${theme.colors.border}`,
            borderRadius: '16px',
            padding: '24px',
            backdropFilter: 'blur(20px)'
          }}>
            <div style={{ fontSize: '12px', color: theme.colors.textSecondary, textTransform: 'uppercase', letterSpacing: '1px', marginBottom: '12px' }}>
              {card.icon} {card.label}
            </div>
            <div style={{ fontSize: '28px', fontWeight: '800', color: card.color }}>
              {card.value}
            </div>
          </div>
        ))}
      </div>

      {/* Card status breakdown */}
      {stats && (
        <div style={{
          display: 'flex', gap: '12px', marginBottom: '30px', flexWrap: 'wrap'
        }}>
          {[
            { label: 'Active', count: stats.active_cards, color: '#00e096' },
            { label: 'Frozen', count: stats.frozen_cards, color: '#3b82f6' },
            { label: 'Blocked', count: stats.blocked_cards, color: '#ff6b6b' },
            { label: 'Closed', count: stats.closed_cards, color: '#666' },
          ].map(s => (
            <div key={s.label} style={{
              padding: '8px 16px', borderRadius: '8px',
              backgroundColor: 'rgba(255,255,255,0.03)', border: `1px solid ${theme.colors.border}`,
              fontSize: '12px', color: s.color, fontWeight: '600'
            }}>
              {s.label}: {s.count}
            </div>
          ))}
        </div>
      )}

      {/* Users Table */}
      <div style={{
        backgroundColor: theme.colors.backgroundCard,
        border: `1px solid ${theme.colors.border}`,
        borderRadius: '16px',
        padding: '24px',
        backdropFilter: 'blur(20px)'
      }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
          <h2 style={{ margin: 0, fontSize: '18px', fontWeight: '700' }}>Users ({filteredUsers.length})</h2>
          <input
            type="text"
            placeholder="Search by email..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{
              padding: '10px 16px',
              backgroundColor: 'rgba(255,255,255,0.05)',
              border: `1px solid ${theme.colors.border}`,
              borderRadius: '8px',
              color: '#fff',
              fontSize: '13px',
              outline: 'none',
              width: '280px'
            }}
          />
        </div>

        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              {['ID', 'Email', 'Balance', 'Cards', 'Status', 'Admin', 'Created'].map(h => (
                <th key={h} style={{
                  textAlign: 'left', padding: '12px 8px', fontSize: '11px',
                  color: theme.colors.textSecondary, fontWeight: '600',
                  textTransform: 'uppercase', letterSpacing: '0.5px',
                  borderBottom: `1px solid ${theme.colors.border}`
                }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filteredUsers.length === 0 ? (
              <tr>
                <td colSpan={7} style={{ padding: '30px', textAlign: 'center', color: theme.colors.textSecondary }}>
                  {search ? 'No users match your search' : 'No users found'}
                </td>
              </tr>
            ) : (
              filteredUsers.map(u => (
                <tr key={u.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                  <td style={{ padding: '12px 8px', fontSize: '13px', color: theme.colors.textSecondary }}>#{u.id}</td>
                  <td style={{ padding: '12px 8px', fontSize: '13px', fontWeight: '600' }}>{u.email}</td>
                  <td style={{ padding: '12px 8px', fontSize: '13px', color: '#00e096', fontWeight: '600' }}>
                    ${parseFloat(u.balance_rub).toFixed(2)}
                  </td>
                  <td style={{ padding: '12px 8px', fontSize: '13px' }}>{u.card_count}</td>
                  <td style={{ padding: '12px 8px' }}>
                    <span style={{
                      fontSize: '10px', padding: '3px 8px', borderRadius: '4px', fontWeight: '700',
                      backgroundColor: u.status === 'ACTIVE' ? 'rgba(0, 224, 150, 0.15)' : 'rgba(255, 107, 107, 0.15)',
                      color: u.status === 'ACTIVE' ? '#00e096' : '#ff6b6b'
                    }}>{u.status}</span>
                  </td>
                  <td style={{ padding: '12px 8px', fontSize: '13px' }}>
                    {u.is_admin ? <span style={{ color: '#f59e0b', fontWeight: '700' }}>⭐ Admin</span> : '—'}
                  </td>
                  <td style={{ padding: '12px 8px', fontSize: '12px', color: theme.colors.textSecondary }}>
                    {u.created_at ? new Date(u.created_at).toLocaleDateString() : '—'}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default AdminPanel;
