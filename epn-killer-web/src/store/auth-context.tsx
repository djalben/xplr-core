import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';

export type UserRole = 'OWNER' | 'MEMBER';
export type UserMode = 'personal';

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

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUserState] = useState<UserProfile>({
    id: '1',
    name: 'Алексей Петров',
    email: 'alex@xplr.io',
    role: 'OWNER',
    avatar: 'АП',
  });

  const userMode: UserMode = 'personal';
  const onboardingComplete = true;

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
