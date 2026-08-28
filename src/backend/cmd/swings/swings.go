package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fundamental-ramen.com/agent-swing/internal/commons"
	mcp_handler "fundamental-ramen.com/agent-swing/internal/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
	"go.yaml.in/yaml/v3"
)

// Version 預設為 dev，編譯時可透過 ldflags 注入 (見 Taskfile.yaml)
var Version = "dev"

func main() {
	setupLogger()
	zap.L().Info("Logger 配置完成")

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

func setupLogger() {
	cfg := zap.NewProductionConfig()

	configPath := commons.GetPathFromHome("etc/zaplog.yaml")
	fmt.Printf("log config path: %s\n", configPath)
	yamlData, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	// zaplog.yaml 覆蓋 zap.Config 基本架構
	if err := yaml.Unmarshal(yamlData, &cfg); err != nil {
		panic(err)
	}

	// 建立 Logger 實例
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	// 設定全域使用
	zap.ReplaceGlobals(logger)

	// 結束前先完成存檔
	defer zap.L().Sync()
}

func newMcpServer() *mcp.Server {
	svr := mcp.NewServer(&mcp.Implementation{Name: "swings", Version: Version}, nil)
	mcp.AddTool(svr, &mcp.Tool{Name: "greet", Description: "say hi"}, mcp_handler.SayHi)
	mcp.AddTool(svr, &mcp.Tool{Name: "mermaidToImage", Description: "Convert mermaid into SVG or PNG."}, mcp_handler.MermaidToImage)
	// mcp.AddTool(svr, &mcp.Tool{Name: "xmlTidy", Description: "xxx"}, mcp_handler.XmlFormat)
	return svr
}
