import { Routes, Route, Navigate } from 'react-router-dom';
import Login from './pages/Login';
import EventList from './pages/EventList';
import BusinessDayList from './pages/BusinessDayList';
import ShiftSlotList from './pages/ShiftSlotList';
import AssignShift from './pages/AssignShift';
import MyShifts from './pages/MyShifts';
import Layout from './components/Layout';

function App() {
  // ログインチェック：member_id が localStorage にあるかどうか
  const isLoggedIn = !!localStorage.getItem('member_id');

  return (
    <Routes>
      {/* ログイン画面 */}
      <Route path="/login" element={<Login />} />

      {/* ログイン必須の画面 */}
      <Route path="/" element={isLoggedIn ? <Layout /> : <Navigate to="/login" replace />}>
        <Route index element={<Navigate to="/events" replace />} />
        <Route path="events" element={<EventList />} />
        <Route path="events/:eventId/business-days" element={<BusinessDayList />} />
        <Route path="business-days/:businessDayId/shift-slots" element={<ShiftSlotList />} />
        <Route path="shift-slots/:slotId/assign" element={<AssignShift />} />
        <Route path="my-shifts" element={<MyShifts />} />
      </Route>

      {/* 404 */}
      <Route path="*" element={<div className="p-8 text-center">404 Not Found</div>} />
    </Routes>
  );
}

export default App;
