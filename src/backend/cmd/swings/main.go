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
	api_handlers "fundamental-ramen.com/agent-swing/internal/handlers/api"
	mcp_handler "fundamental-ramen.com/agent-swing/internal/handlers/mcp"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
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

	// ---- MCP / SSE 服務 (stdlib http.Server, :9001) ----
	// 獨立於 Fiber，直接以標準庫 http 掛載 MCP 的 Streamable HTTP 與 SSE transport，
	// 避免跨框架代理與路徑改寫，降低維護複雜度。
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

	mcpSrv := &http.Server{
		Addr:    ":9001",
		Handler: mux,
	}

	// 啟動 MCP / SSE 服務
	go func() {
		log.Println("MCP/SSE server (streamable + sse) listening on :9001")
		if err := mcpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("mcp/sse server error: %v", err)
		}
	}()

	// ---- API 服務 (Fiber, :9002) ----
	app := fiber.New(fiber.Config{
		AppName: "swings",
	})

	// CORS middleware — 允許前端跨域存取
	app.Use(cors.New())

	// POST /api/mermaid-to-svg — proxy mermaid syntax to kroki and return SVG
	app.Post("/api/mermaid-to-svg", api_handlers.MermaidToImage)

	// 收到 SIGINT / SIGTERM 時優雅關閉，同時關閉兩個服務
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := mcpSrv.Shutdown(ctx); err != nil {
			log.Printf("mcp/sse server shutdown error: %v", err)
		}
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("fiber server shutdown error: %v", err)
		}
	}()

	// Fiber API server 跑在主 goroutine，程式的生命週期跟著它走
	log.Println("API server (fiber) listening on :9002")
	if err := app.Listen(":9002"); err != nil {
		log.Fatalf("fiber server error: %v", err)
	}
	log.Println("API server stopped")
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
