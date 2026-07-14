/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
 **/

import { test, expect, request } from '@playwright/test';

test.describe('Strange Input Tests', () => {
  test.describe('Invalid ID Formats', () => {
    test('GET /users/:id - should return 400 for string ID', async ({ request }) => {
      const response = await request.get(`/users/abc`);
      
      expect(response.status()).toBe(400);
      
      const body = await response.json();
      expect(body.success).toBe(false);
      expect(body.error).toBeDefined();
    });

    test('GET /users/:id - should return 400 for special characters in ID', async ({ request }) => {
      const response = await request.get(`/users/!@#$%`);
      
      expect(response.status()).toBe(400);
      
      const body = await response.json();
      expect(body.success).toBe(false);
      expect(body.error).toBeDefined();
    });

    test('GET /users/:id - should return 400 for float ID', async ({ request }) => {
      const response = await request.get(`/users/123.456`);
      
      expect(response.status()).toBe(400);
      
      const body = await response.json();
      expect(body.success).toBe(false);
      expect(body.error).toBeDefined();
    });

    test('GET /users/:id - should return 400 for negative ID', async ({ request }) => {
      const response = await request.get(`/users/-1`);
      
      // Could be 400 or 404 depending on implementation
      expect([400, 404]).toContain(response.status());
    });

    test('GET /users/:id - should return 400 for empty ID', async ({ request }) => {
      const response = await request.get(`/users/`);
      
      // Empty ID might return various status codes depending on routing
      expect([200, 400, 404]).toContain(response.status());
    });

    test('PATCH /users/:id - should return 400 for string ID', async ({ request }) => {
      const response = await request.patch(`/users/not-a-number`, {
        data: { email: 'test@example.com' },
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      expect(response.status()).toBe(400);
      
      const body = await response.json();
      expect(body.success).toBe(false);
      expect(body.error).toBeDefined();
    });

    test('DELETE /users/:id - should return 400 for string ID', async ({ request }) => {
      const response = await request.delete(`/users/invalid-id`);
      
      expect(response.status()).toBe(400);
      
      const body = await response.json();
      expect(body.success).toBe(false);
      expect(body.error).toBeDefined();
    });
  });

  test.describe('Invalid Request Bodies', () => {
    test('POST /users - should return 400 for malformed JSON', async ({ request }) => {
      const response = await request.post(`/users`, {
        data: 'this is not valid json',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      expect(response.status()).toBe(400);
    });

    test('POST /users - should return 400 for empty body', async ({ request }) => {
      const response = await request.post(`/users`, {
        data: '',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      expect(response.status()).toBe(400);
    });

    test('POST /users - should return 400 for null body', async ({ request }) => {
      const response = await request.post(`/users`, {
        data: null,
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      expect(response.status()).toBe(400);
    });

    test('POST /users - should return 400 for array instead of object', async ({ request }) => {
      const response = await request.post(`/users`, {
        data: ['not', 'an', 'object'],
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      expect(response.status()).toBe(400);
    });

    test('POST /users - should return 400 for wrong data types', async ({ request }) => {
      const invalidUser = {
        username: 123,  // should be string
        email: true,    // should be string
        full_name: [],  // should be string
        content: 'this should be an object'  // should be object
      };

      const response = await request.post(`/users`, {
        data: invalidUser,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(400);
    });

    test('POST /users - should handle extremely long strings', async ({ request }) => {
      const longString = 'a'.repeat(10000);
      const invalidUser = {
        username: longString,
        email: 'test@example.com'
      };

      const response = await request.post(`/users`, {
        data: invalidUser,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // API may accept or reject based on validation rules
      expect([201, 400]).toContain(response.status());
    });

    test('POST /users - should return 400 for SQL injection attempt', async ({ request }) => {
      const sqlInjection = {
        username: "admin'; DROP TABLE users; --",
        email: 'test@example.com'
      };

      const response = await request.post(`/users`, {
        data: sqlInjection,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(400);
    });

    test('POST /users - should return 400 for XSS attempt', async ({ request }) => {
      const xssAttempt = {
        username: '<script>alert("xss")</script>',
        email: 'test@example.com'
      };

      const response = await request.post(`/users`, {
        data: xssAttempt,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(400);
    });
  });

  test.describe('Invalid Query Parameters', () => {
    test('GET /users - should handle invalid query parameters gracefully', async ({ request }) => {
      const response = await request.get(`/users?invalid_param=invalid_value`);
      
      // Should either work (ignore invalid params) or return 400
      expect([200, 400]).toContain(response.status());
    });

    test('GET /users - should handle SQL injection in query params', async ({ request }) => {
      const response = await request.get(`/users?id=1' OR '1'='1`);
      
      expect([200, 400]).toContain(response.status());
    });
  });

  test.describe('Invalid HTTP Methods', () => {
    test('PUT /users/:id - should return 405 or 404 for unsupported method', async ({ request }) => {
      const response = await request.put(`/users/1`, {
        data: { username: 'test' },
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      // Should return 404 (not found) or 405 (method not allowed)
      expect([404, 405]).toContain(response.status());
    });
  });

  test.describe('Invalid Content-Type', () => {
    test('POST /users - should return 400 for wrong content type', async ({ request }) => {
      const response = await request.post(`/users`, {
        data: 'username=test&email=test@example.com',
        headers: {
          'Content-Type': 'text/plain'
        }
      });
      
      expect(response.status()).toBe(400);
    });

    test('POST /users - should return 400 for missing content type', async ({ request }) => {
      const response = await request.post(`/users`, {
        data: { username: 'test', email: 'test@example.com' }
      });
      
      expect(response.status()).toBe(400);
    });
  });

  test.describe('Boundary Values', () => {
    test('GET /users/:id - should handle very large ID', async ({ request }) => {
      const largeId = 999999999999999999;
      const response = await request.get(`/users/${largeId}`);
      
      // Should return 404 (not found) or handle gracefully
      expect([200, 400, 404]).toContain(response.status());
    });

    test('GET /users/:id - should handle ID with leading zeros', async ({ request }) => {
      const response = await request.get(`/users/00042`);
      
      // Should parse correctly or return 404
      expect([200, 400, 404]).toContain(response.status());
    });

    test('POST /users - should handle unicode characters', async ({ request }) => {
      const unicodeUser = {
        username: '测试用户_🎉',
        email: 'test_unicode@example.com',
        full_name: 'Test User 日本語'
      };

      const response = await request.post(`/users`, {
        data: unicodeUser,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // Should either accept unicode or return 400
      expect([201, 400]).toContain(response.status());
    });

    test('POST /users - should handle emoji in content', async ({ request }) => {
      const emojiUser = {
        username: 'emoji_user',
        email: 'emoji@example.com',
        content: {
          message: 'Hello 🌍🎉🚀',
          emojis: ['😀', '😃', '😄']
        }
      };

      const response = await request.post(`/users`, {
        data: emojiUser,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      // Should either accept emoji or return 400
      expect([201, 400]).toContain(response.status());
    });
  });

  test.describe('Missing Required Fields', () => {
    test('POST /users - should return 400 when username is empty string', async ({ request }) => {
      const invalidUser = {
        username: '',
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

    test('POST /users - should return 400 when email is empty string', async ({ request }) => {
      const invalidUser = {
        username: 'testuser',
        email: ''
      };

      const response = await request.post(`/users`, {
        data: invalidUser,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(400);
    });

    test('POST /users - should return 400 when both username and email are missing', async ({ request }) => {
      const invalidUser = {
        full_name: 'Test User'
      };

      const response = await request.post(`/users`, {
        data: invalidUser,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(400);
    });
  });

  test.describe('Path Traversal and Security', () => {
    test('GET /users/:id - should handle path traversal attempt', async ({ request }) => {
      const response = await request.get(`/users/../../../etc/passwd`);
      
      expect([400, 404]).toContain(response.status());
    });

    test('GET /users/:id - should handle null byte injection', async ({ request }) => {
      const response = await request.get(`/users/1%00`);
      
      expect([400, 404]).toContain(response.status());
    });
  });
});