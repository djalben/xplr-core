import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ModeProvider } from './store/mode-context';
import { RatesProvider } from './store/rates-context';
import { AuthProvider } from './store/auth-context';
import { AuthPage } from './pages/auth';
import { DashboardPage } from './pages/dashboard';
import { CardsPage } from './pages/cards';
import { CardIssuePage } from './pages/card-issue';
import { HistoryPage } from './pages/history';
import { ReferralsPage } from './pages/referrals';
import { SettingsPage } from './pages/settings';
import { SupportPage } from './pages/support';
import { LandingPage } from './pages/landing';
import { AdminRatesPage } from './pages/admin-rates';
import { ForgotPasswordPage } from './pages/forgot-password';
import { ResetPasswordPage } from './pages/reset-password';
import { NewsPage } from './pages/news';
import { StorePage } from './pages/store';
import { PurchasesPage } from './pages/purchases';
import { PWAInstallPrompt } from './components/pwa-install-prompt';
import { NeuralBackground } from './components/neural-background';
import { useAuth } from './store/auth-context';
import { AdminApp } from './pages/admin';

interface GuardProps {
  children: React.ReactNode;
}

/* ── Requires token ── */
const ProtectedRoute: React.FC<GuardProps> = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) return <Navigate to="/auth" replace />;
  return <>{children}</>;
};

/* ── Requires admin role ── */
const AdminRoute: React.FC<GuardProps> = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) {
    return <Navigate to="/auth" replace />;
  }
  const { isAdmin, authReady, user } = useAuth();

  if (!authReady) {
    return null;
  }

  if (!isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }
  return <>{children}</>;
};

const RootRedirect: React.FC = () => {
  const token = localStorage.getItem('token');
  if (token) return <Navigate to="/dashboard" replace />;
  return <Navigate to="/landing" replace />;
};

function App() {
  return (
    <AuthProvider>
    <ModeProvider>
    <RatesProvider>
      <NeuralBackground />
      <PWAInstallPrompt />
      <Routes>
        <Route path="/" element={<RootRedirect />} />
        <Route path="/landing" element={<LandingPage />} />
        <Route path="/auth" element={<AuthPage />} />
        {/* Keep old routes working */}
        <Route path="/login" element={<Navigate to="/auth" replace />} />
        <Route path="/register" element={<Navigate to="/auth" replace />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/reset-password" element={<ResetPasswordPage />} />

        <Route path="/dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
        <Route path="/cards" element={<ProtectedRoute><CardsPage /></ProtectedRoute>} />
        <Route path="/card-issue" element={<ProtectedRoute><CardIssuePage /></ProtectedRoute>} />
        <Route path="/history" element={<ProtectedRoute><HistoryPage /></ProtectedRoute>} />
        <Route path="/finance" element={<Navigate to="/history" replace />} />
        <Route path="/referrals" element={<ProtectedRoute><ReferralsPage /></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
        <Route path="/support" element={<ProtectedRoute><SupportPage /></ProtectedRoute>} />
        <Route path="/news" element={<ProtectedRoute><NewsPage /></ProtectedRoute>} />
        <Route path="/store" element={<ProtectedRoute><StorePage /></ProtectedRoute>} />
        <Route path="/purchases" element={<ProtectedRoute><PurchasesPage /></ProtectedRoute>} />
        <Route path="/admin/rates" element={<ProtectedRoute><AdminRatesPage /></ProtectedRoute>} />
        {/* Admin area (separate pages) */}
        <Route path="/admin/*" element={<AdminRoute><AdminApp /></AdminRoute>} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </RatesProvider>
    </ModeProvider>
    </AuthProvider>
  );
}

export default App;
