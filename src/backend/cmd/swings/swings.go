package main

import (
	"context"
	"log"
	"net/http"

	mcp_handler "fundamental-ramen.com/agent-swing/internal/handlers"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

	// HTTP (Streamable + SSE) 用 goroutine 跑，不然會卡住主流程
	go func() {
		log.Println("HTTP server (streamable + sse) listening on :8080")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Stdio 放在主 goroutine，程式的生命週期跟著它走
	if err := svr.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func newMcpServer() *mcp.Server {
	svr := mcp.NewServer(&mcp.Implementation{Name: "stdio", Version: "v1.0.0"}, nil)
	mcp.AddTool(svr, &mcp.Tool{Name: "greet", Description: "say hi"}, mcp_handler.SayHi)
	mcp.AddTool(svr, &mcp.Tool{Name: "xmlTidy", Description: "xxx"}, mcp_handler.XmlTidy)
	return svr
}
