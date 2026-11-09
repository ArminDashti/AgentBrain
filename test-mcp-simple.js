#!/usr/bin/env node

/**
 * Simple MCP server test using stdio
 */

import { spawn } from 'child_process';
import { once } from 'events';

const API_URL = process.env.API_URL || 'http://127.0.0.1:8220';

async function testMCPServer(command, args, env, serverName) {
  console.log(`\n${'='.repeat(60)}`);
  console.log(`Testing ${serverName}`);
  console.log('='.repeat(60));

  const proc = spawn(command, args, {
    env: { ...process.env, ...env },
    stdio: ['pipe', 'pipe', 'pipe'],
  });

  let stderrOutput = '';
  proc.stderr.on('data', (data) => {
    stderrOutput += data.toString();
  });

  const results = [];
  let messageId = 1;

  async function sendRequest(method, params) {
    const request = {
      jsonrpc: '2.0',
      id: messageId++,
      method,
      params: params || {},
    };
    
    console.log(`\n→ ${method}`);
    proc.stdin.write(JSON.stringify(request) + '\n');

    return new Promise((resolve, reject) => {
      let responseData = '';
      const timeout = setTimeout(() => {
        reject(new Error(`Timeout waiting for response to ${method}`));
      }, 5000);

      const dataHandler = (chunk) => {
        responseData += chunk.toString();
        const lines = responseData.split('\n');
        
        for (let i = 0; i < lines.length - 1; i++) {
          const line = lines[i].trim();
          if (!line) continue;
          
          try {
            const response = JSON.parse(line);
            if (response.id === request.id) {
              clearTimeout(timeout);
              proc.stdout.off('data', dataHandler);
              
              if (response.error) {
                console.log(`← ERROR: ${response.error.message}`);
                resolve({ success: false, error: response.error });
              } else {
                console.log(`← SUCCESS`);
                resolve({ success: true, result: response.result });
              }
              return;
            }
          } catch (e) {
            // Not complete JSON yet
          }
        }
        responseData = lines[lines.length - 1];
      };

      proc.stdout.on('data', dataHandler);
    });
  }

  try {
    // Initialize
    let result = await sendRequest('initialize', {
      protocolVersion: '2024-11-05',
      capabilities: {},
      clientInfo: { name: 'test-client', version: '1.0.0' },
    });
    results.push({ test: 'initialize', pass: result.success });

    // List tools
    result = await sendRequest('tools/list', {});
    results.push({ test: 'list_tools', pass: result.success && result.result?.tools?.length > 0 });
    
    if (result.success) {
      console.log(`  Found ${result.result.tools.length} tools:`);
      result.result.tools.forEach(t => console.log(`    - ${t.name}`));
    }

    // Call a simple tool
    if (serverName.includes('agent-brain-mcp')) {
      result = await sendRequest('tools/call', {
        name: 'health_check',
        arguments: {},
      });
      results.push({ test: 'health_check', pass: result.success });

      result = await sendRequest('tools/call', {
        name: 'list_knowledge',
        arguments: { limit: 5 },
      });
      results.push({ test: 'list_knowledge', pass: result.success });

      // Test invalid call
      result = await sendRequest('tools/call', {
        name: 'nonexistent_tool',
        arguments: {},
      });
      results.push({ test: 'invalid_tool_fails', pass: !result.success });

    } else if (serverName.includes('mcp-prompt')) {
      result = await sendRequest('tools/call', {
        name: 'send-human-prompt',
        arguments: {
          prompt: 'Test prompt',
          technically_revised: 'Technical test',
        },
      });
      results.push({ test: 'send_human_prompt', pass: result.success });

      // Test missing required field
      result = await sendRequest('tools/call', {
        name: 'send-human-prompt',
        arguments: {},
      });
      results.push({ test: 'missing_prompt_fails', pass: !result.success });
    }

  } catch (error) {
    console.error(`\n✗ Error: ${error.message}`);
    results.push({ test: 'overall', pass: false, error: error.message });
  } finally {
    proc.stdin.end();
    proc.kill();
  }

  // Print results
  console.log(`\n--- Results for ${serverName} ---`);
  let passed = 0;
  let failed = 0;
  
  results.forEach(r => {
    const status = r.pass ? '✓' : '✗';
    console.log(`${status} ${r.test}`);
    if (r.error) console.log(`  Error: ${r.error}`);
    r.pass ? passed++ : failed++;
  });
  
  console.log(`\nTotal: ${results.length} tests, ${passed} passed, ${failed} failed`);
  
  if (stderrOutput.trim()) {
    console.log(`\nStderr output:\n${stderrOutput}`);
  }

  return failed === 0;
}

async function main() {
  console.log('AgentBrain MCP Servers Simple Test');
  console.log(`API URL: ${API_URL}\n`);

  const mcp1Pass = await testMCPServer(
    'node',
    ['/agent/repos/AgentBrain/AgentBrain/dist/index.js'],
    {
      AGENT_BRAIN_API_URL: API_URL,
      AGENT_BRAIN_USERNAME: 'armin',
      AGENT_BRAIN_PASSWORD: 'dopadopa123',
    },
    'agent-brain-mcp'
  );

  const mcp2Pass = await testMCPServer(
    'node',
    ['/agent/repos/AgentBrain/mcp-prompt/dist/index.js'],
    {
      AGENT_BRAIN_API_URL: API_URL,
      AGENT_BRAIN_USERNAME: 'armin',
      AGENT_BRAIN_PASSWORD: 'dopadopa123',
    },
    'mcp-prompt'
  );

  console.log(`\n\n${'='.repeat(60)}`);
  console.log('Overall Summary');
  console.log('='.repeat(60));
  console.log(`agent-brain-mcp: ${mcp1Pass ? '✓ PASS' : '✗ FAIL'}`);
  console.log(`mcp-prompt: ${mcp2Pass ? '✓ PASS' : '✗ FAIL'}`);

  process.exit(mcp1Pass && mcp2Pass ? 0 : 1);
}

main().catch(error => {
  console.error('Fatal error:', error);
  process.exit(1);
});
