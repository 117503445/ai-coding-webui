# Claude Code WebUI - Architecture Design

## Overview

A web-based UI for Claude Code that provides streaming chat with thinking process
visibility, tool call details, and session persistence. The backend is fully stateless;
all conversation history is managed by Claude Code's built-in session storage
(`~/.claude/projects/`).

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Browser (React)                           │
│                                                                     │
│  ┌──────────┐  ┌──────────────┐  ┌───────────┐  ┌───────────────┐ │
│  │InputArea  │  │ MessageList  │  │ StatusBar │  │ localStorage  │ │
│  │(slash cmd)│  │ (streaming)  │  │(conn/work)│  │(session cache)│ │
│  └─────┬────┘  └──────▲───────┘  └─────▲─────┘  └───────▲───────┘ │
│        │               │               │                 │         │
│        ▼               │               │                 │         │
│  ┌─────────────────────┴───────────────┴─────────────────┘       │ │
│  │              zustand Store (chatStore + connectionStore)       │ │
│  └─────────────────────┬─────────────────────────────────────────┘ │
│                        │                                           │
│              ┌─────────▼─────────┐                                 │
│              │  WebSocket Client │                                 │
│              │ (expo. backoff    │                                 │
│              │  reconnect)       │                                 │
│              └─────────┬─────────┘                                 │
└────────────────────────┼───────────────────────────────────────────┘
                         │ ws://host:8080/ws
                         │
┌────────────────────────┼───────────────────────────────────────────┐
│                  Go Backend (:8080)  [STATELESS]                   │
│                        │                                           │
│              ┌─────────▼─────────┐                                 │
│              │  WebSocket Handler │◄──── coder/websocket           │
│              │  (cmd/rpc/ws.go)  │                                 │
│              └─────────┬─────────┘                                 │
│                        │                                           │
│              ┌─────────▼─────────┐       ┌──────────────────────┐ │
│              │  Claude Manager   │       │  Connect RPC Handler │ │
│              │  (cmd/rpc/claude  │       │  (Healthz, GetStatus │ │
│              │   .go)            │       │   Abort)             │ │
│              └─────────┬─────────┘       └──────────────────────┘ │
│                        │                                           │
│              ┌─────────▼─────────┐       ┌──────────────────────┐ │
│              │  os/exec spawn    │       │  go:embed frontend   │ │
│              │  claude CLI       │       │  (SPA fallback)      │ │
│              └─────────┬─────────┘       └──────────────────────┘ │
└────────────────────────┼───────────────────────────────────────────┘
                         │
              ┌──────────▼──────────┐
              │    claude CLI       │
              │                     │
              │  claude -p "msg"    │
              │    --output-format  │
              │      stream-json   │
              │    --verbose        │
              │    --resume <id>    │
              └──────────┬──────────┘
                         │
              ┌──────────▼──────────┐
              │  ~/.claude/projects │
              │  /<project-hash>/   │
              │    sessions/        │
              │      <session>.json │
              │    memory/          │
              │      CLAUDE.md      │
              └─────────────────────┘
```

## Communication Protocol

### WebSocket Messages (JSON over ws://host:8080/ws)

#### Frontend -> Backend

```
┌──────────────────────────────────────────────────┐
│  { "type": "chat",    "content": "user message", │
│    "session_id": "abc-123" }                     │
│  -- send chat message, resume session if id set  │
├──────────────────────────────────────────────────┤
│  { "type": "command", "command": "/new" }        │
│  -- execute slash command                        │
├──────────────────────────────────────────────────┤
│  { "type": "abort" }                             │
│  -- kill current claude process                  │
└──────────────────────────────────────────────────┘
```

#### Backend -> Frontend

```
┌──────────────────────────────────────────────────┐
│  { "type": "status",                             │
│    "status": "idle" | "working",                 │
│    "detail": "thinking..." }                     │
├──────────────────────────────────────────────────┤
│  { "type": "stream",                             │
│    "event": { <claude stream-json event> } }     │
│  -- forwarded from claude CLI stdout             │
├──────────────────────────────────────────────────┤
│  { "type": "complete",                           │
│    "session_id": "abc-123",                      │
│    "is_error": false }                           │
├──────────────────────────────────────────────────┤
│  { "type": "error", "message": "..." }           │
└──────────────────────────────────────────────────┘
```

### Connect RPC (HTTP)

```
ClaudeService {
  Healthz()   -> version, uptime
  GetStatus() -> idle/working, session_id
  Abort()     -> success/fail
}
```

## Multi-turn Chat Flow

```
  Frontend                  Backend                   claude CLI
     │                         │                          │
     │── chat(msg1) ──────────>│                          │
     │                         │── spawn: claude -p msg1  │
     │                         │   --stream-json          │
     │                         │   --verbose ────────────>│
     │                         │                          │
     │<── stream(event) ──────-│<── stdout line ──────────│
     │<── stream(event) ──────-│<── stdout line ──────────│
     │<── complete(sid=X) ────-│<── exit 0 ───────────────│
     │                         │                          │
     │  [stores sid=X in       │                          │
     │   localStorage]         │                          │
     │                         │                          │
     │── chat(msg2, sid=X) ──->│                          │
     │                         │── spawn: claude -p msg2  │
     │                         │   --resume X             │
     │                         │   --stream-json ────────>│
     │                         │                          │
     │<── stream(event) ──────-│<── stdout line ──────────│
     │<── complete(sid=X) ────-│<── exit 0 ───────────────│
     │                         │                          │
```

## Session Persistence

```
┌─────────────┐    ┌──────────────────┐    ┌─────────────────────┐
│  Frontend   │    │    Backend       │    │  ~/.claude/projects  │
│             │    │  (stateless)     │    │                     │
│ localStorage│    │                  │    │  sessions managed   │
│  - session  │    │  reads nothing   │    │  by claude CLI      │
│    _id      │    │  stores nothing  │    │  automatically      │
│  - messages │    │                  │    │                     │
│    (cache)  │    │  just relays     │    │  --resume uses      │
│             │    │  between browser │    │  these files        │
│             │    │  and claude CLI  │    │                     │
└──────┬──────┘    └──────────────────┘    └──────────┬──────────┘
       │                                              │
       │  On page refresh:                            │
       │  1. Load messages from localStorage          │
       │  2. Reconnect WebSocket                      │
       │  3. Next chat uses --resume <session_id>     │
       │     which reads from ─────────────────────>  │
       └──────────────────────────────────────────────┘
```

## Build Pipeline

```
  ┌──────────────┐     ┌───────────────────┐     ┌─────────────────┐
  │ 1. pnpm build│────>│ 2. cp fe/dist ->  │────>│ 3. go build     │
  │    (fe/)     │     │ cmd/rpc/          │     │    ./cmd/rpc    │
  │              │     │ frontend_dist/    │     │    (with embed) │
  └──────────────┘     └───────────────────┘     └────────┬────────┘
                                                          │
                                                  ┌───────▼───────┐
                                                  │ data/rpc/rpc  │
                                                  │ (single bin)  │
                                                  └───────────────┘
```

## Frontend Component Tree

```
App
 └─ ChatContainer
     ├─ StatusBar
     │   ├─ Logo + Title
     │   ├─ ConnectionIndicator (spinner on disconnect)
     │   └─ WorkingStatus (pulse animation when busy)
     │
     ├─ MessageList
     │   └─ MessageItem (per message)
     │       ├─ ThinkingBlock (collapsible, italic)
     │       ├─ ToolCallBlock (card, expandable)
     │       └─ MarkdownRenderer (with copy button)
     │
     ├─ ConnectionOverlay (fullscreen on disconnect)
     │
     └─ InputArea
         ├─ TextArea (Shift+Enter newline, Enter send)
         ├─ SlashCommandMenu (popup on "/" prefix)
         └─ SendButton / AbortButton (toggle by state)
```

## Slash Commands

| Command    | Action                                      |
|------------|---------------------------------------------|
| `/new`     | Clear messages, reset session_id, new chat  |
| `/compact` | Send compact request to claude CLI          |
| `/cost`    | Show token usage (from claude response)     |
| `/clear`   | Same as /new                                |
| `/help`    | Show available commands                     |

## E2E Test Architecture

```
  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
  │  main.py     │────>│  Playwright  │────>│  Browser     │
  │  test runner │     │  (chromium)  │     │  (localhost)  │
  └──────┬───────┘     └──────────────┘     └──────┬───────┘
         │                                         │
         │  Manages:                               │  Tests against:
         │  - Start/stop backend                   │  - Chat UI
         │  - Screenshot each step                 │  - WebSocket msgs
         │  - Collect trace.zip                    │  - Slash commands
         │  - Write test.log                       │  - Session persist
         │                                         │
         ▼                                         ▼
  data/e2e/cases/                          localhost:8080
    YYYYMMDD-HHMMSS-casename/              (Go backend)
      screenshots/
      logs/
        trace.zip
        test.log
```
