import React, { useState, useRef, useCallback, useEffect } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import apiClient from '../services/axios';
import { useRates } from '../store/rates-context';
import { useAuth } from '../store/auth-context';
import {
  LayoutDashboard,
  CreditCard,
  Receipt,
  Gift,
  ChevronRight,
  Menu,
  X,
  LogOut,
  Settings,
  Bell,
  DollarSign,
  HelpCircle,
  Lock,
  Newspaper,
  ShoppingBag
} from 'lucide-react';
import { getUnreadNewsCount } from '../services/news';

const Logo = ({ onTripleClick }: { onTripleClick?: () => void }) => {
  const clickRef = useRef<{ count: number; timer: ReturnType<typeof setTimeout> | null }>({ count: 0, timer: null });

  const handleClick = useCallback(() => {
    const c = clickRef.current;
    c.count += 1;
    if (c.timer) clearTimeout(c.timer);
    if (c.count >= 3) {
      c.count = 0;
      onTripleClick?.();
    } else {
      c.timer = setTimeout(() => { c.count = 0; }, 600);
    }
  }, [onTripleClick]);

  return (
    <div className="flex items-center gap-3 px-2 select-none" onClick={handleClick}>
      <div className="relative w-11 h-11 rounded-xl gradient-accent flex items-center justify-center overflow-hidden shadow-lg shadow-blue-500/30">
        <div className="absolute inset-0 bg-gradient-to-tr from-transparent via-white/20 to-transparent" />
        <span className="text-white font-bold text-xl tracking-tighter">X</span>
      </div>
      <span className="text-2xl font-bold tracking-tight gradient-text">XPLR</span>
    </div>
  );
};

const CurrencyRates = () => {
  const { rates } = useRates();
  return (
    <div className="rate-ticker flex items-center gap-3 text-sm">
      <DollarSign className="w-4 h-4 text-blue-400" />
      <span className="text-slate-400">USD: <strong className="text-white">{rates.usd.toFixed(2)}₽</strong></span>
      <span className="text-slate-400">EUR: <strong className="text-white">{rates.eur.toFixed(2)}₽</strong></span>
    </div>
  );
};


interface NavItemProps {
  href: string;
  icon: React.ReactNode;
  label: string;
  isActive: boolean;
  badge?: number;
  onClick?: () => void;
}

const NavItem = ({ href, icon, label, isActive, badge, onClick }: NavItemProps) => (
  <Link to={href} onClick={onClick}>
    <div
      className={`group flex items-center gap-3 px-4 py-3 rounded-xl cursor-pointer transition-all duration-150 ${
        isActive
          ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/10 text-blue-400 border border-blue-500/30 shadow-lg shadow-blue-500/10'
          : 'text-slate-400 hover:text-white hover:bg-white/5'
      }`}
    >
      <span className={`transition-colors ${isActive ? 'text-blue-400' : 'text-slate-500 group-hover:text-slate-300'}`}>
        {icon}
      </span>
      <span className="flex-1 text-sm font-medium">{label}</span>
      {badge && badge > 0 ? (
        <span className="min-w-[20px] h-5 px-1.5 flex items-center justify-center rounded-full bg-red-500 text-white text-[10px] font-bold leading-none">
          {badge > 99 ? '99+' : badge}
        </span>
      ) : isActive ? <ChevronRight className="w-4 h-4 text-blue-400" /> : null}
    </div>
  </Link>
);

const UserProfile = () => {
  const navigate = useNavigate();
  const { user, logout: authLogout } = useAuth();
  const { t } = useTranslation();

  const handleLogout = () => {
    authLogout();
    navigate('/login');
  };

  return (
    <div className="glass-card p-4">
      <div className="flex items-center gap-3 mb-3">
        <div className="w-11 h-11 rounded-xl gradient-accent flex items-center justify-center shadow-lg shadow-blue-500/25">
          <span className="text-white font-semibold text-sm">{user?.avatar || 'U'}</span>
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-white truncate">{user?.name || 'User'}</p>
          <p className="text-xs text-slate-500 truncate">{user?.role === 'OWNER' ? t('nav.owner') : t('nav.member')}</p>
        </div>
        <div className="relative">
          <Bell className="w-5 h-5 text-slate-500 hover:text-slate-300 cursor-pointer transition-colors" />
        </div>
      </div>
      <div className="flex gap-2">
        <Link to="/settings" className="flex-1">
          <button className="w-full flex items-center justify-center gap-2 py-2 text-xs text-slate-400 hover:text-white hover:bg-white/5 rounded-lg transition-colors">
            <Settings className="w-6 h-6 text-blue-400 animate-[spin_6s_linear_infinite]" />
            {t('nav.settings')}
          </button>
        </Link>
        <button
          onClick={handleLogout}
          className="flex-1 flex items-center justify-center gap-2 py-2 text-xs text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
        >
          <LogOut className="w-4 h-4" />
          {t('nav.logout')}
        </button>
      </div>
    </div>
  );
};

const BottomNavItem = ({ href, icon, label, isActive, badge }: { href: string; icon: React.ReactNode; label: string; isActive: boolean; badge?: number }) => (
  <Link to={href} className="flex-1 min-w-0">
    <div className={`relative flex flex-col items-center justify-center h-full py-1.5 px-0.5 rounded-xl transition-colors ${
      isActive ? 'text-blue-400' : 'text-slate-500'
    }`}>
      <span className="relative shrink-0">
        {icon}
        {badge && badge > 0 ? (
          <span className="absolute -top-1.5 -right-2.5 min-w-[14px] h-3.5 px-0.5 flex items-center justify-center rounded-full bg-violet-500 text-white text-[8px] font-bold leading-none">
            {badge > 99 ? '99+' : badge}
          </span>
        ) : null}
      </span>
      <span className="text-[9px] xs:text-[10px] mt-0.5 font-medium leading-tight text-center truncate w-full">{label}</span>
    </div>
  </Link>
);

// ── Staff PIN Modal ──
const StaffPinModal = ({ open, onClose }: { open: boolean; onClose: () => void }) => {
  const [pin, setPin] = useState('');
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(false);

  if (!open) return null;

  const openAdminInNewTab = () => {
    // Keep `noopener` off so `postMessage` has a reliable target on all browsers.
    // We still guard by `origin` on the receiving side.
    const w = window.open('/admin/entry', '_blank');
    if (!w) {
      // Popup blocked; fall back to same-tab navigation.
      window.location.assign('/admin/entry');
      return;
    }

    const payload = { type: 'xplr_admin_grant' as const };

    const bc = new BroadcastChannel('xplr_admin');
    const sendGrant = () => {
      try {
        bc.postMessage(payload);
      } catch {
        // ignore
      }
      try {
        w.postMessage(payload, window.location.origin);
      } catch {
        // ignore
      }
    };

    // Send immediately, then keep sending for a few seconds to avoid race conditions.
    sendGrant();
    let triesLeft = 80; // ~12s @ 150ms
    const interval = window.setInterval(() => {
      sendGrant();
      triesLeft -= 1;
      if (triesLeft <= 0) {
        window.clearInterval(interval);
        bc.close();
      }
    }, 150);

    // Also react to explicit readiness from the new tab.
    const onReady = (event: MessageEvent) => {
      const data = event.data as { type?: string } | null;
      if (!data || data.type !== 'xplr_admin_ready') return;
      sendGrant();
    };
    bc.addEventListener('message', onReady);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(false);
    try {
      const res = await apiClient.post('/verify-staff-pin', { pin });
      if (res.data?.access === 'granted') {
        setPin('');
        onClose();
        openAdminInNewTab();
      } else {
        setError(true);
        setPin('');
      }
    } catch {
      setError(true);
      setPin('');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={onClose}>
      <form
        onSubmit={handleSubmit}
        onClick={e => e.stopPropagation()}
        className="glass-card p-6 w-full max-w-xs mx-4 space-y-4"
      >
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-red-500 to-purple-600 flex items-center justify-center">
            <Lock className="w-5 h-5 text-white" />
          </div>
          <div>
            <p className="text-sm font-bold text-white">Staff Access</p>
            <p className="text-xs text-slate-500">Enter PIN to continue</p>
          </div>
        </div>
        <input
          type="password"
          inputMode="numeric"
          pattern="[0-9]*"
          maxLength={4}
          value={pin}
          onChange={e => { setPin(e.target.value.replace(/[^\d]/g, '').slice(0, 4)); setError(false); }}
          placeholder="PIN"
          autoFocus
          className={`w-full px-4 py-3 bg-white/5 border rounded-xl text-white text-center text-lg tracking-[0.3em] font-mono placeholder-slate-600 outline-none transition-colors ${
            error ? 'border-red-500/60 shake' : 'border-white/10 focus:border-blue-500/50'
          }`}
        />
        {error && <p className="text-xs text-red-400 text-center">Invalid PIN</p>}
        <button
          type="submit"
          disabled={loading || pin.length !== 4}
          className="w-full py-2.5 bg-gradient-to-r from-blue-500 to-purple-600 text-white rounded-xl text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-60"
        >
          {loading ? '...' : 'Authenticate'}
        </button>
      </form>
    </div>
  );
};

export const Sidebar = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const [unreadNews, setUnreadNews] = useState(0);
  const { isAdmin } = useAuth();

  const { t } = useTranslation();

  // Fetch unread news count on mount + poll every 60s
  useEffect(() => {
    const fetchUnread = () => {
      getUnreadNewsCount().then(r => setUnreadNews(r.count)).catch(() => {});
    };
    fetchUnread();
    const interval = setInterval(fetchUnread, 60000);
    return () => clearInterval(interval);
  }, []);

  // Reset unread count when user navigates to /news
  useEffect(() => {
    if (location.pathname === '/news') {
      setUnreadNews(0);
    }
  }, [location.pathname]);

  const handleLogoTripleClick = useCallback(() => {
    if (isAdmin) setStaffModalOpen(true);
  }, [isAdmin]);

  const navItems = [
    { href: '/dashboard', icon: <LayoutDashboard className="w-5 h-5" />, label: t('nav.home'), badge: 0 },
    { href: '/cards', icon: <CreditCard className="w-5 h-5" />, label: t('nav.cards'), badge: 0 },
    { href: '/history', icon: <Receipt className="w-5 h-5" />, label: 'История', badge: 0 },
    { href: '/referrals', icon: <Gift className="w-5 h-5" />, label: t('nav.referrals'), badge: 0 },
    { href: '/store', icon: <ShoppingBag className="w-5 h-5" />, label: 'Магазин', badge: 0 },
    { href: '/news', icon: <Newspaper className="w-5 h-5" />, label: 'Новости', badge: unreadNews },
    { href: '/support', icon: <HelpCircle className="w-5 h-5" />, label: t('nav.support'), badge: 0 },
  ];

  const bottomNavItems = navItems.slice(0, 6);

  return (
    <>
      {/* Desktop Sidebar */}
      <aside className="fixed left-0 top-0 w-64 h-screen hidden lg:flex flex-col z-50">
        <div className="absolute inset-0 bg-[#0a0a0f]/80 backdrop-blur-xl border-r border-white/[0.08]" />
        <div className="absolute inset-0 bg-gradient-to-b from-blue-500/[0.03] via-transparent to-purple-500/[0.02]" />

        <div className="relative flex flex-col h-full p-4">
          <div className="mb-6 pt-2">
            <Logo onTripleClick={handleLogoTripleClick} />
          </div>
          <div className="mb-4">
            <CurrencyRates />
          </div>
          <nav className="flex-1 min-h-0 overflow-y-auto space-y-1 scrollbar-thin">
            {navItems.map(item => (
              <NavItem
                key={item.href}
                href={item.href}
                icon={item.icon}
                label={item.label}
                isActive={location.pathname === item.href}
                badge={item.badge}
              />
            ))}
          </nav>
          <div className="mt-auto pt-4">
            <p className="text-[10px] text-slate-600 text-center mb-3 tracking-wide">{t('brand.tagline')}</p>
            <UserProfile />
          </div>
        </div>
      </aside>

      {/* Mobile Header */}
      <header className="lg:hidden fixed top-0 left-0 right-0 z-50 bg-[#0a0a0f]/90 backdrop-blur-xl border-b border-white/[0.08]" style={{ paddingTop: 'env(safe-area-inset-top, 0px)' }}>
        <div className="flex items-center justify-between px-4 py-3">
          <Logo onTripleClick={handleLogoTripleClick} />
          <div className="flex items-center gap-3">
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="p-2.5 hover:bg-white/5 rounded-xl transition-colors"
            >
              {mobileMenuOpen ? <X className="w-6 h-6 text-white" /> : <Menu className="w-6 h-6 text-white" />}
            </button>
          </div>
        </div>
      </header>

      {/* Mobile Menu Overlay */}
      {mobileMenuOpen && (
        <div className="lg:hidden fixed inset-0 z-40 pt-28">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setMobileMenuOpen(false)} />
          <div className="relative glass-strong mx-4 p-4">
            <div className="mb-4">
              <CurrencyRates />
            </div>
            <nav className="space-y-1">
              {navItems.map(item => (
                <NavItem
                  key={item.href}
                  href={item.href}
                  icon={item.icon}
                  label={item.label}
                  isActive={location.pathname === item.href}
                  onClick={() => setMobileMenuOpen(false)}
                />
              ))}
            </nav>
            <div className="mt-4 pt-4 border-t border-white/10">
              <UserProfile />
            </div>
          </div>
        </div>
      )}

      {/* Mobile Bottom Navigation */}
      <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-[#0a0a0f]/90 backdrop-blur-xl border-t border-white/[0.08]" style={{ paddingBottom: 'env(safe-area-inset-bottom, 0px)' }}>
        <div className="flex items-stretch justify-around px-1 py-1">
          {bottomNavItems.map(item => (
            <BottomNavItem
              key={item.href}
              href={item.href}
              icon={React.cloneElement(item.icon as React.ReactElement<any>, { className: 'w-4 h-4' })}
              label={item.label}
              isActive={location.pathname === item.href}
              badge={item.badge}
            />
          ))}
        </div>
      </nav>

      {/* Staff PIN Modal */}
      <StaffPinModal open={staffModalOpen} onClose={() => setStaffModalOpen(false)} />
    </>
  );
};
