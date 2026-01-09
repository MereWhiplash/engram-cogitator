// internal/mcptypes/types.go
// Package mcptypes contains shared MCP tool input/output types.
// These are used by both the direct MCP server (tools) and the shim proxy.
package mcptypes

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// AddInput defines the input schema for ec_add
type AddInput struct {
	Type      string `json:"type" jsonschema:"required" jsonschema_description:"Type of memory: decision, learning, or pattern"`
	Area      string `json:"area" jsonschema:"required" jsonschema_description:"Domain area (e.g. auth, permissions, ui, api)"`
	Content   string `json:"content" jsonschema:"required" jsonschema_description:"The actual content to remember"`
	Rationale string `json:"rationale,omitempty" jsonschema_description:"Why this matters or additional context"`
}

// AddOutput defines the output schema for ec_add
type AddOutput struct {
	Memory *types.Memory `json:"memory"`
}

// SearchInput defines the input schema for ec_search
type SearchInput struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"Search query to find relevant memories"`
	Limit int    `json:"limit,omitempty" jsonschema_description:"Maximum number of results (default: 5)"`
	Type  string `json:"type,omitempty" jsonschema_description:"Filter by type (decision, learning, or pattern)"`
	Area  string `json:"area,omitempty" jsonschema_description:"Filter by domain area"`
}

// SearchOutput defines the output schema for ec_search
type SearchOutput struct {
	Memories []types.Memory `json:"memories"`
}

// InvalidateInput defines the input schema for ec_invalidate
type InvalidateInput struct {
	ID           int64 `json:"id" jsonschema:"required" jsonschema_description:"ID of the memory to invalidate"`
	SupersededBy int64 `json:"superseded_by,omitempty" jsonschema_description:"ID of the memory that supersedes this one"`
}

// InvalidateOutput defines the output schema for ec_invalidate
type InvalidateOutput struct {
	Message string `json:"message"`
}

// ListInput defines the input schema for ec_list
type ListInput struct {
	Limit          int    `json:"limit,omitempty" jsonschema_description:"Maximum number of results (default: 10)"`
	Type           string `json:"type,omitempty" jsonschema_description:"Filter by type (decision, learning, or pattern)"`
	Area           string `json:"area,omitempty" jsonschema_description:"Filter by domain area"`
	IncludeInvalid bool   `json:"include_invalid,omitempty" jsonschema_description:"Include invalidated entries (default: false)"`
}

// ListOutput defines the output schema for ec_list
type ListOutput struct {
	Memories []types.Memory `json:"memories"`
}

// TextResult creates a successful MCP result with text content
func TextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// ErrorResult creates an error MCP result
func ErrorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}

// Tool definitions (shared between server and shim)
var (
	AddTool = &mcp.Tool{
		Name:        "ec_add",
		Description: "Add a new memory entry (decision, learning, or pattern)",
	}

	SearchTool = &mcp.Tool{
		Name:        "ec_search",
		Description: "Search memories by semantic similarity",
	}

	InvalidateTool = &mcp.Tool{
		Name:        "ec_invalidate",
		Description: "Invalidate a memory entry (soft delete)",
	}

	ListTool = &mcp.Tool{
		Name:        "ec_list",
		Description: "List recent memory entries",
	}
)
