package main

import (
	"context"
	"net"
	"net/http"

	"connectrpc.com/connect"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"

	"github.com/117503445/ai-coding-webui/pkg/rpc/rpcconnect"
)

func ListenAndServe(ctx context.Context, port string) error {
	mux := http.NewServeMux()
	server := NewServer()

	path, handler := rpcconnect.NewTemplateServiceHandler(
		server,
		connect.WithInterceptors(NewCtxInterceptor()),
	)
	mux.Handle(path, handler)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		MaxAge:         86400,
	})

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := listener.Close(); closeErr != nil {
			log.Ctx(ctx).Error().Err(closeErr).Msg("failed to close listener")
		}
	}()

	log.Ctx(ctx).Info().Str("addr", listener.Addr().String()).Msg("listening")
	return http.Serve(listener, c.Handler(mux))
}
