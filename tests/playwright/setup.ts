/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
import { request, FullConfig } from '@playwright/test';

export default async function globalSetup(config: FullConfig) {
  const baseURL = config.projects[0].use.baseURL || 'http://localhost:3000';
  console.log(`🔍 Testing connection to ${baseURL}...`);
  
  try {
    const context = await request.newContext({ baseURL });
    const response = await context.get('/users', { timeout: 5000 });
    
    if (response.status() < 500) {
      console.log(`✅ Server is reachable at ${baseURL} (status: ${response.status()})`);
      return;
    }
    
    throw new Error(`❌ Server returned error status: ${response.status()}`);
  } catch (error) {
    console.error(`❌ Cannot connect to server at ${baseURL}`);
    console.error(`   Error: ${error instanceof Error ? error.message : error}`);
    throw error;
  }
}
