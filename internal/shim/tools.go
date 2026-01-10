// internal/shim/tools.go
package shim

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/mcptypes"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// APIClient defines the interface for the central API client
type APIClient interface {
	Add(ctx context.Context, memType, area, content, rationale string) (*types.Memory, error)
	Search(ctx context.Context, query string, limit int, memType, area string) ([]types.Memory, error)
	List(ctx context.Context, limit int, memType, area string, includeInvalid bool) ([]types.Memory, error)
	Invalidate(ctx context.Context, id int64, supersededBy *int64) error
}

// Handler holds shim dependencies
type Handler struct {
	client APIClient
}

// NewHandler creates a new shim handler
func NewHandler(c APIClient) *Handler {
	return &Handler{client: c}
}

// Register adds all EC tools to the MCP server
func Register(server *mcp.Server, h *Handler) {
	mcp.AddTool(server, mcptypes.AddTool, h.Add)
	mcp.AddTool(server, mcptypes.SearchTool, h.Search)
	mcp.AddTool(server, mcptypes.InvalidateTool, h.Invalidate)
	mcp.AddTool(server, mcptypes.ListTool, h.List)
}

func (h *Handler) Add(ctx context.Context, _ *mcp.CallToolRequest, input mcptypes.AddInput) (*mcp.CallToolResult, mcptypes.AddOutput, error) {
	if input.Type == "" || input.Area == "" || input.Content == "" {
		return mcptypes.ErrorResult("type, area, and content are required"), mcptypes.AddOutput{}, nil
	}

	memory, err := h.client.Add(ctx, input.Type, input.Area, input.Content, input.Rationale)
	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to store memory: %v", err)), mcptypes.AddOutput{}, nil
	}

	result, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to format response: %v", err)), mcptypes.AddOutput{}, nil
	}
	return mcptypes.TextResult(fmt.Sprintf("Memory added successfully:\n%s", string(result))), mcptypes.AddOutput{Memory: memory}, nil
}

func (h *Handler) Search(ctx context.Context, _ *mcp.CallToolRequest, input mcptypes.SearchInput) (*mcp.CallToolResult, mcptypes.SearchOutput, error) {
	if input.Query == "" {
		return mcptypes.ErrorResult("query is required"), mcptypes.SearchOutput{}, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}

	memories, err := h.client.Search(ctx, input.Query, limit, input.Type, input.Area)
	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to search: %v", err)), mcptypes.SearchOutput{}, nil
	}

	if len(memories) == 0 {
		return mcptypes.TextResult("No matching memories found."), mcptypes.SearchOutput{Memories: []types.Memory{}}, nil
	}

	result, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to format response: %v", err)), mcptypes.SearchOutput{}, nil
	}
	return mcptypes.TextResult(string(result)), mcptypes.SearchOutput{Memories: memories}, nil
}

func (h *Handler) Invalidate(ctx context.Context, _ *mcp.CallToolRequest, input mcptypes.InvalidateInput) (*mcp.CallToolResult, mcptypes.InvalidateOutput, error) {
	if input.ID == 0 {
		return mcptypes.ErrorResult("id is required"), mcptypes.InvalidateOutput{}, nil
	}

	var supersededBy *int64
	if input.SupersededBy > 0 {
		supersededBy = &input.SupersededBy
	}

	if err := h.client.Invalidate(ctx, input.ID, supersededBy); err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to invalidate: %v", err)), mcptypes.InvalidateOutput{}, nil
	}

	msg := fmt.Sprintf("Memory %d has been invalidated.", input.ID)
	if supersededBy != nil {
		msg += fmt.Sprintf(" Superseded by memory %d.", *supersededBy)
	}

	return mcptypes.TextResult(msg), mcptypes.InvalidateOutput{Message: msg}, nil
}

func (h *Handler) List(ctx context.Context, _ *mcp.CallToolRequest, input mcptypes.ListInput) (*mcp.CallToolResult, mcptypes.ListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	memories, err := h.client.List(ctx, limit, input.Type, input.Area, input.IncludeInvalid)
	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to list: %v", err)), mcptypes.ListOutput{}, nil
	}

	if len(memories) == 0 {
		return mcptypes.TextResult("No memories found."), mcptypes.ListOutput{Memories: []types.Memory{}}, nil
	}

	result, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to format response: %v", err)), mcptypes.ListOutput{}, nil
	}
	return mcptypes.TextResult(string(result)), mcptypes.ListOutput{Memories: memories}, nil
}
