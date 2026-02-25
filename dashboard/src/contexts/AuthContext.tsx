import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { auth, type User } from '../api/client';

interface AuthState {
    user: User | null;
    loading: boolean;
    login: (email: string, password: string) => Promise<void>;
    register: (email: string, password: string, name: string) => Promise<void>;
    logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        auth.me().then(res => setUser(res.data)).catch(() => setUser(null)).finally(() => setLoading(false));
    }, []);

    const login = async (email: string, password: string) => {
        await auth.login(email, password);
        const res = await auth.me();
        setUser(res.data);
    };

    const register = async (email: string, password: string, name: string) => {
        await auth.register(email, password, name);
        const res = await auth.me();
        setUser(res.data);
    };

    const logout = async () => {
        try {
            await auth.logout();
        } catch (e) {
            console.error('Logout failed', e);
        }
        setUser(null);
    };

    return (
        <AuthContext.Provider value={{ user, loading, login, register, logout }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
    return ctx;
}
