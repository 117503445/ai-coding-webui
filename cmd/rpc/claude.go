package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/rs/zerolog/log"
)

type ClaudeManager struct {
	mu        sync.Mutex
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	status    string // "idle" or "working"
	detail    string
	sessionID string
	workDir   string
}

func NewClaudeManager(workDir string) *ClaudeManager {
	return &ClaudeManager{
		status:  "idle",
		workDir: workDir,
	}
}

type ClaudeStatus struct {
	Status    string `json:"status"`
	Detail    string `json:"detail"`
	SessionID string `json:"session_id"`
}

func (m *ClaudeManager) GetStatus() ClaudeStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return ClaudeStatus{
		Status:    m.status,
		Detail:    m.detail,
		SessionID: m.sessionID,
	}
}

func (m *ClaudeManager) IsWorking() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status == "working"
}

func (m *ClaudeManager) Abort() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
		m.status = "idle"
		m.detail = ""
		return true
	}
	return false
}

// StreamEvent is a parsed line from claude's stream-json output.
type StreamEvent struct {
	Raw json.RawMessage
}

// Run spawns a claude CLI process and streams events to the callback.
// sessionID can be empty for a new session.
func (m *ClaudeManager) Run(ctx context.Context, message string, sessionID string, onEvent func(eventJSON json.RawMessage), onComplete func(sid string, err error)) {
	m.mu.Lock()
	if m.status == "working" {
		m.mu.Unlock()
		onComplete("", fmt.Errorf("already working"))
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.status = "working"
	m.detail = "starting claude..."
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			m.status = "idle"
			m.detail = ""
			m.cancel = nil
			m.mu.Unlock()
		}()

		sid, err := m.runClaude(runCtx, message, sessionID, onEvent)
		if sid != "" {
			m.mu.Lock()
			m.sessionID = sid
			m.mu.Unlock()
		}
		onComplete(sid, err)
	}()
}

func (m *ClaudeManager) runClaude(ctx context.Context, message string, sessionID string, onEvent func(eventJSON json.RawMessage)) (string, error) {
	args := []string{
		"-p", message,
		"--output-format", "stream-json",
		"--verbose",
	}

	if sessionID != "" {
		args = append(args, "--resume", sessionID)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = m.workDir
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start claude: %w", err)
	}

	m.mu.Lock()
	m.cmd = cmd
	m.detail = "processing..."
	m.mu.Unlock()

	var resultSessionID string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		onEvent(json.RawMessage(append([]byte(nil), line...)))

		sid := extractSessionID(line)
		if sid != "" {
			resultSessionID = sid
		}

		detail := extractDetail(line)
		if detail != "" {
			m.mu.Lock()
			m.detail = detail
			m.mu.Unlock()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Warn().Err(err).Msg("scanner error")
	}

	waitErr := cmd.Wait()

	m.mu.Lock()
	m.cmd = nil
	m.mu.Unlock()

	if ctx.Err() != nil {
		return resultSessionID, fmt.Errorf("aborted")
	}

	return resultSessionID, waitErr
}

func extractSessionID(line []byte) string {
	var obj struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(line, &obj); err == nil && obj.SessionID != "" {
		return obj.SessionID
	}
	return ""
}

func extractDetail(line []byte) string {
	var obj struct {
		Type  string `json:"type"`
		Event struct {
			Type         string `json:"type"`
			ContentBlock struct {
				Type string `json:"type"`
				Name string `json:"name"`
			} `json:"content_block"`
		} `json:"event"`
	}
	if err := json.Unmarshal(line, &obj); err != nil {
		return ""
	}

	if obj.Type == "stream_event" && obj.Event.Type == "content_block_start" {
		switch obj.Event.ContentBlock.Type {
		case "thinking":
			return "thinking..."
		case "text":
			return "writing..."
		case "tool_use":
			name := obj.Event.ContentBlock.Name
			if name != "" {
				return "using tool: " + name
			}
			return "using tool..."
		}
	}
	return ""
}
