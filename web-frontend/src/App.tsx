import { Routes, Route, Navigate } from 'react-router-dom';
import AdminLogin from './pages/AdminLogin';
import AdminInvitation from './pages/AdminInvitation';
import AcceptInvitation from './pages/AcceptInvitation';
import EventList from './pages/EventList';
import BusinessDayList from './pages/BusinessDayList';
import ShiftSlotList from './pages/ShiftSlotList';
import AssignShift from './pages/AssignShift';
import Members from './pages/Members';
import RoleList from './pages/RoleList';
import AttendanceList from './pages/AttendanceList';
import AttendanceDetail from './pages/AttendanceDetail';
import ScheduleList from './pages/ScheduleList';
import ScheduleDetail from './pages/ScheduleDetail';
import TemplateList from './pages/TemplateList';
import TemplateForm from './pages/TemplateForm';
import TemplateDetail from './pages/TemplateDetail';
import Settings from './pages/Settings';
import Layout from './components/Layout';
import AttendanceResponse from './pages/public/AttendanceResponse';
import ScheduleResponse from './pages/public/ScheduleResponse';

/**
 * ログイン状態をチェック
 * JWT トークンが存在するかどうかで判定
 */
function isAuthenticated(): boolean {
  const authToken = localStorage.getItem('auth_token');
  if (!authToken) {
    return false;
  }

  // JWT の有効期限をチェック（簡易版: ペイロードのexpを確認）
  try {
    const payload = JSON.parse(atob(authToken.split('.')[1]));
    const exp = payload.exp * 1000; // UNIX timestamp to milliseconds
    if (Date.now() >= exp) {
      // トークン期限切れ → ログアウト処理
      localStorage.removeItem('auth_token');
      localStorage.removeItem('admin_id');
      localStorage.removeItem('tenant_id');
      localStorage.removeItem('admin_role');
      return false;
    }
    return true;
  } catch {
    // パースエラー → 無効なトークン
    localStorage.removeItem('auth_token');
    return false;
  }
}

function App() {
  const isLoggedIn = isAuthenticated();

  return (
    <Routes>
      {/* 管理者ログイン */}
      <Route path="/admin/login" element={<AdminLogin />} />
      <Route path="/login" element={<Navigate to="/admin/login" replace />} />

      {/* 招待受理（認証不要） */}
      <Route path="/invite/:token" element={<AcceptInvitation />} />

      {/* 公開ページ（認証不要） */}
      <Route path="/p/attendance/:token" element={<AttendanceResponse />} />
      <Route path="/p/schedule/:token" element={<ScheduleResponse />} />

      {/* ログイン必須の画面 */}
      <Route path="/" element={isLoggedIn ? <Layout /> : <Navigate to="/admin/login" replace />}>
        <Route index element={<Navigate to="/events" replace />} />
        <Route path="events" element={<EventList />} />
        <Route path="events/:eventId/business-days" element={<BusinessDayList />} />
        <Route path="events/:eventId/templates" element={<TemplateList />} />
        <Route path="events/:eventId/templates/new" element={<TemplateForm />} />
        <Route path="events/:eventId/templates/:templateId" element={<TemplateDetail />} />
        <Route path="events/:eventId/templates/:templateId/edit" element={<TemplateForm />} />
        <Route path="business-days/:businessDayId/shift-slots" element={<ShiftSlotList />} />
        <Route path="shift-slots/:slotId/assign" element={<AssignShift />} />
        <Route path="members" element={<Members />} />
        <Route path="roles" element={<RoleList />} />
        <Route path="attendance" element={<AttendanceList />} />
        <Route path="attendance/:collectionId" element={<AttendanceDetail />} />
        <Route path="schedules" element={<ScheduleList />} />
        <Route path="schedules/:scheduleId" element={<ScheduleDetail />} />
        <Route path="admin/invite" element={<AdminInvitation />} />
        <Route path="settings" element={<Settings />} />
      </Route>

      {/* 404 */}
      <Route path="*" element={<div className="p-8 text-center">404 Not Found</div>} />
    </Routes>
  );
}

export default App;
