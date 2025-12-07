import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { authAPI } from '../services/api';

const AuthContext = createContext(null);

// Helper to decode JWT and get expiry time
const getTokenExpiry = (token) => {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.exp * 1000; // Convert to milliseconds
  } catch {
    return null;
  }
};

// Check if token is expired or will expire soon (within 2 minutes)
const isTokenExpiringSoon = (token, bufferMs = 2 * 60 * 1000) => {
  const expiry = getTokenExpiry(token);
  if (!expiry) return true;
  return Date.now() >= expiry - bufferMs;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [accessToken, setAccessToken] = useState(null);
  const [refreshToken, setRefreshToken] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const refreshTimerRef = useRef(null);
  const isRefreshingRef = useRef(false);

  // Refresh the access token
  const refreshAccessToken = useCallback(async () => {
    const storedRefresh = refreshToken || localStorage.getItem('refreshToken');
    if (!storedRefresh || isRefreshingRef.current) return null;

    isRefreshingRef.current = true;
    try {
      console.log('Refreshing access token...');
      const response = await authAPI.refreshToken(storedRefresh);

      if (response.access_token) {
        console.log('Token refreshed successfully');
        setAccessToken(response.access_token);
        localStorage.setItem('accessToken', response.access_token);

        if (response.refresh_token) {
          setRefreshToken(response.refresh_token);
          localStorage.setItem('refreshToken', response.refresh_token);
        }

        return response.access_token;
      } else {
        console.error('Token refresh failed:', response);
        logout();
        return null;
      }
    } catch (error) {
      console.error('Token refresh error:', error);
      logout();
      return null;
    } finally {
      isRefreshingRef.current = false;
    }
  }, [refreshToken]);

  // Schedule token refresh before expiry
  const scheduleTokenRefresh = useCallback((token) => {
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current);
    }

    const expiry = getTokenExpiry(token);
    if (!expiry) return;

    // Refresh 2 minutes before expiry
    const refreshTime = expiry - Date.now() - 2 * 60 * 1000;

    if (refreshTime > 0) {
      console.log(`Scheduling token refresh in ${Math.round(refreshTime / 1000)} seconds`);
      refreshTimerRef.current = setTimeout(async () => {
        const newToken = await refreshAccessToken();
        if (newToken) {
          scheduleTokenRefresh(newToken);
        }
      }, refreshTime);
    } else {
      // Token is already expiring soon, refresh now
      refreshAccessToken().then(newToken => {
        if (newToken) {
          scheduleTokenRefresh(newToken);
        }
      });
    }
  }, [refreshAccessToken]);

  // Get a valid access token (refresh if needed)
  const getValidAccessToken = useCallback(async () => {
    const currentToken = accessToken || localStorage.getItem('accessToken');

    if (!currentToken) return null;

    if (isTokenExpiringSoon(currentToken)) {
      console.log('Token expiring soon, refreshing...');
      return await refreshAccessToken();
    }

    return currentToken;
  }, [accessToken, refreshAccessToken]);

  useEffect(() => {
    // Check for stored auth data on mount
    const storedToken = localStorage.getItem('accessToken');
    const storedRefresh = localStorage.getItem('refreshToken');
    const storedUser = localStorage.getItem('user');

    if (storedToken && storedUser) {
      // Check if token is expired
      if (isTokenExpiringSoon(storedToken, 0)) {
        // Token is expired, try to refresh
        if (storedRefresh) {
          setRefreshToken(storedRefresh);
          setUser(JSON.parse(storedUser));
          refreshAccessToken().then(newToken => {
            if (newToken) {
              setIsAuthenticated(true);
              scheduleTokenRefresh(newToken);
            }
            setLoading(false);
          });
          return;
        } else {
          // No refresh token, clear everything
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
        }
      } else {
        setAccessToken(storedToken);
        setRefreshToken(storedRefresh);
        setUser(JSON.parse(storedUser));
        setIsAuthenticated(true);
        scheduleTokenRefresh(storedToken);
      }
    }
    setLoading(false);
  }, [refreshAccessToken, scheduleTokenRefresh]);

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (refreshTimerRef.current) {
        clearTimeout(refreshTimerRef.current);
      }
    };
  }, []);

  const login = (userData, tokens) => {
    setUser(userData);
    setAccessToken(tokens.accessToken);
    setRefreshToken(tokens.refreshToken);
    setIsAuthenticated(true);

    localStorage.setItem('accessToken', tokens.accessToken);
    localStorage.setItem('refreshToken', tokens.refreshToken);
    localStorage.setItem('user', JSON.stringify(userData));

    // Schedule token refresh
    scheduleTokenRefresh(tokens.accessToken);
  };

  const logout = () => {
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current);
    }
    setUser(null);
    setAccessToken(null);
    setRefreshToken(null);
    setIsAuthenticated(false);

    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
  };

  const updateUser = (updatedUserData) => {
    const newUser = { ...user, ...updatedUserData };
    setUser(newUser);
    localStorage.setItem('user', JSON.stringify(newUser));
  };

  return (
    <AuthContext.Provider value={{
      user,
      accessToken,
      refreshToken,
      isAuthenticated,
      loading,
      login,
      logout,
      updateUser,
      getValidAccessToken,
      refreshAccessToken,
    }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

