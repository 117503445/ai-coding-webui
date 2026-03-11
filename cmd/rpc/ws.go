package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/rs/zerolog"
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
		log.Error().Err(err).Msg("websocket accept failed")
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()
	logger := log.Ctx(ctx).With().Str("component", "ws").Logger()
	logger.Info().Msg("client connected")

	h.sendStatus(ctx, conn)

	for {
		var msg WSMessage
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				logger.Info().Msg("client disconnected")
			} else {
				logger.Warn().Err(err).Msg("read error")
			}
			return
		}

		switch msg.Type {
		case "chat":
			h.handleChat(ctx, conn, msg, &logger)
		case "command":
			h.handleCommand(ctx, conn, msg, &logger)
		case "abort":
			h.handleAbort(ctx, conn, &logger)
		default:
			logger.Warn().Str("type", msg.Type).Msg("unknown message type")
		}
	}
}

func (h *WSHandler) handleChat(ctx context.Context, conn *websocket.Conn, msg WSMessage, logger *zerolog.Logger) {
	if msg.Content == "" {
		h.sendError(ctx, conn, "empty message")
		return
	}

	if h.claude.IsWorking() {
		h.sendError(ctx, conn, "claude is already working")
		return
	}

	logger.Info().Str("session_id", msg.SessionID).Msg("chat request")

	h.sendJSON(ctx, conn, WSResponse{Type: "status", Status: "working", Detail: "starting..."})

	var writeMu sync.Mutex

	h.claude.Run(ctx, msg.Content, msg.SessionID,
		func(eventJSON json.RawMessage) {
			raw := json.RawMessage(eventJSON)
			writeMu.Lock()
			defer writeMu.Unlock()
			h.sendJSON(ctx, conn, WSResponse{Type: "stream", Event: &raw})
		},
		func(sid string, err error) {
			writeMu.Lock()
			defer writeMu.Unlock()
			if err != nil {
				logger.Warn().Err(err).Msg("claude run error")
				h.sendJSON(ctx, conn, WSResponse{
					Type:      "complete",
					SessionID: sid,
					IsError:   true,
					Message:   err.Error(),
				})
			} else {
				h.sendJSON(ctx, conn, WSResponse{
					Type:      "complete",
					SessionID: sid,
				})
			}
			h.sendStatus(ctx, conn)
		},
	)
}

func (h *WSHandler) handleCommand(ctx context.Context, conn *websocket.Conn, msg WSMessage, logger *zerolog.Logger) {
	logger.Info().Str("command", msg.Command).Msg("command request")

	switch msg.Command {
	case "/new", "/clear", "/reset":
		h.claude.Abort()
		h.sendJSON(ctx, conn, WSResponse{
			Type:    "command_result",
			Message: "session cleared",
		})
		h.sendStatus(ctx, conn)
	default:
		h.sendError(ctx, conn, "unknown command: "+msg.Command)
	}
}

func (h *WSHandler) handleAbort(ctx context.Context, conn *websocket.Conn, logger *zerolog.Logger) {
	logger.Info().Msg("abort request")
	success := h.claude.Abort()
	if success {
		h.sendJSON(ctx, conn, WSResponse{Type: "status", Status: "idle", Detail: "aborted"})
	} else {
		h.sendJSON(ctx, conn, WSResponse{Type: "status", Status: "idle", Detail: "nothing to abort"})
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
		log.Warn().Err(err).Msg("write error")
	}
}
