import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import { isAuthenticated } from './utils/auth';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import GamePage from './pages/GamePage';

function App() {
    return (
      <Router>
        <Routes>
          <Route path="/" exact element={<HomePage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/game" element={isAuthenticated() ? <GamePage /> : <LoginPage />} />
        </Routes>
      </Router>
    );
  }

export default App;