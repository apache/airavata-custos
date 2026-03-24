// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apache/airavata-custos/signer/internal/config"
)

// Server wraps the HTTP server with graceful shutdown.
type Server struct {
	httpServer      *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

func New(cfg config.ServerConfig, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      handler,
			ReadTimeout:  time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeoutSeconds) * time.Second,
		},
		logger:          logger,
		shutdownTimeout: time.Duration(cfg.ShutdownTimeoutSeconds) * time.Second,
	}
}

// ListenAndServe starts the HTTP server and blocks until shutdown signal is received.
func (s *Server) ListenAndServe() error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting HTTP server", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		s.logger.Info("received shutdown signal", "signal", sig.String())
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	s.logger.Info("shutting down server", "timeout", s.shutdownTimeout.String())
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}
