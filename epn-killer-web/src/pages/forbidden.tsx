import { DashboardLayout } from '../components/dashboard-layout';
import { ShieldX } from 'lucide-react';
import { Link } from 'react-router-dom';

export const ForbiddenPage = () => (
  <DashboardLayout>
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center px-4">
      <div className="w-20 h-20 rounded-2xl bg-red-500/10 flex items-center justify-center mb-6">
        <ShieldX className="w-10 h-10 text-red-400" />
      </div>
      <h1 className="text-4xl font-bold text-white mb-3">403</h1>
      <p className="text-xl text-slate-300 mb-2">Доступ ограничен</p>
      <p className="text-slate-500 mb-8 max-w-md">
        Этот раздел доступен только владельцу аккаунта. Обратитесь к администратору для получения доступа.
      </p>
      <Link
        to="/dashboard"
        className="px-6 py-3 gradient-accent text-white font-medium rounded-xl transition-all shadow-lg shadow-blue-500/25 hover:shadow-blue-500/40"
      >
        Вернуться на главную
      </Link>
    </div>
  </DashboardLayout>
);
