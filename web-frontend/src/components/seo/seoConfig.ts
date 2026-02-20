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
    title: 'VRCShift - VRChat イベント向けシフト管理システム | 月額200円から',
    description:
      'VRChat イベントのシフト管理を簡単に。メンバーの空き時間調整、シフト表作成、出欠確認がワンストップで。月額200円から始められます。',
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
    title: 'プラン・料金 | VRCShift - VRChat シフト管理',
    description:
      'VRCShiftの料金プランと新規登録。月額200円でシフト管理を始められます。VRChatイベント向けの出欠・シフト管理ツール。',
    path: '/subscribe',
  },
} as const;
