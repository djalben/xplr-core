import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ModeProvider } from './store/mode-context';
import { RatesProvider } from './store/rates-context';
import { AuthProvider, useAuth } from './store/auth-context';
import { AuthPage } from './pages/auth';
import { OnboardingPage } from './pages/onboarding';
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
import { ForbiddenPage } from './pages/forbidden';
import { PWAInstallPrompt } from './components/pwa-install-prompt';

interface GuardProps {
  children: React.ReactNode;
}

/* ── Requires token ── */
const ProtectedRoute: React.FC<GuardProps> = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) return <Navigate to="/auth" replace />;
  return <>{children}</>;
};

/* ── Requires token + onboarding complete ── */
const OnboardedRoute: React.FC<GuardProps> = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) return <Navigate to="/auth" replace />;
  const { onboardingComplete } = useAuth();
  if (!onboardingComplete) return <Navigate to="/onboarding" replace />;
  return <>{children}</>;
};

/* ── Owner-only (personal tab) — blocked for MEMBER ── */
const OwnerRoute: React.FC<GuardProps> = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) return <Navigate to="/auth" replace />;
  const { isOwner, onboardingComplete } = useAuth();
  if (!onboardingComplete) return <Navigate to="/onboarding" replace />;
  if (!isOwner) return <Navigate to="/forbidden" replace />;
  return <>{children}</>;
};

/* ── Business-only routes (teams, api) — blocked in personal mode ── */
const BusinessRoute: React.FC<GuardProps> = ({ children }) => {
  const token = localStorage.getItem('token');
  if (!token) return <Navigate to="/auth" replace />;
  const { userMode, onboardingComplete } = useAuth();
  if (!onboardingComplete) return <Navigate to="/onboarding" replace />;
  if (userMode === 'personal') return <Navigate to="/forbidden" replace />;
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
      <PWAInstallPrompt />
      <Routes>
        <Route path="/" element={<RootRedirect />} />
        <Route path="/landing" element={<LandingPage />} />
        <Route path="/auth" element={<AuthPage />} />
        <Route path="/onboarding" element={<ProtectedRoute><OnboardingPage /></ProtectedRoute>} />
        {/* Keep old routes working */}
        <Route path="/login" element={<Navigate to="/auth" replace />} />
        <Route path="/register" element={<Navigate to="/auth" replace />} />

        <Route path="/dashboard" element={<OnboardedRoute><DashboardPage /></OnboardedRoute>} />
        <Route path="/cards" element={<OnboardedRoute><CardsPage /></OnboardedRoute>} />
        <Route path="/card-issue" element={<OnboardedRoute><CardIssuePage /></OnboardedRoute>} />
        <Route path="/finance" element={<OnboardedRoute><FinancePage /></OnboardedRoute>} />
        <Route path="/teams" element={<BusinessRoute><TeamsPage /></BusinessRoute>} />
        <Route path="/referrals" element={<OnboardedRoute><ReferralsPage /></OnboardedRoute>} />
        <Route path="/api" element={<BusinessRoute><ApiPage /></BusinessRoute>} />
        <Route path="/settings" element={<OnboardedRoute><SettingsPage /></OnboardedRoute>} />
        <Route path="/support" element={<OnboardedRoute><SupportPage /></OnboardedRoute>} />
        <Route path="/admin/rates" element={<OnboardedRoute><AdminRatesPage /></OnboardedRoute>} />
        <Route path="/forbidden" element={<ProtectedRoute><ForbiddenPage /></ProtectedRoute>} />
      </Routes>
    </RatesProvider>
    </ModeProvider>
    </AuthProvider>
  );
}

export default App;
