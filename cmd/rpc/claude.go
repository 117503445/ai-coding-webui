package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
	m.detail = "启动 claude..."
	m.mu.Unlock()

	go func() {
		sid, err := m.runClaude(runCtx, message, sessionID, onEvent)

		m.mu.Lock()
		if sid != "" {
			m.sessionID = sid
		}
		m.status = "idle"
		m.detail = ""
		m.cancel = nil
		m.mu.Unlock()

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

	log.Ctx(ctx).Info().
		Str("workDir", m.workDir).
		Strs("args", args).
		Str("sessionID", sessionID).
		Msg("启动 claude CLI 子进程")

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = m.workDir
	cmd.Env = os.Environ()
	cmd.Stdin = nil // 不连接 stdin，避免 claude CLI 等待 TTY 输入导致挂起

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("创建 stdout pipe 失败")
		return "", fmt.Errorf("stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("创建 stderr pipe 失败")
		return "", fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("启动 claude 进程失败")
		return "", fmt.Errorf("start claude: %w", err)
	}

	log.Ctx(ctx).Info().Int("pid", cmd.Process.Pid).Msg("claude 进程已启动")

	// 在后台消费 stderr，避免管道缓冲区满导致子进程阻塞
	go func() {
		stderrScanner := bufio.NewScanner(stderrPipe)
		stderrScanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
		for stderrScanner.Scan() {
			log.Ctx(ctx).Debug().Str("src", "claude-stderr").Msg(stderrScanner.Text())
		}
	}()

	m.mu.Lock()
	m.cmd = cmd
	m.detail = "处理中..."
	m.mu.Unlock()

	var resultSessionID string
	lineCount := 0
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		lineCount++

		log.Ctx(ctx).Trace().
			Int("line", lineCount).
			Int("bytes", len(line)).
			Msg("收到 claude stdout 行")

		onEvent(json.RawMessage(append([]byte(nil), line...)))

		sid := extractSessionID(line)
		if sid != "" {
			resultSessionID = sid
			log.Ctx(ctx).Debug().Str("sessionID", sid).Msg("提取到 session_id")
		}

		detail := extractDetail(line)
		if detail != "" {
			m.mu.Lock()
			m.detail = detail
			m.mu.Unlock()
			log.Ctx(ctx).Debug().Str("detail", detail).Msg("更新工作详情")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("scanner 读取错误")
	}

	waitErr := cmd.Wait()

	m.mu.Lock()
	m.cmd = nil
	m.mu.Unlock()

	if ctx.Err() != nil {
		log.Ctx(ctx).Info().Int("lines", lineCount).Msg("claude 进程被终止 (abort)")
		return resultSessionID, fmt.Errorf("aborted")
	}

	if waitErr != nil {
		log.Ctx(ctx).Warn().Err(waitErr).Int("lines", lineCount).Msg("claude 进程退出异常")
	} else {
		log.Ctx(ctx).Info().
			Int("lines", lineCount).
			Str("sessionID", resultSessionID).
			Msg("claude 进程正常完成")
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
			return "思考中..."
		case "text":
			return "撰写中..."
		case "tool_use":
			name := obj.Event.ContentBlock.Name
			if name != "" {
				return "使用工具: " + name
			}
			return "使用工具..."
		}
	}

	// 顶层事件类型也做详情提示
	var topLevel struct {
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
	}
	if json.Unmarshal(line, &topLevel) == nil {
		switch topLevel.Type {
		case "system":
			return "初始化..."
		case "result":
			if topLevel.Subtype == "success" {
				return "完成"
			}
			return "结果: " + strings.TrimPrefix(topLevel.Subtype, "")
		}
	}

	return ""
}
