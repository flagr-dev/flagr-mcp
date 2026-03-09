package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	apiKey := os.Getenv("FLAGR_API_KEY")
	if apiKey == "" {
		log.Fatal("FLAGR_API_KEY environment variable is required (use an org API key: sk_org_...)")
	}

	baseURL := os.Getenv("FLAGR_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.flagr.dev"
	}

	client := NewClient(apiKey, baseURL)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "flagr",
		Version: "0.1.0",
	}, nil)

	registerTools(server, client)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func textResult(data json.RawMessage) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

func errResult(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}, nil, nil
}

func registerTools(server *mcp.Server, client *Client) {

	// -----------------------------------------------------------------------
	// list_projects
	// -----------------------------------------------------------------------
	type ListProjectsArgs struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all projects in the Flagr organization.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ ListProjectsArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.ListProjects(ctx)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})

	// -----------------------------------------------------------------------
	// list_environments
	// -----------------------------------------------------------------------
	type ListEnvironmentsArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_environments",
		Description: "List all environments for a project.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a ListEnvironmentsArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.ListEnvironments(ctx, a.ProjectID)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})

	// -----------------------------------------------------------------------
	// list_flags
	// -----------------------------------------------------------------------
	type ListFlagsArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_flags",
		Description: "List all feature flags in a project.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a ListFlagsArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.ListFlags(ctx, a.ProjectID)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})

	// -----------------------------------------------------------------------
	// get_flag_state
	// -----------------------------------------------------------------------
	type GetFlagStateArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
		FlagID    string `json:"flag_id"    jsonschema:"The flag UUID"`
		EnvID     string `json:"env_id"     jsonschema:"The environment UUID"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_flag_state",
		Description: "Get the current state of a feature flag in a specific environment. Returns state (enabled/disabled/partially_enabled) and the tenant list.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a GetFlagStateArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.GetFlagState(ctx, a.ProjectID, a.FlagID, a.EnvID)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})

	// -----------------------------------------------------------------------
	// set_flag_state
	// -----------------------------------------------------------------------
	type SetFlagStateArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
		FlagID    string `json:"flag_id"    jsonschema:"The flag UUID"`
		EnvID     string `json:"env_id"     jsonschema:"The environment UUID"`
		State     string `json:"state"      jsonschema:"New state: enabled, disabled, or partially_enabled"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_flag_state",
		Description: "Set the state of a feature flag in a specific environment. Use 'enabled' or 'disabled' for full rollout control. Use 'partially_enabled' to limit to specific tenants — manage the tenant list with add_tenant and remove_tenant.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a SetFlagStateArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.SetFlagState(ctx, a.ProjectID, a.FlagID, a.EnvID, a.State)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})

	// -----------------------------------------------------------------------
	// get_flag_history
	// -----------------------------------------------------------------------
	type GetFlagHistoryArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
		FlagID    string `json:"flag_id"    jsonschema:"The flag UUID"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_flag_history",
		Description: "Get the state change history (audit log) for a feature flag across all environments, newest first. Useful for incident investigation.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a GetFlagHistoryArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.GetFlagHistory(ctx, a.ProjectID, a.FlagID)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})

	// -----------------------------------------------------------------------
	// list_tenants
	// -----------------------------------------------------------------------
	type ListTenantsArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
		FlagID    string `json:"flag_id"    jsonschema:"The flag UUID"`
		EnvID     string `json:"env_id"     jsonschema:"The environment UUID"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tenants",
		Description: "List all tenant IDs in the partial rollout list for a feature flag in a specific environment.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a ListTenantsArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.GetFlagState(ctx, a.ProjectID, a.FlagID, a.EnvID)
		if err != nil {
			return errResult(err)
		}
		// Extract just the enabled_list from the full flag state response.
		var state struct {
			EnabledList json.RawMessage `json:"enabled_list"`
		}
		if err := json.Unmarshal(data, &state); err != nil {
			return textResult(data)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf(`{"enabled_list":%s}`, state.EnabledList)},
			},
		}, nil, nil
	})

	// -----------------------------------------------------------------------
	// add_tenant
	// -----------------------------------------------------------------------
	type AddTenantArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
		FlagID    string `json:"flag_id"    jsonschema:"The flag UUID"`
		EnvID     string `json:"env_id"     jsonschema:"The environment UUID"`
		TenantID  string `json:"tenant_id"  jsonschema:"The tenant ID to add to the partial rollout list (UUID recommended)"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_tenant",
		Description: "Add a tenant ID to the partial rollout list for a feature flag. The flag must be in partially_enabled state for this to affect evaluation.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a AddTenantArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.AddTenant(ctx, a.ProjectID, a.FlagID, a.EnvID, a.TenantID)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})

	// -----------------------------------------------------------------------
	// remove_tenant
	// -----------------------------------------------------------------------
	type RemoveTenantArgs struct {
		ProjectID string `json:"project_id" jsonschema:"The project UUID"`
		FlagID    string `json:"flag_id"    jsonschema:"The flag UUID"`
		EnvID     string `json:"env_id"     jsonschema:"The environment UUID"`
		TenantID  string `json:"tenant_id"  jsonschema:"The tenant ID to remove from the partial rollout list"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_tenant",
		Description: "Remove a tenant ID from the partial rollout list for a feature flag.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, a RemoveTenantArgs) (*mcp.CallToolResult, any, error) {
		data, err := client.RemoveTenant(ctx, a.ProjectID, a.FlagID, a.EnvID, a.TenantID)
		if err != nil {
			return errResult(err)
		}
		return textResult(data)
	})
}
