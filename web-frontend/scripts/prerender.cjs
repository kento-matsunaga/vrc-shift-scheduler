/**
 * Post-build prerender script
 *
 * Injects static SEO content into built HTML files so that search engine
 * crawlers can read meaningful content without executing JavaScript.
 *
 * React's createRoot will replace the injected content on hydration,
 * but crawlers (Google, Bing, Discord, Twitter/X) see the static HTML.
 *
 * Usage: node scripts/prerender.cjs (run after `vite build`)
 */
const fs = require('fs');
const path = require('path');

const DIST_DIR = path.join(__dirname, '..', 'dist');
const BASE_URL = 'https://vrcshift.com';

// ---------------------------------------------------------------------------
// SEO content for each public page
// ---------------------------------------------------------------------------

const LANDING_CONTENT = `<main>
<h1>VRChatイベントのシフト管理を、もっと簡単に。</h1>
<p>Excelやスプレッドシートでの煩雑なシフト管理から解放。出欠収集からシフト調整まで、一括で管理できます。</p>

<section>
<h2>主な機能</h2>
<p>VRChatイベント運営に必要な機能を搭載。シフト管理の手間を大幅に削減します。</p>
<ul>
<li><strong>イベント・シフト管理</strong> — イベントごとにシフト枠を作成。日付・時間・必要人数を柔軟に設定できます。</li>
<li><strong>出欠収集</strong> — 公開URLを共有するだけで出欠を収集。メンバーはログイン不要で回答できます。</li>
<li><strong>メンバー管理</strong> — メンバーの登録・管理、ロール設定、グループ分けで効率的に運営。</li>
<li><strong>シフト調整</strong> — 収集した出欠をもとにシフトを調整。ドラッグ&ドロップで簡単割り当て。</li>
<li><strong>日程調整</strong> — 複数候補日から最適な日程を決定。メンバーの都合を可視化。</li>
<li><strong>マルチテナント</strong> — 複数のイベント・チームを一元管理。運営規模に合わせて拡張可能。</li>
</ul>
</section>

<section>
<h2>料金プラン</h2>
<p>シンプルな料金体系。隠れた費用はありません。</p>
<p>月額200円（初期キャンペーン価格、通常500円/月 — 60% OFF）</p>
<ul>
<li>全機能利用可能</li>
<li>イベント数無制限</li>
<li>メンバー数無制限</li>
<li>シフト枠数無制限</li>
<li>出欠収集機能</li>
<li>日程調整機能</li>
</ul>
<a href="/subscribe">今すぐ始める</a>
</section>

<section>
<h2>よくある質問</h2>
<dl>
<dt>VRC Shift Schedulerとは何ですか？</dt>
<dd>VRChatイベント向けのシフト管理システムです。出欠収集、シフト割り当て、日程調整などの機能を提供しています。</dd>
<dt>料金はいくらですか？</dt>
<dd>月額200円でフル機能をご利用いただけます。初期費用や追加料金はありません。</dd>
<dt>どのような機能がありますか？</dt>
<dd>イベント・シフト枠の作成、メンバーの出欠収集、シフトの割り当て・調整、日程調整機能などがあります。</dd>
<dt>支払い方法は何がありますか？</dt>
<dd>クレジットカード（Stripe経由）またはBOOTHでのライセンスキー購入に対応しています。</dd>
</dl>
</section>

<footer>
<nav>
<a href="/">トップ</a>
<a href="/terms">利用規約</a>
<a href="/privacy">プライバシーポリシー</a>
<a href="/subscribe">新規登録</a>
</nav>
<p>&copy; 2025–${new Date().getFullYear()} VRC Shift Scheduler</p>
</footer>
</main>`;

const TERMS_CONTENT = `<main>
<h1>利用規約</h1>
<p>最終更新日: 2025年1月22日</p>

<section>
<h2>第1条（総則）</h2>
<p>本利用規約（以下「本規約」）は、VRC Shift Scheduler（以下「本サービス」）の利用に関する条件を定めるものです。本サービスを利用するすべてのユーザー（以下「ユーザー」）は、本規約に同意したものとみなします。</p>
</section>

<section>
<h2>第2条（定義）</h2>
<ul>
<li>「本サービス」: VRC Shift Schedulerが提供するシフト管理サービス</li>
<li>「ユーザー」: 本サービスを利用する個人または団体</li>
<li>「テナント」: ユーザーが作成する組織単位</li>
<li>「メンバー」: テナントに登録されるイベント参加者</li>
</ul>
</section>

<section>
<h2>第3条（アカウント）</h2>
<ul>
<li>ユーザーは、正確な情報を提供してアカウントを登録するものとします。</li>
<li>ユーザーは、自己のアカウント情報（メールアドレス、パスワード等）を適切に管理する責任を負います。</li>
<li>アカウントの不正利用により生じた損害について、運営者は一切の責任を負いません。</li>
<li>1人のユーザーが複数のアカウントを作成することは原則として禁止します。</li>
</ul>
</section>

<section>
<h2>第4条（料金・支払い）</h2>
<ul>
<li>本サービスの利用には、所定の月額料金が発生します。</li>
<li>料金の支払いは、Stripeまたはその他の指定された方法により行います。</li>
<li>月額料金は自動更新され、解約手続きを行わない限り継続して課金されます。</li>
<li>返金をご希望の場合は、お問い合わせください。</li>
<li>料金の変更がある場合は、事前に通知いたします。</li>
</ul>
</section>

<section>
<h2>第5条（サービス内容）</h2>
<p>本サービスは、以下の機能を提供します。</p>
<ul>
<li>イベント・シフト枠の作成・管理</li>
<li>メンバーの出欠収集</li>
<li>シフトの割り当て・調整</li>
<li>日程調整機能</li>
<li>その他、運営者が提供する機能</li>
</ul>
</section>

<section>
<h2>第6条（禁止事項）</h2>
<p>ユーザーは、以下の行為を行ってはなりません。</p>
<ul>
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
<h2>第7条（知的財産権）</h2>
<ul>
<li>本サービスに関する著作権、商標権その他の知的財産権は、運営者に帰属します。</li>
<li>ユーザーが本サービスに登録したコンテンツの権利は、ユーザーに帰属します。</li>
</ul>
</section>

<section>
<h2>第8条（免責事項）</h2>
<ul>
<li>運営者は、本サービスの完全性、正確性、有用性等について保証しません。</li>
<li>運営者は、システムメンテナンス、障害等により本サービスが中断・停止した場合の損害について責任を負いません。</li>
<li>運営者は、ユーザーのデータ消失について、故意または重過失がない限り責任を負いません。</li>
<li>VRChat、Stripe、BOOTHその他の外部サービスに起因する問題について、運営者は責任を負いません。</li>
</ul>
</section>

<section>
<h2>第9条（契約終了）</h2>
<ul>
<li>ユーザーは、所定の手続きにより、いつでも本サービスを解約できます。</li>
<li>運営者は、ユーザーが本規約に違反した場合、事前の通知なくアカウントを停止または削除できます。</li>
<li>解約後のデータは、一定期間経過後に削除されます。</li>
</ul>
</section>

<section>
<h2>第10条（規約の変更）</h2>
<p>運営者は、必要に応じて本規約を変更できます。重要な変更がある場合は、本サービス上で通知いたします。変更後も本サービスを継続して利用した場合、ユーザーは変更後の規約に同意したものとみなします。</p>
</section>

<section>
<h2>第11条（準拠法・管轄）</h2>
<ul>
<li>本規約は、日本法に準拠します。</li>
<li>本サービスに関する紛争については、東京地方裁判所を第一審の専属的合意管轄裁判所とします。</li>
</ul>
</section>

<section>
<h2>第12条（お問い合わせ）</h2>
<p>本規約に関するお問い合わせは、以下までご連絡ください。</p>
<p>X (Twitter): @Noa_Fortevita</p>
</section>

<nav><a href="/">トップページに戻る</a></nav>
</main>`;

const PRIVACY_CONTENT = `<main>
<h1>プライバシーポリシー</h1>
<p>最終更新日: 2025年1月22日</p>
<p>VRC Shift Scheduler（以下「本サービス」）は、ユーザーのプライバシーを尊重し、個人情報の保護に努めます。本プライバシーポリシーでは、本サービスにおける個人情報の取り扱いについて説明します。</p>

<section>
<h2>1. 収集する情報</h2>

<h3>1.1 アカウント情報</h3>
<ul>
<li>メールアドレス</li>
<li>パスワード（ハッシュ化して保存）</li>
</ul>

<h3>1.2 テナント・イベント情報</h3>
<ul>
<li>テナント名</li>
<li>イベント名・日時・説明</li>
<li>シフト枠情報</li>
<li>メンバー情報（表示名、出欠情報等）</li>
</ul>

<h3>1.3 決済情報</h3>
<ul>
<li>Stripe顧客ID・サブスクリプションID</li>
<li>BOOTHライセンスキー（ハッシュ化して保存）</li>
</ul>
<p>※クレジットカード情報は本サービスでは保存せず、Stripeが安全に管理します。</p>

<h3>1.4 利用情報</h3>
<ul>
<li>アクセス日時</li>
<li>IPアドレス</li>
<li>ブラウザ・デバイス情報</li>
</ul>
</section>

<section>
<h2>2. 情報の利用目的</h2>
<p>収集した情報は、以下の目的で利用します。</p>
<ul>
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
<h2>3. 第三者への提供</h2>
<p>本サービスは、以下の場合を除き、ユーザーの個人情報を第三者に提供しません。</p>
<ul>
<li>ユーザーの同意がある場合</li>
<li>法令に基づく開示請求があった場合</li>
<li>人の生命・身体・財産の保護に必要な場合</li>
</ul>
</section>

<section>
<h2>4. 外部サービスとの連携</h2>

<h3>4.1 Stripe</h3>
<p>決済処理のため、Stripeにユーザー情報（メールアドレス等）を提供します。</p>

<h3>4.2 BOOTH</h3>
<p>ライセンス認証のため、BOOTHで発行されたライセンスキーを使用します。</p>
</section>

<section>
<h2>5. データの保護</h2>
<p>本サービスは、ユーザーの情報を保護するため、以下の対策を実施しています。</p>
<ul>
<li>通信の暗号化（HTTPS/TLS）</li>
<li>パスワードのハッシュ化（bcrypt）</li>
<li>データベースへのアクセス制限</li>
<li>定期的なセキュリティ監査</li>
</ul>
</section>

<section>
<h2>6. データの保存</h2>
<ul>
<li>ユーザーのデータは、サービス利用中および解約後一定期間保存されます。</li>
<li>解約後、一定期間経過後にデータは削除されます。</li>
<li>法令で保存が義務付けられている情報は、所定の期間保存します。</li>
</ul>
</section>

<section>
<h2>7. ユーザーの権利</h2>
<p>ユーザーは、以下の権利を有します。</p>
<ul>
<li>自己の個人情報の開示を請求する権利</li>
<li>個人情報の訂正・削除を請求する権利</li>
<li>個人情報の利用停止を請求する権利</li>
</ul>
<p>これらの請求を行う場合は、お問い合わせ先までご連絡ください。</p>
</section>

<section>
<h2>8. Cookie</h2>
<p>本サービスでは、ユーザー認証およびセッション管理のためにCookieを使用します。Cookieを無効にした場合、本サービスの一部機能が利用できなくなる場合があります。</p>
</section>

<section>
<h2>9. ポリシーの変更</h2>
<p>本プライバシーポリシーは、必要に応じて変更されることがあります。重要な変更がある場合は、本サービス上で通知いたします。</p>
</section>

<section>
<h2>10. お問い合わせ</h2>
<p>本プライバシーポリシーに関するお問い合わせは、以下までご連絡ください。</p>
<p>X (Twitter): @Noa_Fortevita</p>
</section>

<nav><a href="/">トップページに戻る</a></nav>
</main>`;

const SUBSCRIBE_CONTENT = `<main>
<h1>新規登録</h1>
<p>アカウントを作成して、シフト管理を始めましょう</p>
<p>初期キャンペーン: 月額200円</p>

<section>
<h2>アカウント情報の入力</h2>
<p>メールアドレス、パスワード、組織名（イベント名など）、表示名を入力して登録できます。</p>
<p>登録後、Stripe決済ページに進みます。</p>
</section>

<nav>
<a href="/">トップページに戻る</a>
<a href="/terms">利用規約</a>
<a href="/privacy">プライバシーポリシー</a>
</nav>
</main>`;

// ---------------------------------------------------------------------------
// Page configurations
// ---------------------------------------------------------------------------

const PAGES = [
  {
    route: '/',
    title: 'VRCShift - VRChat イベント向けシフト管理システム | 月額200円から',
    description:
      'VRChat イベントのシフト管理を簡単に。メンバーの空き時間調整、シフト表作成、出欠確認がワンストップで。月額200円から始められます。',
    content: LANDING_CONTENT,
  },
  {
    route: '/terms',
    title: '利用規約 | VRC Shift Scheduler',
    description:
      'VRC Shift Schedulerの利用規約です。本サービスをご利用いただく前に必ずお読みください。',
    content: TERMS_CONTENT,
  },
  {
    route: '/privacy',
    title: 'プライバシーポリシー | VRC Shift Scheduler',
    description:
      'VRC Shift Schedulerのプライバシーポリシーです。個人情報の取り扱いについてご確認ください。',
    content: PRIVACY_CONTENT,
  },
  {
    route: '/subscribe',
    title: 'プラン・料金 | VRCShift - VRChat シフト管理',
    description:
      'VRCShiftの料金プランと新規登録。月額200円でシフト管理を始められます。VRChatイベント向けの出欠・シフト管理ツール。',
    content: SUBSCRIBE_CONTENT,
  },
];

// ---------------------------------------------------------------------------
// Prerender logic
// ---------------------------------------------------------------------------

function replaceMetaTag(html, attr, name, newContent) {
  // 属性順序に依存しない正規表現: attr="name" と content="..." が任意の順で出現
  const re = new RegExp(
    `<meta\\s+(?:${attr}="${name}"\\s+content="[^"]*"|content="[^"]*"\\s+${attr}="${name}")\\s*/?>`,
    'g'
  );
  return html.replace(re, `<meta ${attr}="${name}" content="${newContent}" />`);
}

function prerender() {
  const templatePath = path.join(DIST_DIR, 'index.html');
  if (!fs.existsSync(templatePath)) {
    console.error('Error: dist/index.html not found. Run `vite build` first.');
    process.exit(1);
  }

  const template = fs.readFileSync(templatePath, 'utf-8');

  for (const page of PAGES) {
    let html = template;

    // Update <title>
    html = html.replace(/<title>[^<]*<\/title>/, `<title>${page.title}</title>`);

    // Update <meta name="description">
    html = replaceMetaTag(html, 'name', 'description', page.description);

    // Update canonical URL
    const canonical = `${BASE_URL}${page.route === '/' ? '/' : page.route}`;
    html = html.replace(
      /<link rel="canonical" href="[^"]*" \/>/,
      `<link rel="canonical" href="${canonical}" />`
    );

    // Update Open Graph tags
    html = replaceMetaTag(html, 'property', 'og:title', page.title);
    html = replaceMetaTag(html, 'property', 'og:description', page.description);
    html = replaceMetaTag(html, 'property', 'og:url', canonical);

    // Update Twitter Card tags
    html = replaceMetaTag(html, 'name', 'twitter:title', page.title);
    html = replaceMetaTag(html, 'name', 'twitter:description', page.description);

    // Inject SEO content into <div id="root">
    html = html.replace(
      '<div id="root"></div>',
      `<div id="root">${page.content}</div>`
    );

    // Determine output path
    const outputDir =
      page.route === '/' ? DIST_DIR : path.join(DIST_DIR, page.route.slice(1));

    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    const outputFile = path.join(outputDir, 'index.html');
    fs.writeFileSync(outputFile, html);
    console.log(`  prerendered: ${page.route} -> ${path.relative(DIST_DIR, outputFile)}`);
  }

  console.log('Prerender complete!');
}

prerender();
