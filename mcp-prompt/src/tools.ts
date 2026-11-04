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
  server.tool(
    'send-human-prompt',
    'Create a human prompt record in AgentBrain (raw prompt plus optional revised text and harness). Returns the new prompt id.',
    {
      prompt: z.string().min(1).describe('Raw human prompt text'),
      technically_revised: z
        .string()
        .optional()
        .describe('Technically revised variant (empty if not yet rewritten)'),
      english_revised: z
        .string()
        .optional()
        .describe('English-revised variant (empty if not yet rewritten)'),
      harness: z.string().optional().describe('Harness / routing label (optional)'),
    },
    async ({ prompt, technically_revised, english_revised, harness }) => {
      try {
        const result = await client.createPrompt({
          prompt,
          technically_revised,
          english_revised,
          harness,
        })
        return textResult(result)
      } catch (err) {
        return errorResult(err)
      }
    },
  )
}
