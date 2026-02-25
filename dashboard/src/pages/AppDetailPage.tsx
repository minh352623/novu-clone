import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apps, type App as AppType, type Environment, type ApiKey, type Webhook } from '../api/client';

export default function AppDetailPage() {
    const { appId } = useParams();
    const navigate = useNavigate();
    const [app, setApp] = useState<AppType | null>(null);
    const [envs, setEnvs] = useState<Environment[]>([]);
    const [keys, setKeys] = useState<Record<string, ApiKey[]>>({});
    const [webhooks, setWebhooks] = useState<Webhook[]>([]);
    const [activeTab, setActiveTab] = useState('environments');

    useEffect(() => {
        if (!appId) return;
        apps.get(appId).then(r => setApp(r.data)).catch(() => navigate('/apps'));
        apps.environments(appId).then(r => {
            const envList = Array.isArray(r.data) ? r.data : [];
            setEnvs(envList);
            envList.forEach(env => {
                apps.apiKeys(env.id).then(kr => {
                    setKeys(prev => ({ ...prev, [env.id]: Array.isArray(kr.data) ? kr.data : [] }));
                }).catch(() => { });
            });
        }).catch(() => { });
        apps.webhooks(appId).then(r => setWebhooks(Array.isArray(r.data) ? r.data : [])).catch(() => { });
    }, [appId]);

    const tabs = [
        { key: 'environments', label: 'Environments' },
        { key: 'webhooks', label: 'Webhooks' },
    ];

    return (
        <div style={{ animation: 'fadeIn var(--t-normal)' }}>
            <div className="page-header">
                <div>
                    <button className="btn btn-outline btn-sm" onClick={() => navigate('/apps')} style={{ marginBottom: 'var(--s-3)' }}>
                        ← Back
                    </button>
                    <h1>{app?.name || 'Loading...'}</h1>
                </div>
                <span className={`badge ${app?.status === 'active' ? 'badge-success' : 'badge-neutral'}`}>{app?.status}</span>
            </div>

            <div style={{ display: 'flex', gap: 'var(--s-2)', marginBottom: 'var(--s-6)', borderBottom: '1px solid var(--c-border)', paddingBottom: 'var(--s-1)' }}>
                {tabs.map(t => (
                    <button
                        key={t.key}
                        className="btn"
                        style={{
                            borderBottom: activeTab === t.key ? '2px solid var(--c-accent)' : '2px solid transparent',
                            borderRadius: 0,
                            fontWeight: activeTab === t.key ? 600 : 400,
                            color: activeTab === t.key ? 'var(--c-text)' : 'var(--c-text-secondary)',
                        }}
                        onClick={() => setActiveTab(t.key)}
                    >
                        {t.label}
                    </button>
                ))}
            </div>

            {activeTab === 'environments' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--s-4)' }}>
                    {envs.map(env => (
                        <div key={env.id} className="card">
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--s-4)' }}>
                                <div>
                                    <h3 style={{ fontWeight: 600 }}>{env.name}</h3>
                                    <span style={{ fontSize: 'var(--text-xs)', color: 'var(--c-text-muted)' }}>
                                        RPM: {env.rate_limit_rpm} · Daily: {env.rate_limit_daily}
                                    </span>
                                </div>
                            </div>
                            <h4 style={{ fontSize: 'var(--text-sm)', fontWeight: 600, color: 'var(--c-text-secondary)', marginBottom: 'var(--s-3)' }}>API Keys</h4>
                            {(keys[env.id] || []).length === 0 ? (
                                <p style={{ fontSize: 'var(--text-sm)', color: 'var(--c-text-muted)' }}>No API keys</p>
                            ) : (
                                <table>
                                    <thead>
                                        <tr><th>Prefix</th><th>Status</th><th>Created</th></tr>
                                    </thead>
                                    <tbody>
                                        {(keys[env.id] || []).map(k => (
                                            <tr key={k.id}>
                                                <td><code style={{ background: 'var(--c-bg)', padding: '2px 6px', borderRadius: '4px' }}>{k.key_prefix}•••</code></td>
                                                <td><span className={`badge ${k.is_active ? 'badge-success' : 'badge-danger'}`}>{k.is_active ? 'active' : 'revoked'}</span></td>
                                                <td style={{ fontSize: 'var(--text-xs)', color: 'var(--c-text-muted)' }}>{new Date(k.created_at).toLocaleDateString()}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            )}
                        </div>
                    ))}
                    {envs.length === 0 && <div className="card" style={{ textAlign: 'center', padding: 'var(--s-8)', color: 'var(--c-text-muted)' }}>No environments</div>}
                </div>
            )}

            {activeTab === 'webhooks' && (
                <div className="card">
                    {webhooks.length === 0 ? (
                        <p style={{ textAlign: 'center', padding: 'var(--s-8)', color: 'var(--c-text-muted)' }}>No webhooks configured</p>
                    ) : (
                        <table>
                            <thead><tr><th>URL</th><th>Events</th><th>Status</th></tr></thead>
                            <tbody>
                                {webhooks.map(w => (
                                    <tr key={w.id}>
                                        <td style={{ fontFamily: 'monospace', fontSize: 'var(--text-xs)' }}>{w.url}</td>
                                        <td>{w.events?.map(e => <span key={e} className="badge badge-neutral" style={{ marginRight: 4 }}>{e}</span>)}</td>
                                        <td><span className={`badge ${w.is_active ? 'badge-success' : 'badge-danger'}`}>{w.is_active ? 'active' : 'inactive'}</span></td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>
            )}
        </div>
    );
}
