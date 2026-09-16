package api_handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

const (
	aiConnTimeout = 1500 * time.Millisecond
	aiRespTimeout = 30 * time.Second
)

// AiChatRequest is the request body for POST /api/ai-chat.
type AiChatRequest struct {
	SystemPrompt string `json:"system_prompt"`
	Content      string `json:"content"`
	Model        string `json:"model,omitempty"` // optional, defaults to "chat"
}

// AiChatResponse is the response body for POST /api/ai-chat.
type AiChatResponse struct {
	Content string `json:"content"`
}

// AiChat handles POST /api/ai-chat — proxies a chat request to an
// OpenAI-compatible server, distinguishing connection timeout from
// response timeout in the error response.
func AiChat(c fiber.Ctx) error {
	var req AiChatRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.SystemPrompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "system_prompt is required"})
	}
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "content is required"})
	}

	serverURL := os.Getenv("AI_SERVER_URL")
	apiKey := os.Getenv("AI_API_KEY")
	if serverURL == "" || apiKey == "" {
		zap.L().Error("AI server not configured", zap.String("hint", "set AI_SERVER_URL and AI_API_KEY in .env"))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "AI server is not configured"})
	}

	model := req.Model
	if model == "" {
		model = "chat"
	}

	// Build OpenAI-compatible chat completion payload
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": req.SystemPrompt},
			{"role": "user", "content": req.Content},
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to build request"})
	}

	// Transport with a short dial timeout for connection phase
	dialer := &net.Dialer{Timeout: aiConnTimeout}
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: aiConnTimeout,
	}

	// Context timeout governs the response phase (reading the body)
	reqCtx, cancel := context.WithTimeout(c.Context(), aiRespTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost,
		serverURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Transport: transport}
	resp, err := client.Do(httpReq)
	if err != nil {
		// Classify the timeout
		switch classifyTimeout(err) {
		case "connection_timeout":
			zap.L().Warn("AI chat: connection timeout",
				zap.String("url", serverURL),
				zap.Duration("timeout", aiConnTimeout),
			)
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
				"error":  "connection_timeout",
				"detail": fmt.Sprintf("failed to connect to AI server within %v", aiConnTimeout),
			})
		case "response_timeout":
			zap.L().Warn("AI chat: response timeout",
				zap.String("url", serverURL),
				zap.Duration("timeout", aiRespTimeout),
			)
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
				"error":  "response_timeout",
				"detail": fmt.Sprintf("AI server did not respond within %v", aiRespTimeout),
			})
		default:
			zap.L().Error("AI chat request failed", zap.Error(err))
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to call AI server"})
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
				"error":  "response_timeout",
				"detail": fmt.Sprintf("AI server did not respond within %v", aiRespTimeout),
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to read AI response"})
	}

	if resp.StatusCode != http.StatusOK {
		zap.L().Error("AI server returned error", zap.Int("status", resp.StatusCode), zap.String("body", string(respBody)))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":  fmt.Sprintf("AI server returned %d", resp.StatusCode),
			"detail": string(respBody),
		})
	}

	// Parse OpenAI-compatible response
	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &aiResp); err != nil {
		zap.L().Error("failed to parse AI response", zap.Error(err))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to parse AI response"})
	}
	if len(aiResp.Choices) == 0 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "AI returned no choices"})
	}

	zap.L().Info("ai-chat ok", zap.String("model", model), zap.Int("resp_len", len(respBody)))
	return c.JSON(AiChatResponse{Content: aiResp.Choices[0].Message.Content})
}

// classifyTimeout inspects the error and returns a category string:
// "connection_timeout", "response_timeout", or "unknown".
//
// Logic:
//   - If the error chain contains context.DeadlineExceeded, the overall
//     request context expired → response timeout.
//   - If the error is a net.Error with Timeout()==true (produced by the
//     dialer), it means the TCP/TLS connection could not be established
//     in time → connection timeout.
//
// Because aiConnTimeout (1.5s) << aiRespTimeout (30s), a context deadline
// exceeded can only happen after the connection succeeded, making the
// classification unambiguous.
func classifyTimeout(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "response_timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "connection_timeout"
	}
	return "unknown"
}
