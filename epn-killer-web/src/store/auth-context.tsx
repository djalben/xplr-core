import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react';

export type UserRole = 'OWNER' | 'MEMBER';
export type UserMode = 'personal' | 'business';

export interface UserProfile {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatar: string;
}

interface AuthContextType {
  user: UserProfile | null;
  role: UserRole;
  isOwner: boolean;
  isMember: boolean;
  userMode: UserMode;
  onboardingComplete: boolean;
  setUser: (user: UserProfile) => void;
  setUserMode: (mode: UserMode) => void;
  completeOnboarding: (mode: UserMode) => void;
  logout: () => void;
}

const ROLE_KEY = 'xplr_user_role';
const MODE_KEY = 'xplr_user_mode';
const ONBOARDING_KEY = 'xplr_onboarding_complete';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUserState] = useState<UserProfile>(() => {
    const savedRole = localStorage.getItem(ROLE_KEY);
    const role: UserRole = savedRole === 'MEMBER' ? 'MEMBER' : 'OWNER';
    return {
      id: '1',
      name: 'Алексей Петров',
      email: 'alex@xplr.io',
      role,
      avatar: 'АП',
    };
  });

  const [userMode, setUserModeState] = useState<UserMode>(() => {
    const saved = localStorage.getItem(MODE_KEY);
    return saved === 'personal' || saved === 'business' ? saved : 'business';
  });

  const [onboardingComplete, setOnboardingComplete] = useState<boolean>(() => {
    return localStorage.getItem(ONBOARDING_KEY) === 'true';
  });

  useEffect(() => { localStorage.setItem(ROLE_KEY, user.role); }, [user.role]);
  useEffect(() => { localStorage.setItem(MODE_KEY, userMode); }, [userMode]);
  useEffect(() => { localStorage.setItem(ONBOARDING_KEY, String(onboardingComplete)); }, [onboardingComplete]);

  const setUser = useCallback((newUser: UserProfile) => {
    setUserState(newUser);
  }, []);

  const setUserMode = useCallback((mode: UserMode) => {
    setUserModeState(mode);
  }, []);

  const completeOnboarding = useCallback((mode: UserMode) => {
    setUserModeState(mode);
    setOnboardingComplete(true);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem(ROLE_KEY);
    localStorage.removeItem(MODE_KEY);
    localStorage.removeItem(ONBOARDING_KEY);
    setUserState(prev => ({ ...prev, role: 'OWNER' }));
    setOnboardingComplete(false);
  }, []);

  const role = user.role;
  const isOwner = role === 'OWNER';
  const isMember = role === 'MEMBER';

  return (
    <AuthContext.Provider value={{ user, role, isOwner, isMember, userMode, onboardingComplete, setUser, setUserMode, completeOnboarding, logout }}>
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
