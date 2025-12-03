import { Outlet, Link, useNavigate } from 'react-router-dom';

export default function Layout() {
  const navigate = useNavigate();
  const memberName = localStorage.getItem('member_name') || '未設定';

  const handleLogout = () => {
    localStorage.removeItem('member_id');
    localStorage.removeItem('member_name');
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center">
            <div className="flex items-center space-x-8">
              <h1 className="text-2xl font-bold text-gray-900">VRC Shift Scheduler</h1>
              <nav className="hidden md:flex space-x-4">
                <Link to="/events" className="text-gray-600 hover:text-gray-900">
                  イベント
                </Link>
                <Link to="/my-shifts" className="text-gray-600 hover:text-gray-900">
                  自分のシフト
                </Link>
              </nav>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">{memberName}</span>
              <button onClick={handleLogout} className="btn-secondary text-sm">
                ログアウト
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* メインコンテンツ */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>

      {/* フッター */}
      <footer className="bg-white border-t mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <p className="text-center text-sm text-gray-500">
            ⚠️ これは α 版のテストです。データは予告なく消える可能性があります。
          </p>
        </div>
      </footer>
    </div>
  );
}

