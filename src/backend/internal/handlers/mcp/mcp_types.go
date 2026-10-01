package mcp_handler

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

// PdfExtractTextInput is the input for the PdfExtractText tool.
type PdfExtractTextInput struct {
	Path string `json:"path" jsonschema:"Absolute or relative path to the PDF file on disk"`
}

// PdfExtractTextOutput is the extracted text content of a PDF.
type PdfExtractTextOutput struct {
	Path    string `json:"path" jsonschema:"The path of the processed PDF file"`
	Content string `json:"content" jsonschema:"The extracted text content"`
}

// DuckDuckGoSearchInput is the input for the DuckDuckGoSearch tool.
type DuckDuckGoSearchInput struct {
	Query string `json:"query" jsonschema:"The search query to look up on DuckDuckGo"`
}

// DuckDuckGoSearchResult represents a single related result or topic.
type DuckDuckGoSearchResult struct {
	Name string `json:"name,omitempty" jsonschema:"Optional group or topic name"`
	Text string `json:"text,omitempty" jsonschema:"Description or snippet text"`
	URL  string `json:"url,omitempty" jsonschema:"Related link URL"`
}

// DuckDuckGoSearchOutput is the result of a DuckDuckGo search.
type DuckDuckGoSearchOutput struct {
	Heading        string                   `json:"heading,omitempty" jsonschema:"Main heading of the instant answer"`
	Answer         string                   `json:"answer,omitempty" jsonschema:"Direct short answer, if available"`
	AbstractText   string                   `json:"abstract,omitempty" jsonschema:"Summary or abstract text"`
	AbstractSource string                   `json:"abstractSource,omitempty" jsonschema:"Source of the abstract"`
	AbstractURL    string                   `json:"abstractUrl,omitempty" jsonschema:"URL of the abstract source"`
	Results        []DuckDuckGoSearchResult `json:"results" jsonschema:"Related results and topics"`
}

// MediaWikiSearchInput is the input for the MediaWikiSearch tool.
type MediaWikiSearchInput struct {
	Query string `json:"query" jsonschema:"The search query to look up on MediaWiki"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of results to return (default 10, max 50)"`
}

// MediaWikiSearchResult represents a single search hit from MediaWiki.
type MediaWikiSearchResult struct {
	Title    string `json:"title" jsonschema:"Page title"`
	Fragment string `json:"fragment,omitempty" jsonschema:"Snippet or excerpt from the page"`
	Url      string `json:"url,omitempty" jsonschema:"Full URL to the page"`
}

// MediaWikiSearchOutput is the result of a MediaWiki search.
type MediaWikiSearchOutput struct {
	TotalHits int                     `json:"totalHits,omitempty" jsonschema:"Total number of matching pages"`
	Results   []MediaWikiSearchResult `json:"results" jsonschema:"Matching pages"`
}

// MarkdownTableAlignInput is the input for the MarkdownTableAlign tool.
type MarkdownTableAlignInput struct {
	Markdown string `json:"markdown" jsonschema:"Markdown content containing one or more tables to align"`
}

// MarkdownTableAlignOutput is the result after aligning table columns.
type MarkdownTableAlignOutput struct {
	Markdown string `json:"markdown" jsonschema:"Markdown content with aligned table column widths"`
}

// BibleFetchInput is the input for the BibleFetch tool.
type BibleFetchInput struct {
	Book    string `json:"book" jsonschema:"Book abbreviation in Chinese (e.g. 創=Genesis, 詩=Psalms, 箴=Proverbs, 太=Matthew, 啟=Revelation)"`
	Chapter int    `json:"chapter" jsonschema:"Chapter number (1-based)"`
}

// BibleVerse represents a single verse.
type BibleVerse struct {
	Verse   int    `json:"verse" jsonschema:"Verse number within the chapter"`
	Content string `json:"content" jsonschema:"Verse text"`
}

// BibleFetchOutput is the result of fetching a Bible chapter.
type BibleFetchOutput struct {
	Book    string       `json:"book" jsonschema:"Full book name in Chinese"`
	Chapter int          `json:"chapter" jsonschema:"Chapter number"`
	Verses  []BibleVerse `json:"verses" jsonschema:"All verses in the chapter"`
}
