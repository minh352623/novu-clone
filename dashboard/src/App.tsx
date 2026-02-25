import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { TenantProvider, useTenant } from './contexts/TenantContext';
import Sidebar from './components/Sidebar';
import LoginPage from './pages/LoginPage';
import OnboardingPage from './pages/OnboardingPage';
import DashboardPage from './pages/DashboardPage';
import MembersPage from './pages/MembersPage';
import AppsPage from './pages/AppsPage';
import AppDetailPage from './pages/AppDetailPage';
import BillingPage from './pages/BillingPage';
import GDPRPage from './pages/GDPRPage';

function AppShell() {
  const { loading, myTenants } = useTenant();

  if (loading) {
    return <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', color: 'var(--c-text-muted)' }}>Loading...</div>;
  }

  // No tenants → onboarding
  if (myTenants.length === 0) {
    return <OnboardingPage />;
  }

  return (
    <div style={{ display: 'flex', minHeight: '100vh' }}>
      <Sidebar />
      <main style={{ marginLeft: 'var(--sidebar-w)', flex: 1, padding: 'var(--s-8)', maxWidth: 960 }}>
        <Outlet />
      </main>
    </div>
  );
}

function ProtectedLayout() {
  const { user, loading } = useAuth();
  if (loading) return <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', color: 'var(--c-text-muted)' }}>Loading...</div>;
  if (!user) return <Navigate to="/login" replace />;
  return (
    <TenantProvider>
      <AppShell />
    </TenantProvider>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route element={<ProtectedLayout />}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/members" element={<MembersPage />} />
            <Route path="/apps" element={<AppsPage />} />
            <Route path="/apps/:appId" element={<AppDetailPage />} />
            <Route path="/billing" element={<BillingPage />} />
            <Route path="/gdpr" element={<GDPRPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
