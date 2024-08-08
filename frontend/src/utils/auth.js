import '../axiosConfig';
import axios from 'axios';

// Function to get new access token using the refresh token
async function refreshTokens() {
    const refreshToken = localStorage.getItem('refresh_token');
    
    if (!refreshToken) {
        throw new Error('No refresh token available');
    }

    try {
        const response = await axios.post('/refresh-token', { refresh_token: refreshToken });
        const { access_token } = response.data;
        
        localStorage.setItem('token', access_token);

        return access_token;
    } catch (error) {
        console.error('Token refresh failed:', error);
        // Invalidate all tokens and redirect to home page
        logout();
        throw error;
    }
}

export async function isAuthenticated() {
    let token = localStorage.getItem('token');
    
    if (!token) {
        return false;
    }
    
    try {
        const response = await axios.get('/verify-token', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.status === 200) {
            return true;
        } else {
            // Token is invalid, try refreshing tokens
            token = await refreshTokens();
            return token ? true : false;
        }
    } catch (error) {
        console.error('Token verification failed:', error);
        // Attempt to refresh tokens if possible
        try {
            await refreshTokens();
            return true;
        } catch (err) {
            return false;
        }
    }
}

export function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('refresh_token');
    window.location = "/";
}