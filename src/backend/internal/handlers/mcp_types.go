package mcp_handler

type SayHiInput struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type SayHiOutput struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

type TidyInput struct {
	XmlContent string `json:"xml"`
}

type TidyOutput struct {
	XmlContent string `json:"xml"`
}
