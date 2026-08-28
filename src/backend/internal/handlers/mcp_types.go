package mcp_handler

type MermaidToImageInput struct {
	Syntax string `json:"syntax" jsonschema:"todo"`
	Format string `json:"format" jsonschema:"todo"`
}

type MermaidToImageOutput struct {
	Image string `json:"image" jsonschema:"todo"`
}

type SayHiInput struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type SayHiOutput struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

type XmlFormatterInput struct {
	XmlContent string `json:"xml"`
}

type XmlFormatterOutput struct {
	XmlContent string `json:"xml"`
}
