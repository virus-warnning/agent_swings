package mcp_handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unsafe"

	"github.com/go-xmlfmt/xmlfmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// mermaidServiceURL is the base URL of the kroki-mermaid container.
const mermaidServiceURL = "http://localhost:8000/mermaid"

// mermaidFormatMime maps a supported output format to its MIME type.
var mermaidFormatMime = map[string]string{
	"svg": "image/svg+xml",
	"png": "image/png",
}

func MermaidToImage(ctx context.Context, req *mcp.CallToolRequest, input MermaidToImageInput) (
	*mcp.CallToolResult,
	any,
	error,
) {
	// Read and normalize the requested output format.
	format := strings.ToLower(strings.TrimSpace(input.Format))
	_, ok := mermaidFormatMime[format]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported format %q: must be one of svg, png", input.Format)
	}

	zap.L().Info("收到 MCP 請求",
		zap.String("format", input.Format),
		zap.String("syntax", input.Syntax),
	)

	// Send the mermaid source to the kroki-mermaid service.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, mermaidServiceURL+"/"+format, strings.NewReader(input.Syntax))
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("call mermaid service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, nil, fmt.Errorf("mermaid service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read image: %w", err)
	}

	var result *mcp.CallToolResult
	if input.Format == "svg" {
		svgSource := unsafe.String(unsafe.SliceData(imageBytes), len(imageBytes))
		result = &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: svgSource,
				},
			},
		}
		zap.L().Info("完成 SVG 轉檔", zap.String("svg", svgSource))
	} else {
		// base64Data := base64.StdEncoding.EncodeToString(imageBytes)
		result = &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.ImageContent{
					Data:     imageBytes,
					MIMEType: "image/png",
				},
			},
		}
	}

	return result, nil, nil
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input SayHiInput) (
	*mcp.CallToolResult,
	SayHiOutput,
	error,
) {
	return nil, SayHiOutput{Greeting: "Hi " + input.Name}, nil
}

func XmlFormat(ctx context.Context, req *mcp.CallToolRequest, input XmlFormatterInput) (
	*mcp.CallToolResult,
	XmlFormatterOutput,
	error,
) {
	// Format the XML with two-space indentation.
	formatted := xmlfmt.FormatXML(input.XmlContent, "", "  ")

	return nil, XmlFormatterOutput{XmlContent: formatted}, nil
}
