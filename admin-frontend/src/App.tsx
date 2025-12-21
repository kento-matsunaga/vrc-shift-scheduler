import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import LicenseKeys from './pages/LicenseKeys';
import Tenants from './pages/Tenants';
import TenantDetail from './pages/TenantDetail';
import AuditLogs from './pages/AuditLogs';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<LicenseKeys />} />
        <Route path="tenants" element={<Tenants />} />
        <Route path="tenants/:tenantId" element={<TenantDetail />} />
        <Route path="audit-logs" element={<AuditLogs />} />
      </Route>
    </Routes>
  );
}

export default App;
