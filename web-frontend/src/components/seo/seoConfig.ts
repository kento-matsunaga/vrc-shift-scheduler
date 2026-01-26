/**
 * SEO Configuration Constants
 */
export const SEO_CONFIG = {
  siteName: 'VRC Shift Scheduler',
  baseUrl: 'https://vrcshift.com',
  locale: 'ja_JP',
  twitterHandle: '@Noa_Fortevita',
  twitterUrl: 'https://x.com/Noa_Fortevita',
  defaultMeta: {
    title: 'VRC Shift Scheduler - VRChatイベントのシフト管理ツール',
    description:
      'VRChatイベント向けのシフト管理システム。出欠収集からシフト調整まで一括管理。月額200円から。',
    ogImage: '/images/og/default.png',
  },
  organization: {
    name: 'VRC Shift Scheduler',
    url: 'https://vrcshift.com',
    logo: 'https://vrcshift.com/images/og/default.png',
  },
} as const;

/**
 * Page-specific SEO configurations
 */
export const PAGE_SEO = {
  landing: {
    title: 'VRC Shift Scheduler - VRChatイベントのシフト管理ツール',
    description:
      'VRChatイベント向けのシフト管理システム。出欠収集からシフト調整まで一括管理。Discord連携対応。月額200円から。',
    path: '/',
  },
  terms: {
    title: '利用規約 | VRC Shift Scheduler',
    description:
      'VRC Shift Schedulerの利用規約です。本サービスをご利用いただく前に必ずお読みください。',
    path: '/terms',
  },
  privacy: {
    title: 'プライバシーポリシー | VRC Shift Scheduler',
    description:
      'VRC Shift Schedulerのプライバシーポリシーです。個人情報の取り扱いについてご確認ください。',
    path: '/privacy',
  },
  subscribe: {
    title: '新規登録 | VRC Shift Scheduler',
    description:
      'VRC Shift Schedulerに新規登録して、シフト管理を始めましょう。月額200円でフル機能をご利用いただけます。',
    path: '/subscribe',
  },
} as const;
