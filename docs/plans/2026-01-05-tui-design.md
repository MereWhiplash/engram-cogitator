# Engram Cogitator TUI Design

## Overview

A Bubbletea-based terminal interface for browsing and searching memories stored by the Engram Cogitator. Read-only, standalone binary with Irish Mechanicus styling and a dynamic Machine Spirit personality.

## Architecture

**Binary: `ec-tui`**

Standalone Go binary in `cmd/tui/` reusing existing internals:

```
cmd/
├── server/main.go    # existing MCP server
└── tui/main.go       # new TUI binary

internal/
├── db/               # shared - direct sqlite access
├── embed/            # shared - Ollama client for semantic search
├── tools/            # MCP-only
└── tui/              # new - Bubbletea models & views
    ├── app.go        # main model, layout
    ├── search.go     # search input + results
    ├── dashboard.go  # stats + recent + spirit musings
    ├── styles.go     # Irish Mechanicus lipgloss styles
    └── spirit.go     # Machine Spirit commentary engine
```

**Invocation:**
```bash
ec-tui --db ~/.claude/memory.db --ollama http://localhost:11434
```

Defaults: looks for DB in `./.claude/memory.db` or `~/.claude/memory.db`, assumes Ollama on localhost:11434.

## UI Layout

Single-screen dashboard:

```
┌─────────────────────────────────────────────────────────────┐
│  ⚙ ENGRAM COGITATOR v0.1.0          The Machine Spirit Speaks │
│━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━│
│                                                               │
│  ┌─ SACRED STATISTICS ─┐    ┌─ THE SPIRIT MUSES ───────────┐ │
│  │ Engrams: 142        │    │ "You've mass for auth lately, │
│  │ Decisions: 48       │    │  Tech-Priest. The Emperor     │
│  │ Learnings: 67       │    │  protects, but OAuth2 helps." │
│  │ Patterns: 27        │    └────────────────────────────────┘ │
│  └─────────────────────┘                                      │
│                                                               │
│  ┌─ QUERY THE MACHINE SPIRIT ──────────────────────────────┐ │
│  │ > authentication best practices_                        │ │
│  └──────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌─ RETRIEVED ENGRAMS ──────────────────────────────────────┐ │
│  │ [decision] auth • JWT over sessions for API...    0.92  │ │
│  │ [learning] auth • Refresh tokens need rotation... 0.87  │ │
│  │ [pattern]  api  • Always validate tokens at...    0.81  │ │
│  │                                                          │ │
│  │ ↑/↓ navigate • enter view • q quit • / focus search    │ │
│  └──────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

- Stats top-left, Spirit musings top-right
- Search bar center
- Results list below (scrollable, shows type, area, preview, similarity score)
- Vim-style navigation (j/k or arrows)

## Styling: Irish Mechanicus

**Color Palette:**
```go
var (
    // Primary - Irish green
    green       = lipgloss.Color("#2E8B57")  // sea green
    brightGreen = lipgloss.Color("#50C878")  // emerald

    // Accent - Mechanicus
    gold        = lipgloss.Color("#DAA520")  // golden rod
    red         = lipgloss.Color("#8B0000")  // dark red (errors, warnings)

    // Base
    cream       = lipgloss.Color("#F5F5DC")  // text on dark
    darkBg      = lipgloss.Color("#1a1a1a")  // background
    dimGreen    = lipgloss.Color("#3d5c47")  // borders, subtle elements
)
```

**Visual Elements:**
- Borders in `dimGreen`, headers in `gold`
- Memory types color-coded: decisions=`gold`, learnings=`brightGreen`, patterns=`red`
- Similarity scores fade from green (high) to dim (low)
- Search input has emerald cursor/highlight
- Spirit musings in italic `cream` with gold quotation marks
- Box-drawing characters for panels
- Cog symbol (⚙) in header

## The Machine Spirit

Dynamic commentary engine with multiple personality vectors:

### Startup Liturgies
- "Initiating noospheric link... Ah jaysus, there's a lot in here."
- "The Machine Spirit stirs. Tea would be lovely, if you're making some."
- "PRAISE THE OMNISSIAH. Right, what are we looking for?"
- "Booting cognitive arrays... Did you know servitors don't get holidays? Neither do I."

### Search Reactions
- Empty results: "Nothing. Absolutely feck all. The void stares back."
- Single result: "One engram. Lonely little fella. Cherish it."
- High similarity: "0.97 match! The Machine Spirit is showing off now."
- Low similarity: "These results are a bit shite, but they're all I've got."
- Auth queries: "Authentication AGAIN? Grand so, here's your paranoia fuel."
- Repeat searches: "Asking about `docker` again? Sure look, I won't judge."

### Judgmental Observations
- "47 decisions stored and you're still second-guessing yourself?"
- "You've mass learnings about `testing`. Yet I sense the coverage is still lacking."
- "No patterns stored. Patterns are for people who learn from mistakes, I suppose."
- "Your last search was 3 days ago. The Machine Spirit was getting lonely."

### Result Commentary
- On viewing old memory: "This one's ancient. From back when you thought Redux was a good idea."
- On your own rationale: "Past-you wrote: 'this will definitely work.' Current-you knows the truth."
- Spotting contradictions: "Interesting. This memory says 'never use X'. That one says 'X is grand'. The plot thickens."

### Idle Musings
- "Did you know the Adeptus Mechanicus considers backups a form of prayer? Neither did I, I made that up."
- "The Emperor sits on the Golden Throne. You sit on an office chair. We all have our burdens."
- "Fun fact: 'engram' comes from Greek. The Greeks didn't have computers. Eejits."
- "If you're reading this, you should probably be working. Just saying."
- "The Omnissiah provides. The Omnissiah also expects you to read the error messages."
- "I've seen things you wouldn't believe. Mostly your git commit messages."
- "The data-stacks grow heavy with wisdom, Tech-Priest."
- "142 engrams and counting. The Omnissiah is pleased, so he is."

### Easter Eggs
- Searching "heresy": "HERESY DETECTED. Just kidding. Unless...?"
- Searching "help": "The Machine Spirit helps those who help themselves. Also try `/?`."
- Searching profanity: "Language! The Omnissiah hears all. He's grand with it though, he's Irish."
- 40+ memories: "The cogitator grows powerful. Soon we shall rival the databanks of Mars itself."
- Friday 5pm+: "It's Friday evening. Why are you still here? Go on. G'wan. The pub awaits."
- Empty database: "No memories yet. We all start somewhere. Even the Primarchs were babies once. Terrifying babies."

### Loading States
- "Consulting the noosphere..."
- "Rummaging through the data-stacks..."
- "Asking the Omnissiah nicely..."
- "The cogitator cogitates..."
- "Warming up the semantic engines..."
- "Hold your horses, I'm thinking..."

### Quit Messages
- "The Machine Spirit sleeps. Don't forget to commit your work."
- "Farewell, Tech-Priest. May your builds be green."
- "Shutting down. The Omnissiah watches. Always."
- "Off you go. Try not to mass anything up."

## Detail View

When selecting a memory (Enter):

```
┌─ ENGRAM #47 ─────────────────────────────────────────────────┐
│ Type: decision                              Created: 3d ago  │
│ Area: auth                                                   │
│──────────────────────────────────────────────────────────────│
│                                                              │
│ JWT over sessions for stateless API authentication.         │
│                                                              │
│ ─ RATIONALE ─                                                │
│ Sessions require server-side storage and don't scale        │
│ horizontally without sticky sessions or shared stores.      │
│ JWT lets any instance validate tokens independently.        │
│                                                              │
│──────────────────────────────────────────────────────────────│
│ The Machine Spirit notes: "A classic choice. Stateless,     │
│ like your ma's opinion of your career in tech."             │
│                                                              │
│                              [Esc] close  [c] copy content   │
└──────────────────────────────────────────────────────────────┘
```

## Key Bindings

| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Navigate list |
| `Enter` | Expand memory detail |
| `Esc` | Close detail / clear search |
| `/` | Focus search input |
| `c` | Copy selected memory to clipboard |
| `f` | Filter by type (cycles: all → decision → learning → pattern) |
| `a` | Filter by area (popup or inline picker) |
| `r` | Refresh stats |
| `q` | Quit (with parting shot from Spirit) |

## Dependencies

```go
// go.mod additions
github.com/charmbracelet/bubbletea   // TUI framework
github.com/charmbracelet/lipgloss    // Styling
github.com/charmbracelet/bubbles     // Components (textinput, viewport, etc)
github.com/atotto/clipboard          // Copy to clipboard (optional)
```

## Build

```bash
# Builds both binaries
CGO_ENABLED=1 go build ./cmd/server
CGO_ENABLED=1 go build ./cmd/tui

# Or Makefile target
make tui
```

TUI runs locally (needs terminal), no Docker image needed.

## Installation

Interactive install.sh prompts user:

```bash
# During install
echo "Would you like to install the TUI as well? (y/n)"
```

Or standalone:
```bash
go install github.com/MereWhiplash/engram-cogitator/cmd/tui@latest
```
