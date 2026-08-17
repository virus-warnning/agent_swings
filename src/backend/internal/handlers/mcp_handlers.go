package mcp_handler

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input SayHiInput) (
	*mcp.CallToolResult,
	SayHiOutput,
	error,
) {
	return nil, SayHiOutput{Greeting: "Hi " + input.Name}, nil
}

func XmlTidy(ctx context.Context, req *mcp.CallToolRequest, input TidyInput) (
	*mcp.CallToolResult,
	TidyOutput,
	error,
) {
	// TODO: tidy
	return nil, TidyOutput{XmlContent: "<text>sucks</text>"}, nil
}
