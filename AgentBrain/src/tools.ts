import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { z } from 'zod'
import { BrainClient } from './client.js'

function textResult(data: unknown) {
  return { content: [{ type: 'text' as const, text: JSON.stringify(data, null, 2) }] }
}

function errorResult(err: unknown) {
  const message = err instanceof Error ? err.message : String(err)
  return { content: [{ type: 'text' as const, text: message }], isError: true as const }
}

export function registerTools(server: McpServer, client: BrainClient): void {
  server.tool('health_check', 'Check agent-brain-api health (SQLite).', {}, async () => {
    try {
      return textResult(await client.health())
    } catch (err) {
      return errorResult(err)
    }
  })

  server.tool(
    'list_sessions',
    'List Sessions with context % and token totals.',
    { q: z.string().optional().describe('Optional name search') },
    async ({ q }) => {
      try {
        return textResult(await client.listSessions(q))
      } catch (err) {
        return errorResult(err)
      }
    },
  )

  server.tool(
    'store_human_prompt',
    'Store a human prompt in agent-brain (prompts table).',
    {
      prompt: z.string().min(1).describe('Raw human prompt text'),
      technically_revised: z.string().optional().describe('Technically revised prompt'),
      english_revised: z.string().optional().describe('English-revised prompt'),
      harness: z.string().optional().describe('Optional harness label'),
    },
    async ({ prompt, technically_revised, english_revised, harness }) => {
      try {
        return textResult(
          await client.createPrompt({
            prompt,
            technically_revised,
            english_revised,
            harness,
          }),
        )
      } catch (err) {
        return errorResult(err)
      }
    },
  )
}
