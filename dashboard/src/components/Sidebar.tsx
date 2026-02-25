import { NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useTenant } from '../contexts/TenantContext';
import './Sidebar.css';

const NAV_ITEMS = [
    { path: '/', label: 'Dashboard', icon: '⬡' },
    { path: '/members', label: 'Members', icon: '◉' },
    { path: '/apps', label: 'Apps', icon: '⬢' },
    { path: '/billing', label: 'Billing', icon: '◈' },
    { path: '/gdpr', label: 'GDPR', icon: '◎' },
];

export default function Sidebar() {
    const { user, logout } = useAuth();
    const { activeTenant, myTenants, setActiveTenant } = useTenant();
    const navigate = useNavigate();

    const handleLogout = () => { logout(); navigate('/login'); };

    return (
        <aside className="sidebar">
            <div className="sidebar-header">
                <div className="sidebar-logo">◆ Converda</div>
                <select
                    className="tenant-picker"
                    value={activeTenant?.id || ''}
                    onChange={e => {
                        const t = myTenants.find(t => t.id === e.target.value);
                        if (t) setActiveTenant(t);
                    }}
                >
                    {myTenants.map(t => (
                        <option key={t.id} value={t.id}>{t.name}</option>
                    ))}
                </select>
            </div>

            <nav className="sidebar-nav">
                {NAV_ITEMS.map(item => (
                    <NavLink
                        key={item.path}
                        to={item.path}
                        className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
                    >
                        <span className="nav-icon">{item.icon}</span>
                        {item.label}
                    </NavLink>
                ))}
            </nav>

            <div className="sidebar-footer">
                <div className="user-info">
                    <div className="user-avatar">{user?.email?.[0]?.toUpperCase() || '?'}</div>
                    <div className="user-details">
                        <span className="user-name">{user?.full_name || user?.email}</span>
                        <span className="user-email">{user?.email}</span>
                    </div>
                </div>
                <button className="btn btn-outline btn-sm" onClick={handleLogout}>Logout</button>
            </div>
        </aside>
    );
}
