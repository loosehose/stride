import React, { createContext, useContext, useState, useEffect } from 'react';

export const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [apiKey, setApiKey] = useState('');
    const [loading, setLoading] = useState(false); // Define loading state
    const [error, setError] = useState(''); // Define error state

    useEffect(() => {
        const storedApiKey = localStorage.getItem('apiKey');
        if (storedApiKey) {
            setApiKey(storedApiKey);
            setIsAuthenticated(true);
        }
    }, []);

    const validateDigitalOceanToken = async (token) => {
        setLoading(true);
        setError('');
        try {
            const response = await fetch('https://api.digitalocean.com/v2/account', {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });
            if (response.ok) {
                setIsAuthenticated(true);
                setApiKey(token);
                localStorage.setItem('apiKey', token);
                setLoading(false);
            } else {
                setIsAuthenticated(false);
                setError('Invalid API key. Please try again.');
                setLoading(false);
            }
        } catch (error) {
            console.error('Error validating token:', error);
            setError('Failed to validate API key.');
            setIsAuthenticated(false);
            setLoading(false);
        }
    };

    const login = async (inputApiKey) => {
        await validateDigitalOceanToken(inputApiKey);
    };

    const logout = () => {
        setIsAuthenticated(false);
        setApiKey('');
        localStorage.removeItem('apiKey');
    };

    return (
        <AuthContext.Provider value={{ isAuthenticated, login, logout, apiKey, loading, error }}>
            {children}
        </AuthContext.Provider>
    );
};
