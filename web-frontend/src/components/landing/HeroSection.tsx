import { Link } from 'react-router-dom';
import { AnimatedSection } from './AnimatedSection';
import { InteractiveDemo } from './InteractiveDemo';

export function HeroSection() {
  return (
    <section id="top" className="relative min-h-screen flex items-center justify-center pt-20 sm:pt-24 pb-16 sm:pb-20 px-4 sm:px-6">
      <div className="max-w-6xl mx-auto">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* 左側: テキストコンテンツ */}
          <div className="text-center lg:text-left">
            {/* Badge */}
            <AnimatedSection>
              <div
                className="inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm mb-6"
                style={{
                  background: 'rgba(139, 92, 246, 0.15)',
                  border: '1px solid rgba(139, 92, 246, 0.3)',
                }}
              >
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-violet-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-violet-500" />
                </span>
                <span className="text-violet-300">初期キャンペーン実施中</span>
              </div>
            </AnimatedSection>

            {/* Main Heading */}
            <AnimatedSection delay={100}>
              <h1
                className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-bold mb-6 leading-tight"
                style={{
                  background: 'linear-gradient(135deg, #ffffff 0%, #a5b4fc 50%, #8B5CF6 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  backgroundClip: 'text',
                }}
              >
                <span className="inline sm:block">VRChatイベントの</span>
                <span className="inline sm:block">シフト管理を、</span>
                <span
                  className="inline sm:block"
                  style={{
                    background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 50%, #c4b5fd 100%)',
                    WebkitBackgroundClip: 'text',
                    backgroundClip: 'text',
                  }}
                >
                  もっと簡単に。
                </span>
              </h1>
            </AnimatedSection>

            {/* Subheading */}
            <AnimatedSection delay={200}>
              <p className="text-sm sm:text-base lg:text-lg text-gray-400 mb-6 sm:mb-8 max-w-lg mx-auto lg:mx-0 leading-relaxed">
                Excelやスプレッドシートでの煩雑なシフト管理から解放。出欠収集からシフト調整まで、一括で管理できます。
              </p>
            </AnimatedSection>

            {/* CTA Buttons */}
            <AnimatedSection delay={300}>
              <div className="flex flex-col sm:flex-row items-center lg:items-start justify-center lg:justify-start gap-4 mb-6">
                <Link
                  to="/subscribe"
                  className="w-full sm:w-auto px-8 py-4 rounded-full font-semibold text-base sm:text-lg transition-all duration-300 hover:scale-105 active:scale-95 text-white text-center min-h-[48px] flex items-center justify-center"
                  style={{
                    background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
                    boxShadow: '0 8px 40px rgba(79, 70, 229, 0.5)',
                  }}
                >
                  今すぐ始める
                  <span className="text-violet-200 text-sm font-normal ml-2">月額200円〜</span>
                </Link>
              </div>
            </AnimatedSection>

            {/* Login Link */}
            <AnimatedSection delay={400}>
              <p className="text-gray-500 text-sm">
                すでにアカウントをお持ちの方は{' '}
                <Link
                  to="/admin/login"
                  className="text-violet-400 hover:text-violet-300 transition-colors underline underline-offset-4"
                >
                  ログイン
                </Link>
              </p>
            </AnimatedSection>
          </div>

          {/* 右側: インタラクティブデモ */}
          <AnimatedSection delay={300} className="hidden lg:flex justify-center">
            <InteractiveDemo />
          </AnimatedSection>
        </div>

        {/* モバイル用: インタラクティブデモ */}
        <AnimatedSection delay={500} className="lg:hidden mt-12 flex justify-center">
          <InteractiveDemo />
        </AnimatedSection>
      </div>
    </section>
  );
}
