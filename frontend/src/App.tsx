import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from '@/hooks/useAuth';
import { RequireAuth } from '@/components/shared/RequireAuth';
import { LoginPage } from '@/pages/LoginPage';
import { OutletLayout } from '@/layouts/OutletLayout';
import { SSOLayout } from '@/layouts/SSOLayout';
import { Dashboard } from '@/pages/outlet/Dashboard';
import { Analysis } from '@/pages/outlet/Analysis';
import { Profile } from '@/pages/outlet/Profile';
import { DataEntry } from '@/pages/outlet/DataEntry';
import { TerritoryView } from '@/pages/sso/TerritoryView';
import { SetTargets } from '@/pages/sso/SetTargets';
import { UserManagement } from '@/pages/admin/UserManagement';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function RoleRedirect() {
  const { user } = useAuth();

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  switch (user.role) {
    case 'ro_manager':
      return <Navigate to="/outlet/dashboard" replace />;
    case 'territory_manager':
      return <Navigate to="/sso/territory" replace />;
    case 'admin':
      return <Navigate to="/admin" replace />;
    default:
      return <Navigate to="/login" replace />;
  }
}

function AppRoutes() {
  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<LoginPage />} />

      {/* Role-based redirect from root */}
      <Route path="/" element={<RoleRedirect />} />

      {/* Outlet routes (ro_manager) */}
      <Route
        path="/outlet"
        element={
          <RequireAuth allowedRoles={['ro_manager', 'territory_manager', 'admin']}>
            <OutletLayout />
          </RequireAuth>
        }
      >
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="analysis" element={<Analysis />} />
        <Route path="profile" element={<Profile />} />
        <Route path="data-entry" element={<DataEntry />} />
      </Route>

      {/* SSO routes (territory_manager) */}
      <Route
        path="/sso"
        element={
          <RequireAuth allowedRoles={['territory_manager', 'admin']}>
            <SSOLayout />
          </RequireAuth>
        }
      >
        <Route path="territory" element={<TerritoryView />} />
        <Route path="targets" element={<SetTargets />} />
      </Route>

      {/* Admin routes (admin only) */}
      <Route
        path="/admin"
        element={
          <RequireAuth allowedRoles={['admin']}>
            <UserManagement />
          </RequireAuth>
        }
      />

      {/* Unauthorized page */}
      <Route
        path="/unauthorized"
        element={
          <div className="flex h-screen items-center justify-center">
            <div className="text-center">
              <h1 className="text-4xl font-bold text-red-600">403</h1>
              <p className="mt-2 text-gray-600">You do not have permission to access this page.</p>
            </div>
          </div>
        }
      />

      {/* Catch all */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <AppRoutes />
        </AuthProvider>
      </QueryClientProvider>
    </BrowserRouter>
  );
}
