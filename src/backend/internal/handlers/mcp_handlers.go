package mcp_handler

import (
	"context"

	"github.com/go-xmlfmt/xmlfmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
