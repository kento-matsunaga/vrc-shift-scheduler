import { Link } from 'react-router-dom';
import { AnimatedSection } from './AnimatedSection';

const pricingFeatures = [
  '全機能利用可能',
  'イベント数無制限',
  'メンバー数無制限',
  'シフト枠数無制限',
  '出欠収集機能',
  '日程調整機能',
];

export function PricingSection() {
  return (
    <section id="pricing" className="relative py-16 sm:py-24 px-4 sm:px-6">
      <div className="max-w-4xl mx-auto">
        <AnimatedSection>
          <div className="text-center mb-10 sm:mb-16">
            <span className="text-violet-400 text-xs sm:text-sm font-medium tracking-wider uppercase mb-4 block">
              Pricing
            </span>
            <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold mb-4 text-white">料金プラン</h2>
            <p className="text-gray-400 text-sm sm:text-base">シンプルな料金体系。隠れた費用はありません。</p>
          </div>
        </AnimatedSection>

        <AnimatedSection delay={200}>
          <div
            className="relative max-w-sm sm:max-w-md mx-auto rounded-2xl sm:rounded-3xl overflow-hidden"
            style={{
              background:
                'linear-gradient(135deg, rgba(79, 70, 229, 0.15) 0%, rgba(139, 92, 246, 0.08) 100%)',
              border: '1px solid rgba(139, 92, 246, 0.25)',
              boxShadow: '0 25px 80px rgba(79, 70, 229, 0.2)',
            }}
          >
            {/* Campaign Badge */}
            <div
              className="py-2.5 sm:py-3 px-4 sm:px-6 text-center text-xs sm:text-sm font-medium text-white"
              style={{
                background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
              }}
            >
              初期キャンペーン実施中
            </div>

            <div className="p-5 sm:p-8 text-center">
              {/* Price */}
              <div className="mb-6 sm:mb-8">
                <div className="flex items-baseline justify-center gap-1 sm:gap-2">
                  <span className="text-4xl sm:text-5xl md:text-6xl font-bold text-white">¥200</span>
                  <span className="text-gray-400 text-sm sm:text-base">/月</span>
                </div>
                <div className="text-gray-500 mt-2 text-xs sm:text-sm">
                  <span className="line-through">通常 ¥500/月</span>
                  <span className="text-violet-400 ml-2">60% OFF</span>
                </div>
              </div>

              {/* Features List */}
              <ul className="space-y-3 sm:space-y-4 mb-6 sm:mb-8 text-left">
                {pricingFeatures.map((feature) => (
                  <li key={feature} className="flex items-center gap-2.5 sm:gap-3">
                    <div
                      className="w-4 h-4 sm:w-5 sm:h-5 rounded-full flex items-center justify-center flex-shrink-0"
                      style={{
                        background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
                      }}
                    >
                      <svg
                        className="w-2.5 h-2.5 sm:w-3 sm:h-3 text-white"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={3}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </div>
                    <span className="text-gray-300 text-sm sm:text-base">{feature}</span>
                  </li>
                ))}
              </ul>

              {/* CTA */}
              <Link
                to="/subscribe"
                className="block w-full py-3.5 sm:py-4 rounded-lg sm:rounded-xl font-semibold text-base sm:text-lg transition-all duration-300 hover:scale-[1.02] active:scale-[0.98] text-white min-h-[48px] flex items-center justify-center"
                style={{
                  background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
                  boxShadow: '0 8px 30px rgba(79, 70, 229, 0.4)',
                }}
              >
                今すぐ始める
              </Link>
            </div>
          </div>
        </AnimatedSection>
      </div>
    </section>
  );
}
