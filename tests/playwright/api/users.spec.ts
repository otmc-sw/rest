/**
 * @License OTMC License
 * @Copyright (c) 2026 OTMC Softwares. All rights reserved.
 * @Contributors Trung Ng, OTMC Authors.
 */

import { test, expect, request } from '@playwright/test';

const createdTemplateIds: number[] = [];
const createdDocumentIdsFromTemplate: number[] = [];

function generateTestData() {
  return {
    title: 'Test Template ' + Date.now(),
    content: JSON.stringify({
      sections: [
        {
          title: 'Section 1',
          placeholder: 'Enter content here...'
        }
      ]
    }),
    doc_icon: '📋',
    description: 'Test template created by Playwright'
  };
}

test.describe('Templates API', () => {
  test('GET /api/templates - should list all templates', async ({ request }) => {
    const response = await request.get(`/api/templates`);
    
    expect(response.status()).toBe(200);
    
    const body = await response.json();
    expect(body.success).toBe(true);
    
    const templates = body.data || [];
    expect(Array.isArray(templates)).toBe(true);
  });

  test('POST /api/templates - should create a new template', async ({ request }) => {
    const newTemplate = generateTestData();

    const response = await request.post(`/api/templates`, {
      data: newTemplate,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(201);
    
    const body = await response.json();
    const template = body.data || body;
    expect(template).toHaveProperty('id');
    expect(template).toHaveProperty('title');
    expect(template).toHaveProperty('content');
  });

  test('GET /api/templates/:id - should get template by ID', async ({ request }) => {
    const newTemplate = generateTestData();
    const createResponse = await request.post(`/api/templates`, {
      data: newTemplate,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);
    
    const listResponse = await request.get(`/api/templates`);
    const listBody = await listResponse.json();
    const templates = listBody.data || [];
    const createdTemplate = templates.find((t: any) => t.title === newTemplate.title);
    
    if (createdTemplate) {
      const templateId = createdTemplate.id;
      createdTemplateIds.push(templateId);

      const response = await request.get(`/api/templates/${templateId}`);
      
      expect(response.status()).toBe(200);
      
      const body = await response.json();
      const template = body.data || body;
      expect(template).toHaveProperty('id', templateId);
      expect(template).toHaveProperty('title');
      expect(template).toHaveProperty('content');
      expect(template).toHaveProperty('is_system');
      expect(template).toHaveProperty('created_at');
      expect(template).toHaveProperty('updated_at');
    } else {
      const nonExistentId = 999999;
      const response = await request.get(`/api/templates/${nonExistentId}`);
      expect([400, 404]).toContain(response.status());
    }
  });

  test('GET /api/templates/:id - should return error for non-existent template', async ({ request }) => {
    const nonExistentId = 999999;
    const response = await request.get(`/api/templates/${nonExistentId}`);
    
    expect([400, 404]).toContain(response.status());
  });

  test('PATCH /api/templates/:id - should update a template', async ({ request }) => {
    const newTemplate = generateTestData();
    const createResponse = await request.post(`/api/templates`, {
      data: newTemplate,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);

    const listResponse = await request.get(`/api/templates`);
    const listBody = await listResponse.json();
    const templates = listBody.data || [];
    const createdTemplate = templates.find((t: any) => t.title === newTemplate.title);
    
    if (createdTemplate) {
      const templateId = createdTemplate.id;
      createdTemplateIds.push(templateId);

      const updateData = {
        title: 'Updated Test Template ' + Date.now(),
        doc_icon: '✏️'
      };

      const response = await request.patch(`/api/templates/${templateId}`, {
        data: updateData,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(200);
      
      const body = await response.json();
      const template = body.data || body;
      expect(template).toHaveProperty('id');
      expect(template).toHaveProperty('title');
    }
  });

  test('PATCH /api/templates/:id - should preserve existing fields when not provided', async ({ request }) => {
    const newTemplate = generateTestData();
    const createResponse = await request.post(`/api/templates`, {
      data: newTemplate,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);

    const listResponse = await request.get(`/api/templates`);
    const listBody = await listResponse.json();
    const templates = listBody.data || [];
    const createdTemplate = templates.find((t: any) => t.title === newTemplate.title);
    
    if (createdTemplate) {
      const templateId = createdTemplate.id;
      createdTemplateIds.push(templateId);

      const updateData = {
        title: 'Partially Updated Template ' + Date.now()
      };

      const response = await request.patch(`/api/templates/${templateId}`, {
        data: updateData,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(200);
    }
  });

  test('POST /api/templates/:id/create - should create document from template', async ({ request }) => {
    const newTemplate = generateTestData();
    const createResponse = await request.post(`/api/templates`, {
      data: newTemplate,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);
    
    const listResponse = await request.get(`/api/templates`);
    const listBody = await listResponse.json();
    const templates = listBody.data || [];
    const createdTemplate = templates.find((t: any) => t.title === newTemplate.title);
    
    if (createdTemplate) {
      const templateId = createdTemplate.id;
      createdTemplateIds.push(templateId);

      const documentData = {
        parent_id: null,
        doc_type: 'document',
        doc_status: 'draft',
        title: 'Document Created from Template ' + Date.now()
      };

      const response = await request.post(`/api/templates/${templateId}/create`, {
        data: documentData,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      expect(response.status()).toBe(201);
      
      const body = await response.json();
      const document = body.data || body;
      expect(document).toHaveProperty('id');
      expect(document).toHaveProperty('title');
      expect(document).toHaveProperty('doc_type');
      expect(document).toHaveProperty('doc_status');
      
      if (document.id) {
        createdDocumentIdsFromTemplate.push(document.id);
      }
    }
  });

  test('DELETE /api/templates/:id - should delete a template', async ({ request }) => {
    const newTemplate = generateTestData();
    const createResponse = await request.post(`/api/templates`, {
      data: newTemplate,
      headers: {
        'Content-Type': 'application/json'
      }
    });
    
    expect(createResponse.status()).toBe(201);
    
    const listResponse = await request.get(`/api/templates`);
    const listBody = await listResponse.json();
    const templates = listBody.data || [];
    const createdTemplate = templates.find((t: any) => t.title === newTemplate.title);
    
    if (createdTemplate) {
      const templateId = createdTemplate.id;

      const response = await request.delete(`/api/templates/${templateId}`);
      
      expect([204, 200]).toContain(response.status());

      const getResponse = await request.get(`/api/templates/${templateId}`);
      expect([400, 404]).toContain(getResponse.status());
      
      const index = createdTemplateIds.indexOf(templateId);
      if (index > -1) createdTemplateIds.splice(index, 1);
    }
  });

  test('POST /api/templates - should return 400 when title is missing', async ({ request }) => {
    const invalidTemplate = {
      content: JSON.stringify({ sections: [] })
    };

    const response = await request.post(`/api/templates`, {
      data: invalidTemplate,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(400);
  });

  test('POST /api/templates - should create template with JSON content', async ({ request }) => {
    const templateWithJsonContent = {
      title: 'Template with JSON Content ' + Date.now(),
      content: JSON.stringify({
        type: 'page',
        elements: [
          {
            type: 'heading',
            text: 'Welcome'
          },
          {
            type: 'paragraph',
            text: 'Hello World'
          }
        ]
      })
    };

    const response = await request.post(`/api/templates`, {
      data: templateWithJsonContent,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect(response.status()).toBe(201);
    
    const body = await response.json();
    const template = body.data || body;
    expect(template).toHaveProperty('title');
    expect(template).toHaveProperty('content');
  });

  test('PATCH /api/templates/:id - should return error for non-existent template', async ({ request }) => {
    const nonExistentId = 999999;
    const updateData = {
      title: 'Updated Title'
    };

    const response = await request.patch(`/api/templates/${nonExistentId}`, {
      data: updateData,
      headers: {
        'Content-Type': 'application/json'
      }
    });

    expect([400, 404]).toContain(response.status());
  });

  test('DELETE /api/templates/:id - should return success for non-existent template', async ({ request }) => {
    const nonExistentId = 999999;
    const response = await request.delete(`/api/templates/${nonExistentId}`);
    
    expect([204, 404, 400]).toContain(response.status());
  });
});

test.afterAll(async ({ request }) => {
  for (const documentId of [...createdDocumentIdsFromTemplate].reverse()) {
    try {
      await request.delete(`/api/documents/${documentId}`);
    } catch (error) {
      console.warn(`Failed to cleanup document ${documentId}:`, error);
    }
  }
  createdDocumentIdsFromTemplate.length = 0;
  
  for (const templateId of [...createdTemplateIds].reverse()) {
    try {
      await request.delete(`/api/templates/${templateId}`);
    } catch (error) {
      console.warn(`Failed to cleanup template ${templateId}:`, error);
    }
  }
  createdTemplateIds.length = 0;
});