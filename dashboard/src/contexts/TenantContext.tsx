import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { tenants, type Tenant } from '../api/client';
import { useAuth } from './AuthContext';

interface TenantState {
    activeTenant: Tenant | null;
    myTenants: Tenant[];
    setActiveTenant: (t: Tenant) => void;
    loading: boolean;
    refetch: () => Promise<void>;
}

const TenantContext = createContext<TenantState | null>(null);

export function TenantProvider({ children }: { children: ReactNode }) {
    const { user } = useAuth();
    const [myTenants, setMyTenants] = useState<Tenant[]>([]);
    const [activeTenant, setActiveTenant] = useState<Tenant | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchTenants = useCallback(async () => {
        if (!user) { setLoading(false); return; }
        setLoading(true);
        try {
            const res = await tenants.myTenants();
            const list = Array.isArray(res.data) ? res.data : [];
            setMyTenants(list);
            if (list.length > 0) setActiveTenant(list[0]);
        } catch {
            // ignore
        } finally {
            setLoading(false);
        }
    }, [user]);

    useEffect(() => { fetchTenants(); }, [fetchTenants]);

    return (
        <TenantContext.Provider value={{ activeTenant, myTenants, setActiveTenant, loading, refetch: fetchTenants }}>
            {children}
        </TenantContext.Provider>
    );
}

export function useTenant() {
    const ctx = useContext(TenantContext);
    if (!ctx) throw new Error('useTenant must be used inside TenantProvider');
    return ctx;
}
