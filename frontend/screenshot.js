import puppeteer from 'puppeteer';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const baseUrl = 'http://localhost:3000';
const pages = [
  { name: '首页', url: '/' },
  { name: '命名空间列表', url: '/namespaces' },
  { name: '部署向导', url: '/deploy' },
  { name: '监控面板', url: '/monitor' }
];

async function takeScreenshots() {
  console.log('🚀 启动无头浏览器...');
  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1920, height: 1080 });

  const outputDir = path.join(__dirname, '../screenshots');
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  for (const p of pages) {
    try {
      console.log(`📸 正在截取：${p.name} (${baseUrl}${p.url})...`);
      await page.goto(`${baseUrl}${p.url}`, { waitUntil: 'networkidle2', timeout: 15000 });
      
      // 等待 Vue 组件渲染
      await page.waitForSelector('body', { timeout: 5000 });
      
      const screenshotPath = path.join(outputDir, `${p.name.replace(/\s+/g, '_')}.png`);
      await page.screenshot({ path: screenshotPath, fullPage: true });
      console.log(`✅ 截图已保存：${screenshotPath}`);
    } catch (err) {
      console.error(`❌ 截取 ${p.name} 失败:`, err.message);
    }
  }

  await browser.close();
  console.log('🎉 所有截图完成！');
}

takeScreenshots().catch(console.error);
