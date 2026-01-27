import {
  Header,
  HeroSection,
  FeaturesSection,
  PricingSection,
  CTASection,
  Footer,
} from '../components/landing';
import { SEO, PAGE_SEO } from '../components/seo';
import { ReleaseStatusProvider } from '../hooks/useReleaseStatus';

// FAQ data for structured data
const FAQ_DATA = [
  {
    question: 'VRC Shift Schedulerとは何ですか？',
    answer:
      'VRChatイベント向けのシフト管理システムです。出欠収集、シフト割り当て、日程調整などの機能を提供しています。',
  },
  {
    question: '料金はいくらですか？',
    answer:
      '月額200円でフル機能をご利用いただけます。初期費用や追加料金はありません。',
  },
  {
    question: 'どのような機能がありますか？',
    answer:
      'イベント・シフト枠の作成、メンバーの出欠収集、シフトの割り当て・調整、日程調整機能などがあります。',
  },
  {
    question: '支払い方法は何がありますか？',
    answer:
      'クレジットカード（Stripe経由）またはBOOTHでのライセンスキー購入に対応しています。',
  },
];

export default function Landing() {
  return (
    <ReleaseStatusProvider>
    <div
      className="min-h-screen text-white overflow-x-hidden"
      style={{
        background: 'linear-gradient(180deg, #0a0a0f 0%, #0f0f1a 50%, #0a0a0f 100%)',
        fontFamily: '"Noto Sans JP", "Inter", system-ui, sans-serif',
        minHeight: '-webkit-fill-available',
      }}
    >
      {/* Background Effects */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        {/* Gradient Orbs */}
        <div
          className="absolute w-[800px] h-[800px] rounded-full"
          style={{
            top: '-20%',
            right: '-10%',
            background: 'radial-gradient(circle, rgba(79, 70, 229, 0.15) 0%, transparent 70%)',
            filter: 'blur(100px)',
            animation: 'float 20s ease-in-out infinite',
          }}
        />
        <div
          className="absolute w-[600px] h-[600px] rounded-full"
          style={{
            bottom: '10%',
            left: '-15%',
            background: 'radial-gradient(circle, rgba(139, 92, 246, 0.12) 0%, transparent 70%)',
            filter: 'blur(80px)',
            animation: 'float 15s ease-in-out infinite reverse',
          }}
        />
        {/* Grid Pattern */}
        <div
          className="absolute inset-0 opacity-[0.03]"
          style={{
            backgroundImage:
              'linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)',
            backgroundSize: '60px 60px',
          }}
        />
      </div>

      {/* SEO - React 19 hoists meta tags to <head> */}
      <SEO
        title={PAGE_SEO.landing.title}
        description={PAGE_SEO.landing.description}
        path={PAGE_SEO.landing.path}
        jsonLd={{
          organization: true,
          webSite: true,
          softwareApplication: true,
          faq: FAQ_DATA,
        }}
      />

      {/* Content */}
      <div className="relative z-10">
        <Header />
        <HeroSection />
        <FeaturesSection />
        <PricingSection />
        <CTASection />
        <Footer />
      </div>

      {/* Global Styles */}
      <style>{`
        @keyframes float {
          0%, 100% { transform: translate(0, 0); }
          50% { transform: translate(-20px, 20px); }
        }

        html {
          scroll-behavior: smooth;
          -webkit-text-size-adjust: 100%;
        }

        /* iOS Safari 100vh fix */
        @supports (-webkit-touch-callout: none) {
          .min-h-screen {
            min-height: -webkit-fill-available;
          }
        }

        /* Safe area for notched devices */
        .safe-area-top {
          padding-top: env(safe-area-inset-top, 0);
        }
        .safe-area-bottom {
          padding-bottom: env(safe-area-inset-bottom, 0);
        }

        ::selection {
          background: rgba(139, 92, 246, 0.3);
        }

        /* Prevent text selection on touch */
        .no-select {
          -webkit-user-select: none;
          user-select: none;
        }

        /* Smooth scrolling for iOS */
        .scroll-container {
          -webkit-overflow-scrolling: touch;
        }
      `}</style>
    </div>
    </ReleaseStatusProvider>
  );
}
