import { useParams, Navigate } from 'react-router-dom';
import { SettingsLayout } from '../components/settings/SettingsLayout';
import { OrganizationSettings } from '../components/settings/OrganizationSettings';
import { AccountSettings } from '../components/settings/AccountSettings';
import { BillingSettings } from '../components/settings/BillingSettings';
import { PermissionsSettings } from '../components/settings/PermissionsSettings';
import { EventsSettings } from '../components/settings/EventsSettings';
import { ImportSettings } from '../components/settings/ImportSettings';
import { SEO } from '../components/seo';

const sectionComponents = {
  organization: { component: OrganizationSettings, title: '組織情報' },
  account: { component: AccountSettings, title: 'アカウント' },
  billing: { component: BillingSettings, title: '課金管理' },
  permissions: { component: PermissionsSettings, title: '権限設定' },
  events: { component: EventsSettings, title: 'イベント管理' },
  import: { component: ImportSettings, title: 'データ取込' },
};

export default function Settings() {
  const { section } = useParams<{ section?: string }>();

  // デフォルトは organization
  if (!section) {
    return <Navigate to="/settings/organization" replace />;
  }

  const config = sectionComponents[section as keyof typeof sectionComponents];
  if (!config) {
    return <Navigate to="/settings/organization" replace />;
  }

  const SectionComponent = config.component;

  return (
    <>
      <SEO noindex={true} />
      <SettingsLayout title={config.title}>
        <SectionComponent />
      </SettingsLayout>
    </>
  );
}
