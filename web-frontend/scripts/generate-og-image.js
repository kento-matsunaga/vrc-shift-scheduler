/**
 * OG Image Generator Script
 * Generates a 1200x630 PNG image for social media sharing
 *
 * Run: node scripts/generate-og-image.js
 */

import { createCanvas, registerFont } from 'canvas';
import { writeFileSync, mkdirSync, existsSync } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Register Japanese font
const fontPath = join(__dirname, 'fonts', 'NotoSansJP-Regular.otf');
if (existsSync(fontPath)) {
  registerFont(fontPath, { family: 'Noto Sans JP' });
}

// Image dimensions (standard OG image size)
const WIDTH = 1200;
const HEIGHT = 630;

// Create canvas
const canvas = createCanvas(WIDTH, HEIGHT);
const ctx = canvas.getContext('2d');

// Background gradient
const gradient = ctx.createLinearGradient(0, 0, 0, HEIGHT);
gradient.addColorStop(0, '#0a0a0f');
gradient.addColorStop(0.5, '#0f0f1a');
gradient.addColorStop(1, '#0a0a0f');
ctx.fillStyle = gradient;
ctx.fillRect(0, 0, WIDTH, HEIGHT);

// Add subtle grid pattern
ctx.strokeStyle = 'rgba(255, 255, 255, 0.03)';
ctx.lineWidth = 1;
const gridSize = 60;
for (let x = 0; x <= WIDTH; x += gridSize) {
  ctx.beginPath();
  ctx.moveTo(x, 0);
  ctx.lineTo(x, HEIGHT);
  ctx.stroke();
}
for (let y = 0; y <= HEIGHT; y += gridSize) {
  ctx.beginPath();
  ctx.moveTo(0, y);
  ctx.lineTo(WIDTH, y);
  ctx.stroke();
}

// Add gradient orb effect (top right)
const orbGradient1 = ctx.createRadialGradient(WIDTH - 100, -100, 0, WIDTH - 100, -100, 400);
orbGradient1.addColorStop(0, 'rgba(79, 70, 229, 0.15)');
orbGradient1.addColorStop(1, 'transparent');
ctx.fillStyle = orbGradient1;
ctx.fillRect(0, 0, WIDTH, HEIGHT);

// Add gradient orb effect (bottom left)
const orbGradient2 = ctx.createRadialGradient(100, HEIGHT + 50, 0, 100, HEIGHT + 50, 300);
orbGradient2.addColorStop(0, 'rgba(139, 92, 246, 0.12)');
orbGradient2.addColorStop(1, 'transparent');
ctx.fillStyle = orbGradient2;
ctx.fillRect(0, 0, WIDTH, HEIGHT);

// Icon background (rounded rectangle simulation)
const iconX = WIDTH / 2 - 40;
const iconY = 180;
const iconSize = 80;
const iconGradient = ctx.createLinearGradient(iconX, iconY, iconX + iconSize, iconY + iconSize);
iconGradient.addColorStop(0, '#4F46E5');
iconGradient.addColorStop(1, '#8B5CF6');

// Draw rounded rectangle for icon
ctx.fillStyle = iconGradient;
const radius = 16;
ctx.beginPath();
ctx.moveTo(iconX + radius, iconY);
ctx.lineTo(iconX + iconSize - radius, iconY);
ctx.quadraticCurveTo(iconX + iconSize, iconY, iconX + iconSize, iconY + radius);
ctx.lineTo(iconX + iconSize, iconY + iconSize - radius);
ctx.quadraticCurveTo(iconX + iconSize, iconY + iconSize, iconX + iconSize - radius, iconY + iconSize);
ctx.lineTo(iconX + radius, iconY + iconSize);
ctx.quadraticCurveTo(iconX, iconY + iconSize, iconX, iconY + iconSize - radius);
ctx.lineTo(iconX, iconY + radius);
ctx.quadraticCurveTo(iconX, iconY, iconX + radius, iconY);
ctx.closePath();
ctx.fill();

// Calendar icon (drawn with shapes)
ctx.fillStyle = '#ffffff';
const calX = iconX + 18;
const calY = iconY + 18;
const calW = 44;
const calH = 44;

// Calendar body
ctx.fillRect(calX, calY + 8, calW, calH - 8);

// Calendar header
ctx.fillStyle = '#ffffff';
ctx.fillRect(calX, calY + 8, calW, 12);

// Calendar rings
ctx.fillStyle = '#8B5CF6';
ctx.fillRect(calX + 10, calY, 4, 14);
ctx.fillRect(calX + 30, calY, 4, 14);

// Calendar grid lines (horizontal)
ctx.fillStyle = '#4F46E5';
ctx.fillRect(calX, calY + 24, calW, 1);
ctx.fillRect(calX, calY + 34, calW, 1);

// Calendar grid lines (vertical)
ctx.fillRect(calX + 14, calY + 20, 1, calH - 12);
ctx.fillRect(calX + 29, calY + 20, 1, calH - 12);

// Title
ctx.fillStyle = '#ffffff';
ctx.font = 'bold 56px "Noto Sans JP", sans-serif';
ctx.textAlign = 'center';
ctx.fillText('VRC Shift Scheduler', WIDTH / 2, 340);

// Tagline (Japanese)
ctx.fillStyle = '#a0a0b0';
ctx.font = '28px "Noto Sans JP", sans-serif';
ctx.fillText('VRChatイベントのシフト管理を、もっと簡単に。', WIDTH / 2, 400);

// Price badge
const badgeWidth = 200;
const badgeHeight = 40;
const badgeX = WIDTH / 2 - badgeWidth / 2;
const badgeY = 450;

ctx.fillStyle = 'rgba(139, 92, 246, 0.15)';
ctx.strokeStyle = 'rgba(139, 92, 246, 0.3)';
ctx.lineWidth = 1;

// Badge rounded rectangle
ctx.beginPath();
ctx.moveTo(badgeX + 20, badgeY);
ctx.lineTo(badgeX + badgeWidth - 20, badgeY);
ctx.quadraticCurveTo(badgeX + badgeWidth, badgeY, badgeX + badgeWidth, badgeY + 20);
ctx.lineTo(badgeX + badgeWidth, badgeY + badgeHeight - 20);
ctx.quadraticCurveTo(badgeX + badgeWidth, badgeY + badgeHeight, badgeX + badgeWidth - 20, badgeY + badgeHeight);
ctx.lineTo(badgeX + 20, badgeY + badgeHeight);
ctx.quadraticCurveTo(badgeX, badgeY + badgeHeight, badgeX, badgeY + badgeHeight - 20);
ctx.lineTo(badgeX, badgeY + 20);
ctx.quadraticCurveTo(badgeX, badgeY, badgeX + 20, badgeY);
ctx.closePath();
ctx.fill();
ctx.stroke();

// Badge text (Japanese)
ctx.fillStyle = '#c4b5fd';
ctx.font = '20px "Noto Sans JP", sans-serif';
ctx.fillText('月額200円から', WIDTH / 2, badgeY + 27);

// URL at bottom
ctx.fillStyle = '#6b7280';
ctx.font = '18px sans-serif';
ctx.fillText('vrcshift.com', WIDTH / 2, HEIGHT - 40);

// Output directory
const outputDir = join(__dirname, '..', 'public', 'images', 'og');
if (!existsSync(outputDir)) {
  mkdirSync(outputDir, { recursive: true });
}

// Save as PNG
const outputPath = join(outputDir, 'default.png');
const buffer = canvas.toBuffer('image/png');
writeFileSync(outputPath, buffer);

console.log(`OG image generated: ${outputPath}`);
console.log(`Dimensions: ${WIDTH}x${HEIGHT}px`);
