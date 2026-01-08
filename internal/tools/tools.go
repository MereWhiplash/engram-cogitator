package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// Handler holds dependencies for tool handlers
type Handler struct {
	svc *service.Service
}

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
func Register(server *mcp.Server, svc *service.Service) {
	h := &Handler{svc: svc}

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

	memory, err := h.svc.Add(ctx, types.MemoryType(input.Type), input.Area, input.Content, input.Rationale)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to store memory: %v", err)), AddOutput{}, nil
	}

	result, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("failed to format response: %v", err)), AddOutput{}, nil
	}
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

	memories, err := h.svc.Search(ctx, input.Query, limit, types.MemoryType(input.Type), input.Area)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to search: %v", err)), SearchOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No matching memories found."), SearchOutput{Memories: []types.Memory{}}, nil
	}

	result, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("failed to format response: %v", err)), SearchOutput{}, nil
	}
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

	if err := h.svc.Invalidate(ctx, input.ID, supersededBy); err != nil {
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

	memories, err := h.svc.List(ctx, limit, types.MemoryType(input.Type), input.Area, input.IncludeInvalid)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to list: %v", err)), ListOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No memories found."), ListOutput{Memories: []types.Memory{}}, nil
	}

	result, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("failed to format response: %v", err)), ListOutput{}, nil
	}
	return textResult(string(result)), ListOutput{Memories: memories}, nil
}
