package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MereWhiplash/engram-cogitator/internal/mcptypes"
	"github.com/MereWhiplash/engram-cogitator/internal/service"
	"github.com/MereWhiplash/engram-cogitator/internal/types"
)

// Handler holds dependencies for tool handlers
type Handler struct {
	svc  *service.Service
	repo string // project identity for this session
}

// Register adds all EC tools to the MCP server.
// repo can be empty for backward compatibility (per-project DB mode).
func Register(server *mcp.Server, svc *service.Service) {
	RegisterWithRepo(server, svc, "")
}

// RegisterWithRepo adds all EC tools with project context.
// When repo is set, memories are tagged with the project identity.
func RegisterWithRepo(server *mcp.Server, svc *service.Service, repo string) {
	h := &Handler{svc: svc, repo: repo}

	mcp.AddTool(server, mcptypes.AddTool, h.Add)
	mcp.AddTool(server, mcptypes.SearchTool, h.Search)
	mcp.AddTool(server, mcptypes.InvalidateTool, h.Invalidate)
	mcp.AddTool(server, mcptypes.ListTool, h.List)
}

func (h *Handler) Add(ctx context.Context, req *mcp.CallToolRequest, input mcptypes.AddInput) (*mcp.CallToolResult, mcptypes.AddOutput, error) {
	if input.Type == "" || input.Area == "" || input.Content == "" {
		return mcptypes.ErrorResult("type, area, and content are required"), mcptypes.AddOutput{}, nil
	}

	var memory *types.Memory
	var err error

	if h.repo != "" {
		// Global mode: inject project context
		memory, err = h.svc.AddWithContext(ctx, service.AddParams{
			Type:    input.Type,
			Area:    input.Area,
			Content: input.Content,
			Rationale: input.Rationale,
			Repo:    h.repo,
		})
	} else {
		// Legacy mode: per-project DB, no repo tag
		memory, err = h.svc.Add(ctx, types.MemoryType(input.Type), input.Area, input.Content, input.Rationale)
	}

	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to store memory: %v", err)), mcptypes.AddOutput{}, nil
	}

	result, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to format response: %v", err)), mcptypes.AddOutput{}, nil
	}
	return mcptypes.TextResult(fmt.Sprintf("Memory added successfully:\n%s", string(result))), mcptypes.AddOutput{Memory: memory}, nil
}

func (h *Handler) Search(ctx context.Context, req *mcp.CallToolRequest, input mcptypes.SearchInput) (*mcp.CallToolResult, mcptypes.SearchOutput, error) {
	if input.Query == "" {
		return mcptypes.ErrorResult("query is required"), mcptypes.SearchOutput{}, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}

	var memories []types.Memory
	var err error

	if h.repo != "" {
		// Global mode: filter by current project
		memories, err = h.svc.SearchWithRepo(ctx, input.Query, limit, input.Type, input.Area, h.repo)
	} else {
		// Legacy mode: per-project DB, no filter
		memories, err = h.svc.Search(ctx, input.Query, limit, types.MemoryType(input.Type), input.Area)
	}

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

func (h *Handler) Invalidate(ctx context.Context, req *mcp.CallToolRequest, input mcptypes.InvalidateInput) (*mcp.CallToolResult, mcptypes.InvalidateOutput, error) {
	if input.ID == 0 {
		return mcptypes.ErrorResult("id is required"), mcptypes.InvalidateOutput{}, nil
	}

	var supersededBy *int64
	if input.SupersededBy > 0 {
		supersededBy = &input.SupersededBy
	}

	if err := h.svc.Invalidate(ctx, input.ID, supersededBy); err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to invalidate: %v", err)), mcptypes.InvalidateOutput{}, nil
	}

	msg := fmt.Sprintf("Memory %d has been invalidated.", input.ID)
	if supersededBy != nil {
		msg += fmt.Sprintf(" Superseded by memory %d.", *supersededBy)
	}

	return mcptypes.TextResult(msg), mcptypes.InvalidateOutput{Message: msg}, nil
}

func (h *Handler) List(ctx context.Context, req *mcp.CallToolRequest, input mcptypes.ListInput) (*mcp.CallToolResult, mcptypes.ListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	var memories []types.Memory
	var err error

	if h.repo != "" {
		// Global mode: filter by current project
		memories, err = h.svc.ListWithRepo(ctx, limit, input.Type, input.Area, h.repo, input.IncludeInvalid, 0)
	} else {
		// Legacy mode: per-project DB, no filter
		memories, err = h.svc.List(ctx, limit, types.MemoryType(input.Type), input.Area, input.IncludeInvalid)
	}

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
