// internal/shim/tools.go
package shim

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/client"
	"github.com/MereWhiplash/engram-cogitator/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handler holds shim dependencies
type Handler struct {
	client *client.Client
}

// NewHandler creates a new shim handler
func NewHandler(c *client.Client) *Handler {
	return &Handler{client: c}
}

// Input/Output types (same as tools package)
type AddInput struct {
	Type      string `json:"type" jsonschema:"required"`
	Area      string `json:"area" jsonschema:"required"`
	Content   string `json:"content" jsonschema:"required"`
	Rationale string `json:"rationale,omitempty"`
}

type AddOutput struct {
	Memory *storage.Memory `json:"memory"`
}

type SearchInput struct {
	Query string `json:"query" jsonschema:"required"`
	Limit int    `json:"limit,omitempty"`
	Type  string `json:"type,omitempty"`
	Area  string `json:"area,omitempty"`
}

type SearchOutput struct {
	Memories []storage.Memory `json:"memories"`
}

type InvalidateInput struct {
	ID           int64 `json:"id" jsonschema:"required"`
	SupersededBy int64 `json:"superseded_by,omitempty"`
}

type InvalidateOutput struct {
	Message string `json:"message"`
}

type ListInput struct {
	Limit          int    `json:"limit,omitempty"`
	Type           string `json:"type,omitempty"`
	Area           string `json:"area,omitempty"`
	IncludeInvalid bool   `json:"include_invalid,omitempty"`
}

type ListOutput struct {
	Memories []storage.Memory `json:"memories"`
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		IsError: true,
	}
}

// Register adds all EC tools to the MCP server
func Register(server *mcp.Server, h *Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_add",
		Description: "Add a new memory entry (decision, learning, or pattern)",
	}, h.Add)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_search",
		Description: "Search memories by semantic similarity",
	}, h.Search)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_invalidate",
		Description: "Invalidate a memory entry (soft delete)",
	}, h.Invalidate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ec_list",
		Description: "List recent memory entries",
	}, h.List)
}

func (h *Handler) Add(ctx context.Context, req *mcp.CallToolRequest, input AddInput) (*mcp.CallToolResult, AddOutput, error) {
	if input.Type == "" || input.Area == "" || input.Content == "" {
		return errorResult("type, area, and content are required"), AddOutput{}, nil
	}

	memory, err := h.client.Add(ctx, input.Type, input.Area, input.Content, input.Rationale)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to store memory: %v", err)), AddOutput{}, nil
	}

	result, _ := json.MarshalIndent(memory, "", "  ")
	return textResult(fmt.Sprintf("Memory added successfully:\n%s", string(result))), AddOutput{Memory: memory}, nil
}

func (h *Handler) Search(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
	if input.Query == "" {
		return errorResult("query is required"), SearchOutput{}, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}

	memories, err := h.client.Search(ctx, input.Query, limit, input.Type, input.Area)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to search: %v", err)), SearchOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No matching memories found."), SearchOutput{Memories: []storage.Memory{}}, nil
	}

	result, _ := json.MarshalIndent(memories, "", "  ")
	return textResult(string(result)), SearchOutput{Memories: memories}, nil
}

func (h *Handler) Invalidate(ctx context.Context, req *mcp.CallToolRequest, input InvalidateInput) (*mcp.CallToolResult, InvalidateOutput, error) {
	if input.ID == 0 {
		return errorResult("id is required"), InvalidateOutput{}, nil
	}

	var supersededBy *int64
	if input.SupersededBy > 0 {
		supersededBy = &input.SupersededBy
	}

	if err := h.client.Invalidate(ctx, input.ID, supersededBy); err != nil {
		return errorResult(fmt.Sprintf("failed to invalidate: %v", err)), InvalidateOutput{}, nil
	}

	msg := fmt.Sprintf("Memory %d has been invalidated.", input.ID)
	if supersededBy != nil {
		msg += fmt.Sprintf(" Superseded by memory %d.", *supersededBy)
	}

	return textResult(msg), InvalidateOutput{Message: msg}, nil
}

func (h *Handler) List(ctx context.Context, req *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, ListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	memories, err := h.client.List(ctx, limit, input.Type, input.Area, input.IncludeInvalid)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to list: %v", err)), ListOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No memories found."), ListOutput{Memories: []storage.Memory{}}, nil
	}

	result, _ := json.MarshalIndent(memories, "", "  ")
	return textResult(string(result)), ListOutput{Memories: memories}, nil
}
