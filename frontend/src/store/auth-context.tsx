import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react';
import apiClient from '../api/axios';

export type UserRole = 'OWNER' | 'MEMBER';
export type UserMode = 'personal';

export interface UserProfile {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar: string;
  isAdmin?: boolean;
  serverRole?: string;
}

interface AuthContextType {
  user: UserProfile | null;
  role: UserRole;
  isOwner: boolean;
  isMember: boolean;
  isAdmin: boolean;
  authReady: boolean;
  userMode: UserMode;
  onboardingComplete: boolean;
  setUser: (user: UserProfile) => void;
  updateUserName: (name: string) => void;
  setUserMode: (mode: UserMode) => void;
  completeOnboarding: (mode: UserMode) => void;
  refreshSession: () => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUserState] = useState<UserProfile>({
    id: '0',
    name: '',
    email: '',
    role: 'OWNER',
    avatar: '',
    isAdmin: false,
    serverRole: 'user',
  });
  const [authReady, setAuthReady] = useState(false);

  const userMode: UserMode = 'personal';
  const onboardingComplete = true;

  // Helper: apply server data to local user state
  const applyServerData = useCallback((d: Record<string, unknown>) => {
    const adminFlag = d.is_admin === true || d.role === 'admin';
    const displayName = (d.display_name as string) || '';
    const emailStr = (d.email as string) || '';
    const resolvedName = displayName || emailStr.split('@')[0] || '';
    console.log('[AUTH-CONTEXT] Server data received:', {
      id: d.id, email: d.email, display_name: d.display_name, is_admin: d.is_admin, role: d.role,
      computed_isAdmin: adminFlag, resolvedName,
    });
    setUserState(prev => ({
      ...prev,
      id: String(d.id ?? prev.id),
      email: emailStr || prev.email,
      name: resolvedName || prev.name,
      avatar: (resolvedName || emailStr || '??').substring(0, 2).toUpperCase(),
      isAdmin: adminFlag,
      serverRole: (d.role as string) || 'user',
    }));
  }, []);

  // Fetch /user/me on mount (if token exists)
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      console.log('[AUTH-CONTEXT] No token found, marking authReady');
      setAuthReady(true);
      return;
    }
    console.log('[AUTH-CONTEXT] Token found, fetching /user/me...');
    apiClient.get('/user/me').then(res => {
      applyServerData(res.data);
    }).catch(err => {
      console.error('[AUTH-CONTEXT] ❌ /user/me FAILED:', err.response?.status, err.response?.data || err.message);
      // Try to refresh the token to get fresh data
      apiClient.post('/auth/refresh-token').then(res => {
        const d = res.data;
        if (d.token) {
          localStorage.setItem('token', d.token);
          console.log('[AUTH-CONTEXT] ✅ Token refreshed, applying user data');
          applyServerData(d.user);
        }
      }).catch(refreshErr => {
        console.error('[AUTH-CONTEXT] ❌ Token refresh also failed:', refreshErr.message);
      });
    }).finally(() => {
      setAuthReady(true);
    });
  }, [applyServerData]);

  const setUser = useCallback((newUser: UserProfile) => {
    setUserState(newUser);
  }, []);

  const updateUserName = useCallback((name: string) => {
    setUserState(prev => ({
      ...prev,
      name: name || prev.name,
      avatar: (name || prev.name || '??').substring(0, 2).toUpperCase(),
    }));
  }, []);

  const setUserMode = useCallback((_mode: UserMode) => {}, []);
  const completeOnboarding = useCallback((_mode: UserMode) => {}, []);

  // refreshSession: call /auth/refresh-token, update local token + state
  const refreshSession = useCallback(async () => {
    try {
      const res = await apiClient.post('/auth/refresh-token');
      const d = res.data;
      if (d.token) {
        localStorage.setItem('token', d.token);
        applyServerData(d.user);
        console.log('[AUTH-CONTEXT] ✅ Session refreshed');
      }
    } catch (err) {
      console.error('[AUTH-CONTEXT] refreshSession failed:', err);
    }
  }, [applyServerData]);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    sessionStorage.removeItem('_xplr_staff');
    setUserState(prev => ({ ...prev, isAdmin: false, serverRole: 'user' }));
  }, []);

  const role = user.role;
  const isOwner = role === 'OWNER';
  const isMember = role === 'MEMBER';
  const isAdmin = user.isAdmin === true;

  return (
    <AuthContext.Provider value={{ user, role, isOwner, isMember, isAdmin, authReady, userMode, onboardingComplete, setUser, updateUserName, setUserMode, completeOnboarding, refreshSession, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
