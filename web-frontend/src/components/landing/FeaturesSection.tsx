import { AnimatedSection } from './AnimatedSection';
import { FeatureCard } from './FeatureCard';

const features = [
  {
    icon: '📅',
    title: 'イベント・シフト管理',
    description: 'イベントごとにシフト枠を作成。日付・時間・必要人数を柔軟に設定できます。',
  },
  {
    icon: '📝',
    title: '出欠収集',
    description: '公開URLを共有するだけで出欠を収集。メンバーはログイン不要で回答できます。',
  },
  {
    icon: '👥',
    title: 'メンバー管理',
    description: 'メンバーの登録・管理、ロール設定、グループ分けで効率的に運営。',
  },
  {
    icon: '🎯',
    title: 'シフト調整',
    description: '収集した出欠をもとにシフトを調整。ドラッグ&ドロップで簡単割り当て。',
  },
  {
    icon: '🗓️',
    title: '日程調整',
    description: '複数候補日から最適な日程を決定。メンバーの都合を可視化。',
  },
  {
    icon: '🏢',
    title: 'マルチテナント',
    description: '複数のイベント・チームを一元管理。運営規模に合わせて拡張可能。',
  },
];

export function FeaturesSection() {
  return (
    <section id="features" className="relative py-16 sm:py-24 px-4 sm:px-6">
      <div className="max-w-6xl mx-auto">
        <AnimatedSection>
          <div className="text-center mb-10 sm:mb-16">
            <span className="text-violet-400 text-xs sm:text-sm font-medium tracking-wider uppercase mb-4 block">
              Features
            </span>
            <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold mb-4 text-white">主な機能</h2>
            <p className="text-gray-400 max-w-2xl mx-auto text-sm sm:text-base">
              VRChatイベント運営に必要な機能を搭載。シフト管理の手間を大幅に削減します。
            </p>
          </div>
        </AnimatedSection>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5">
          {features.map((feature, i) => (
            <FeatureCard key={feature.title} {...feature} delay={i * 100} />
          ))}
        </div>
      </div>
    </section>
  );
}
