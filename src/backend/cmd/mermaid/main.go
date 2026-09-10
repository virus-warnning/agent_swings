package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// apiBaseURL is the base URL of the swings API server (Fiber, :9002).
// 可透過環境變數 SWINGS_API 覆蓋，例如 http://localhost:9002
const defaultAPIBaseURL = "http://localhost:9002"

// defaultSyntax 是未提供參數時使用的範例 mermaid 語法
const defaultSyntax = `graph TD
    A(中文也可以嗎) --> B(Is it?)
    B -->|Yes| C(可以 der)
    B -->|No| D(End)`

func main() {
	// 語法來源：第一個參數優先，否則使用預設範例
	syntax := defaultSyntax
	if len(os.Args) > 1 {
		syntax = strings.Join(os.Args[1:], " ")
	}

	baseURL := os.Getenv("SWINGS_API")
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}

	svg, err := mermaidToSVG(baseURL, syntax)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mermaid-to-svg failed: %v\n", err)
		os.Exit(1)
	}

	// 將 SVG 原始碼輸出到 console
	fmt.Print(svg)
}

// mermaidToSVG 呼叫 /api/mermaid-to-svg，傳入 mermaid 語法並回傳 SVG 原始碼
func mermaidToSVG(baseURL, syntax string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/mermaid-to-svg", strings.NewReader(syntax))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return string(body), nil
}
