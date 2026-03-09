# flagr-mcp

MCP server for [Flagr](https://flagr.dev) — control your feature flags directly from Claude Desktop, Cursor, or any MCP-compatible AI assistant.

## What it does

Exposes your Flagr organization's feature flags as MCP tools, so your AI assistant can:

- List projects, environments, and flags
- Check and change flag states (enabled / disabled / partially enabled)
- Manage per-tenant partial rollout lists
- Inspect flag change history for incident investigation

The binary runs locally and communicates with Flagr's Management API over HTTPS. Your API key never leaves your machine.

## Prerequisites

- A [Flagr](https://flagr.dev) account with an org API key (`sk_org_...`)
- Go 1.23+ **or** a pre-built binary (see Releases)

## Installation

### Build from source

```bash
go install github.com/flagr-dev/flagr-mcp@latest
```

### Download a release binary

Grab the latest binary for your platform from the [Releases](https://github.com/flagr-dev/flagr-mcp/releases) page and place it somewhere on your `$PATH`.

## Claude Desktop setup

Add the following to your `claude_desktop_config.json`:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "flagr": {
      "command": "flagr-mcp",
      "env": {
        "FLAGR_API_KEY": "sk_org_your_key_here"
      }
    }
  }
}
```

If `flagr-mcp` is not on your `$PATH`, use the absolute path:

```json
{
  "mcpServers": {
    "flagr": {
      "command": "/usr/local/bin/flagr-mcp",
      "env": {
        "FLAGR_API_KEY": "sk_org_your_key_here"
      }
    }
  }
}
```

Restart Claude Desktop. You should see a hammer icon indicating the MCP tools are available.

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `FLAGR_API_KEY` | Yes | — | Org API key from Flagr dashboard (`sk_org_...`) |
| `FLAGR_BASE_URL` | No | `https://api.flagr.dev` | Override for self-hosted or staging |

## Available tools

| Tool | Description |
|---|---|
| `list_projects` | List all projects in your organization |
| `list_environments` | List environments for a project |
| `list_flags` | List all feature flags in a project |
| `get_flag_state` | Get a flag's current state and tenant list in a specific environment |
| `set_flag_state` | Set a flag to `enabled`, `disabled`, or `partially_enabled` |
| `get_flag_history` | Get the audit log for a flag (newest first) — useful for incident investigation |
| `list_tenants` | List tenant IDs in the partial rollout list for a flag |
| `add_tenant` | Add a tenant to the partial rollout list |
| `remove_tenant` | Remove a tenant from the partial rollout list |

## Example prompts

```
Which feature flags are currently enabled in the production environment?
```

```
Disable the "new-checkout" flag in production immediately.
```

```
Add tenant "acme-corp" to the "dark-mode" flag's partial rollout in staging.
```

```
Show me the change history for the "payments-v2" flag — we had an incident last night.
```

## Getting your API key

1. Sign in at [flagr.dev](https://flagr.dev)
2. Go to **Settings → API Keys**
3. Click **Create API Key** — copy the key now (shown once)
4. Paste it into your MCP config as `FLAGR_API_KEY`

## Security

- The binary only reads `FLAGR_API_KEY` from the environment — never from disk or arguments.
- All requests go to `api.flagr.dev` (or your custom `FLAGR_BASE_URL`) over HTTPS.
- The binary makes no outbound connections other than to the Flagr API.
