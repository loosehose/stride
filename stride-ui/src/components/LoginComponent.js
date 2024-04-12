import React, { useState } from 'react';
import { useAuth } from './AuthContext';
import './LoginComponent.css';

const LoginComponent = () => {
    const { login, loginStatus } = useAuth();
    const [apiKey, setApiKey] = useState('');

    const handleLogin = () => {
        login(apiKey);
    };

    return (
        <div>
            <input
                type="text"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Enter DigitalOcean API Key"
            />
            <button
                onClick={handleLogin}
                style={{ backgroundColor: loginStatus.isSuccess ? 'green' : undefined }}
                className={loginStatus.isSuccess === false ? 'shake' : ''}
            >
                {loginStatus.isLoading ? 'Loading...' : loginStatus.isSuccess ? 'Success' : 'Login'}
            </button>
        </div>
    );
};

export default LoginComponent;
