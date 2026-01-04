package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/db"
	"github.com/MereWhiplash/engram-cogitator/internal/embed"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Handler holds dependencies for tool handlers
type Handler struct {
	db       *db.DB
	embedder *embed.Client
}

// AddInput defines the input schema for ec_add
type AddInput struct {
	Type      string `json:"type" jsonschema:"required,enum=decision|learning|pattern,description=Type of memory: decision, learning, or pattern"`
	Area      string `json:"area" jsonschema:"required,description=Domain area (e.g. auth, permissions, ui, api)"`
	Content   string `json:"content" jsonschema:"required,description=The actual content to remember"`
	Rationale string `json:"rationale,omitempty" jsonschema:"description=Why this matters or additional context"`
}

// AddOutput defines the output schema for ec_add
type AddOutput struct {
	Memory *db.Memory `json:"memory"`
}

// SearchInput defines the input schema for ec_search
type SearchInput struct {
	Query string `json:"query" jsonschema:"required,description=Search query to find relevant memories"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum number of results (default: 5)"`
	Type  string `json:"type,omitempty" jsonschema:"enum=decision|learning|pattern,description=Filter by type"`
	Area  string `json:"area,omitempty" jsonschema:"description=Filter by domain area"`
}

// SearchOutput defines the output schema for ec_search
type SearchOutput struct {
	Memories []db.Memory `json:"memories"`
}

// InvalidateInput defines the input schema for ec_invalidate
type InvalidateInput struct {
	ID           int64 `json:"id" jsonschema:"required,description=ID of the memory to invalidate"`
	SupersededBy int64 `json:"superseded_by,omitempty" jsonschema:"description=ID of the memory that supersedes this one"`
}

// InvalidateOutput defines the output schema for ec_invalidate
type InvalidateOutput struct {
	Message string `json:"message"`
}

// ListInput defines the input schema for ec_list
type ListInput struct {
	Limit          int    `json:"limit,omitempty" jsonschema:"description=Maximum number of results (default: 10)"`
	Type           string `json:"type,omitempty" jsonschema:"enum=decision|learning|pattern,description=Filter by type"`
	Area           string `json:"area,omitempty" jsonschema:"description=Filter by domain area"`
	IncludeInvalid bool   `json:"include_invalid,omitempty" jsonschema:"description=Include invalidated entries (default: false)"`
}

// ListOutput defines the output schema for ec_list
type ListOutput struct {
	Memories []db.Memory `json:"memories"`
}

// Helper functions for creating tool results
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
func Register(server *mcp.Server, database *db.DB, embedder *embed.Client) {
	h := &Handler{db: database, embedder: embedder}

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

// Add handles the ec_add tool
func (h *Handler) Add(ctx context.Context, req *mcp.CallToolRequest, input AddInput) (*mcp.CallToolResult, AddOutput, error) {
	if input.Type == "" || input.Area == "" || input.Content == "" {
		return errorResult("type, area, and content are required"), AddOutput{}, nil
	}

	// Generate embedding for the content
	textToEmbed := fmt.Sprintf("%s: %s", input.Area, input.Content)
	if input.Rationale != "" {
		textToEmbed += " " + input.Rationale
	}

	embedding, err := h.embedder.EmbedForStorage(textToEmbed)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to generate embedding: %v", err)), AddOutput{}, nil
	}

	// Store in database
	memory, err := h.db.Add(input.Type, input.Area, input.Content, input.Rationale, embedding)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to store memory: %v", err)), AddOutput{}, nil
	}

	result, _ := json.MarshalIndent(memory, "", "  ")
	return textResult(fmt.Sprintf("Memory added successfully:\n%s", string(result))), AddOutput{Memory: memory}, nil
}

// Search handles the ec_search tool
func (h *Handler) Search(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, SearchOutput, error) {
	if input.Query == "" {
		return errorResult("query is required"), SearchOutput{}, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}

	// Generate embedding for the query
	embedding, err := h.embedder.EmbedForSearch(input.Query)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to generate embedding: %v", err)), SearchOutput{}, nil
	}

	// Search database
	memories, err := h.db.Search(embedding, limit, input.Type, input.Area)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to search: %v", err)), SearchOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No matching memories found."), SearchOutput{Memories: []db.Memory{}}, nil
	}

	result, _ := json.MarshalIndent(memories, "", "  ")
	return textResult(string(result)), SearchOutput{Memories: memories}, nil
}

// Invalidate handles the ec_invalidate tool
func (h *Handler) Invalidate(ctx context.Context, req *mcp.CallToolRequest, input InvalidateInput) (*mcp.CallToolResult, InvalidateOutput, error) {
	if input.ID == 0 {
		return errorResult("id is required"), InvalidateOutput{}, nil
	}

	var supersededBy *int64
	if input.SupersededBy > 0 {
		supersededBy = &input.SupersededBy
	}

	if err := h.db.Invalidate(input.ID, supersededBy); err != nil {
		return errorResult(fmt.Sprintf("failed to invalidate: %v", err)), InvalidateOutput{}, nil
	}

	msg := fmt.Sprintf("Memory %d has been invalidated.", input.ID)
	if supersededBy != nil {
		msg += fmt.Sprintf(" Superseded by memory %d.", *supersededBy)
	}

	return textResult(msg), InvalidateOutput{Message: msg}, nil
}

// List handles the ec_list tool
func (h *Handler) List(ctx context.Context, req *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, ListOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	memories, err := h.db.List(limit, input.Type, input.Area, input.IncludeInvalid)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to list: %v", err)), ListOutput{}, nil
	}

	if len(memories) == 0 {
		return textResult("No memories found."), ListOutput{Memories: []db.Memory{}}, nil
	}

	result, _ := json.MarshalIndent(memories, "", "  ")
	return textResult(string(result)), ListOutput{Memories: memories}, nil
}
