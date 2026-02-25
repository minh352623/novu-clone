import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTenant } from '../contexts/TenantContext';
import { apps, type App } from '../api/client';

export default function AppsPage() {
    const { activeTenant } = useTenant();
    const [appList, setAppList] = useState<App[]>([]);
    const [showCreate, setShowCreate] = useState(false);
    const [newName, setNewName] = useState('');
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    const fetchApps = () => {
        if (!activeTenant) return;
        apps.list(activeTenant.id)
            .then(r => setAppList(Array.isArray(r.data) ? r.data : []))
            .catch(() => { })
            .finally(() => setLoading(false));
    };

    useEffect(fetchApps, [activeTenant]);

    const handleCreate = async () => {
        if (!activeTenant || !newName.trim()) return;
        try {
            await apps.create(activeTenant.id, newName.trim());
            setShowCreate(false);
            setNewName('');
            fetchApps();
        } catch (e: unknown) {
            alert(e instanceof Error ? e.message : 'Failed');
        }
    };

    return (
        <div style={{ animation: 'fadeIn var(--t-normal)' }}>
            <div className="page-header">
                <h1>Apps</h1>
                <button className="btn btn-primary" onClick={() => setShowCreate(true)}>+ Create App</button>
            </div>

            {loading ? (
                <div className="card" style={{ padding: 'var(--s-8)', textAlign: 'center', color: 'var(--c-text-muted)' }}>Loading...</div>
            ) : (
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 'var(--s-4)' }}>
                    {appList.map(app => (
                        <div
                            key={app.id}
                            className="card"
                            style={{ cursor: 'pointer', transition: 'all var(--t-fast)' }}
                            onClick={() => navigate(`/apps/${app.id}`)}
                            onMouseEnter={e => (e.currentTarget.style.borderColor = 'var(--c-accent)')}
                            onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--c-border)')}
                        >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--s-3)' }}>
                                <h3 style={{ fontSize: 'var(--text-lg)', fontWeight: 600 }}>{app.name}</h3>
                                <span className={`badge ${app.status === 'active' ? 'badge-success' : 'badge-neutral'}`}>{app.status}</span>
                            </div>
                            <p style={{ fontSize: 'var(--text-xs)', color: 'var(--c-text-muted)' }}>
                                Created {new Date(app.created_at).toLocaleDateString()}
                            </p>
                        </div>
                    ))}
                    {appList.length === 0 && (
                        <div className="card" style={{ gridColumn: '1 / -1', textAlign: 'center', padding: 'var(--s-12)', color: 'var(--c-text-muted)' }}>
                            <p style={{ fontSize: 'var(--text-lg)', marginBottom: 'var(--s-3)' }}>No apps yet</p>
                            <button className="btn btn-primary" onClick={() => setShowCreate(true)}>Create your first app</button>
                        </div>
                    )}
                </div>
            )}

            {showCreate && (
                <div className="modal-overlay" onClick={() => setShowCreate(false)}>
                    <div className="modal-content" onClick={e => e.stopPropagation()}>
                        <h2>Create App</h2>
                        <div className="form-group">
                            <label>App Name</label>
                            <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="My Chatbot" />
                        </div>
                        <div style={{ display: 'flex', gap: 'var(--s-3)', justifyContent: 'flex-end' }}>
                            <button className="btn btn-outline" onClick={() => setShowCreate(false)}>Cancel</button>
                            <button className="btn btn-primary" onClick={handleCreate}>Create</button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
