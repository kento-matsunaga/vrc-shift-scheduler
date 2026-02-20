import { SEO_CONFIG } from './seoConfig';
import { JsonLd, type JsonLdSchema } from './JsonLd';
import { schemas } from './schemas';

interface SEOProps {
  title?: string;
  description?: string;
  keywords?: string;
  path?: string;
  ogImage?: string;
  ogType?: 'website' | 'article';
  noindex?: boolean;
  jsonLd?: {
    organization?: boolean;
    webSite?: boolean;
    softwareApplication?: boolean;
    faq?: { question: string; answer: string }[];
    breadcrumbs?: { name: string; url?: string }[];
  };
}

/**
 * SEO Component using React 19 native document metadata
 *
 * React 19 automatically hoists <title>, <meta>, and <link> tags
 * from component trees to the document <head>.
 */
export function SEO({
  title = SEO_CONFIG.defaultMeta.title,
  description = SEO_CONFIG.defaultMeta.description,
  keywords = 'VRChat,シフト管理,イベント管理,出欠確認,スケジュール',
  path = '/',
  ogImage = SEO_CONFIG.defaultMeta.ogImage,
  ogType = 'website',
  noindex = false,
  jsonLd,
}: SEOProps) {
  const canonicalUrl = `${SEO_CONFIG.baseUrl}${path}`;
  const ogImageUrl = ogImage.startsWith('http')
    ? ogImage
    : `${SEO_CONFIG.baseUrl}${ogImage}`;

  // Build JSON-LD schemas
  const jsonLdSchemas: JsonLdSchema[] = [];

  if (jsonLd?.organization) {
    jsonLdSchemas.push(schemas.organization());
  }
  if (jsonLd?.webSite) {
    jsonLdSchemas.push(schemas.webSite());
  }
  if (jsonLd?.softwareApplication) {
    jsonLdSchemas.push(schemas.softwareApplication());
  }
  if (jsonLd?.faq && jsonLd.faq.length > 0) {
    jsonLdSchemas.push(schemas.faqPage(jsonLd.faq));
  }
  if (jsonLd?.breadcrumbs && jsonLd.breadcrumbs.length > 0) {
    jsonLdSchemas.push(schemas.breadcrumbList(jsonLd.breadcrumbs));
  }

  return (
    <>
      {/* Basic Meta Tags - React 19 hoists these to <head> */}
      <title>{title}</title>
      <meta name="description" content={description} />
      <meta name="keywords" content={keywords} />
      <meta name="author" content={SEO_CONFIG.siteName} />
      <meta name="theme-color" content="#4F46E5" />

      {/* Robots */}
      {noindex && <meta name="robots" content="noindex, nofollow" />}

      {/* Canonical URL */}
      <link rel="canonical" href={canonicalUrl} />

      {/* Open Graph */}
      <meta property="og:title" content={title} />
      <meta property="og:description" content={description} />
      <meta property="og:type" content={ogType} />
      <meta property="og:url" content={canonicalUrl} />
      <meta property="og:image" content={ogImageUrl} />
      <meta property="og:image:width" content="1200" />
      <meta property="og:image:height" content="630" />
      <meta property="og:image:alt" content={title} />
      <meta property="og:site_name" content={SEO_CONFIG.siteName} />
      <meta property="og:locale" content={SEO_CONFIG.locale} />

      {/* Twitter Card */}
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:site" content={SEO_CONFIG.twitterHandle} />
      <meta name="twitter:creator" content={SEO_CONFIG.twitterHandle} />
      <meta name="twitter:title" content={title} />
      <meta name="twitter:description" content={description} />
      <meta name="twitter:image" content={ogImageUrl} />

      {/* JSON-LD Structured Data */}
      {jsonLdSchemas.length > 0 && (
        <JsonLd data={jsonLdSchemas.length === 1 ? jsonLdSchemas[0] : jsonLdSchemas} />
      )}
    </>
  );
}

export default SEO;
