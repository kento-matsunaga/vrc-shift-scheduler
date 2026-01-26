import { Link } from 'react-router-dom';
import { SEO, PAGE_SEO, SEO_CONFIG } from '../components/seo';

export default function Privacy() {
  return (
    <div
      className="min-h-screen text-white"
      style={{
        background: 'linear-gradient(180deg, #0a0a0f 0%, #0f0f1a 50%, #0a0a0f 100%)',
        fontFamily: '"Noto Sans JP", "Inter", system-ui, sans-serif',
      }}
    >
      {/* SEO */}
      <SEO
        title={PAGE_SEO.privacy.title}
        description={PAGE_SEO.privacy.description}
        path={PAGE_SEO.privacy.path}
        jsonLd={{
          breadcrumbs: [
            { name: 'ホーム', url: SEO_CONFIG.baseUrl },
            { name: 'プライバシーポリシー' },
          ],
        }}
      />

      {/* Header */}
      <header
        className="fixed top-0 left-0 right-0 z-50 py-4"
        style={{
          background: 'rgba(10, 10, 15, 0.85)',
          backdropFilter: 'blur(20px)',
          WebkitBackdropFilter: 'blur(20px)',
          borderBottom: '1px solid rgba(139, 92, 246, 0.1)',
        }}
      >
        <div className="max-w-4xl mx-auto px-4 sm:px-6 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 sm:gap-3 group">
            <div
              className="w-8 h-8 sm:w-10 sm:h-10 rounded-xl flex items-center justify-center"
              style={{
                background: 'linear-gradient(135deg, #4F46E5 0%, #8B5CF6 100%)',
              }}
            >
              <span className="text-base sm:text-xl">📅</span>
            </div>
            <span className="font-bold text-sm sm:text-lg text-white">VRC Shift Scheduler</span>
          </Link>
          <Link
            to="/"
            className="text-gray-400 hover:text-white transition-colors text-sm"
          >
            トップに戻る
          </Link>
        </div>
      </header>

      {/* Content */}
      <main className="pt-24 pb-16 px-4 sm:px-6">
        <div className="max-w-3xl mx-auto">
          <h1 className="text-2xl sm:text-3xl font-bold mb-8 text-center">プライバシーポリシー</h1>

          <div className="prose prose-invert prose-sm sm:prose-base max-w-none space-y-8">
            <p className="text-gray-400 text-sm">最終更新日: 2025年1月22日</p>

            <p className="text-gray-300 leading-relaxed">
              VRC Shift Scheduler（以下「本サービス」）は、ユーザーのプライバシーを尊重し、個人情報の保護に努めます。
              本プライバシーポリシーでは、本サービスにおける個人情報の取り扱いについて説明します。
            </p>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">1. 収集する情報</h2>
              <p className="text-gray-300 leading-relaxed">本サービスでは、以下の情報を収集します。</p>

              <h3 className="text-base font-semibold text-white mt-4 mb-2">1.1 アカウント情報</h3>
              <ul className="list-disc list-inside text-gray-300 space-y-1">
                <li>メールアドレス</li>
                <li>パスワード（ハッシュ化して保存）</li>
              </ul>

              <h3 className="text-base font-semibold text-white mt-4 mb-2">1.2 テナント・イベント情報</h3>
              <ul className="list-disc list-inside text-gray-300 space-y-1">
                <li>テナント名</li>
                <li>イベント名・日時・説明</li>
                <li>シフト枠情報</li>
                <li>メンバー情報（表示名、出欠情報等）</li>
              </ul>

              <h3 className="text-base font-semibold text-white mt-4 mb-2">1.3 決済情報</h3>
              <ul className="list-disc list-inside text-gray-300 space-y-1">
                <li>Stripe顧客ID・サブスクリプションID</li>
                <li>BOOTHライセンスキー（ハッシュ化して保存）</li>
              </ul>
              <p className="text-gray-400 text-sm mt-2">
                ※クレジットカード情報は本サービスでは保存せず、Stripeが安全に管理します。
              </p>

              <h3 className="text-base font-semibold text-white mt-4 mb-2">1.4 利用情報</h3>
              <ul className="list-disc list-inside text-gray-300 space-y-1">
                <li>アクセス日時</li>
                <li>IPアドレス</li>
                <li>ブラウザ・デバイス情報</li>
              </ul>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">2. 情報の利用目的</h2>
              <p className="text-gray-300 leading-relaxed">収集した情報は、以下の目的で利用します。</p>
              <ul className="list-disc list-inside text-gray-300 space-y-2 mt-2">
                <li>本サービスの提供・運営</li>
                <li>ユーザー認証・本人確認</li>
                <li>料金の請求・決済処理</li>
                <li>お問い合わせへの対応</li>
                <li>サービスの改善・新機能の開発</li>
                <li>重要なお知らせ（メンテナンス、規約変更等）の通知</li>
                <li>不正利用の防止・セキュリティ対策</li>
              </ul>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">3. 第三者への提供</h2>
              <p className="text-gray-300 leading-relaxed">
                本サービスは、以下の場合を除き、ユーザーの個人情報を第三者に提供しません。
              </p>
              <ul className="list-disc list-inside text-gray-300 space-y-2 mt-2">
                <li>ユーザーの同意がある場合</li>
                <li>法令に基づく開示請求があった場合</li>
                <li>人の生命・身体・財産の保護に必要な場合</li>
              </ul>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">4. 外部サービスとの連携</h2>
              <p className="text-gray-300 leading-relaxed">本サービスは、以下の外部サービスと連携しています。</p>

              <h3 className="text-base font-semibold text-white mt-4 mb-2">4.1 Stripe</h3>
              <p className="text-gray-300">
                決済処理のため、Stripeにユーザー情報（メールアドレス等）を提供します。
                Stripeのプライバシーポリシーは{' '}
                <a
                  href="https://stripe.com/jp/privacy"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-violet-400 hover:text-violet-300 transition-colors"
                >
                  こちら
                </a>
                をご確認ください。
              </p>

              <h3 className="text-base font-semibold text-white mt-4 mb-2">4.2 BOOTH</h3>
              <p className="text-gray-300">
                ライセンス認証のため、BOOTHで発行されたライセンスキーを使用します。
              </p>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">5. データの保護</h2>
              <p className="text-gray-300 leading-relaxed">本サービスは、ユーザーの情報を保護するため、以下の対策を実施しています。</p>
              <ul className="list-disc list-inside text-gray-300 space-y-2 mt-2">
                <li>通信の暗号化（HTTPS/TLS）</li>
                <li>パスワードのハッシュ化（bcrypt）</li>
                <li>データベースへのアクセス制限</li>
                <li>定期的なセキュリティ監査</li>
              </ul>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">6. データの保存</h2>
              <ol className="list-decimal list-inside text-gray-300 space-y-2">
                <li>ユーザーのデータは、サービス利用中および解約後一定期間保存されます。</li>
                <li>解約後、一定期間経過後にデータは削除されます。</li>
                <li>法令で保存が義務付けられている情報は、所定の期間保存します。</li>
              </ol>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">7. ユーザーの権利</h2>
              <p className="text-gray-300 leading-relaxed">ユーザーは、以下の権利を有します。</p>
              <ul className="list-disc list-inside text-gray-300 space-y-2 mt-2">
                <li>自己の個人情報の開示を請求する権利</li>
                <li>個人情報の訂正・削除を請求する権利</li>
                <li>個人情報の利用停止を請求する権利</li>
              </ul>
              <p className="text-gray-300 mt-2">
                これらの請求を行う場合は、お問い合わせ先までご連絡ください。
              </p>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">8. Cookie</h2>
              <p className="text-gray-300 leading-relaxed">
                本サービスでは、ユーザー認証およびセッション管理のためにCookieを使用します。
                Cookieを無効にした場合、本サービスの一部機能が利用できなくなる場合があります。
              </p>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">9. ポリシーの変更</h2>
              <p className="text-gray-300 leading-relaxed">
                本プライバシーポリシーは、必要に応じて変更されることがあります。
                重要な変更がある場合は、本サービス上で通知いたします。
              </p>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">10. お問い合わせ</h2>
              <p className="text-gray-300 leading-relaxed">
                本プライバシーポリシーに関するお問い合わせは、以下までご連絡ください。
              </p>
              <p className="mt-2">
                <a
                  href="https://x.com/Noa_Fortevita"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-violet-400 hover:text-violet-300 transition-colors"
                >
                  X (Twitter): @Noa_Fortevita
                </a>
              </p>
            </section>
          </div>

          <div className="mt-12 text-center">
            <Link
              to="/"
              className="inline-flex items-center gap-2 text-violet-400 hover:text-violet-300 transition-colors"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              トップページに戻る
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
