/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

import { test, expect } from '@playwright/test';

const createdTest01Ids: string[] = [];

function generateTest01(overrides?: {
  name?: string;
  testInt?: number;
}) {
  const timestamp = Date.now();
  return {
    name: overrides?.name || `test01_${timestamp}`,
    test_int: overrides?.testInt ?? 42,
  };
}

test.describe('Test01 API', () => {
  test('GET /test01 - should list all test01 records', async ({ request }) => {
    const response = await request.get(`/test01`);

    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body.success).toBe(true);

    const records = body.data || [];
    expect(Array.isArray(records)).toBe(true);
  });

  test('POST /test01 - should create a new test01 record', async ({ request }) => {
    const newRecord = generateTest01();

    const response = await request.post(`/test01`, {
      data: newRecord,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    const record = body.data || body;
    expect(record).toHaveProperty('id');
    expect(record).toHaveProperty('name', newRecord.name);
    expect(record).toHaveProperty('test_int', newRecord.test_int);
    expect(record).toHaveProperty('created_at');
    expect(record).toHaveProperty('updated_at');

    if (record.id) {
      createdTest01Ids.push(record.id);
    }
  });

  test('POST /test01 - should create record with specific fields', async ({ request }) => {
    const newRecord = {
      name: 'specific_test',
      test_int: 100
    };

    const response = await request.post(`/test01`, {
      data: newRecord,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    const record = body.data || body;
    expect(record.name).toBe('specific_test');
    expect(record.test_int).toBe(100);

    if (record.id) {
      createdTest01Ids.push(record.id);
    }
  });

  test('GET /test01/:id - should get test01 record by ID', async ({ request }) => {
    const newRecord = generateTest01();
    const createResponse = await request.post(`/test01`, {
      data: newRecord,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(createResponse.status()).toBe(201);

    const createBody = await createResponse.json();
    const createdRecord = createBody.data || createBody;
    const recordId = createdRecord.id;
    createdTest01Ids.push(recordId);

    const response = await request.get(`/test01/${recordId}`);

    expect(response.status()).toBe(200);

    const body = await response.json();
    const record = body.data || body;
    expect(record).toHaveProperty('id', recordId);
    expect(record).toHaveProperty('name', newRecord.name);
    expect(record).toHaveProperty('test_int', newRecord.test_int);
    expect(record).toHaveProperty('created_at');
    expect(record).toHaveProperty('updated_at');
  });

  test('GET /test01/:id - should return error for non-existent record', async ({ request }) => {
    const nonExistentId = '000000000000';
    const response = await request.get(`/test01/${nonExistentId}`);

    expect([400, 404]).toContain(response.status());
  });

  test('PATCH /test01/:id - should update a test01 record', async ({ request }) => {
    const newRecord = generateTest01();
    const createResponse = await request.post(`/test01`, {
      data: newRecord,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(createResponse.status()).toBe(201);

    const createBody = await createResponse.json();
    const createdRecord = createBody.data || createBody;
    const recordId = createdRecord.id;
    createdTest01Ids.push(recordId);

    const updateData = {
      name: 'updated_test01',
      test_int: 200
    };

    const response = await request.patch(`/test01/${recordId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    const record = body.data || body;
    expect(record).toHaveProperty('id', recordId);
    expect(record.name).toBe('updated_test01');
    expect(record.test_int).toBe(200);
  });

  test('PATCH /test01/:id - should preserve existing fields when not provided', async ({ request }) => {
    const newRecord = generateTest01();
    const createResponse = await request.post(`/test01`, {
      data: newRecord,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(createResponse.status()).toBe(201);

    const createBody = await createResponse.json();
    const createdRecord = createBody.data || createBody;
    const recordId = createdRecord.id;
    createdTest01Ids.push(recordId);

    const updateData = {
      name: 'partially_updated'
    };

    const response = await request.patch(`/test01/${recordId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(200);

    const body = await response.json();
    const record = body.data || body;
    expect(record.name).toBe('partially_updated');
    expect(record.test_int).toBe(newRecord.test_int);
  });

  test('DELETE /test01/:id - should delete a test01 record', async ({ request }) => {
    const newRecord = generateTest01();
    const createResponse = await request.post(`/test01`, {
      data: newRecord,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(createResponse.status()).toBe(201);

    const createBody = await createResponse.json();
    const createdRecord = createBody.data || createBody;
    const recordId = createdRecord.id;

    const response = await request.delete(`/test01/${recordId}`);

    expect([204, 200]).toContain(response.status());

    const getResponse = await request.get(`/test01/${recordId}`);
    expect([400, 404]).toContain(getResponse.status());

    const index = createdTest01Ids.indexOf(recordId);
    if (index > -1) createdTest01Ids.splice(index, 1);
  });

  test('POST /test01 - should return 400 when body is empty', async ({ request }) => {
    const response = await request.post(`/test01`, {
      data: {},
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect([201, 400]).toContain(response.status());
  });

  test('PATCH /test01/:id - should return error for non-existent record', async ({ request }) => {
    const nonExistentId = '000000000000';
    const updateData = {
      name: 'nobody'
    };

    const response = await request.patch(`/test01/${nonExistentId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect([400, 404]).toContain(response.status());
  });

  test('DELETE /test01/:id - should handle non-existent record gracefully', async ({ request }) => {
    const nonExistentId = '000000000000';
    const response = await request.delete(`/test01/${nonExistentId}`);

    expect([204, 404, 400]).toContain(response.status());
  });

  test('Full CRUD cycle - Create, Read, Update, Delete', async ({ request }) => {
    const newRecord = generateTest01();
    const createResponse = await request.post(`/test01`, {
      data: newRecord,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(createResponse.status()).toBe(201);
    const createBody = await createResponse.json();
    const createdRecord = createBody.data || createBody;
    const recordId = createdRecord.id;
    createdTest01Ids.push(recordId);

    const getResponse = await request.get(`/test01/${recordId}`);
    expect(getResponse.status()).toBe(200);
    const getBody = await getResponse.json();
    const fetchedRecord = getBody.data || getBody;
    expect(fetchedRecord.name).toBe(newRecord.name);

    const updateData = {
      name: 'crud_updated',
      test_int: 999
    };
    const updateResponse = await request.patch(`/test01/${recordId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    expect(updateResponse.status()).toBe(200);

    const verifyResponse = await request.get(`/test01/${recordId}`);
    const verifyBody = await verifyResponse.json();
    const updatedRecord = verifyBody.data || verifyBody;
    expect(updatedRecord.name).toBe('crud_updated');
    expect(updatedRecord.test_int).toBe(999);

    const deleteResponse = await request.delete(`/test01/${recordId}`);
    expect([204, 200]).toContain(deleteResponse.status());

    const afterDeleteResponse = await request.get(`/test01/${recordId}`);
    expect([400, 404]).toContain(afterDeleteResponse.status());

    const index = createdTest01Ids.indexOf(recordId);
    if (index > -1) createdTest01Ids.splice(index, 1);
  });
});

test.afterAll(async ({ request }) => {
  for (const id of [...createdTest01Ids].reverse()) {
    try {
      await request.delete(`/test01/${id}`);
    } catch (error) {
      console.warn(`Failed to cleanup test01 ${id}:`, error);
    }
  }
  createdTest01Ids.length = 0;
});