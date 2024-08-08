import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Route, Routes, Navigate } from 'react-router-dom';
import { isAuthenticated } from './utils/auth';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import GamePage from './pages/GamePage';

function App() {
    const [auth, setAuth] = useState(null);

    useEffect(() => {
        async function checkAuth() {
            const isAuth = await isAuthenticated();
            setAuth(isAuth);
        }

        checkAuth();
    }, []);

    if (auth === null) {
        return <div>Loading...</div>;
    }

    return (
        <Router>
            <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/login" element={auth ? <Navigate to="/game" /> : <LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />
                <Route path="/game" element={auth ? <GamePage /> : <Navigate to="/login" />} />
            </Routes>
        </Router>
    );
}

export default App;
