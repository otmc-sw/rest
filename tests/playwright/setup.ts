/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
 * 
 * Global setup: wait for backend server to be ready before running tests.
 */
import { request, FullConfig } from '@playwright/test';

const MAX_RETRIES = 5;
const RETRY_INTERVAL_MS = 2000;

async function waitForServer(baseURL: string): Promise<void> {
  const context = await request.newContext({ baseURL });
  
  for (let attempt = 1; attempt <= MAX_RETRIES; attempt++) {
    try {
      const response = await context.get('/users', { timeout: 5000 });
      if (response.status() === 200) {
        console.log(`✅ Backend server is ready at ${baseURL} (attempt ${attempt})`);
        return;
      }
      console.log(`⚠️  Backend returned status ${response.status()}, retrying... (${attempt}/${MAX_RETRIES})`);
    } catch (error) {
      console.log(`⏳ Waiting for backend at ${baseURL}... (${attempt}/${MAX_RETRIES})`);
    }
    
    if (attempt < MAX_RETRIES) {
      await new Promise(resolve => setTimeout(resolve, RETRY_INTERVAL_MS));
    }
  }
  
  throw new Error(`❌ Backend server at ${baseURL} is not available after ${MAX_RETRIES} retries`);
}

export default async function globalSetup(config: FullConfig) {
  const baseURL = config.projects[0].use.baseURL || 'http://localhost:5006';
  console.log(`🔍 Checking backend server availability at ${baseURL}...`);
  await waitForServer(baseURL);
  console.log('🚀 All systems ready. Starting tests...');
}