package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/117503445/ai-coding-webui/internal/buildinfo"
	"github.com/117503445/ai-coding-webui/pkg/rpc"
	"github.com/117503445/ai-coding-webui/pkg/rpc/rpcconnect"
)

func NewCtxInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			requestID := ""
			if !req.Spec().IsClient {
				requestID = req.Header().Get("X-Request-ID")
				if requestID == "" {
					requestID = uuid.NewString()
				}
				ctx = WithContext(ctx, AppContext{RequestID: requestID})
				ctx = log.With().Str("request_id", requestID).Logger().WithContext(ctx)
				log.Ctx(ctx).Debug().Str("procedure", req.Spec().Procedure).Msg("request received")
			}

			resp, err := next(ctx, req)
			if err != nil {
				return nil, err
			}
			if resp != nil {
				resp.Header().Set("X-Request-ID", requestID)
			}
			log.Ctx(ctx).Debug().Msg("request done")
			return resp, nil
		}
	}
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Healthz(ctx context.Context, _ *connect.Request[rpc.HealthzRequest]) (*connect.Response[rpc.ApiResponse], error) {
	appCtx := GetAppContext(ctx)
	resp := connect.NewResponse(&rpc.ApiResponse{
		Code:    0,
		Message: "success",
		Payload: &rpc.ApiResponse_Healthz{
			Healthz: &rpc.HealthzResponse{
				Version:   buildinfo.GitVersion,
				RequestId: appCtx.RequestID,
			},
		},
	})
	return resp, nil
}

var _ rpcconnect.TemplateServiceHandler = (*Server)(nil)
