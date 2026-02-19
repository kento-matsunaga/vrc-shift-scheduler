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

/**
 * Union type for all JSON-LD schemas
 */
export type JsonLdSchema =
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
 *
 * XSS対策: HTMLタグ文字をUnicodeエスケープ
 */
export function JsonLd({ data }: JsonLdProps) {
  const jsonLdString = JSON.stringify(data)
    .replace(/</g, '\\u003c')
    .replace(/>/g, '\\u003e');

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: jsonLdString }}
    />
  );
}
