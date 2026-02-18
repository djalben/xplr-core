import React, { useState } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { useMode } from '../store/mode-context';
import {
  LayoutDashboard,
  CreditCard,
  Receipt,
  Users,
  Gift,
  Code,
  ChevronRight,
  User,
  Menu,
  X,
  LogOut,
  Settings,
  Bell,
  DollarSign,
  HelpCircle
} from 'lucide-react';

const Logo = () => (
  <div className="flex items-center gap-3 px-2">
    <div className="relative w-11 h-11 rounded-xl gradient-accent flex items-center justify-center overflow-hidden shadow-lg shadow-blue-500/30">
      <div className="absolute inset-0 bg-gradient-to-tr from-transparent via-white/20 to-transparent" />
      <span className="text-white font-bold text-xl tracking-tighter">X</span>
    </div>
    <span className="text-2xl font-bold tracking-tight gradient-text">XPLR</span>
  </div>
);

const CurrencyRates = () => (
  <div className="rate-ticker flex items-center gap-3 text-sm">
    <DollarSign className="w-4 h-4 text-blue-400" />
    <span className="text-slate-400">USD: <strong className="text-white">89.45₽</strong></span>
    <span className="text-slate-400">EUR: <strong className="text-white">97.82₽</strong></span>
  </div>
);

const ModeToggle = () => {
  const { mode, setMode } = useMode();

  return (
    <div className="glass-card p-1.5 flex relative">
      <div
        className="absolute top-1.5 bottom-1.5 w-[calc(50%-6px)] gradient-accent rounded-xl transition-all duration-300 ease-out shadow-lg shadow-blue-500/25"
        style={{ left: mode === 'PERSONAL' ? '6px' : 'calc(50%)' }}
      />
      <button
        onClick={() => setMode('PERSONAL')}
        className={`relative z-10 flex-1 px-4 py-2.5 text-xs font-bold tracking-wide rounded-xl transition-colors duration-300 ${
          mode === 'PERSONAL' ? 'text-white' : 'text-slate-500 hover:text-slate-300'
        }`}
      >
        ЛИЧНЫЙ
      </button>
      <button
        onClick={() => setMode('ARBITRAGE')}
        className={`relative z-10 flex-1 px-4 py-2.5 text-xs font-bold tracking-wide rounded-xl transition-colors duration-300 ${
          mode === 'ARBITRAGE' ? 'text-white' : 'text-slate-500 hover:text-slate-300'
        }`}
      >
        АРБИТРАЖ
      </button>
    </div>
  );
};

interface NavItemProps {
  href: string;
  icon: React.ReactNode;
  label: string;
  isActive: boolean;
  onClick?: () => void;
}

const NavItem = ({ href, icon, label, isActive, onClick }: NavItemProps) => (
  <Link to={href} onClick={onClick}>
    <div
      className={`group flex items-center gap-3 px-4 py-3 rounded-xl cursor-pointer transition-all duration-200 ${
        isActive
          ? 'bg-gradient-to-r from-blue-500/20 to-purple-500/10 text-blue-400 border border-blue-500/30 shadow-lg shadow-blue-500/10'
          : 'text-slate-400 hover:text-white hover:bg-white/5'
      }`}
    >
      <span className={`transition-colors ${isActive ? 'text-blue-400' : 'text-slate-500 group-hover:text-slate-300'}`}>
        {icon}
      </span>
      <span className="flex-1 text-sm font-medium">{label}</span>
      {isActive && <ChevronRight className="w-4 h-4 text-blue-400" />}
    </div>
  </Link>
);

const UserProfile = () => {
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    <div className="glass-card p-4">
      <div className="flex items-center gap-3 mb-3">
        <div className="w-11 h-11 rounded-xl gradient-accent flex items-center justify-center shadow-lg shadow-blue-500/25">
          <User className="w-5 h-5 text-white" />
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-white truncate">Пользователь</p>
          <p className="text-xs text-slate-500 truncate">XPLR</p>
        </div>
        <div className="relative">
          <Bell className="w-5 h-5 text-slate-500 hover:text-slate-300 cursor-pointer transition-colors" />
        </div>
      </div>
      <div className="flex gap-2">
        <Link to="/settings" className="flex-1">
          <button className="w-full flex items-center justify-center gap-2 py-2 text-xs text-slate-400 hover:text-white hover:bg-white/5 rounded-lg transition-colors">
            <Settings className="w-4 h-4" />
            Настройки
          </button>
        </Link>
        <button
          onClick={handleLogout}
          className="flex-1 flex items-center justify-center gap-2 py-2 text-xs text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
        >
          <LogOut className="w-4 h-4" />
          Выйти
        </button>
      </div>
    </div>
  );
};

const BottomNavItem = ({ href, icon, label, isActive }: { href: string; icon: React.ReactNode; label: string; isActive: boolean }) => (
  <Link to={href}>
    <div className={`flex flex-col items-center justify-center min-h-[44px] min-w-[44px] py-2 px-3 rounded-xl transition-colors ${
      isActive ? 'text-blue-400' : 'text-slate-500'
    }`}>
      {icon}
      <span className="text-[10px] mt-1 font-medium">{label}</span>
    </div>
  </Link>
);

export const Sidebar = () => {
  const location = useLocation();
  const { mode } = useMode();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navItems = [
    { href: '/dashboard', icon: <LayoutDashboard className="w-5 h-5" />, label: 'Главная', showIn: ['PERSONAL', 'ARBITRAGE'] },
    { href: '/cards', icon: <CreditCard className="w-5 h-5" />, label: 'Карты', showIn: ['PERSONAL', 'ARBITRAGE'] },
    { href: '/finance', icon: <Receipt className="w-5 h-5" />, label: 'Финансы', showIn: ['PERSONAL', 'ARBITRAGE'] },
    { href: '/teams', icon: <Users className="w-5 h-5" />, label: 'Команды', showIn: ['ARBITRAGE'] },
    { href: '/referrals', icon: <Gift className="w-5 h-5" />, label: 'Партнёрка', showIn: ['PERSONAL', 'ARBITRAGE'] },
    { href: '/api', icon: <Code className="w-5 h-5" />, label: 'API', showIn: ['ARBITRAGE'] },
    { href: '/support', icon: <HelpCircle className="w-5 h-5" />, label: 'Поддержка', showIn: ['PERSONAL', 'ARBITRAGE'] },
  ];

  const filteredNavItems = navItems.filter(item => item.showIn.includes(mode));
  const bottomNavItems = filteredNavItems.slice(0, 5);

  return (
    <>
      {/* Desktop Sidebar */}
      <aside className="fixed left-0 top-0 bottom-0 w-64 hidden lg:flex flex-col z-50">
        <div className="absolute inset-0 bg-[#0a0a0f]/80 backdrop-blur-xl border-r border-white/[0.08]" />
        <div className="absolute inset-0 bg-gradient-to-b from-blue-500/[0.03] via-transparent to-purple-500/[0.02]" />

        <div className="relative flex flex-col h-full p-4">
          <div className="mb-6 pt-2">
            <Logo />
          </div>
          <div className="mb-4">
            <CurrencyRates />
          </div>
          <div className="mb-6">
            <ModeToggle />
          </div>
          <nav className="flex-1 space-y-1">
            {filteredNavItems.map(item => (
              <NavItem
                key={item.href}
                href={item.href}
                icon={item.icon}
                label={item.label}
                isActive={location.pathname === item.href}
              />
            ))}
          </nav>
          <div className="mt-auto pt-4">
            <UserProfile />
          </div>
        </div>
      </aside>

      {/* Mobile Header */}
      <header className="lg:hidden fixed top-0 left-0 right-0 z-50 bg-[#0a0a0f]/90 backdrop-blur-xl border-b border-white/[0.08]" style={{ paddingTop: 'env(safe-area-inset-top, 0px)' }}>
        <div className="flex items-center justify-between px-4 py-3">
          <Logo />
          <div className="flex items-center gap-3">
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="p-2.5 hover:bg-white/5 rounded-xl transition-colors"
            >
              {mobileMenuOpen ? <X className="w-6 h-6 text-white" /> : <Menu className="w-6 h-6 text-white" />}
            </button>
          </div>
        </div>
        <div className="px-4 pb-3">
          <ModeToggle />
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
              {filteredNavItems.map(item => (
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
        <div className="flex items-center justify-around px-2 py-1.5">
          {bottomNavItems.map(item => (
            <BottomNavItem
              key={item.href}
              href={item.href}
              icon={React.cloneElement(item.icon as React.ReactElement<any>, { className: 'w-5 h-5' })}
              label={item.label}
              isActive={location.pathname === item.href}
            />
          ))}
        </div>
      </nav>
    </>
  );
};
