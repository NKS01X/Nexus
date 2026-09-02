import puppeteer from 'puppeteer-core';
import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';

const CHROMIUM_PATH = process.env.CHROMIUM_PATH || '/snap/bin/chromium';
const BASE_URL = process.env.PORTAL_URL || 'http://localhost:8084';
const FRAMES_DIR = '/tmp/nexus_frames';

if (!fs.existsSync(FRAMES_DIR)) {
  fs.mkdirSync(FRAMES_DIR, { recursive: true });
} else {
  fs.readdirSync(FRAMES_DIR).forEach(f => fs.unlinkSync(path.join(FRAMES_DIR, f)));
}

async function capture() {
  console.log('Launching browser...');
  const browser = await puppeteer.launch({
    executablePath: CHROMIUM_PATH,
    args: [
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--enable-unsafe-swiftshader',
      '--window-size=1440,900'
    ],
    defaultViewport: { width: 1440, height: 900 }
  });

  const page = await browser.newPage();
  let frameIdx = 0;

  async function takeFrames(count = 1, delay = 500) {
    for (let i = 0; i < count; i++) {
      const num = String(frameIdx++).padStart(4, '0');
      const filename = path.join(FRAMES_DIR, `frame_${num}.png`);
      await page.screenshot({ path: filename, type: 'png' });
      if (delay > 0) await new Promise(r => setTimeout(r, delay));
    }
  }

  // 1. Landing Page
  console.log('Capturing Landing Page...');
  await page.goto(`${BASE_URL}/`, { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 2000));
  await takeFrames(4, 800);

  // Scroll Landing Page
  await page.evaluate(() => window.scrollBy({ top: 500, behavior: 'smooth' }));
  await new Promise(r => setTimeout(r, 1000));
  await takeFrames(3, 600);

  await page.evaluate(() => window.scrollBy({ top: 700, behavior: 'smooth' }));
  await new Promise(r => setTimeout(r, 1000));
  await takeFrames(3, 600);

  await page.evaluate(() => window.scrollBy({ top: 900, behavior: 'smooth' }));
  await new Promise(r => setTimeout(r, 1000));
  await takeFrames(3, 600);

  // 2. Stores Page
  console.log('Capturing Stores Page...');
  await page.goto(`${BASE_URL}/merchants`, { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 1500));
  await takeFrames(4, 600);

  // 3. Review Queue Page
  console.log('Capturing Review Queue Page...');
  await page.goto(`${BASE_URL}/approvals`, { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 1500));
  await takeFrames(4, 600);

  // 4. Test Lab Page
  console.log('Capturing Test Lab Page...');
  await page.goto(`${BASE_URL}/redteam`, { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 1500));
  await takeFrames(4, 600);

  // 5. AI Checkout Demo Page
  console.log('Capturing AI Checkout Demo Page...');
  await page.goto(`${BASE_URL}/ai-purchase`, { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 1500));
  await takeFrames(4, 600);

  // 6. Aegis 3D Demo Animation Page
  console.log('Capturing 3D Aegis Demo Page...');
  await page.goto(`${BASE_URL}/aegis-demo`, { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 1500));
  await takeFrames(6, 1000);

  // Return home
  await page.goto(`${BASE_URL}/`, { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 1000));
  await page.evaluate(() => window.scrollTo(0, 0));
  await takeFrames(2, 500);

  await browser.close();
  console.log(`Captured ${frameIdx} frames.`);

  // Assemble frames into animated WebP using convert
  console.log('Creating animated WebP image assets/demo.webp...');
  const outputWebp = path.join(process.cwd(), 'assets/demo.webp');
  execSync(`convert -delay 70 -loop 0 ${FRAMES_DIR}/frame_*.png ${outputWebp}`);
  console.log('Successfully created assets/demo.webp!');
}

capture().catch(err => {
  console.error('Error recording demo:', err);
  process.exit(1);
});
