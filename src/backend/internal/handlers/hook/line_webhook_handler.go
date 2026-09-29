package hook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// LineWebhookHandler handles LINE Messaging API webhook events.
// Verifies the x-line-signature header before processing the event.
func LineWebhookHandler(c fiber.Ctx) error {
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	if channelSecret == "" {
		zap.L().Error("LINE_CHANNEL_SECRET is not set")
		return c.Status(fiber.StatusServiceUnavailable).SendString("server misconfigured")
	}

	rawBody := c.Body()

	// Verify signature
	signature := c.Get("x-line-signature")
	if !verifyLineSignature(rawBody, signature, channelSecret) {
		zap.L().Warn("LINE Webhook signature verification failed", zap.String("signature", signature))
		return c.Status(fiber.StatusUnauthorized).SendString("invalid signature")
	}

	// TODO: process LINE events (message, follow, unfollow, etc.)
	zap.L().Info("received LINE Webhook event", zap.String("body", string(rawBody)))

	return c.SendStatus(fiber.StatusOK)
}

// verifyLineSignature verifies the HMAC-SHA256 signature from LINE Webhook.
// Per LINE official docs: signature = hex(HMAC-SHA256(request_body, channel_secret))
func verifyLineSignature(body []byte, signature, channelSecret string) bool {
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}
