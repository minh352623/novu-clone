import { useEffect, useState } from 'react';
import { useTenant } from '../contexts/TenantContext';
import { plans, type PricingPlan } from '../api/client';

export default function BillingPage() {
    const { activeTenant } = useTenant();
    const [allPlans, setAllPlans] = useState<PricingPlan[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        plans.list().then(r => setAllPlans(Array.isArray(r.data) ? r.data : [])).catch(() => { }).finally(() => setLoading(false));
    }, []);

    const currentPlan = activeTenant?.pricing_plan;

    return (
        <div style={{ animation: 'fadeIn var(--t-normal)' }}>
            <div className="page-header">
                <h1>Billing & Plan</h1>
            </div>

            <div className="card" style={{ marginBottom: 'var(--s-6)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <div>
                        <h2 style={{ fontSize: 'var(--text-xl)', fontWeight: 600 }}>Current Plan</h2>
                        <p style={{ color: 'var(--c-text-secondary)', fontSize: 'var(--text-sm)', marginTop: 'var(--s-1)' }}>
                            {currentPlan?.name || 'Free'} · {currentPlan ? `$${currentPlan.price}` : '$0.00'}
                        </p>
                    </div>
                    <span className="badge badge-success">{activeTenant?.status || 'active'}</span>
                </div>
            </div>

            <h3 style={{ fontSize: 'var(--text-lg)', fontWeight: 600, marginBottom: 'var(--s-4)' }}>Plan Comparison</h3>

            {loading ? (
                <div className="card" style={{ padding: 'var(--s-8)', textAlign: 'center', color: 'var(--c-text-muted)' }}>Loading plans...</div>
            ) : (
                <div className="card" style={{ overflow: 'auto' }}>
                    <table>
                        <thead>
                            <tr>
                                <th>Feature</th>
                                {allPlans.map(p => (
                                    <th key={p.id} style={{ textAlign: 'center' }}>
                                        {p.name}
                                        {currentPlan?.id === p.id && <span style={{ display: 'block', fontSize: '10px', color: 'var(--c-success)' }}>CURRENT</span>}
                                    </th>
                                ))}
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td>Price</td>
                                {allPlans.map(p => <td key={p.id} style={{ textAlign: 'center', fontWeight: 600 }}>${p.price}</td>)}
                            </tr>
                            <tr>
                                <td>Max Apps</td>
                                {allPlans.map(p => <td key={p.id} style={{ textAlign: 'center' }}>{p.max_apps || '∞'}</td>)}
                            </tr>
                            <tr>
                                <td>Max Members</td>
                                {allPlans.map(p => <td key={p.id} style={{ textAlign: 'center' }}>{p.max_members || '∞'}</td>)}
                            </tr>
                            <tr>
                                <td>Max Workflows</td>
                                {allPlans.map(p => <td key={p.id} style={{ textAlign: 'center' }}>{p.max_workflows || '∞'}</td>)}
                            </tr>
                            <tr>
                                <td>Messages/Month</td>
                                {allPlans.map(p => <td key={p.id} style={{ textAlign: 'center' }}>{p.max_messages_per_month ? p.max_messages_per_month.toLocaleString() : '∞'}</td>)}
                            </tr>
                            <tr>
                                <td>Rate Limit (RPM)</td>
                                {allPlans.map(p => <td key={p.id} style={{ textAlign: 'center' }}>{p.rate_limit_rpm || '∞'}</td>)}
                            </tr>
                            <tr>
                                <td>Monthly Credits</td>
                                {allPlans.map(p => <td key={p.id} style={{ textAlign: 'center' }}>{p.monthly_credits.toLocaleString()}</td>)}
                            </tr>
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
}
