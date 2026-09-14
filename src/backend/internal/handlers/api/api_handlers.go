package api_handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// mermaidServiceURL is the base URL of the kroki-mermaid container.
const mermaidServiceURL = "http://localhost:8000/mermaid"

func MermaidToImage(c fiber.Ctx) error {
	body := c.Body()
	if len(body) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "request body is required"})
	}

	syntax := string(body)
	ctx := c.Context()
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost,
		mermaidServiceURL+"/svg", strings.NewReader(syntax))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	httpReq.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		zap.L().Error("call kroki failed", zap.Error(err))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to call kroki service"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		zap.L().Error("kroki returned error", zap.Int("status", resp.StatusCode), zap.String("body", string(respBody)))
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": fmt.Sprintf("kroki returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody))),
		})
	}

	svg, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read svg"})
	}

	zap.L().Info("mermaid-to-svg ok", zap.Int("size", len(svg)))
	c.Response().Header.Set("Content-Type", "image/svg+xml")
	return c.Send(svg)
}
