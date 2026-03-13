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
  setUserMode: (mode: UserMode) => void;
  completeOnboarding: (mode: UserMode) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUserState] = useState<UserProfile>({
    id: '1',
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

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      setAuthReady(true);
      return;
    }
    apiClient.get('/user/me').then(res => {
      const d = res.data;
      setUserState(prev => ({
        ...prev,
        id: String(d.id),
        email: d.email || prev.email,
        name: d.email?.split('@')[0] || prev.name,
        avatar: (d.email || '??').substring(0, 2).toUpperCase(),
        isAdmin: d.is_admin === true || d.role === 'admin',
        serverRole: d.role || 'user',
      }));
    }).catch(() => {}).finally(() => {
      setAuthReady(true);
    });
  }, []);

  const setUser = useCallback((newUser: UserProfile) => {
    setUserState(newUser);
  }, []);

  const setUserMode = useCallback((_mode: UserMode) => {}, []);
  const completeOnboarding = useCallback((_mode: UserMode) => {}, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
  }, []);

  const role = user.role;
  const isOwner = role === 'OWNER';
  const isMember = role === 'MEMBER';
  const isAdmin = user.isAdmin === true;

  return (
    <AuthContext.Provider value={{ user, role, isOwner, isMember, isAdmin, authReady, userMode, onboardingComplete, setUser, setUserMode, completeOnboarding, logout }}>
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
