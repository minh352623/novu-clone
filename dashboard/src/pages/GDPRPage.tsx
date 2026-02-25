import { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useTenant } from '../contexts/TenantContext';
import { gdpr } from '../api/client';

const CATEGORIES = [
    { key: 'profile', label: 'Profile Data', desc: 'User information, memberships' },
    { key: 'messages', label: 'Messages', desc: 'Chat messages and threads' },
    { key: 'subscriptions', label: 'Subscriptions', desc: 'Newsletter and subscriber data' },
    { key: 'audit_logs', label: 'Audit Logs', desc: 'Activity history (export only)' },
];

export default function GDPRPage() {
    const { user } = useAuth();
    const { activeTenant } = useTenant();
    const [selected, setSelected] = useState<string[]>([]);
    const [exportResult, setExportResult] = useState<string | null>(null);
    const [eraseResult, setEraseResult] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [showEraseConfirm, setShowEraseConfirm] = useState(false);
    const [confirmText, setConfirmText] = useState('');

    const toggle = (cat: string) => {
        setSelected(prev => prev.includes(cat) ? prev.filter(c => c !== cat) : [...prev, cat]);
    };

    const handleExport = async () => {
        if (!activeTenant || !user) return;
        setLoading(true);
        try {
            const res = await gdpr.export(activeTenant.id, user.id, selected.length > 0 ? selected : undefined);
            setExportResult(JSON.stringify(res.data, null, 2));
        } catch (e: unknown) {
            alert(e instanceof Error ? e.message : 'Export failed');
        } finally {
            setLoading(false);
        }
    };

    const handleErase = async () => {
        if (!activeTenant || !user || confirmText !== 'DELETE') return;
        setLoading(true);
        try {
            const res = await gdpr.erase(activeTenant.id, user.id, selected.length > 0 ? selected : undefined);
            setEraseResult(JSON.stringify(res.data, null, 2));
            setShowEraseConfirm(false);
            setConfirmText('');
        } catch (e: unknown) {
            alert(e instanceof Error ? e.message : 'Erase failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{ animation: 'fadeIn var(--t-normal)' }}>
            <div className="page-header">
                <h1>GDPR Compliance</h1>
            </div>

            <div className="card" style={{ marginBottom: 'var(--s-6)' }}>
                <h3 style={{ fontSize: 'var(--text-lg)', fontWeight: 600, marginBottom: 'var(--s-4)' }}>
                    Select Data Categories
                </h3>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 'var(--s-3)' }}>
                    {CATEGORIES.map(cat => (
                        <label
                            key={cat.key}
                            style={{
                                display: 'flex', gap: 'var(--s-3)', padding: 'var(--s-4)', border: '1px solid',
                                borderColor: selected.includes(cat.key) ? 'var(--c-accent)' : 'var(--c-border)',
                                borderRadius: 'var(--r-md)', cursor: 'pointer', transition: 'all var(--t-fast)',
                                background: selected.includes(cat.key) ? 'var(--c-surface-hover)' : 'transparent',
                            }}
                        >
                            <input type="checkbox" checked={selected.includes(cat.key)} onChange={() => toggle(cat.key)} />
                            <div>
                                <div style={{ fontWeight: 500, fontSize: 'var(--text-sm)' }}>{cat.label}</div>
                                <div style={{ fontSize: 'var(--text-xs)', color: 'var(--c-text-muted)' }}>{cat.desc}</div>
                            </div>
                        </label>
                    ))}
                </div>
                <p style={{ fontSize: 'var(--text-xs)', color: 'var(--c-text-muted)', marginTop: 'var(--s-3)' }}>
                    Leave all unchecked to include all categories.
                </p>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--s-4)', marginBottom: 'var(--s-6)' }}>
                <div className="card">
                    <h3 style={{ fontWeight: 600, marginBottom: 'var(--s-3)' }}>📥 Export Data</h3>
                    <p style={{ fontSize: 'var(--text-sm)', color: 'var(--c-text-secondary)', marginBottom: 'var(--s-4)' }}>
                        Download all your personal data as JSON.
                    </p>
                    <button className="btn btn-primary" onClick={handleExport} disabled={loading}>
                        {loading ? 'Exporting...' : 'Export My Data'}
                    </button>
                </div>

                <div className="card" style={{ borderColor: '#fecaca' }}>
                    <h3 style={{ fontWeight: 600, marginBottom: 'var(--s-3)', color: 'var(--c-danger)' }}>🗑️ Erase Data</h3>
                    <p style={{ fontSize: 'var(--text-sm)', color: 'var(--c-text-secondary)', marginBottom: 'var(--s-4)' }}>
                        Permanently anonymize / delete your data. This action cannot be undone.
                    </p>
                    <button className="btn btn-danger" onClick={() => setShowEraseConfirm(true)} disabled={loading}>
                        Request Erasure
                    </button>
                </div>
            </div>

            {exportResult && (
                <div className="card" style={{ marginBottom: 'var(--s-6)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--s-3)' }}>
                        <h3 style={{ fontWeight: 600 }}>Export Result</h3>
                        <button className="btn btn-outline btn-sm" onClick={() => {
                            const blob = new Blob([exportResult], { type: 'application/json' });
                            const url = URL.createObjectURL(blob);
                            const a = document.createElement('a');
                            a.href = url; a.download = 'gdpr-export.json'; a.click();
                        }}>
                            Download JSON
                        </button>
                    </div>
                    <pre style={{ background: 'var(--c-bg)', padding: 'var(--s-4)', borderRadius: 'var(--r-md)', fontSize: 'var(--text-xs)', maxHeight: 300, overflow: 'auto', fontFamily: 'monospace' }}>
                        {exportResult}
                    </pre>
                </div>
            )}

            {eraseResult && (
                <div className="card" style={{ borderColor: '#fecaca' }}>
                    <h3 style={{ fontWeight: 600, color: 'var(--c-danger)', marginBottom: 'var(--s-3)' }}>Erasure Complete</h3>
                    <pre style={{ background: '#fef2f2', padding: 'var(--s-4)', borderRadius: 'var(--r-md)', fontSize: 'var(--text-xs)', fontFamily: 'monospace' }}>
                        {eraseResult}
                    </pre>
                </div>
            )}

            {showEraseConfirm && (
                <div className="modal-overlay" onClick={() => setShowEraseConfirm(false)}>
                    <div className="modal-content" onClick={e => e.stopPropagation()}>
                        <h2 style={{ color: 'var(--c-danger)' }}>⚠️ Confirm Data Erasure</h2>
                        <p style={{ fontSize: 'var(--text-sm)', color: 'var(--c-text-secondary)', margin: 'var(--s-4) 0' }}>
                            This will permanently anonymize your data. Type <strong>DELETE</strong> to confirm.
                        </p>
                        <div className="form-group">
                            <input value={confirmText} onChange={e => setConfirmText(e.target.value)} placeholder="Type DELETE" />
                        </div>
                        <div style={{ display: 'flex', gap: 'var(--s-3)', justifyContent: 'flex-end' }}>
                            <button className="btn btn-outline" onClick={() => setShowEraseConfirm(false)}>Cancel</button>
                            <button className="btn btn-danger" onClick={handleErase} disabled={confirmText !== 'DELETE' || loading}>
                                {loading ? 'Processing...' : 'Erase Permanently'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
