import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import LicenseKeys from './pages/LicenseKeys';
import Tenants from './pages/Tenants';
import TenantDetail from './pages/TenantDetail';
import AuditLogs from './pages/AuditLogs';
import Announcements from './pages/Announcements';
import Tutorials from './pages/Tutorials';
import SystemSettings from './pages/SystemSettings';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<LicenseKeys />} />
        <Route path="tenants" element={<Tenants />} />
        <Route path="tenants/:tenantId" element={<TenantDetail />} />
        <Route path="audit-logs" element={<AuditLogs />} />
        <Route path="announcements" element={<Announcements />} />
        <Route path="tutorials" element={<Tutorials />} />
        <Route path="system" element={<SystemSettings />} />
      </Route>
    </Routes>
  );
}

export default App;
