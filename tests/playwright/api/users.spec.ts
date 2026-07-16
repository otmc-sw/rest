/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/

import { test, expect, request } from '@playwright/test';

const createdUserIds: number[] = [];

function generateTestUser(overrides?: {
  username?: string;
  email?: string;
  fullName?: string;
  enabled?: boolean;
  testInt?: number;
  content?: object;
}) {
  const timestamp = Date.now();
  return {
    username: overrides?.username || `testuser_${timestamp}`,
    email: overrides?.email || `test_${timestamp}@example.com`,
    full_name: overrides?.fullName || `Test User ${timestamp}`,
    enabled: overrides?.enabled ?? true,
    test_int: overrides?.testInt ?? 42,
    content: overrides?.content || {
      preferences: {
        theme: 'dark',
        notifications: true
      },
      metadata: {
        source: 'playwright_test'
      }
    }
  };
}

test.describe('Users API', () => {
  test('GET /users - should list all users', async ({ request }) => {
    const response = await request.get(`/users`);
    
    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body.success).toBe(true);
    
    const users = body.data || [];
    expect(Array.isArray(users)).toBe(true);
  });

  test('POST /users - should create a new user', async ({ request }) => {
    const newUser = generateTestUser();

    const response = await request.post(`/users`, {
      data: newUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(201);
    
    const body = await response.json();
    const user = body.data || body;
    expect(user).toHaveProperty('id');
    expect(user).toHaveProperty('username');
    expect(user).toHaveProperty('email');
    expect(user).toHaveProperty('full_name');
    expect(user).toHaveProperty('enabled');
    expect(user).toHaveProperty('test_int');
    expect(user).toHaveProperty('created_at');
    expect(user).toHaveProperty('updated_at');
    
    if (user.id) {
      createdUserIds.push(user.id);
    }
  });

  test('POST /users - should create user with all fields', async ({ request }) => {
    const newUser = {
      username: 'complete_user',
      email: 'complete@example.com',
      full_name: 'Complete Test User',
      enabled: true,
      test_int: 100,
      content: {
        address: '123 Test St',
        phone: '123-456-7890',
        preferences: {
          newsletter: true,
          notifications: false
        }
      }
    };

    const response = await request.post(`/users`, {
      data: newUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(201);
    
    const body = await response.json();
    const user = body.data || body;
    expect(user.username).toBe('complete_user');
    expect(user.email).toBe('complete@example.com');
    expect(user.full_name).toBe('Complete Test User');
    
    if (user.id) {
      createdUserIds.push(user.id);
    }
  });

  test('GET /users/:id - should get user by ID', async ({ request }) => {
    const newUser = generateTestUser();
    const createResponse = await request.post(`/users`, {
      data: newUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);
    
    const createBody = await createResponse.json();
    const createdUser = createBody.data || createBody;
    const userId = createdUser.id;
    createdUserIds.push(userId);

    const response = await request.get(`/users/${userId}`);
    
    expect(response.status()).toBe(200);
    
    const body = await response.json();
    const user = body.data || body;
    expect(user).toHaveProperty('id', userId);
    expect(user).toHaveProperty('username', newUser.username);
    expect(user).toHaveProperty('email', newUser.email);
    expect(user).toHaveProperty('full_name', newUser.full_name);
    expect(user).toHaveProperty('enabled');
    expect(user).toHaveProperty('test_int');
    expect(user).toHaveProperty('created_at');
    expect(user).toHaveProperty('updated_at');
  });

  test('GET /users/:id - should return error for non-existent user', async ({ request }) => {
    const nonExistentId = 999999;
    const response = await request.get(`/users/${nonExistentId}`);
    
    expect([400, 404]).toContain(response.status());
  });

  test('PATCH /users/:id - should update a user', async ({ request }) => {
    const newUser = generateTestUser();
    const createResponse = await request.post(`/users`, {
      data: newUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);

    const createBody = await createResponse.json();
    const createdUser = createBody.data || createBody;
    const userId = createdUser.id;
    createdUserIds.push(userId);

    const updateData = {
      email: 'updated@example.com',
      full_name: 'Updated Name'
    };

    const response = await request.patch(`/users/${userId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(200);
    
    const body = await response.json();
    const user = body.data || body;
    expect(user).toHaveProperty('id', userId);
    expect(user.email).toBe('updated@example.com');
    expect(user.full_name).toBe('Updated Name');
  });

  test('PATCH /users/:id - should preserve existing fields when not provided', async ({ request }) => {
    const newUser = generateTestUser();
    const createResponse = await request.post(`/users`, {
      data: newUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);

    const createBody = await createResponse.json();
    const createdUser = createBody.data || createBody;
    const userId = createdUser.id;
    createdUserIds.push(userId);

    const updateData = {
      full_name: 'Partially Updated'
    };

    const response = await request.patch(`/users/${userId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(200);
    
    const body = await response.json();
    const user = body.data || body;
    expect(user.username).toBe(newUser.username);
    expect(user.email).toBe(newUser.email);
  });

  test('DELETE /users/:id - should delete a user', async ({ request }) => {
    const newUser = generateTestUser();
    const createResponse = await request.post(`/users`, {
      data: newUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);
    
    const createBody = await createResponse.json();
    const createdUser = createBody.data || createBody;
    const userId = createdUser.id;

    const response = await request.delete(`/users/${userId}`);
    
    expect([204, 200]).toContain(response.status());

    const getResponse = await request.get(`/users/${userId}`);
    expect([400, 404]).toContain(getResponse.status());
    
    const index = createdUserIds.indexOf(userId);
    if (index > -1) createdUserIds.splice(index, 1);
  });

  test('POST /users - should return 400 when username is missing', async ({ request }) => {
    const invalidUser = {
      email: 'test@example.com'
    };

    const response = await request.post(`/users`, {
      data: invalidUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(400);
  });

  test('POST /users - should return 400 when email is invalid', async ({ request }) => {
    const invalidUser = {
      username: 'testuser',
      email: 'not-an-email'
    };

    const response = await request.post(`/users`, {
      data: invalidUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(400);
  });

  test('POST /users - should reject duplicate username', async ({ request }) => {
    const username = 'duplicate_user_test';
    const user1 = {
      username: username,
      email: 'first@example.com'
    };

    const response1 = await request.post(`/users`, {
      data: user1,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response1.status()).toBe(201);
    
    const body1 = await response1.json();
    const createdUser1 = body1.data || body1;
    if (createdUser1.id) {
      createdUserIds.push(createdUser1.id);
    }

    const user2 = {
      username: username,
      email: 'second@example.com'
    };

    const response2 = await request.post(`/users`, {
      data: user2,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response2.status()).not.toBe(201);
  });

  test('POST /users - should reject duplicate email', async ({ request }) => {
    const email = 'duplicate_email@example.com';
    const user1 = {
      username: 'user1',
      email: email
    };

    const response1 = await request.post(`/users`, {
      data: user1,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response1.status()).toBe(201);
    
    const body1 = await response1.json();
    const createdUser1 = body1.data || body1;
    if (createdUser1.id) {
      createdUserIds.push(createdUser1.id);
    }

    const user2 = {
      username: 'user2',
      email: email
    };

    const response2 = await request.post(`/users`, {
      data: user2,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response2.status()).not.toBe(201);
  });

  test('PATCH /users/:id - should return error for non-existent user', async ({ request }) => {
    const nonExistentId = 999999;
    const updateData = {
      email: 'updated@example.com'
    };

    const response = await request.patch(`/users/${nonExistentId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect([400, 404]).toContain(response.status());
  });

  test('DELETE /users/:id - should return success for non-existent user', async ({ request }) => {
    const nonExistentId = 999999;
    const response = await request.delete(`/users/${nonExistentId}`);
    
    expect([204, 404, 400]).toContain(response.status());
  });

  test('POST /users - should create user with JSON content', async ({ request }) => {
    const userWithContent = {
      username: 'content_user',
      email: 'content@example.com',
      full_name: 'Content User',
      enabled: false,
      test_int: 999,
      content: {
        settings: {
          theme: 'light',
          language: 'en'
        },
        tags: ['developer', 'tester'],
        nested: {
          deep: {
            value: 42
          }
        }
      }
    };

    const response = await request.post(`/users`, {
      data: userWithContent,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(201);
    
    const body = await response.json();
    const user = body.data || body;
    expect(user).toHaveProperty('username', 'content_user');
    expect(user).toHaveProperty('content');
    
    if (user.id) {
      createdUserIds.push(user.id);
    }
  });

  test('Full CRUD cycle - Create, Read, Update, Delete', async ({ request }) => {
    const newUser = generateTestUser();
    const createResponse = await request.post(`/users`, {
      data: newUser,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(createResponse.status()).toBe(201);
    const createBody = await createResponse.json();
    const createdUser = createBody.data || createBody;
    const userId = createdUser.id;
    createdUserIds.push(userId);

    const getResponse = await request.get(`/users/${userId}`);
    expect(getResponse.status()).toBe(200);
    const getBody = await getResponse.json();
    const fetchedUser = getBody.data || getBody;
    expect(fetchedUser.username).toBe(newUser.username);

    const updateData = {
      full_name: 'CRUD Updated Name'
    };
    const updateResponse = await request.patch(`/users/${userId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    expect(updateResponse.status()).toBe(200);

    const verifyResponse = await request.get(`/users/${userId}`);
    const verifyBody = await verifyResponse.json();
    const updatedUser = verifyBody.data || verifyBody;
    expect(updatedUser.full_name).toBe('CRUD Updated Name');

    const deleteResponse = await request.delete(`/users/${userId}`);
    expect([204, 200]).toContain(deleteResponse.status());

    const afterDeleteResponse = await request.get(`/users/${userId}`);
    expect([400, 404]).toContain(afterDeleteResponse.status());
    
    const index = createdUserIds.indexOf(userId);
    if (index > -1) createdUserIds.splice(index, 1);
  });
});

test.afterAll(async ({ request }) => {
  for (const userId of [...createdUserIds].reverse()) {
    try {
      await request.delete(`/users/${userId}`);
    } catch (error) {
      console.warn(`Failed to cleanup user ${userId}:`, error);
    }
  }
  createdUserIds.length = 0;
});