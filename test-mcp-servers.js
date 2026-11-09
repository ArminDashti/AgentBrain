#!/usr/bin/env node

/**
 * Test script for AgentBrain MCP servers
 * Tests both agent-brain-mcp and mcp-prompt servers
 */

import { spawn } from 'child_process';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

const API_URL = process.env.API_URL || 'http://127.0.0.1:8220';
const USERNAME = 'armin';
const PASSWORD = 'dopadopa123';

class MCPTester {
  constructor(command, args, env, name) {
    this.command = command;
    this.args = args;
    this.env = env;
    this.name = name;
    this.client = null;
    this.transport = null;
  }

  async connect() {
    console.log(`\n=== Connecting to ${this.name} ===`);
    this.transport = new StdioClientTransport({
      command: this.command,
      args: this.args,
      env: { ...process.env, ...this.env },
    });

    this.client = new Client({
      name: `${this.name}-test-client`,
      version: '1.0.0',
    }, {
      capabilities: {
        roots: {},
      },
    });

    await this.client.connect(this.transport);
    console.log(`✓ Connected to ${this.name}`);
  }

  async listTools() {
    console.log(`\n--- Listing tools from ${this.name} ---`);
    const response = await this.client.request({ method: 'tools/list' }, null);
    console.log(`✓ Found ${response.tools.length} tools:`);
    response.tools.forEach(tool => {
      console.log(`  - ${tool.name}: ${tool.description || '(no description)'}`);
    });
    return response.tools;
  }

  async callTool(name, args) {
    console.log(`\n--- Calling tool: ${name} ---`);
    console.log(`Arguments:`, JSON.stringify(args, null, 2));
    try {
      const response = await this.client.request(
        { method: 'tools/call' },
        { name, arguments: args }
      );
      console.log(`✓ Tool call succeeded`);
      console.log(`Response:`, JSON.stringify(response, null, 2));
      return { success: true, response };
    } catch (error) {
      console.log(`✗ Tool call failed:`, error.message);
      return { success: false, error: error.message };
    }
  }

  async disconnect() {
    if (this.client) {
      await this.client.close();
      console.log(`✓ Disconnected from ${this.name}\n`);
    }
  }
}

async function testAgentBrainMCP() {
  const tester = new MCPTester(
    'node',
    ['/agent/repos/AgentBrain/AgentBrain/dist/index.js'],
    {
      AGENT_BRAIN_API_URL: API_URL,
      AGENT_BRAIN_USERNAME: USERNAME,
      AGENT_BRAIN_PASSWORD: PASSWORD,
    },
    'agent-brain-mcp'
  );

  const results = [];

  try {
    await tester.connect();
    
    // List tools
    const tools = await tester.listTools();
    results.push({ test: 'list_tools', pass: tools.length > 0 });

    // Test health_check
    let result = await tester.callTool('health_check', {});
    results.push({ test: 'health_check', pass: result.success });

    // Test login
    result = await tester.callTool('login', { username: USERNAME, password: PASSWORD });
    results.push({ test: 'login_valid', pass: result.success });

    // Test login with invalid credentials
    result = await tester.callTool('login', { username: 'bad', password: 'bad' });
    results.push({ test: 'login_invalid', pass: !result.success });

    // Test store_knowledge
    result = await tester.callTool('store_knowledge', {
      kind: 'general_knowledge',
      title: 'MCP Test Knowledge',
      body: 'This is a test from MCP',
      tags: ['test', 'mcp'],
      store_target: 'sqlite',
    });
    results.push({ test: 'store_knowledge', pass: result.success });
    const knowledgeId = result.success ? result.response.content[0]?.text : null;

    // Test search_knowledge
    result = await tester.callTool('search_knowledge', {
      query: 'MCP Test',
      mode: 'structured',
    });
    results.push({ test: 'search_knowledge', pass: result.success });

    // Test list_knowledge
    result = await tester.callTool('list_knowledge', { limit: 10 });
    results.push({ test: 'list_knowledge', pass: result.success });

    // Test get_knowledge
    if (knowledgeId) {
      try {
        const parsed = JSON.parse(knowledgeId);
        if (parsed.id) {
          result = await tester.callTool('get_knowledge', { id: parsed.id });
          results.push({ test: 'get_knowledge', pass: result.success });
          
          // Test delete_knowledge
          result = await tester.callTool('delete_knowledge', { id: parsed.id });
          results.push({ test: 'delete_knowledge', pass: result.success });
        }
      } catch (e) {
        results.push({ test: 'get_knowledge', pass: false, error: 'Failed to parse knowledge ID' });
      }
    }

    // Test with invalid inputs
    result = await tester.callTool('store_knowledge', {
      kind: 'invalid_kind',
      title: 'Test',
      body: 'Test',
      store_target: 'sqlite',
    });
    results.push({ test: 'store_knowledge_invalid_kind', pass: !result.success });

    result = await tester.callTool('store_knowledge', {
      kind: 'general_knowledge',
      title: '',
      body: 'Test',
      store_target: 'sqlite',
    });
    results.push({ test: 'store_knowledge_empty_title', pass: !result.success });

    // Test API down scenario (disconnect and reconnect with bad URL)
    await tester.disconnect();

    const badTester = new MCPTester(
      'node',
      ['/agent/repos/AgentBrain/AgentBrain/dist/index.js'],
      {
        AGENT_BRAIN_API_URL: 'http://127.0.0.1:9999',
        AGENT_BRAIN_USERNAME: USERNAME,
        AGENT_BRAIN_PASSWORD: PASSWORD,
      },
      'agent-brain-mcp (bad URL)'
    );

    await badTester.connect();
    result = await badTester.callTool('health_check', {});
    results.push({ test: 'api_down_error_message', pass: !result.success });
    await badTester.disconnect();

  } catch (error) {
    console.error(`\n✗ Test failed with error:`, error);
    results.push({ test: 'overall', pass: false, error: error.message });
  } finally {
    await tester.disconnect();
  }

  return results;
}

async function testMCPPrompt() {
  const tester = new MCPTester(
    'node',
    ['/agent/repos/AgentBrain/mcp-prompt/dist/index.js'],
    {
      AGENT_BRAIN_API_URL: API_URL,
      AGENT_BRAIN_USERNAME: USERNAME,
      AGENT_BRAIN_PASSWORD: PASSWORD,
    },
    'mcp-prompt'
  );

  const results = [];

  try {
    await tester.connect();

    // List tools
    const tools = await tester.listTools();
    results.push({ test: 'list_tools', pass: tools.length > 0 });

    // Test send-human-prompt
    let result = await tester.callTool('send-human-prompt', {
      prompt: 'Test prompt from MCP',
      technically_revised: 'Technically revised test',
      english_revised: 'English revised test',
      harness: 'test harness',
    });
    results.push({ test: 'send_human_prompt', pass: result.success });

    // Test with minimal data
    result = await tester.callTool('send-human-prompt', {
      prompt: 'Minimal test',
    });
    results.push({ test: 'send_human_prompt_minimal', pass: result.success });

    // Test with missing prompt
    result = await tester.callTool('send-human-prompt', {
      technically_revised: 'No prompt provided',
    });
    results.push({ test: 'send_human_prompt_missing_prompt', pass: !result.success });

    // Test API down
    await tester.disconnect();

    const badTester = new MCPTester(
      'node',
      ['/agent/repos/AgentBrain/mcp-prompt/dist/index.js'],
      {
        AGENT_BRAIN_API_URL: 'http://127.0.0.1:9999',
        AGENT_BRAIN_USERNAME: USERNAME,
        AGENT_BRAIN_PASSWORD: PASSWORD,
      },
      'mcp-prompt (bad URL)'
    );

    await badTester.connect();
    result = await badTester.callTool('send-human-prompt', { prompt: 'Test' });
    results.push({ test: 'api_down_error', pass: !result.success });
    await badTester.disconnect();

  } catch (error) {
    console.error(`\n✗ Test failed with error:`, error);
    results.push({ test: 'overall', pass: false, error: error.message });
  } finally {
    await tester.disconnect();
  }

  return results;
}

function printSummary(name, results) {
  console.log(`\n\n${'='.repeat(60)}`);
  console.log(`${name} Test Results`);
  console.log('='.repeat(60));
  
  let passed = 0;
  let failed = 0;

  results.forEach(result => {
    const status = result.pass ? '✓ PASS' : '✗ FAIL';
    console.log(`${status}: ${result.test}`);
    if (result.error) {
      console.log(`  Error: ${result.error}`);
    }
    if (result.pass) {
      passed++;
    } else {
      failed++;
    }
  });

  console.log(`\nTotal: ${results.length} tests, ${passed} passed, ${failed} failed`);
  return failed === 0;
}

async function main() {
  console.log('AgentBrain MCP Servers Test Suite');
  console.log('==================================\n');
  console.log(`API URL: ${API_URL}`);
  console.log(`Username: ${USERNAME}`);

  // Test agent-brain-mcp
  const mcpResults = await testAgentBrainMCP();
  const mcpPass = printSummary('agent-brain-mcp', mcpResults);

  // Test mcp-prompt
  const promptResults = await testMCPPrompt();
  const promptPass = printSummary('mcp-prompt', promptResults);

  // Overall summary
  console.log(`\n\n${'='.repeat(60)}`);
  console.log('Overall Summary');
  console.log('='.repeat(60));
  console.log(`agent-brain-mcp: ${mcpPass ? '✓ PASS' : '✗ FAIL'}`);
  console.log(`mcp-prompt: ${promptPass ? '✓ PASS' : '✗ FAIL'}`);

  process.exit(mcpPass && promptPass ? 0 : 1);
}

main().catch(error => {
  console.error('Fatal error:', error);
  process.exit(1);
});
