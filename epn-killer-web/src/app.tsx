import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ModeProvider } from './store/mode-context';
import { RatesProvider } from './store/rates-context';
import { AuthPage } from './pages/auth';
import { DashboardPage } from './pages/dashboard';
import { CardsPage } from './pages/cards';
import { CardIssuePage } from './pages/card-issue';
import { FinancePage } from './pages/finance';
import { TeamsPage } from './pages/teams';
import { ReferralsPage } from './pages/referrals';
import { ApiPage } from './pages/api';
import { SettingsPage } from './pages/settings';
import { SupportPage } from './pages/support';
import { LandingPage } from './pages/landing';
import { AdminRatesPage } from './pages/admin-rates';
import { PWAInstallPrompt } from './components/pwa-install-prompt';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) {
    return <Navigate to="/auth" replace />;
  }
  return <>{children}</>;
};

const RootRedirect: React.FC = () => {
  const token = localStorage.getItem('token');
  if (token) {
    return <Navigate to="/dashboard" replace />;
  }
  return <Navigate to="/landing" replace />;
};

function App() {
  return (
    <ModeProvider>
    <RatesProvider>
      <PWAInstallPrompt />
      <Routes>
        <Route path="/" element={<RootRedirect />} />
        <Route path="/landing" element={<LandingPage />} />
        <Route path="/auth" element={<AuthPage />} />
        {/* Keep old routes working */}
        <Route path="/login" element={<Navigate to="/auth" replace />} />
        <Route path="/register" element={<Navigate to="/auth" replace />} />

        <Route path="/dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
        <Route path="/cards" element={<ProtectedRoute><CardsPage /></ProtectedRoute>} />
        <Route path="/card-issue" element={<ProtectedRoute><CardIssuePage /></ProtectedRoute>} />
        <Route path="/finance" element={<ProtectedRoute><FinancePage /></ProtectedRoute>} />
        <Route path="/teams" element={<ProtectedRoute><TeamsPage /></ProtectedRoute>} />
        <Route path="/referrals" element={<ProtectedRoute><ReferralsPage /></ProtectedRoute>} />
        <Route path="/api" element={<ProtectedRoute><ApiPage /></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
        <Route path="/support" element={<ProtectedRoute><SupportPage /></ProtectedRoute>} />
        <Route path="/admin/rates" element={<ProtectedRoute><AdminRatesPage /></ProtectedRoute>} />
      </Routes>
    </RatesProvider>
    </ModeProvider>
  );
}

export default App;
