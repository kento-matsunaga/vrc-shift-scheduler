import { SEO_CONFIG } from './seoConfig';

/**
 * Organization Schema
 */
export interface OrganizationSchema {
  '@context': 'https://schema.org';
  '@type': 'Organization';
  name: string;
  url: string;
  logo: string;
  sameAs?: string[];
}

/**
 * WebSite Schema
 */
export interface WebSiteSchema {
  '@context': 'https://schema.org';
  '@type': 'WebSite';
  name: string;
  url: string;
  description?: string;
  inLanguage?: string;
}

/**
 * SoftwareApplication Schema
 */
export interface SoftwareApplicationSchema {
  '@context': 'https://schema.org';
  '@type': 'SoftwareApplication';
  name: string;
  applicationCategory: string;
  operatingSystem: string;
  offers: {
    '@type': 'Offer';
    price: string;
    priceCurrency: string;
    priceValidUntil?: string;
  };
  description?: string;
}

/**
 * FAQPage Schema
 */
export interface FAQPageSchema {
  '@context': 'https://schema.org';
  '@type': 'FAQPage';
  mainEntity: {
    '@type': 'Question';
    name: string;
    acceptedAnswer: {
      '@type': 'Answer';
      text: string;
    };
  }[];
}

/**
 * BreadcrumbList Schema
 */
export interface BreadcrumbListSchema {
  '@context': 'https://schema.org';
  '@type': 'BreadcrumbList';
  itemListElement: {
    '@type': 'ListItem';
    position: number;
    name: string;
    item?: string;
  }[];
}

type JsonLdSchema =
  | OrganizationSchema
  | WebSiteSchema
  | SoftwareApplicationSchema
  | FAQPageSchema
  | BreadcrumbListSchema;

interface JsonLdProps {
  data: JsonLdSchema | JsonLdSchema[];
}

/**
 * JSON-LD Structured Data Component
 */
export function JsonLd({ data }: JsonLdProps) {
  const jsonLdString = JSON.stringify(data);

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: jsonLdString }}
    />
  );
}

/**
 * Pre-built schema generators
 */
export const schemas = {
  organization: (): OrganizationSchema => ({
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: SEO_CONFIG.organization.name,
    url: SEO_CONFIG.organization.url,
    logo: SEO_CONFIG.organization.logo,
    sameAs: [SEO_CONFIG.twitterUrl],
  }),

  webSite: (): WebSiteSchema => ({
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    name: SEO_CONFIG.siteName,
    url: SEO_CONFIG.baseUrl,
    description: SEO_CONFIG.defaultMeta.description,
    inLanguage: 'ja',
  }),

  softwareApplication: (): SoftwareApplicationSchema => ({
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: SEO_CONFIG.siteName,
    applicationCategory: 'BusinessApplication',
    operatingSystem: 'Web Browser',
    description: SEO_CONFIG.defaultMeta.description,
    offers: {
      '@type': 'Offer',
      price: '200',
      priceCurrency: 'JPY',
    },
  }),

  faqPage: (
    faqs: { question: string; answer: string }[]
  ): FAQPageSchema => ({
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: faqs.map((faq) => ({
      '@type': 'Question',
      name: faq.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: faq.answer,
      },
    })),
  }),

  breadcrumbList: (
    items: { name: string; url?: string }[]
  ): BreadcrumbListSchema => ({
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      ...(item.url && { item: item.url }),
    })),
  }),
};
