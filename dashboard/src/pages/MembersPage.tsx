import { useEffect, useState } from 'react';
import { useTenant } from '../contexts/TenantContext';
import { members, roles, type TenantMember, type Role } from '../api/client';

export default function MembersPage() {
    const { activeTenant } = useTenant();
    const [memberList, setMemberList] = useState<TenantMember[]>([]);
    const [roleList, setRoleList] = useState<Role[]>([]);
    const [showInvite, setShowInvite] = useState(false);
    const [inviteEmail, setInviteEmail] = useState('');
    const [inviteRole, setInviteRole] = useState('');
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        if (!activeTenant) return;
        Promise.all([
            members.list(activeTenant.id).then(r => setMemberList(Array.isArray(r.data) ? r.data : [])),
            roles.list().then(r => setRoleList(Array.isArray(r.data) ? r.data : [])),
        ]).finally(() => setLoading(false));
    }, [activeTenant]);

    const handleInvite = async () => {
        if (!activeTenant || !inviteEmail) return;
        try {
            await members.invite(activeTenant.id, inviteEmail, inviteRole || undefined);
            setShowInvite(false);
            setInviteEmail('');
            alert('Invitation sent!');
        } catch (e: unknown) {
            alert(e instanceof Error ? e.message : 'Failed to invite');
        }
    };

    return (
        <div style={{ animation: 'fadeIn var(--t-normal)' }}>
            <div className="page-header">
                <h1>Members</h1>
                <button className="btn btn-primary" onClick={() => setShowInvite(true)}>+ Invite</button>
            </div>

            <div className="card">
                {loading ? (
                    <div style={{ padding: 'var(--s-6)', textAlign: 'center', color: 'var(--c-text-muted)' }}>Loading...</div>
                ) : (
                    <table>
                        <thead>
                            <tr>
                                <th>User</th>
                                <th>Email</th>
                                <th>Role</th>
                                <th>Joined</th>
                            </tr>
                        </thead>
                        <tbody>
                            {memberList.map(m => (
                                <tr key={m.id}>
                                    <td style={{ fontWeight: 500 }}>{m.user?.full_name || '—'}</td>
                                    <td style={{ color: 'var(--c-text-secondary)' }}>{m.user?.email || '—'}</td>
                                    <td><span className="badge badge-neutral">{m.role?.name || 'member'}</span></td>
                                    <td style={{ color: 'var(--c-text-muted)', fontSize: 'var(--text-xs)' }}>
                                        {new Date(m.created_at).toLocaleDateString()}
                                    </td>
                                </tr>
                            ))}
                            {memberList.length === 0 && (
                                <tr><td colSpan={4} style={{ textAlign: 'center', color: 'var(--c-text-muted)', padding: 'var(--s-8)' }}>No members yet</td></tr>
                            )}
                        </tbody>
                    </table>
                )}
            </div>

            {showInvite && (
                <div className="modal-overlay" onClick={() => setShowInvite(false)}>
                    <div className="modal-content" onClick={e => e.stopPropagation()}>
                        <h2>Invite Member</h2>
                        <div className="form-group">
                            <label>Email</label>
                            <input
                                type="email"
                                value={inviteEmail}
                                onChange={e => setInviteEmail(e.target.value)}
                                placeholder="colleague@company.com"
                            />
                        </div>
                        <div className="form-group">
                            <label>Role</label>
                            <select value={inviteRole} onChange={e => setInviteRole(e.target.value)}>
                                <option value="">Default</option>
                                {roleList.map(r => <option key={r.id} value={r.id}>{r.name}</option>)}
                            </select>
                        </div>
                        <div style={{ display: 'flex', gap: 'var(--s-3)', justifyContent: 'flex-end' }}>
                            <button className="btn btn-outline" onClick={() => setShowInvite(false)}>Cancel</button>
                            <button className="btn btn-primary" onClick={handleInvite}>Send Invite</button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
