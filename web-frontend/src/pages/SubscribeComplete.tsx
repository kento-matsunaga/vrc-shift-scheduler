import { Link } from 'react-router-dom';
import { useDocumentTitle } from '../hooks/useDocumentTitle';

export default function SubscribeComplete() {
  useDocumentTitle('登録完了');

  return (
    <div
      className="min-h-screen text-white flex items-center justify-center"
      style={{
        background: 'linear-gradient(180deg, #0a0a0f 0%, #0f0f1a 50%, #0a0a0f 100%)',
        fontFamily: '"Noto Sans JP", "Inter", system-ui, sans-serif',
      }}
    >
      <div className="max-w-md mx-auto px-4 text-center">
        {/* Success Icon */}
        <div
          className="w-20 h-20 rounded-full mx-auto mb-6 flex items-center justify-center"
          style={{
            background: 'linear-gradient(135deg, #10B981 0%, #059669 100%)',
            boxShadow: '0 8px 40px rgba(16, 185, 129, 0.3)',
          }}
        >
          <svg className="w-10 h-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
          </svg>
        </div>

        <h1 className="text-2xl sm:text-3xl font-bold mb-4">
          登録が完了しました！
        </h1>

        <p className="text-gray-400 mb-8 leading-relaxed">
          ご登録ありがとうございます。<br />
          アカウントが有効化されました。<br />
          ログインしてシフト管理を始めましょう。
        </p>

        <Link
          to="/admin/login"
          className="inline-flex items-center gap-2 px-8 py-4 rounded-xl font-semibold text-white transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]"
          style={{
            background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
            boxShadow: '0 8px 40px rgba(79, 70, 229, 0.3)',
          }}
        >
          ログイン画面へ
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
          </svg>
        </Link>

        <div className="mt-8">
          <Link
            to="/"
            className="text-gray-500 hover:text-gray-400 transition-colors text-sm"
          >
            トップページに戻る
          </Link>
        </div>
      </div>
    </div>
  );
}
