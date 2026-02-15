package tools

import (
	"context"
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
		memory, err = h.svc.AddWithContext(ctx, service.AddParams{
			Type:      input.Type,
			Area:      input.Area,
			Content:   input.Content,
			Rationale: input.Rationale,
			Repo:      h.repo,
		})
	} else {
		memory, err = h.svc.Add(ctx, types.MemoryType(input.Type), input.Area, input.Content, input.Rationale)
	}

	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to store memory: %v", err)), mcptypes.AddOutput{}, nil
	}

	result, fmtErr := mcptypes.MemoryAddedResult(memory)
	if fmtErr != nil {
		return mcptypes.ErrorResult(fmtErr.Error()), mcptypes.AddOutput{}, nil
	}
	return result, mcptypes.AddOutput{Memory: memory}, nil
}

func (h *Handler) Search(ctx context.Context, req *mcp.CallToolRequest, input mcptypes.SearchInput) (*mcp.CallToolResult, mcptypes.SearchOutput, error) {
	if input.Query == "" {
		return mcptypes.ErrorResult("query is required"), mcptypes.SearchOutput{}, nil
	}

	limit := mcptypes.DefaultSearchLimit(input.Limit)

	var memories []types.Memory
	var err error

	if h.repo != "" {
		memories, err = h.svc.SearchWithRepo(ctx, input.Query, limit, input.Type, input.Area, h.repo)
	} else {
		memories, err = h.svc.Search(ctx, input.Query, limit, types.MemoryType(input.Type), input.Area)
	}

	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to search: %v", err)), mcptypes.SearchOutput{}, nil
	}

	result, fmtErr := mcptypes.MemoriesResult(memories, "No matching memories found.")
	if fmtErr != nil {
		return mcptypes.ErrorResult(fmtErr.Error()), mcptypes.SearchOutput{}, nil
	}
	if memories == nil {
		memories = []types.Memory{}
	}
	return result, mcptypes.SearchOutput{Memories: memories}, nil
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

	msg := mcptypes.InvalidateMsg(input.ID, supersededBy)
	return mcptypes.TextResult(msg), mcptypes.InvalidateOutput{Message: msg}, nil
}

func (h *Handler) List(ctx context.Context, req *mcp.CallToolRequest, input mcptypes.ListInput) (*mcp.CallToolResult, mcptypes.ListOutput, error) {
	limit := mcptypes.DefaultListLimit(input.Limit)

	var memories []types.Memory
	var err error

	if h.repo != "" {
		memories, err = h.svc.ListWithRepo(ctx, limit, input.Type, input.Area, h.repo, input.IncludeInvalid, 0)
	} else {
		memories, err = h.svc.List(ctx, limit, types.MemoryType(input.Type), input.Area, input.IncludeInvalid)
	}

	if err != nil {
		return mcptypes.ErrorResult(fmt.Sprintf("failed to list: %v", err)), mcptypes.ListOutput{}, nil
	}

	result, fmtErr := mcptypes.MemoriesResult(memories, "No memories found.")
	if fmtErr != nil {
		return mcptypes.ErrorResult(fmtErr.Error()), mcptypes.ListOutput{}, nil
	}
	if memories == nil {
		memories = []types.Memory{}
	}
	return result, mcptypes.ListOutput{Memories: memories}, nil
}
