import { createContext, useContext, useState, useCallback, useEffect, type ReactNode } from 'react';

export type UserRole = 'OWNER' | 'MEMBER';

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
  setUser: (user: UserProfile) => void;
  logout: () => void;
}

const STORAGE_KEY = 'xplr_user_role';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUserState] = useState<UserProfile>(() => {
    const savedRole = localStorage.getItem(STORAGE_KEY);
    const role: UserRole = savedRole === 'MEMBER' ? 'MEMBER' : 'OWNER';
    return {
      id: '1',
      name: 'Алексей Петров',
      email: 'alex@xplr.io',
      role,
      avatar: 'АП',
    };
  });

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, user.role);
  }, [user.role]);

  const setUser = useCallback((newUser: UserProfile) => {
    setUserState(newUser);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    localStorage.removeItem(STORAGE_KEY);
    setUserState(prev => ({ ...prev, role: 'OWNER' }));
  }, []);

  const role = user.role;
  const isOwner = role === 'OWNER';
  const isMember = role === 'MEMBER';

  return (
    <AuthContext.Provider value={{ user, role, isOwner, isMember, setUser, logout }}>
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
