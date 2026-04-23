import React from 'react';
import { Routes, Route, Navigate, Link, useLocation } from 'react-router-dom';
import { AdminLayout } from '../../components/admin-layout';
import { AdminDashboardPage } from './dashboard';
import { AdminUsersPage } from './users';
import { AdminTicketsPage } from './tickets';
import { AdminCommissionsPage } from './commissions';
import { AdminStorePage } from './store';
import { AdminNewsPage } from './news';
import { AdminSystemSettingsPage } from './system-settings';
import { AdminLogsPage } from './logs';

const NavLink = ({ to, label }: { to: string; label: string }) => {
  const loc = useLocation();
  const active = loc.pathname === to;
  return (
    <Link
      to={to}
      className={`px-3 py-2 rounded-lg text-xs border transition-colors ${
        active ? 'bg-blue-500/20 border-blue-500/30 text-blue-300' : 'bg-white/5 border-white/10 text-slate-300 hover:bg-white/10'
      }`}
    >
      {label}
    </Link>
  );
};

const Placeholder = ({ title }: { title: string }) => (
  <div className="glass-card p-6">
    <h1 className="text-xl font-bold text-white mb-2">{title}</h1>
    <p className="text-sm text-slate-400">Страница в разработке. API уже вынесено в отдельные админские ручки.</p>
  </div>
);

export const AdminApp: React.FC = () => {
  return (
    <AdminLayout>
      <div className="stagger-fade-in space-y-6">
        <div className="flex flex-wrap gap-2">
          <NavLink to="/admin/dashboard" label="Dashboard" />
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

