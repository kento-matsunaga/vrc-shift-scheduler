import { Link } from 'react-router-dom';
import { useDocumentTitle } from '../hooks/useDocumentTitle';
import { SEO } from '../components/seo';

export default function SubscribeCancel() {
  useDocumentTitle('決済キャンセル');

  return (
    <>
      <SEO noindex={true} />
      <div
        className="min-h-[100dvh] text-white flex items-center justify-center pt-[env(safe-area-inset-top)] pb-[env(safe-area-inset-bottom)]"
      style={{
        background: 'linear-gradient(180deg, #0a0a0f 0%, #0f0f1a 50%, #0a0a0f 100%)',
        fontFamily: '"Noto Sans JP", "Inter", system-ui, sans-serif',
      }}
    >
      <div className="max-w-md mx-auto px-4 text-center">
        {/* Cancel Icon */}
        <div
          className="w-20 h-20 rounded-full mx-auto mb-6 flex items-center justify-center"
          style={{
            background: 'rgba(255, 255, 255, 0.05)',
            border: '2px solid rgba(255, 255, 255, 0.1)',
          }}
        >
          <svg className="w-10 h-10 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>

        <h1 className="text-2xl sm:text-3xl font-bold mb-4">
          決済がキャンセルされました
        </h1>

        <p className="text-gray-400 mb-8 leading-relaxed">
          決済が完了しませんでした。<br />
          もう一度お試しいただくか、<br />
          問題が続く場合はお問い合わせください。
        </p>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link
            to="/subscribe"
            className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-xl font-semibold text-white transition-all duration-300 hover:scale-[1.02] active:scale-[0.98]"
            style={{
              background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
              boxShadow: '0 8px 40px rgba(79, 70, 229, 0.3)',
            }}
          >
            もう一度試す
          </Link>

          <Link
            to="/"
            className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-xl font-semibold text-white transition-all duration-300 hover:bg-white/10"
            style={{
              background: 'rgba(255, 255, 255, 0.05)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
            }}
          >
            トップに戻る
          </Link>
        </div>

        <div className="mt-8 text-sm text-gray-500">
          お問い合わせ:{' '}
          <a
            href="https://x.com/Noa_Fortevita"
            target="_blank"
            rel="noopener noreferrer"
            className="text-violet-400 hover:text-violet-300 transition-colors"
          >
            @Noa_Fortevita
          </a>
        </div>
      </div>
      </div>
    </>
  );
}
