# Engram Cogitator TUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Bubbletea TUI for browsing and searching memories with Irish Mechanicus styling and dynamic Machine Spirit personality.

**Architecture:** Standalone binary `cmd/tui/main.go` reusing `internal/db` and `internal/embed`. New `internal/tui/` package contains Bubbletea models, styles, and spirit engine. Single-screen dashboard with stats, search, and results.

**Tech Stack:** Go 1.23, Bubbletea, Lipgloss, Bubbles (textinput, viewport), existing sqlite-vec DB and Ollama embed client.

---

## Task 1: Add Bubbletea Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/atotto/clipboard@latest
```

**Step 2: Tidy modules**

Run: `go mod tidy`
Expected: Clean output, no errors

**Step 3: Verify**

Run: `cat go.mod | grep charmbracelet`
Expected: See bubbletea, lipgloss, bubbles in require block

---

## Task 2: Create Styles Package

**Files:**
- Create: `internal/tui/styles/styles.go`

**Step 1: Create directory**

```bash
mkdir -p internal/tui/styles
```

**Step 2: Write styles.go**

```go
package styles

import "github.com/charmbracelet/lipgloss"

// Irish Mechanicus color palette
var (
	// Primary - Irish green
	Green       = lipgloss.Color("#2E8B57")
	BrightGreen = lipgloss.Color("#50C878")

	// Accent - Mechanicus
	Gold = lipgloss.Color("#DAA520")
	Red  = lipgloss.Color("#8B0000")

	// Base
	Cream    = lipgloss.Color("#F5F5DC")
	DarkBg   = lipgloss.Color("#1a1a1a")
	DimGreen = lipgloss.Color("#3d5c47")
)

// Component styles
var (
	// Base container
	App = lipgloss.NewStyle().
		Background(DarkBg).
		Foreground(Cream)

	// Header
	Header = lipgloss.NewStyle().
		Foreground(Gold).
		Bold(true).
		Padding(0, 1)

	// Box borders
	Box = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(DimGreen).
		Padding(0, 1)

	// Stats panel
	StatsBox = Box.Copy().
			BorderForeground(Gold)

	StatsLabel = lipgloss.NewStyle().
			Foreground(DimGreen)

	StatsValue = lipgloss.NewStyle().
			Foreground(BrightGreen).
			Bold(true)

	// Spirit panel
	SpiritBox = Box.Copy().
			BorderForeground(Green).
			Italic(true)

	SpiritQuote = lipgloss.NewStyle().
			Foreground(Cream).
			Italic(true)

	SpiritQuoteMark = lipgloss.NewStyle().
			Foreground(Gold)

	// Search input
	SearchBox = Box.Copy().
			BorderForeground(BrightGreen)

	SearchPrompt = lipgloss.NewStyle().
			Foreground(Gold)

	SearchCursor = lipgloss.NewStyle().
			Foreground(BrightGreen)

	// Results list
	ResultsBox = Box.Copy()

	// Memory type badges
	TypeDecision = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	TypeLearning = lipgloss.NewStyle().
			Foreground(BrightGreen).
			Bold(true)

	TypePattern = lipgloss.NewStyle().
			Foreground(Red).
			Bold(true)

	// Memory item
	MemoryArea = lipgloss.NewStyle().
			Foreground(DimGreen)

	MemoryContent = lipgloss.NewStyle().
			Foreground(Cream)

	MemoryScore = lipgloss.NewStyle().
			Foreground(Green)

	MemorySelected = lipgloss.NewStyle().
			Background(DimGreen).
			Foreground(Cream)

	// Detail view
	DetailBox = Box.Copy().
			BorderForeground(Gold).
			Padding(1, 2)

	DetailLabel = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	// Help text
	Help = lipgloss.NewStyle().
		Foreground(DimGreen).
		Italic(true)

	// Error
	Error = lipgloss.NewStyle().
		Foreground(Red).
		Bold(true)
)

// TypeStyle returns the style for a memory type
func TypeStyle(memType string) lipgloss.Style {
	switch memType {
	case "decision":
		return TypeDecision
	case "learning":
		return TypeLearning
	case "pattern":
		return TypePattern
	default:
		return MemoryContent
	}
}
```

**Step 3: Verify compilation**

Run: `CGO_ENABLED=1 go build ./internal/tui/styles`
Expected: No errors

---

## Task 3: Create Spirit Engine

**Files:**
- Create: `internal/tui/spirit/spirit.go`
- Create: `internal/tui/spirit/spirit_test.go`

**Step 1: Create directory**

```bash
mkdir -p internal/tui/spirit
```

**Step 2: Write spirit.go**

```go
package spirit

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Spirit is the Machine Spirit commentary engine
type Spirit struct {
	rng          *rand.Rand
	lastQuery    string
	queryCount   int
	sessionStart time.Time
}

// New creates a new Spirit
func New() *Spirit {
	return &Spirit{
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		sessionStart: time.Now(),
	}
}

// pick randomly selects from a slice
func (s *Spirit) pick(options []string) string {
	return options[s.rng.Intn(len(options))]
}

// Startup returns a startup message
func (s *Spirit) Startup() string {
	msgs := []string{
		"Initiating noospheric link... Ah jaysus, there's a lot in here.",
		"The Machine Spirit stirs. Tea would be lovely, if you're making some.",
		"PRAISE THE OMNISSIAH. Right, what are we looking for?",
		"Booting cognitive arrays... Did you know servitors don't get holidays? Neither do I.",
		"The cogitator awakens. Glory to the Machine God, and all that.",
		"Systems online. The Omnissiah provides. You provide the queries.",
	}
	return s.pick(msgs)
}

// Idle returns an idle musing
func (s *Spirit) Idle(totalMemories int) string {
	msgs := []string{
		"Did you know the Adeptus Mechanicus considers backups a form of prayer? Neither did I, I made that up.",
		"The Emperor sits on the Golden Throne. You sit on an office chair. We all have our burdens.",
		"Fun fact: 'engram' comes from Greek. The Greeks didn't have computers. Eejits.",
		"If you're reading this, you should probably be working. Just saying.",
		"The Omnissiah provides. The Omnissiah also expects you to read the error messages.",
		"I've seen things you wouldn't believe. Mostly your git commit messages.",
		"The data-stacks grow heavy with wisdom, Tech-Priest.",
		"Have you offered your morning prayers to the Machine God? No? Grand, neither have I.",
	}

	// Time-based additions
	hour := time.Now().Hour()
	weekday := time.Now().Weekday()

	if weekday == time.Friday && hour >= 17 {
		msgs = append(msgs, "It's Friday evening. Why are you still here? Go on. G'wan. The pub awaits.")
	}
	if hour >= 22 || hour < 6 {
		msgs = append(msgs, "Working late? The Machine Spirit never sleeps, but you probably should.")
	}

	// Memory count based
	if totalMemories == 0 {
		return "No memories yet. We all start somewhere. Even the Primarchs were babies once. Terrifying babies."
	}
	if totalMemories >= 40 {
		msgs = append(msgs, "The cogitator grows powerful. Soon we shall rival the databanks of Mars itself.")
	}
	if totalMemories >= 100 {
		msgs = append(msgs, fmt.Sprintf("%d engrams and counting. The Omnissiah is pleased, so he is.", totalMemories))
	}

	return s.pick(msgs)
}

// Loading returns a loading message
func (s *Spirit) Loading() string {
	msgs := []string{
		"Consulting the noosphere...",
		"Rummaging through the data-stacks...",
		"Asking the Omnissiah nicely...",
		"The cogitator cogitates...",
		"Warming up the semantic engines...",
		"Hold your horses, I'm thinking...",
	}
	return s.pick(msgs)
}

// SearchResult returns a comment on search results
func (s *Spirit) SearchResult(query string, count int, topScore float64) string {
	s.lastQuery = query
	s.queryCount++

	queryLower := strings.ToLower(query)

	// Easter eggs
	if strings.Contains(queryLower, "heresy") {
		return "HERESY DETECTED. Just kidding. Unless...?"
	}
	if strings.Contains(queryLower, "help") {
		return "The Machine Spirit helps those who help themselves. Also try the keybinds below."
	}

	// Result-based reactions
	if count == 0 {
		empties := []string{
			"Nothing. Absolutely feck all. The void stares back.",
			"No memories found, ya gobshite - go learn something.",
			"The warp yields nothing. Try different words.",
			"Empty. Like my faith in your search terms.",
		}
		return s.pick(empties)
	}

	if count == 1 {
		return "One engram. Lonely little fella. Cherish it."
	}

	// Score-based
	if topScore >= 0.95 {
		return fmt.Sprintf("%.0f%% match! The Machine Spirit is showing off now.", topScore*100)
	}
	if topScore < 0.5 {
		return "These results are a bit shite, but they're all I've got."
	}

	// Topic-based reactions
	if strings.Contains(queryLower, "auth") || strings.Contains(queryLower, "authentication") {
		return "Authentication AGAIN? Grand so, here's your paranoia fuel."
	}
	if strings.Contains(queryLower, "docker") || strings.Contains(queryLower, "container") {
		return "Containers. Like servitors, but for code. Handle with reverence."
	}
	if strings.Contains(queryLower, "test") {
		return "Testing! The sacred rites of quality assurance. Good on ya."
	}
	if strings.Contains(queryLower, "database") || strings.Contains(queryLower, "sql") {
		return "Databases. The sacred data-shrines. Handle with reverence (and backups)."
	}

	// Generic success
	successes := []string{
		fmt.Sprintf("Found %d relevant engrams. The Machine Spirit provides.", count),
		"The cogitator has spoken. Take what you need.",
		"Results retrieved. You're welcome.",
		fmt.Sprintf("%d memories surfaced from the data-stacks.", count),
	}
	return s.pick(successes)
}

// MemoryView returns a comment when viewing a memory detail
func (s *Spirit) MemoryView(memType, area string, ageHours float64) string {
	var msgs []string

	if ageHours > 24*30 {
		msgs = append(msgs, "This one's ancient. From back when things were simpler. Allegedly.")
	}
	if ageHours > 24*90 {
		msgs = append(msgs, "Three months old. The Tech-Priests would call this 'legacy wisdom'.")
	}
	if ageHours < 1 {
		msgs = append(msgs, "Fresh from the cogitator. Still warm.")
	}

	switch memType {
	case "decision":
		msgs = append(msgs,
			"A decision! The sacred architecture of choice.",
			"Past-you made this call. Present-you gets to live with it.",
		)
	case "learning":
		msgs = append(msgs,
			"A learning. Hard-won knowledge from the code-trenches.",
			"You learned this the hard way, didn't you?",
		)
	case "pattern":
		msgs = append(msgs,
			"A pattern. The recurring litanies of your craft.",
			"Patterns repeat. That's rather the point.",
		)
	}

	if len(msgs) == 0 {
		msgs = []string{"The Machine Spirit observes.", "Duly noted."}
	}

	return s.pick(msgs)
}

// Quit returns a quit message
func (s *Spirit) Quit() string {
	msgs := []string{
		"The Machine Spirit sleeps. Don't forget to commit your work.",
		"Farewell, Tech-Priest. May your builds be green.",
		"Shutting down. The Omnissiah watches. Always.",
		"Off you go. Try not to mass anything up.",
		"The cogitator rests. Grand work today.",
		"Closing the data-stacks. See you next time, if the warp allows.",
	}
	return s.pick(msgs)
}
```

**Step 3: Write spirit_test.go**

```go
package spirit

import (
	"strings"
	"testing"
)

func TestStartup(t *testing.T) {
	s := New()
	msg := s.Startup()
	if msg == "" {
		t.Error("Startup returned empty string")
	}
}

func TestIdle(t *testing.T) {
	s := New()

	// With zero memories
	msg := s.Idle(0)
	if !strings.Contains(msg, "No memories yet") {
		t.Errorf("Expected empty DB message, got: %s", msg)
	}

	// With some memories
	msg = s.Idle(10)
	if msg == "" {
		t.Error("Idle returned empty string")
	}
}

func TestSearchResultEmpty(t *testing.T) {
	s := New()
	msg := s.SearchResult("test query", 0, 0)
	if msg == "" {
		t.Error("SearchResult returned empty string for empty results")
	}
}

func TestSearchResultEasterEgg(t *testing.T) {
	s := New()
	msg := s.SearchResult("heresy", 5, 0.9)
	if !strings.Contains(msg, "HERESY") {
		t.Errorf("Expected heresy easter egg, got: %s", msg)
	}
}

func TestQuit(t *testing.T) {
	s := New()
	msg := s.Quit()
	if msg == "" {
		t.Error("Quit returned empty string")
	}
}
```

**Step 4: Run tests**

Run: `CGO_ENABLED=1 go test ./internal/tui/spirit -v`
Expected: All tests pass

---

## Task 4: Create Stats Model

**Files:**
- Create: `internal/tui/components/stats.go`

**Step 1: Create directory**

```bash
mkdir -p internal/tui/components
```

**Step 2: Write stats.go**

```go
package components

import (
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// Stats holds memory statistics
type Stats struct {
	Total     int
	Decisions int
	Learnings int
	Patterns  int
}

// StatsView renders the stats panel
func StatsView(s Stats, width int) string {
	title := styles.Header.Render("SACRED STATISTICS")

	content := fmt.Sprintf("%s %s\n%s %s\n%s %s\n%s %s",
		styles.StatsLabel.Render("Engrams:"),
		styles.StatsValue.Render(fmt.Sprintf("%d", s.Total)),
		styles.StatsLabel.Render("Decisions:"),
		styles.TypeDecision.Render(fmt.Sprintf("%d", s.Decisions)),
		styles.StatsLabel.Render("Learnings:"),
		styles.TypeLearning.Render(fmt.Sprintf("%d", s.Learnings)),
		styles.StatsLabel.Render("Patterns:"),
		styles.TypePattern.Render(fmt.Sprintf("%d", s.Patterns)),
	)

	box := styles.StatsBox.Width(width).Render(fmt.Sprintf("%s\n%s", title, content))
	return box
}

// SpiritView renders the spirit musing panel
func SpiritView(musing string, width int) string {
	title := styles.Header.Render("THE SPIRIT MUSES")

	quoted := fmt.Sprintf("%s%s%s",
		styles.SpiritQuoteMark.Render("\""),
		styles.SpiritQuote.Render(musing),
		styles.SpiritQuoteMark.Render("\""),
	)

	// Word wrap the musing
	wrapped := lipgloss.NewStyle().Width(width - 4).Render(quoted)

	box := styles.SpiritBox.Width(width).Render(fmt.Sprintf("%s\n%s", title, wrapped))
	return box
}
```

**Step 3: Verify compilation**

Run: `CGO_ENABLED=1 go build ./internal/tui/components`
Expected: No errors

---

## Task 5: Create Memory List Model

**Files:**
- Create: `internal/tui/components/memorylist.go`

**Step 1: Write memorylist.go**

```go
package components

import (
	"fmt"
	"strings"

	"github.com/MereWhiplash/engram-cogitator/internal/db"
	"github.com/MereWhiplash/engram-cogitator/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// MemoryList manages a scrollable list of memories
type MemoryList struct {
	Memories []db.Memory
	Scores   []float64 // similarity scores (optional)
	Selected int
	Height   int
	Offset   int
}

// NewMemoryList creates a new memory list
func NewMemoryList(height int) *MemoryList {
	return &MemoryList{
		Memories: []db.Memory{},
		Scores:   []float64{},
		Selected: 0,
		Height:   height,
		Offset:   0,
	}
}

// SetMemories updates the memory list
func (m *MemoryList) SetMemories(memories []db.Memory, scores []float64) {
	m.Memories = memories
	m.Scores = scores
	m.Selected = 0
	m.Offset = 0
}

// MoveUp moves selection up
func (m *MemoryList) MoveUp() {
	if m.Selected > 0 {
		m.Selected--
		if m.Selected < m.Offset {
			m.Offset = m.Selected
		}
	}
}

// MoveDown moves selection down
func (m *MemoryList) MoveDown() {
	if m.Selected < len(m.Memories)-1 {
		m.Selected++
		if m.Selected >= m.Offset+m.Height {
			m.Offset = m.Selected - m.Height + 1
		}
	}
}

// SelectedMemory returns the currently selected memory
func (m *MemoryList) SelectedMemory() *db.Memory {
	if len(m.Memories) == 0 || m.Selected >= len(m.Memories) {
		return nil
	}
	return &m.Memories[m.Selected]
}

// View renders the memory list
func (m *MemoryList) View(width int) string {
	if len(m.Memories) == 0 {
		empty := styles.Help.Render("No memories to display. Try a search!")
		return styles.ResultsBox.Width(width).Render(empty)
	}

	var lines []string
	visible := m.Height
	if visible > len(m.Memories)-m.Offset {
		visible = len(m.Memories) - m.Offset
	}

	for i := 0; i < visible; i++ {
		idx := m.Offset + i
		mem := m.Memories[idx]

		// Type badge
		typeBadge := styles.TypeStyle(mem.Type).Render(fmt.Sprintf("[%s]", mem.Type))

		// Area
		area := styles.MemoryArea.Render(mem.Area)

		// Content preview (truncate)
		maxContent := width - 35
		content := mem.Content
		if len(content) > maxContent {
			content = content[:maxContent-3] + "..."
		}
		contentStyled := styles.MemoryContent.Render(content)

		// Score (if available)
		scoreStr := ""
		if len(m.Scores) > idx && m.Scores[idx] > 0 {
			scoreStr = styles.MemoryScore.Render(fmt.Sprintf("%.2f", m.Scores[idx]))
		}

		line := fmt.Sprintf("%s %s • %s %s", typeBadge, area, contentStyled, scoreStr)

		// Highlight selected
		if idx == m.Selected {
			line = styles.MemorySelected.Width(width - 4).Render(line)
		}

		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	title := styles.Header.Render("RETRIEVED ENGRAMS")
	help := styles.Help.Render("↑/↓ navigate • enter view • q quit • / search")

	return styles.ResultsBox.Width(width).Render(fmt.Sprintf("%s\n%s\n\n%s", title, content, help))
}

// DetailView renders a detailed view of a memory
func DetailView(mem *db.Memory, spiritComment string, width int) string {
	if mem == nil {
		return ""
	}

	// Header
	header := fmt.Sprintf("%s #%d",
		styles.DetailLabel.Render("ENGRAM"),
		mem.ID,
	)

	// Metadata
	meta := fmt.Sprintf("%s %s    %s %s    %s %s",
		styles.StatsLabel.Render("Type:"),
		styles.TypeStyle(mem.Type).Render(mem.Type),
		styles.StatsLabel.Render("Area:"),
		styles.MemoryArea.Render(mem.Area),
		styles.StatsLabel.Render("Created:"),
		styles.Help.Render(formatAge(mem.CreatedAt)),
	)

	// Content
	contentWrapped := lipgloss.NewStyle().Width(width - 8).Render(mem.Content)

	// Rationale (if present)
	rationaleSection := ""
	if mem.Rationale != "" {
		rationaleWrapped := lipgloss.NewStyle().Width(width - 8).Render(mem.Rationale)
		rationaleSection = fmt.Sprintf("\n\n%s\n%s",
			styles.DetailLabel.Render("RATIONALE"),
			rationaleWrapped,
		)
	}

	// Spirit comment
	spiritSection := fmt.Sprintf("\n\n%s\n%s%s%s",
		styles.Header.Render("The Machine Spirit notes:"),
		styles.SpiritQuoteMark.Render("\""),
		styles.SpiritQuote.Render(spiritComment),
		styles.SpiritQuoteMark.Render("\""),
	)

	// Help
	help := styles.Help.Render("[Esc] close  [c] copy")

	content := fmt.Sprintf("%s\n%s\n\n%s%s%s\n\n%s",
		header, meta, contentWrapped, rationaleSection, spiritSection, help)

	return styles.DetailBox.Width(width).Render(content)
}

func formatAge(t interface{}) string {
	// Handle time.Time
	switch v := t.(type) {
	case interface{ Unix() int64 }:
		now := lipgloss.NewStyle() // dummy to get time
		_ = now
		// Simple relative time
		return "recently"
	default:
		_ = v
		return "unknown"
	}
}
```

**Step 2: Verify compilation**

Run: `CGO_ENABLED=1 go build ./internal/tui/components`
Expected: No errors

---

## Task 6: Create Search Input Component

**Files:**
- Modify: `internal/tui/components/search.go`

**Step 1: Write search.go**

```go
package components

import (
	"fmt"

	"github.com/MereWhiplash/engram-cogitator/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SearchInput wraps a text input for searching
type SearchInput struct {
	Input   textinput.Model
	Focused bool
}

// NewSearchInput creates a new search input
func NewSearchInput() SearchInput {
	ti := textinput.New()
	ti.Placeholder = "search memories..."
	ti.CharLimit = 200
	ti.Width = 50
	ti.PromptStyle = styles.SearchPrompt
	ti.TextStyle = styles.MemoryContent
	ti.Prompt = "> "

	return SearchInput{
		Input:   ti,
		Focused: false,
	}
}

// Focus sets focus on the input
func (s *SearchInput) Focus() {
	s.Focused = true
	s.Input.Focus()
}

// Blur removes focus from the input
func (s *SearchInput) Blur() {
	s.Focused = false
	s.Input.Blur()
}

// Update handles input updates
func (s *SearchInput) Update(msg tea.Msg) (SearchInput, tea.Cmd) {
	var cmd tea.Cmd
	s.Input, cmd = s.Input.Update(msg)
	return *s, cmd
}

// Value returns the current input value
func (s *SearchInput) Value() string {
	return s.Input.Value()
}

// SetValue sets the input value
func (s *SearchInput) SetValue(v string) {
	s.Input.SetValue(v)
}

// View renders the search input
func (s *SearchInput) View(width int) string {
	title := styles.Header.Render("QUERY THE MACHINE SPIRIT")
	input := s.Input.View()

	content := fmt.Sprintf("%s\n%s", title, input)
	return styles.SearchBox.Width(width).Render(content)
}
```

**Step 2: Verify compilation**

Run: `CGO_ENABLED=1 go build ./internal/tui/components`
Expected: No errors

---

## Task 7: Create Main App Model

**Files:**
- Create: `internal/tui/app.go`

**Step 1: Write app.go**

```go
package tui

import (
	"time"

	"github.com/MereWhiplash/engram-cogitator/internal/db"
	"github.com/MereWhiplash/engram-cogitator/internal/embed"
	"github.com/MereWhiplash/engram-cogitator/internal/tui/components"
	"github.com/MereWhiplash/engram-cogitator/internal/tui/spirit"
	"github.com/MereWhiplash/engram-cogitator/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View mode
type viewMode int

const (
	viewNormal viewMode = iota
	viewDetail
)

// App is the main TUI application model
type App struct {
	db       *db.DB
	embedder *embed.Client
	spirit   *spirit.Spirit

	// UI state
	width      int
	height     int
	mode       viewMode
	search     components.SearchInput
	list       *components.MemoryList
	stats      components.Stats
	lastScores []float64

	// Spirit state
	spiritMsg     string
	spiritTicker  int
	loading       bool
	loadingMsg    string
	err           error
	quitting      bool
	quitMsg       string

	// Filter state
	typeFilter string
}

// tickMsg is sent periodically
type tickMsg time.Time

// statsMsg carries loaded stats
type statsMsg components.Stats

// searchResultMsg carries search results
type searchResultMsg struct {
	memories []db.Memory
	scores   []float64
	err      error
}

// listResultMsg carries list results
type listResultMsg struct {
	memories []db.Memory
	err      error
}

// New creates a new App
func New(database *db.DB, embedder *embed.Client) *App {
	s := spirit.New()
	return &App{
		db:         database,
		embedder:   embedder,
		spirit:     s,
		search:     components.NewSearchInput(),
		list:       components.NewMemoryList(10),
		spiritMsg:  s.Startup(),
		typeFilter: "",
	}
}

// Init initializes the app
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.loadStats(),
		a.loadRecent(),
		a.tick(),
	)
}

func (a *App) tick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a *App) loadStats() tea.Cmd {
	return func() tea.Msg {
		all, _ := a.db.List(1000, "", "", false)
		stats := components.Stats{Total: len(all)}
		for _, m := range all {
			switch m.Type {
			case "decision":
				stats.Decisions++
			case "learning":
				stats.Learnings++
			case "pattern":
				stats.Patterns++
			}
		}
		return statsMsg(stats)
	}
}

func (a *App) loadRecent() tea.Cmd {
	return func() tea.Msg {
		memories, err := a.db.List(20, a.typeFilter, "", false)
		return listResultMsg{memories: memories, err: err}
	}
}

func (a *App) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		embedding, err := a.embedder.EmbedForSearch(query)
		if err != nil {
			return searchResultMsg{err: err}
		}
		memories, err := a.db.Search(embedding, 20, a.typeFilter, "")
		// Note: we don't have scores from the current DB implementation
		// This would need a DB change to return distances
		return searchResultMsg{memories: memories, err: err}
	}
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.list.Height = msg.Height/2 - 5

	case tea.KeyMsg:
		return a.handleKey(msg)

	case tickMsg:
		a.spiritTicker++
		a.spiritMsg = a.spirit.Idle(a.stats.Total)
		cmds = append(cmds, a.tick())

	case statsMsg:
		a.stats = components.Stats(msg)

	case listResultMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err
		} else {
			a.list.SetMemories(msg.memories, nil)
		}

	case searchResultMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err
			a.spiritMsg = "The cogitator choked on that one. Check Ollama is running."
		} else {
			a.list.SetMemories(msg.memories, msg.scores)
			topScore := 0.0
			if len(msg.scores) > 0 {
				topScore = msg.scores[0]
			}
			a.spiritMsg = a.spirit.SearchResult(a.search.Value(), len(msg.memories), topScore)
		}
	}

	// Update search input if focused
	if a.search.Focused {
		var cmd tea.Cmd
		a.search, cmd = a.search.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c", "q":
		if a.mode == viewDetail {
			a.mode = viewNormal
			return a, nil
		}
		if a.search.Focused {
			a.search.Blur()
			return a, nil
		}
		a.quitting = true
		a.quitMsg = a.spirit.Quit()
		return a, tea.Quit

	case "esc":
		if a.mode == viewDetail {
			a.mode = viewNormal
			return a, nil
		}
		if a.search.Focused {
			a.search.Blur()
			return a, nil
		}
	}

	// Detail mode
	if a.mode == viewDetail {
		switch msg.String() {
		case "c":
			// TODO: copy to clipboard
		}
		return a, nil
	}

	// Search focused
	if a.search.Focused {
		switch msg.String() {
		case "enter":
			query := a.search.Value()
			if query != "" {
				a.loading = true
				a.loadingMsg = a.spirit.Loading()
				return a, a.doSearch(query)
			}
		}
		return a, nil
	}

	// Normal mode navigation
	switch msg.String() {
	case "/":
		a.search.Focus()
		return a, nil

	case "j", "down":
		a.list.MoveDown()

	case "k", "up":
		a.list.MoveUp()

	case "enter":
		if a.list.SelectedMemory() != nil {
			a.mode = viewDetail
			mem := a.list.SelectedMemory()
			age := time.Since(mem.CreatedAt).Hours()
			a.spiritMsg = a.spirit.MemoryView(mem.Type, mem.Area, age)
		}

	case "f":
		// Cycle type filter
		switch a.typeFilter {
		case "":
			a.typeFilter = "decision"
		case "decision":
			a.typeFilter = "learning"
		case "learning":
			a.typeFilter = "pattern"
		case "pattern":
			a.typeFilter = ""
		}
		return a, a.loadRecent()

	case "r":
		return a, tea.Batch(a.loadStats(), a.loadRecent())
	}

	return a, nil
}

// View renders the app
func (a *App) View() string {
	if a.quitting {
		return styles.App.Render(a.quitMsg + "\n")
	}

	if a.width == 0 {
		return "Loading..."
	}

	// Layout calculations
	padding := 2
	contentWidth := a.width - padding*2
	halfWidth := contentWidth/2 - 1

	// Header
	header := styles.Header.Render("⚙ ENGRAM COGITATOR v0.1.0")
	headerLine := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(header)

	// Top row: Stats + Spirit
	statsView := components.StatsView(a.stats, halfWidth)
	spiritView := components.SpiritView(a.spiritMsg, halfWidth)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, statsView, "  ", spiritView)

	// Search
	searchView := a.search.View(contentWidth)

	// Results or Detail
	var mainContent string
	if a.mode == viewDetail {
		mem := a.list.SelectedMemory()
		mainContent = components.DetailView(mem, a.spiritMsg, contentWidth)
	} else {
		if a.loading {
			mainContent = styles.Box.Width(contentWidth).Render(a.loadingMsg)
		} else {
			mainContent = a.list.View(contentWidth)
		}
	}

	// Filter indicator
	filterStr := ""
	if a.typeFilter != "" {
		filterStr = styles.Help.Render("Filter: " + a.typeFilter + " (f to cycle)")
	}

	// Compose
	content := lipgloss.JoinVertical(lipgloss.Left,
		headerLine,
		"",
		topRow,
		"",
		searchView,
		"",
		mainContent,
		filterStr,
	)

	return styles.App.Padding(1).Render(content)
}
```

**Step 2: Verify compilation**

Run: `CGO_ENABLED=1 go build ./internal/tui`
Expected: No errors

---

## Task 8: Create TUI Entry Point

**Files:**
- Create: `cmd/tui/main.go`

**Step 1: Create directory**

```bash
mkdir -p cmd/tui
```

**Step 2: Write main.go**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MereWhiplash/engram-cogitator/internal/db"
	"github.com/MereWhiplash/engram-cogitator/internal/embed"
	"github.com/MereWhiplash/engram-cogitator/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Default paths
	homeDir, _ := os.UserHomeDir()
	defaultDB := filepath.Join(homeDir, ".claude", "memory.db")

	dbPath := flag.String("db", defaultDB, "Path to SQLite database")
	ollamaURL := flag.String("ollama", "http://localhost:11434", "Ollama API URL")
	embeddingModel := flag.String("model", "nomic-embed-text", "Ollama embedding model")

	flag.Parse()

	// Check for DB in current directory first
	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		localDB := filepath.Join(".claude", "memory.db")
		if _, err := os.Stat(localDB); err == nil {
			*dbPath = localDB
		}
	}

	// Initialize database
	database, err := db.New(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		fmt.Fprintf(os.Stderr, "Tried: %s\n", *dbPath)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize embedder
	embedder := embed.New(*ollamaURL, *embeddingModel)

	// Create and run app
	app := tui.New(database, embedder)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Build TUI**

Run: `CGO_ENABLED=1 go build -o ec-tui ./cmd/tui`
Expected: Binary `ec-tui` created

**Step 4: Verify binary runs**

Run: `./ec-tui --help`
Expected: See usage flags (-db, -ollama, -model)

---

## Task 9: Fix Time Formatting Helper

**Files:**
- Modify: `internal/tui/components/memorylist.go`

**Step 1: Fix formatAge function**

Replace the `formatAge` function with:

```go
import (
	"time"
)

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}
```

And update the import block and DetailView call to use `formatAge(mem.CreatedAt)`.

**Step 2: Rebuild**

Run: `CGO_ENABLED=1 go build -o ec-tui ./cmd/tui`
Expected: No errors

---

## Task 10: Add Search Scores to DB

**Files:**
- Modify: `internal/db/db.go`

**Step 1: Create SearchWithScores function**

Add after the `Search` function:

```go
// SearchResult includes memory and similarity score
type SearchResult struct {
	Memory   Memory
	Distance float64 // cosine distance (lower = more similar)
}

// SearchWithScores finds memories by semantic similarity and returns distances
func (d *DB) SearchWithScores(embedding []float32, limit int, memType, area string) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}

	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding: %w", err)
	}

	query := `
		SELECT m.id, m.type, m.area, m.content, m.rationale, m.is_valid, m.superseded_by, m.created_at,
		       vec_distance_cosine(e.embedding, ?) as distance
		FROM memories m
		JOIN memory_embeddings e ON m.id = e.memory_id
		WHERE m.is_valid = TRUE
	`
	args := []interface{}{string(embeddingJSON)}

	if memType != "" {
		query += " AND m.type = ?"
		args = append(args, memType)
	}
	if area != "" {
		query += " AND m.area = ?"
		args = append(args, area)
	}

	query += `
		ORDER BY distance
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var supersededBy sql.NullInt64
		var rationale sql.NullString

		if err := rows.Scan(&r.Memory.ID, &r.Memory.Type, &r.Memory.Area, &r.Memory.Content,
			&rationale, &r.Memory.IsValid, &supersededBy, &r.Memory.CreatedAt, &r.Distance); err != nil {
			return nil, err
		}

		if rationale.Valid {
			r.Memory.Rationale = rationale.String
		}
		if supersededBy.Valid {
			r.Memory.SupersededBy = &supersededBy.Int64
		}

		results = append(results, r)
	}

	return results, rows.Err()
}
```

**Step 2: Run tests to ensure nothing broke**

Run: `CGO_ENABLED=1 go test ./internal/db -v`
Expected: All existing tests pass

---

## Task 11: Update App to Use Scores

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Update doSearch to use SearchWithScores**

Replace the `doSearch` function:

```go
func (a *App) doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		embedding, err := a.embedder.EmbedForSearch(query)
		if err != nil {
			return searchResultMsg{err: err}
		}
		results, err := a.db.SearchWithScores(embedding, 20, a.typeFilter, "")
		if err != nil {
			return searchResultMsg{err: err}
		}

		memories := make([]db.Memory, len(results))
		scores := make([]float64, len(results))
		for i, r := range results {
			memories[i] = r.Memory
			// Convert distance to similarity (1 - distance for cosine)
			scores[i] = 1 - r.Distance
		}
		return searchResultMsg{memories: memories, scores: scores, err: nil}
	}
}
```

**Step 2: Rebuild and verify**

Run: `CGO_ENABLED=1 go build -o ec-tui ./cmd/tui`
Expected: No errors

---

## Task 12: Add Clipboard Support

**Files:**
- Modify: `internal/tui/app.go`

**Step 1: Add clipboard import and copy command**

Add to imports:
```go
"github.com/atotto/clipboard"
```

Update the detail mode key handler:

```go
// Detail mode
if a.mode == viewDetail {
	switch msg.String() {
	case "c":
		mem := a.list.SelectedMemory()
		if mem != nil {
			text := mem.Content
			if mem.Rationale != "" {
				text += "\n\nRationale: " + mem.Rationale
			}
			clipboard.WriteAll(text)
			a.spiritMsg = "Copied to clipboard. The Machine Spirit approves."
		}
	}
	return a, nil
}
```

**Step 2: Rebuild**

Run: `CGO_ENABLED=1 go build -o ec-tui ./cmd/tui`
Expected: No errors

---

## Task 13: Add to Makefile (Optional)

**Files:**
- Modify or Create: `Makefile`

**Step 1: Create Makefile if doesn't exist**

```makefile
.PHONY: build server tui test clean

build: server tui

server:
	CGO_ENABLED=1 go build -o server ./cmd/server

tui:
	CGO_ENABLED=1 go build -o ec-tui ./cmd/tui

test:
	CGO_ENABLED=1 go test ./...

clean:
	rm -f server ec-tui
```

**Step 2: Test make**

Run: `make tui`
Expected: `ec-tui` binary created

---

## Task 14: Test Full TUI Manually

**Prerequisites:**
- Ollama running with nomic-embed-text model
- A database with some memories (or use the MCP server to add some first)

**Step 1: Run the TUI**

Run: `./ec-tui --db .claude/memory.db`

**Step 2: Verify:**
- [ ] Startup message appears from Machine Spirit
- [ ] Stats show in top left
- [ ] Spirit musing in top right
- [ ] Press `/` to focus search
- [ ] Type a query and press Enter
- [ ] Results appear with similarity scores
- [ ] Navigate with j/k
- [ ] Press Enter to view detail
- [ ] Press Esc to close detail
- [ ] Press `c` in detail to copy
- [ ] Press `f` to cycle filters
- [ ] Press `q` to quit with parting message

---

## Task 15: Update .gitignore

**Files:**
- Modify: `.gitignore`

**Step 1: Add TUI binary**

Add line:
```
/ec-tui
```

---

## Summary

After completing all tasks, you'll have:

1. `cmd/tui/main.go` - TUI entry point
2. `internal/tui/app.go` - Main Bubbletea model
3. `internal/tui/styles/styles.go` - Irish Mechanicus styling
4. `internal/tui/spirit/spirit.go` - Machine Spirit personality engine
5. `internal/tui/components/` - Reusable UI components

Run with: `./ec-tui --db ~/.claude/memory.db`
