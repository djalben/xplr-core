import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { theme, XPLR_STORAGE_MODE, type AppMode } from '../theme/theme';

interface SidebarLayoutProps {
  children: React.ReactNode;
}

const SidebarLayout: React.FC<SidebarLayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();

  const [appMode, setAppMode] = useState<AppMode>(() => {
    const stored = localStorage.getItem(XPLR_STORAGE_MODE);
    return (stored === 'personal' || stored === 'professional') ? stored : 'professional';
  });

  const isProfessional = appMode === 'professional';
  const setMode = (mode: AppMode) => {
    setAppMode(mode);
    localStorage.setItem(XPLR_STORAGE_MODE, mode);
  };

  const [sidebarOpen, setSidebarOpen] = useState(false);

  const menuItems = isProfessional
    ? [
        { key: 'dashboard', label: '📊 Dashboard', path: '/dashboard' },
        { key: 'cards', label: '💳 Cards', path: '/cards' },
        { key: 'history', label: '💸 History', path: '/history' },
        { key: 'teams', label: '👥 Teams', path: '/teams' },
        { key: 'api', label: '🔌 API', path: '/api' },
      ]
    : [
        { key: 'dashboard', label: '📊 Dashboard', path: '/dashboard' },
        { key: 'cards', label: '💳 Cards', path: '/cards' },
        { key: 'history', label: '💸 History', path: '/history' },
      ];

  const activePath = location.pathname;

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem(XPLR_STORAGE_MODE);
    navigate('/login');
  };

  const isMobile = typeof window !== 'undefined' && window.innerWidth < 768;

  return (
    <div style={{
      width: '100vw',
      height: '100vh',
      backgroundColor: theme.colors.background,
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
      {/* Mobile burger */}
      {isMobile && (
        <button
          onClick={() => setSidebarOpen(!sidebarOpen)}
          style={{
            position: 'fixed', top: 12, left: 12, zIndex: 10002,
            padding: '10px 13px',
            backgroundColor: 'rgba(255,255,255,0.06)',
            backdropFilter: 'blur(12px)',
            border: '1px solid rgba(255,255,255,0.1)',
            borderRadius: 10, color: '#fff', fontSize: 18, cursor: 'pointer'
          }}
        >
          {sidebarOpen ? '✕' : '☰'}
        </button>
      )}

      {/* Sidebar */}
      <aside style={{
        width: 260, minWidth: 260,
        backgroundColor: theme.colors.backgroundSecondary,
        backdropFilter: 'blur(16px)',
        borderRight: `1px solid ${theme.colors.border}`,
        padding: 20,
        display: 'flex',
        flexDirection: 'column',
        ...(isMobile ? {
          position: 'fixed' as const, top: 0,
          left: sidebarOpen ? 0 : -280,
          height: '100vh', zIndex: 10001,
          transition: 'left 0.3s ease',
          boxShadow: sidebarOpen ? '4px 0 20px rgba(0,0,0,0.5)' : 'none'
        } : {})
      }}>
        {/* Logo */}
        <div style={{
          fontSize: 20, fontWeight: 800, color: theme.colors.accent,
          marginBottom: 16, letterSpacing: 2,
          display: 'flex', alignItems: 'center', gap: 8
        }}>
          ✦ XPLR
        </div>

        {/* Mode Toggle */}
        <div style={{
          marginBottom: 24, padding: 5,
          backgroundColor: theme.colors.backgroundCard,
          borderRadius: theme.borderRadius.md,
          border: `1px solid ${theme.colors.border}`,
          display: 'flex', gap: 4
        }}>
          {(['professional', 'personal'] as const).map(mode => {
            const isActive = appMode === mode;
            return (
              <button key={mode} type="button" onClick={() => setMode(mode)}
                style={{
                  flex: 1, padding: '10px 12px', border: 'none', borderRadius: 8,
                  fontSize: 11, fontWeight: 700, cursor: 'pointer',
                  textTransform: 'uppercase', letterSpacing: 1, transition: 'all 0.2s',
                  backgroundColor: isActive ? theme.colors.accent : 'transparent',
                  color: isActive ? theme.colors.background : theme.colors.textSecondary
                }}>
                {mode === 'professional' ? 'ARBITRAGE' : 'PERSONAL'}
              </button>
            );
          })}
        </div>

        {/* Menu Items */}
        {menuItems.map(item => {
          const isActive = activePath === item.path;
          return (
            <div key={item.key}
              onClick={() => { navigate(item.path); setSidebarOpen(false); }}
              style={{
                padding: '12px 15px', cursor: 'pointer',
                borderRadius: theme.borderRadius.sm, marginBottom: 3,
                transition: '0.2s', fontSize: 14,
                display: 'flex', alignItems: 'center', gap: 10,
                backgroundColor: isActive ? theme.colors.accentMuted : 'transparent',
                color: isActive ? theme.colors.accent : theme.colors.textSecondary,
                borderLeft: isActive ? `3px solid ${theme.colors.accent}` : '3px solid transparent'
              }}
              onMouseEnter={e => { if (!isActive) { e.currentTarget.style.backgroundColor = theme.colors.backgroundCard; e.currentTarget.style.color = theme.colors.textPrimary; }}}
              onMouseLeave={e => { if (!isActive) { e.currentTarget.style.backgroundColor = 'transparent'; e.currentTarget.style.color = theme.colors.textSecondary; }}}
            >
              {item.label}
            </div>
          );
        })}

        {/* Referrals */}
        <div
          onClick={() => { navigate('/referrals'); setSidebarOpen(false); }}
          style={{
            padding: '12px 15px', cursor: 'pointer', borderRadius: theme.borderRadius.sm,
            marginBottom: 3, transition: '0.2s', fontSize: 14,
            display: 'flex', alignItems: 'center', gap: 10,
            backgroundColor: activePath === '/referrals' ? theme.colors.accentMuted : 'transparent',
            color: activePath === '/referrals' ? theme.colors.accent : theme.colors.textSecondary,
            borderLeft: activePath === '/referrals' ? `3px solid ${theme.colors.accent}` : '3px solid transparent'
          }}
          onMouseEnter={e => { e.currentTarget.style.backgroundColor = theme.colors.backgroundCard; e.currentTarget.style.color = theme.colors.textPrimary; }}
          onMouseLeave={e => { if (activePath !== '/referrals') { e.currentTarget.style.backgroundColor = 'transparent'; e.currentTarget.style.color = theme.colors.textSecondary; }}}
        >
          🎁 Referrals
        </div>

        {/* Spacer */}
        <div style={{ flex: 1 }} />

        {/* Logout */}
        <div onClick={handleLogout}
          style={{
            padding: '12px 15px', cursor: 'pointer', borderRadius: theme.borderRadius.sm,
            fontSize: 14, display: 'flex', alignItems: 'center', gap: 10,
            color: theme.colors.textSecondary
          }}
          onMouseEnter={e => { e.currentTarget.style.backgroundColor = theme.colors.error + '20'; e.currentTarget.style.color = theme.colors.error; }}
          onMouseLeave={e => { e.currentTarget.style.backgroundColor = 'transparent'; e.currentTarget.style.color = theme.colors.textSecondary; }}
        >
          ⚙️ Logout
        </div>
      </aside>

      {/* Main content area */}
      <div style={{
        flex: 1, overflowY: 'auto',
        backgroundColor: 'rgba(255, 255, 255, 0.015)',
        backdropFilter: 'blur(12px)',
        padding: isMobile ? '60px 16px 16px' : 30
      }}>
        {children}
      </div>
    </div>
  );
};

export default SidebarLayout;
