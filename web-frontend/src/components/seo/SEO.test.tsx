import { render, cleanup } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { SEO } from './SEO';

describe('SEO Component', () => {
  afterEach(() => {
    cleanup();
    // head 内のメタタグをクリーンアップ
    document.head.innerHTML = '';
  });

  it('should render noindex meta tag when noindex=true', () => {
    render(<SEO noindex={true} />);
    const robotsMeta = document.querySelector('meta[name="robots"]');
    expect(robotsMeta).not.toBeNull();
    expect(robotsMeta?.getAttribute('content')).toBe('noindex, nofollow');
  });

  it('should not render noindex meta tag when noindex=false', () => {
    render(<SEO noindex={false} />);
    const robotsMeta = document.querySelector('meta[name="robots"]');
    expect(robotsMeta).toBeNull();
  });

  it('should render JSON-LD when softwareApplication is true', () => {
    render(<SEO jsonLd={{ softwareApplication: true }} />);
    const jsonLdScript = document.querySelector('script[type="application/ld+json"]');
    expect(jsonLdScript).not.toBeNull();
  });

  it('should set title correctly', () => {
    render(<SEO title="テストタイトル" />);
    expect(document.title).toContain('テストタイトル');
  });

  it('should set description meta tag correctly', () => {
    render(<SEO description="テスト説明文" />);
    const descMeta = document.querySelector('meta[name="description"]');
    expect(descMeta?.getAttribute('content')).toBe('テスト説明文');
  });
});
