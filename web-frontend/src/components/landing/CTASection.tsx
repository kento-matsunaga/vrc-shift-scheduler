import { Link } from 'react-router-dom';
import { AnimatedSection } from './AnimatedSection';
import { useReleaseStatus } from '../../hooks/useReleaseStatus';

export function CTASection() {
  const { released, isLoading } = useReleaseStatus();
  return (
    <section className="relative py-16 sm:py-24 px-4 sm:px-6">
      <div className="max-w-3xl mx-auto text-center">
        <AnimatedSection>
          <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold mb-6 text-white">
            今すぐVRC Shift Schedulerを始めましょう
          </h2>
          <p className="text-gray-400 mb-8 text-sm sm:text-base">いつでもキャンセル可能。まずは試してみてください。</p>
          {released ? (
            <Link
              to="/subscribe"
              className="inline-flex items-center justify-center gap-2 px-6 sm:px-8 py-4 rounded-full font-semibold text-base sm:text-lg transition-all duration-300 hover:scale-105 active:scale-95 text-white min-h-[48px]"
              style={{
                background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
                boxShadow: '0 8px 40px rgba(79, 70, 229, 0.5)',
              }}
            >
              今すぐ始める
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 7l5 5m0 0l-5 5m5-5H6"
                />
              </svg>
            </Link>
          ) : (
            <div
              className="inline-flex items-center justify-center gap-2 px-6 sm:px-8 py-4 rounded-full font-semibold text-base sm:text-lg text-white min-h-[48px] cursor-not-allowed opacity-75"
              style={{
                background: 'linear-gradient(135deg, #6b7280 0%, #9ca3af 100%)',
              }}
            >
              {isLoading ? '読み込み中...' : 'リリース前です'}
            </div>
          )}
        </AnimatedSection>
      </div>
    </section>
  );
}
