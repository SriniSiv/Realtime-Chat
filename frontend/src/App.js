import React, { useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Login from './components/Login';
import Register from './components/Register';
import ChatRoom from './components/ChatRoom';

// Loading component
const LoadingScreen = () => (
  <div className="loading-screen">
    <div className="loader"></div>
    <p>Loading...</p>
  </div>
);

// Protected route wrapper
const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) return <LoadingScreen />;
  if (!isAuthenticated) return <Navigate to="/login" replace />;

  return children;
};

// Auth pages wrapper (redirects to chat if already logged in)
const AuthRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) return <LoadingScreen />;
  if (isAuthenticated) return <Navigate to="/chat" replace />;

  return children;
};

// Login page with switch to register
const LoginPage = () => {
  const [showRegister, setShowRegister] = useState(false);

  return showRegister ? (
    <Register onSwitchToLogin={() => setShowRegister(false)} />
  ) : (
    <Login onSwitchToRegister={() => setShowRegister(true)} />
  );
};

const App = () => {
  return (
    <BrowserRouter basename="/realtimechat">
      <AuthProvider>
        <Routes>
          <Route path="/login" element={
            <AuthRoute>
              <LoginPage />
            </AuthRoute>
          } />
          <Route path="/chat" element={
            <ProtectedRoute>
              <ChatRoom />
            </ProtectedRoute>
          } />
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
};

export default App;

