export const isAuthenticated = () => {
    // TODO seems a bit soft
    return localStorage.getItem('token') !== null;
};

export function logout() {
    localStorage.removeItem('token');
    window.location = "/";
};