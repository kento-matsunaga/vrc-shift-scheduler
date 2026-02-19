// Settings menu configuration
// Shared between SettingsSidebar and SettingsMobileMenu

export interface MenuItem {
  id: string;
  label: string;
  path: string;
  ownerOnly?: boolean;
  danger?: boolean;
}

export const menuItems: MenuItem[] = [
  { id: 'organization', label: '組織情報', path: '/settings/organization' },
  { id: 'account', label: 'アカウント', path: '/settings/account' },
  { id: 'billing', label: '課金管理', path: '/settings/billing', ownerOnly: true },
  { id: 'permissions', label: '権限設定', path: '/settings/permissions', ownerOnly: true },
  { id: 'events', label: 'イベント管理', path: '/settings/events', danger: true },
  { id: 'import', label: 'データ取込', path: '/settings/import' },
];
