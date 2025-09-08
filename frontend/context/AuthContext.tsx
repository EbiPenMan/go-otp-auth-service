
import React, { createContext, useState, useContext, useEffect, ReactNode, useCallback } from 'react';

interface AuthContextType {
  isAuthenticated: boolean;
  token: string | null;
  login: (token: string) => void;
  logout: () => void;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    try {
      const storedToken = localStorage.getItem('authToken');
      if (storedToken) {
        setToken(storedToken);
      }
    } catch (error) {
      console.error('Failed to access localStorage:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const login = useCallback((newToken: string) => {
    setToken(newToken);
    try {
      localStorage.setItem('authToken', newToken);
    } catch (error) {
      console.error('Failed to set token in localStorage:', error);
    }
  }, []);

  const logout = useCallback(() => {
    setToken(null);
    try {
      localStorage.removeItem('authToken');
    } catch (error) {
      console.error('Failed to remove token from localStorage:', error);
    }
  }, []);

  const value = {
    isAuthenticated: !!token,
    token,
    login,
    logout,
    isLoading
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
