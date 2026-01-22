import { Link } from 'react-router-dom';

export default function Terms() {
  return (
    <div
      className="min-h-screen text-white"
      style={{
        background: 'linear-gradient(180deg, #0a0a0f 0%, #0f0f1a 50%, #0a0a0f 100%)',
        fontFamily: '"Noto Sans JP", "Inter", system-ui, sans-serif',
      }}
    >
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
          <h1 className="text-2xl sm:text-3xl font-bold mb-8 text-center">利用規約</h1>

          <div className="prose prose-invert prose-sm sm:prose-base max-w-none space-y-8">
            <p className="text-gray-400 text-sm">最終更新日: 2025年1月22日</p>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第1条（総則）</h2>
              <p className="text-gray-300 leading-relaxed">
                本利用規約（以下「本規約」）は、VRC Shift Scheduler（以下「本サービス」）の利用に関する条件を定めるものです。
                本サービスを利用するすべてのユーザー（以下「ユーザー」）は、本規約に同意したものとみなします。
              </p>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第2条（定義）</h2>
              <p className="text-gray-300 leading-relaxed">本規約において、以下の用語は次の意味を有します。</p>
              <ul className="list-disc list-inside text-gray-300 space-y-2 mt-2">
                <li>「本サービス」: VRC Shift Schedulerが提供するシフト管理サービス</li>
                <li>「ユーザー」: 本サービスを利用する個人または団体</li>
                <li>「テナント」: ユーザーが作成する組織単位</li>
                <li>「メンバー」: テナントに登録されるイベント参加者</li>
              </ul>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第3条（アカウント）</h2>
              <ol className="list-decimal list-inside text-gray-300 space-y-2">
                <li>ユーザーは、正確な情報を提供してアカウントを登録するものとします。</li>
                <li>ユーザーは、自己のアカウント情報（メールアドレス、パスワード等）を適切に管理する責任を負います。</li>
                <li>アカウントの不正利用により生じた損害について、運営者は一切の責任を負いません。</li>
                <li>1人のユーザーが複数のアカウントを作成することは原則として禁止します。</li>
              </ol>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第4条（料金・支払い）</h2>
              <ol className="list-decimal list-inside text-gray-300 space-y-2">
                <li>本サービスの利用には、所定の月額料金が発生します。</li>
                <li>料金の支払いは、Stripeまたはその他の指定された方法により行います。</li>
                <li>月額料金は自動更新され、解約手続きを行わない限り継続して課金されます。</li>
                <li>返金をご希望の場合は、お問い合わせください。</li>
                <li>料金の変更がある場合は、事前に通知いたします。</li>
              </ol>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第5条（サービス内容）</h2>
              <p className="text-gray-300 leading-relaxed">本サービスは、以下の機能を提供します。</p>
              <ul className="list-disc list-inside text-gray-300 space-y-2 mt-2">
                <li>イベント・シフト枠の作成・管理</li>
                <li>メンバーの出欠収集</li>
                <li>シフトの割り当て・調整</li>
                <li>日程調整機能</li>
                <li>その他、運営者が提供する機能</li>
              </ul>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第6条（禁止事項）</h2>
              <p className="text-gray-300 leading-relaxed">ユーザーは、以下の行為を行ってはなりません。</p>
              <ul className="list-disc list-inside text-gray-300 space-y-2 mt-2">
                <li>法令または公序良俗に違反する行為</li>
                <li>本サービスのシステムへの不正アクセス、妨害行為</li>
                <li>虚偽の情報を登録する行為</li>
                <li>他のユーザーまたは第三者に迷惑をかける行為</li>
                <li>本サービスの運営を妨害する行為</li>
                <li>自動化ツール、スクレイピング等による不正な利用</li>
                <li>その他、運営者が不適切と判断する行為</li>
              </ul>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第7条（知的財産権）</h2>
              <ol className="list-decimal list-inside text-gray-300 space-y-2">
                <li>本サービスに関する著作権、商標権その他の知的財産権は、運営者に帰属します。</li>
                <li>ユーザーが本サービスに登録したコンテンツの権利は、ユーザーに帰属します。</li>
              </ol>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第8条（免責事項）</h2>
              <ol className="list-decimal list-inside text-gray-300 space-y-2">
                <li>運営者は、本サービスの完全性、正確性、有用性等について保証しません。</li>
                <li>運営者は、システムメンテナンス、障害等により本サービスが中断・停止した場合の損害について責任を負いません。</li>
                <li>運営者は、ユーザーのデータ消失について、故意または重過失がない限り責任を負いません。</li>
                <li>VRChat、Stripe、BOOTHその他の外部サービスに起因する問題について、運営者は責任を負いません。</li>
              </ol>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第9条（契約終了）</h2>
              <ol className="list-decimal list-inside text-gray-300 space-y-2">
                <li>ユーザーは、所定の手続きにより、いつでも本サービスを解約できます。</li>
                <li>運営者は、ユーザーが本規約に違反した場合、事前の通知なくアカウントを停止または削除できます。</li>
                <li>解約後のデータは、一定期間経過後に削除されます。</li>
              </ol>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第10条（規約の変更）</h2>
              <p className="text-gray-300 leading-relaxed">
                運営者は、必要に応じて本規約を変更できます。重要な変更がある場合は、本サービス上で通知いたします。
                変更後も本サービスを継続して利用した場合、ユーザーは変更後の規約に同意したものとみなします。
              </p>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第11条（準拠法・管轄）</h2>
              <ol className="list-decimal list-inside text-gray-300 space-y-2">
                <li>本規約は、日本法に準拠します。</li>
                <li>本サービスに関する紛争については、東京地方裁判所を第一審の専属的合意管轄裁判所とします。</li>
              </ol>
            </section>

            <section>
              <h2 className="text-lg sm:text-xl font-bold text-white mb-4">第12条（お問い合わせ）</h2>
              <p className="text-gray-300 leading-relaxed">
                本規約に関するお問い合わせは、以下までご連絡ください。
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
