package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcp_handler "fundamental-ramen.com/agent-swing/internal/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version 預設為 dev，編譯時可透過 ldflags 注入 (見 Taskfile.yaml)
var Version = "dev"

func main() {
	svr := newMcpServer()

	mux := http.NewServeMux()

	// Streamable HTTP transport
	streamableHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return svr },
		nil,
	)
	mux.Handle("/mcp", streamableHandler)

	// SSE transport
	sseHandler := mcp.NewSSEHandler(
		func(*http.Request) *mcp.Server { return svr },
		nil,
	)
	mux.Handle("/sse", sseHandler)

	srv := &http.Server{
		Addr:    ":9001",
		Handler: mux,
	}

	// 收到 SIGINT / SIGTERM 時優雅關閉，讓服務持續工作直到被要求停止
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("http server shutdown error: %v", err)
		}
	}()

	// HTTP server 跑在主 goroutine，程式的生命週期跟著它走
	log.Println("HTTP server (streamable + sse) listening on :9001")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server error: %v", err)
	}
	log.Println("HTTP server stopped")
}

func newMcpServer() *mcp.Server {
	svr := mcp.NewServer(&mcp.Implementation{Name: "swings", Version: Version}, nil)
	mcp.AddTool(svr, &mcp.Tool{Name: "greet", Description: "say hi"}, mcp_handler.SayHi)
	mcp.AddTool(svr, &mcp.Tool{Name: "xmlTidy", Description: "xxx"}, mcp_handler.XmlFormat)
	return svr
}
