import React from 'react';
import { Routes, Route, Navigate, Link, useLocation } from 'react-router-dom';
import { AdminLayout } from '../../components/admin-layout';
import { Activity, Shield } from 'lucide-react';
import { AdminDashboardPage } from './dashboard';
import { AdminUsersPage } from './users';
import { AdminTicketsPage } from './tickets';
import { AdminCommissionsPage } from './commissions';
import { AdminStorePage } from './store';
import { AdminNewsPage } from './news';
import { AdminSystemSettingsPage } from './system-settings';
import { AdminLogsPage } from './logs';

const NavLink = ({ to, label, icon }: { to: string; label: string; icon?: React.ReactNode }) => {
  const loc = useLocation();
  const active = loc.pathname === to;
  return (
    <Link
      to={to}
      className={`inline-flex items-center gap-2 px-4 py-2 rounded-xl text-xs border transition-colors ${
        active ? 'bg-blue-500/20 border-blue-500/30 text-blue-200' : 'bg-white/5 border-white/10 text-slate-300 hover:bg-white/10'
      }`}
    >
      {icon ? <span className={active ? 'text-blue-300' : 'text-slate-400'}>{icon}</span> : null}
      {label}
    </Link>
  );
};

export const AdminApp: React.FC = () => {
  return (
    <AdminLayout>
      <div className="stagger-fade-in space-y-6">
        <div className="flex items-start gap-4">
          <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-fuchsia-500/25 to-purple-500/25 border border-fuchsia-500/30 flex items-center justify-center">
            <Shield className="w-7 h-7 text-fuchsia-200" />
          </div>
          <div className="min-w-0">
            <h1 className="text-2xl md:text-3xl font-bold text-white leading-tight">Admin Panel</h1>
            <p className="text-sm text-slate-400 mt-1">Закрытая зона управления</p>
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <NavLink
            to="/admin/dashboard"
            label="Dashboard"
            icon={
              <span className="w-6 h-6 rounded-lg bg-gradient-to-br from-blue-500/30 to-purple-500/30 border border-blue-500/30 flex items-center justify-center">
                <Activity className="w-3.5 h-3.5" />
              </span>
            }
          />
          <NavLink to="/admin/users" label="Пользователи" />
          <NavLink to="/admin/tickets" label="Тикеты" />
          <NavLink to="/admin/commissions" label="Комиссии" />
          <NavLink to="/admin/system-settings" label="Системные настройки" />
          <NavLink to="/admin/store" label="Магазин" />
          <NavLink to="/admin/news" label="Новости" />
          <NavLink to="/admin/logs" label="Логи" />
        </div>

        <Routes>
          <Route path="/" element={<Navigate to="/admin/dashboard" replace />} />
          <Route path="/dashboard" element={<AdminDashboardPage />} />
          <Route path="/news" element={<AdminNewsPage />} />
          <Route path="/store" element={<AdminStorePage />} />
          <Route path="/commissions" element={<AdminCommissionsPage />} />
          <Route path="/system-settings" element={<AdminSystemSettingsPage />} />
          <Route path="/logs" element={<AdminLogsPage />} />
          <Route path="/tickets" element={<AdminTicketsPage />} />
          <Route path="/users" element={<AdminUsersPage />} />
          <Route path="*" element={<Navigate to="/admin/dashboard" replace />} />
        </Routes>
      </div>
    </AdminLayout>
  );
};

