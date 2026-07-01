#!/usr/bin/env node
// WC-Tournament MCP server (stdio). Read-only: registers the validated tools and
// wires them to a credential-free HTTP read client. No secret is read from the
// environment beyond the (non-secret) read base URL. See docs/adr/0015 and
// docs/features/mcp/spec.md.
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { httpClient } from './api.js'
import { TOOLS } from './tools.js'

const base = process.env.WC_API_BASE
if (!base) {
  console.error('WC_API_BASE is required (a read-reachable WC-Tournament API base URL)')
  process.exit(1)
}
const client = httpClient(base)

const server = new McpServer({ name: 'wc-tournament', version: '0.1.0' })

for (const tool of TOOLS) {
  server.registerTool(
    tool.name,
    {
      title: tool.title,
      description: tool.description,
      inputSchema: tool.schema.shape,
    },
    async (args: unknown) => {
      // Re-validate defensively, then run the read-only handler.
      const parsed = tool.schema.parse(args ?? {})
      const result = await tool.run(parsed, client)
      return { content: [{ type: 'text' as const, text: JSON.stringify(result, null, 2) }] }
    },
  )
}

await server.connect(new StdioServerTransport())
