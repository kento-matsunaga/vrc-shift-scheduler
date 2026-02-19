const fs = require('fs');
const path = require('path');

const BASE_URL = 'https://vrc-shift-scheduler.com';

const pages = [
  { url: '/', priority: '1.0', changefreq: 'weekly' },
  { url: '/subscribe', priority: '0.8', changefreq: 'monthly' },
  { url: '/terms', priority: '0.3', changefreq: 'yearly' },
  { url: '/privacy', priority: '0.3', changefreq: 'yearly' },
];

const today = new Date().toISOString().split('T')[0];

const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${pages.map(page => `  <url>
    <loc>${BASE_URL}${page.url}</loc>
    <lastmod>${today}</lastmod>
    <changefreq>${page.changefreq}</changefreq>
    <priority>${page.priority}</priority>
  </url>`).join('\n')}
</urlset>`;

const outputPath = path.join(__dirname, '../public/sitemap.xml');
fs.writeFileSync(outputPath, sitemap);
console.log('sitemap.xml generated:', outputPath);
