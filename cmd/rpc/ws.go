package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/rs/zerolog/log"
)

type WSMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	Command   string `json:"command,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type WSResponse struct {
	Type      string           `json:"type"`
	Status    string           `json:"status,omitempty"`
	Detail    string           `json:"detail,omitempty"`
	Event     *json.RawMessage `json:"event,omitempty"`
	SessionID string           `json:"session_id,omitempty"`
	IsError   bool             `json:"is_error,omitempty"`
	Message   string           `json:"message,omitempty"`
}

type WSHandler struct {
	claude *ClaudeManager
}

func NewWSHandler(claude *ClaudeManager) *WSHandler {
	return &WSHandler{claude: claude}
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Error().Err(err).Msg("websocket 握手失败")
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()
	ctx = log.Ctx(ctx).With().
		Str("component", "ws").
		Str("remoteAddr", r.RemoteAddr).
		Logger().WithContext(ctx)

	log.Ctx(ctx).Info().Msg("客户端已连接")

	h.sendStatus(ctx, conn)

	for {
		var msg WSMessage
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				log.Ctx(ctx).Info().Msg("客户端正常断开")
			} else {
				log.Ctx(ctx).Warn().Err(err).Msg("读取 WebSocket 消息失败")
			}
			return
		}

		log.Ctx(ctx).Debug().Str("msgType", msg.Type).Msg("收到消息")

		switch msg.Type {
		case "chat":
			h.handleChat(ctx, conn, msg)
		case "command":
			h.handleCommand(ctx, conn, msg)
		case "abort":
			h.handleAbort(ctx, conn)
		default:
			log.Ctx(ctx).Warn().Str("type", msg.Type).Msg("未知消息类型")
		}
	}
}

func (h *WSHandler) handleChat(ctx context.Context, conn *websocket.Conn, msg WSMessage) {
	if msg.Content == "" {
		log.Ctx(ctx).Warn().Msg("收到空消息")
		h.sendError(ctx, conn, "empty message")
		return
	}

	if h.claude.IsWorking() {
		log.Ctx(ctx).Warn().Msg("claude 正在工作中，拒绝新请求")
		h.sendError(ctx, conn, "claude is already working")
		return
	}

	log.Ctx(ctx).Info().
		Str("sessionID", msg.SessionID).
		Int("contentLen", len(msg.Content)).
		Msg("开始处理聊天请求")

	h.sendJSON(ctx, conn, WSResponse{Type: "status", Status: "working", Detail: "启动中..."})

	var writeMu sync.Mutex
	eventCount := 0

	h.claude.Run(ctx, msg.Content, msg.SessionID,
		func(eventJSON json.RawMessage) {
			raw := json.RawMessage(eventJSON)
			writeMu.Lock()
			defer writeMu.Unlock()
			eventCount++
			h.sendJSON(ctx, conn, WSResponse{Type: "stream", Event: &raw})
		},
		func(sid string, err error) {
			writeMu.Lock()
			defer writeMu.Unlock()
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).
					Str("sessionID", sid).
					Int("events", eventCount).
					Msg("claude 执行出错")
				h.sendJSON(ctx, conn, WSResponse{
					Type:      "complete",
					SessionID: sid,
					IsError:   true,
					Message:   err.Error(),
				})
			} else {
				log.Ctx(ctx).Info().
					Str("sessionID", sid).
					Int("events", eventCount).
					Msg("claude 执行完成")
				h.sendJSON(ctx, conn, WSResponse{
					Type:      "complete",
					SessionID: sid,
				})
			}
			h.sendStatus(ctx, conn)
		},
	)
}

func (h *WSHandler) handleCommand(ctx context.Context, conn *websocket.Conn, msg WSMessage) {
	log.Ctx(ctx).Info().Str("command", msg.Command).Msg("执行斜杠命令")

	switch msg.Command {
	case "/new", "/clear", "/reset":
		h.claude.Abort()
		h.sendJSON(ctx, conn, WSResponse{
			Type:    "command_result",
			Message: "session cleared",
		})
		h.sendStatus(ctx, conn)
		log.Ctx(ctx).Info().Str("command", msg.Command).Msg("会话已清除")
	default:
		log.Ctx(ctx).Warn().Str("command", msg.Command).Msg("未知命令")
		h.sendError(ctx, conn, "unknown command: "+msg.Command)
	}
}

func (h *WSHandler) handleAbort(ctx context.Context, conn *websocket.Conn) {
	log.Ctx(ctx).Info().Msg("收到终止请求")
	success := h.claude.Abort()
	if success {
		log.Ctx(ctx).Info().Msg("claude 进程已终止")
		h.sendJSON(ctx, conn, WSResponse{Type: "status", Status: "idle", Detail: "已终止"})
	} else {
		log.Ctx(ctx).Debug().Msg("没有正在运行的 claude 进程")
		h.sendJSON(ctx, conn, WSResponse{Type: "status", Status: "idle", Detail: "无需终止"})
	}
}

func (h *WSHandler) sendStatus(ctx context.Context, conn *websocket.Conn) {
	st := h.claude.GetStatus()
	h.sendJSON(ctx, conn, WSResponse{
		Type:      "status",
		Status:    st.Status,
		Detail:    st.Detail,
		SessionID: st.SessionID,
	})
}

func (h *WSHandler) sendError(ctx context.Context, conn *websocket.Conn, message string) {
	h.sendJSON(ctx, conn, WSResponse{Type: "error", Message: message})
}

func (h *WSHandler) sendJSON(ctx context.Context, conn *websocket.Conn, v any) {
	writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := wsjson.Write(writeCtx, conn, v); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("写入 WebSocket 消息失败")
	}
}
