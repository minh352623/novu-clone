import { useEffect, useState } from 'react';
import { useTenant } from '../contexts/TenantContext';
import { apps, type AppMetrics } from '../api/client';
import './DashboardPage.css';

export default function DashboardPage() {
    const { activeTenant } = useTenant();
    const [metrics, setMetrics] = useState<AppMetrics | null>(null);
    const [appCount, setAppCount] = useState(0);

    useEffect(() => {
        if (!activeTenant) return;
        apps.list(activeTenant.id).then(res => {
            const list = Array.isArray(res.data) ? res.data : [];
            setAppCount(list.length);
            if (list.length > 0) {
                apps.metrics(list[0].id).then(m => setMetrics(m.data)).catch(() => { });
            }
        }).catch(() => { });
    }, [activeTenant]);

    const plan = activeTenant?.pricing_plan;

    return (
        <div className="dashboard-page">
            <div className="page-header">
                <h1>Dashboard</h1>
                <span className={`badge ${activeTenant?.status === 'active' ? 'badge-success' : 'badge-warning'}`}>
                    {activeTenant?.status || 'unknown'}
                </span>
            </div>

            <div className="tenant-info card">
                <div className="tenant-detail">
                    <h2>{activeTenant?.name || '—'}</h2>
                    <span className="tenant-slug">/{activeTenant?.slug}</span>
                </div>
                <div className="tenant-plan">
                    <span className="plan-name">{plan?.name || 'Free'}</span>
                    <span className="plan-price">{plan ? `$${plan.price}/${plan.currency}` : 'Free'}</span>
                </div>
            </div>

            <div className="stats-grid">
                <StatCard label="Total Apps" value={appCount} max={plan?.max_apps || 0} />
                <StatCard label="Messages Today" value={metrics?.messages_today || 0} />
                <StatCard label="Total Threads" value={metrics?.total_threads || 0} />
                <StatCard label="Subscribers" value={metrics?.total_subscribers || 0} />
            </div>

            {plan && plan.max_apps > 0 && (
                <div className="card usage-section">
                    <h3>Plan Usage</h3>
                    <div className="usage-bars">
                        <UsageBar label="Apps" current={appCount} max={plan.max_apps} />
                        <UsageBar label="Members" current={0} max={plan.max_members} />
                        <UsageBar label="Workflows" current={0} max={plan.max_workflows} />
                        <UsageBar label="RPM" current={0} max={plan.rate_limit_rpm} />
                    </div>
                </div>
            )}
        </div>
    );
}

function StatCard({ label, value, max }: { label: string; value: number; max?: number }) {
    return (
        <div className="card stat-card">
            <span className="stat-label">{label}</span>
            <span className="stat-value">{value.toLocaleString()}</span>
            {max !== undefined && max > 0 && (
                <span className="stat-limit">/ {max}</span>
            )}
        </div>
    );
}

function UsageBar({ label, current, max }: { label: string; current: number; max: number }) {
    const pct = max > 0 ? Math.min((current / max) * 100, 100) : 0;
    const cls = pct > 90 ? 'danger' : pct > 70 ? 'warning' : '';
    return (
        <div className="usage-item">
            <div className="usage-header">
                <span>{label}</span>
                <span className="usage-count">{current} / {max}</span>
            </div>
            <div className="progress-bar">
                <div className={`progress-bar-fill ${cls}`} style={{ width: `${pct}%` }} />
            </div>
        </div>
    );
}
