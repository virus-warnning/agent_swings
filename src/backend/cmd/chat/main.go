package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ChatRequest struct {
	SystemPrompt string `json:"system_prompt"`
	Content      string `json:"content"`
	Model        string `json:"model,omitempty"`
}

type ChatResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func main() {
	model := "chat"
	if len(os.Args) > 1 {
		model = os.Args[1]
	}

	payload := ChatRequest{
		SystemPrompt: "你是一位專業翻譯人員。只負責將英文翻譯繁體中文，不添加解釋、備註或額外說明。直接輸出翻譯結果，盡可能用最短時間完成。",
		Content: `SyntaxError: Lexical error on line 1. Unrecognized text.
flowchart TD2  A1(["👩🏻‍💻`,
		Model: model,
	}
	body, _ := json.Marshal(payload)

	fmt.Printf("POST http://localhost:9002/api/ai-chat (model: %s)\n", model)
	fmt.Printf("Request: %s\n", string(body))
	fmt.Println("─────────────────────────────────────")

	start := time.Now()
	resp, err := http.Post("http://localhost:9002/api/ai-chat", "application/json", bytes.NewReader(body))
	elapsed := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status:  %d\n", resp.StatusCode)
	fmt.Printf("Time:    %v\n", elapsed.Round(time.Millisecond))

	var cr ChatResponse
	if err := json.Unmarshal(respBody, &cr); err != nil {
		fmt.Printf("Raw response: %s\n", string(respBody))
		return
	}
	if cr.Error != "" {
		fmt.Printf("Error:  %s\n", cr.Error)
		fmt.Printf("Detail: %s\n", cr.Detail)
	} else {
		fmt.Printf("Reply:  %s\n", cr.Content)
	}
}
