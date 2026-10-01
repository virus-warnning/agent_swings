package mcp_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/go-xmlfmt/xmlfmt"
	"github.com/ledongthuc/pdf"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// mermaidServiceURL is the base URL of the kroki-mermaid container.
// It is a var (rather than const) so tests can point it at a local test server.
var mermaidServiceURL = "http://localhost:8000/mermaid"

// MermaidToImageInput is the input for the MermaidToImage tool.
type MermaidToImageInput struct {
	Syntax string `json:"syntax" jsonschema:"Syntax of mermaid chart"`
	Format string `json:"format" jsonschema:"Output format, svg or png"`
}

// mermaidFormatMime maps a supported output format to its MIME type.
var mermaidFormatMime = map[string]string{
	"svg": "image/svg+xml",
	"png": "image/png",
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input SayHiInput) (
	*mcp.CallToolResult,
	SayHiOutput,
	error,
) {
	return nil, SayHiOutput{Greeting: "Hi " + input.Name}, nil
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

func XmlFormat(ctx context.Context, req *mcp.CallToolRequest, input XmlFormatterInput) (
	*mcp.CallToolResult,
	XmlFormatterOutput,
	error,
) {
	// Format the XML with two-space indentation.
	formatted := xmlfmt.FormatXML(input.XmlContent, "", "  ")

	return nil, XmlFormatterOutput{XmlContent: formatted}, nil
}

// duckduckgoAPIBase is the base URL of the DuckDuckGo Instant Answer API.
// Reference: https://duckduckgo.com/api
// It is a var (rather than const) so tests can point it at a local test server.
var duckduckgoAPIBase = "https://duckduckgo.com/"

// DuckDuckGoSearch queries the DuckDuckGo Instant Answer API with the given query
// and returns the parsed instant answer, including heading, abstract and related
// results (topics and related links).
func DuckDuckGoSearch(ctx context.Context, req *mcp.CallToolRequest, input DuckDuckGoSearchInput) (
	*mcp.CallToolResult,
	DuckDuckGoSearchOutput,
	error,
) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return nil, DuckDuckGoSearchOutput{}, fmt.Errorf("query must not be empty")
	}

	zap.L().Info("收到 DuckDuckGo 搜尋請求", zap.String("query", query))

	// Build the API endpoint with the encoded query.
	apiURL, err := url.Parse(duckduckgoAPIBase)
	if err != nil {
		return nil, DuckDuckGoSearchOutput{}, fmt.Errorf("parse api url: %w", err)
	}
	apiURL.Path = "/i.js"
	q := apiURL.Query()
	q.Set("q", query)
	apiURL.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, DuckDuckGoSearchOutput{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("User-Agent", "swings/1.0 (+mcp)")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, DuckDuckGoSearchOutput{}, fmt.Errorf("call duckduckgo api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, DuckDuckGoSearchOutput{}, fmt.Errorf("duckduckgo api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, DuckDuckGoSearchOutput{}, fmt.Errorf("read response: %w", err)
	}

	var output DuckDuckGoSearchOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return nil, DuckDuckGoSearchOutput{}, fmt.Errorf("decode response: %w", err)
	}

	// The Instant Answer API returns topics/related links with different
	// shapes; normalise them into a flat result list for stable output.
	output.Results = flattenDuckDuckGoResults(body)

	return nil, output, nil
}

// flattenDuckDuckGoResults parses the raw response and builds a flat list of
// related results and topics from the "RelatedTopics" field, which may contain
// either flat objects or nested group objects.
func flattenDuckDuckGoResults(raw []byte) []DuckDuckGoSearchResult {
	var envelope struct {
		RelatedTopics []struct {
			Name string `json:"name"`
			Text string `json:"text"`
			URL  string `json:"url"`
			// Nested group form
			FirstURL string `json:"first_url"`
			Topics   []struct {
				Text string `json:"text"`
				URL  string `json:"url"`
			} `json:"topics"`
		} `json:"RelatedTopics"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil
	}

	results := make([]DuckDuckGoSearchResult, 0, len(envelope.RelatedTopics))
	for _, rt := range envelope.RelatedTopics {
		if rt.Topics != nil {
			// Nested group: one result per sub-topic.
			for _, t := range rt.Topics {
				results = append(results, DuckDuckGoSearchResult{
					Name: rt.Name,
					Text: t.Text,
					URL:  t.URL,
				})
			}
			continue
		}
		if rt.Text != "" || rt.URL != "" {
			results = append(results, DuckDuckGoSearchResult{
				Name: rt.Name,
				Text: rt.Text,
				URL:  rt.URL,
			})
		}
	}
	return results
}

// PdfExtractText reads a PDF file from disk and extracts its text content page by page.
func PdfExtractText(ctx context.Context, req *mcp.CallToolRequest, input PdfExtractTextInput) (
	*mcp.CallToolResult,
	PdfExtractTextOutput,
	error,
) {
	path := strings.TrimSpace(input.Path)
	if path == "" {
		return nil, PdfExtractTextOutput{}, fmt.Errorf("path must not be empty")
	}

	zap.L().Info("收到 PDF 文字萃取請求", zap.String("path", path))

	f, reader, err := pdf.Open(path)
	if err != nil {
		return nil, PdfExtractTextOutput{}, fmt.Errorf("open pdf %q: %w", path, err)
	}
	defer f.Close()

	if reader.NumPage() <= 0 {
		return nil, PdfExtractTextOutput{Path: path, Content: ""}, nil
	}

	var sb strings.Builder
	for p := 1; p <= reader.NumPage(); p++ {
		page := reader.Page(p)
		text, err := page.GetPlainText(nil)
		if err != nil {
			return nil, PdfExtractTextOutput{}, fmt.Errorf("extract text page %d: %w", p, err)
		}
		if p > 1 {
			sb.WriteString("\n")
		}
		sb.WriteString(text)
	}

	return nil, PdfExtractTextOutput{
		Path:    path,
		Content: sb.String(),
	}, nil
}

// mediawikiAPIBase is the base URL of the user's MediaWiki instance.
// It is a var (rather than const) so tests can point it at a local test server.
var mediawikiAPIBase = "https://wiki.fundamental-ramen.com"

// MediaWikiSearch queries the MediaWiki API for pages matching the given query.
func MediaWikiSearch(ctx context.Context, req *mcp.CallToolRequest, input MediaWikiSearchInput) (
	*mcp.CallToolResult,
	MediaWikiSearchOutput,
	error,
) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return nil, MediaWikiSearchOutput{}, fmt.Errorf("query must not be empty")
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	zap.L().Info("收到 MediaWiki 搜尋請求",
		zap.String("query", query),
		zap.Int("limit", limit),
	)

	// Build the MediaWiki API endpoint.
	apiURL, err := url.Parse(mediawikiAPIBase + "/api.php")
	if err != nil {
		return nil, MediaWikiSearchOutput{}, fmt.Errorf("parse api url: %w", err)
	}
	q := apiURL.Query()
	q.Set("action", "query")
	q.Set("list", "search")
	q.Set("srsearch", query)
	q.Set("srlimit", fmt.Sprintf("%d", limit))
	q.Set("format", "json")
	q.Set("origin", "*") // CORS support
	apiURL.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, MediaWikiSearchOutput{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("User-Agent", "swings/1.0 (+mcp)")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, MediaWikiSearchOutput{}, fmt.Errorf("call mediawiki api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, MediaWikiSearchOutput{}, fmt.Errorf("mediawiki api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, MediaWikiSearchOutput{}, fmt.Errorf("read response: %w", err)
	}

	var mwResp struct {
		Query struct {
			SearchInfo struct {
				TotalHits int `json:"totalhits"`
			} `json:"searchinfo"`
			Search []struct {
				Title    string `json:"title"`
				Fragment string `json:"snippet"`
			} `json:"search"`
		} `json:"query"`
	}
	if err := json.Unmarshal(body, &mwResp); err != nil {
		return nil, MediaWikiSearchOutput{}, fmt.Errorf("decode response: %w", err)
	}

	results := make([]MediaWikiSearchResult, 0, len(mwResp.Query.Search))
	for _, item := range mwResp.Query.Search {
		results = append(results, MediaWikiSearchResult{
			Title:    item.Title,
			Fragment: stripHTML(item.Fragment),
			Url:      mediawikiAPIBase + "/wiki/" + url.PathEscape(item.Title),
		})
	}

	return nil, MediaWikiSearchOutput{
		TotalHits: mwResp.Query.SearchInfo.TotalHits,
		Results:   results,
	}, nil
}

// stripHTML removes simple HTML tags (e.g. <span class="searchmatch">) from a string.
func stripHTML(s string) string {
	var sb strings.Builder
	inTag := false
	for _, ch := range s {
		switch {
		case ch == '<':
			inTag = true
		case ch == '>':
			inTag = false
		default:
			if !inTag {
				sb.WriteRune(ch)
			}
		}
	}
	return sb.String()
}

// MarkdownTableAlign aligns the column widths of all Markdown tables in the given content.
// Non-ASCII characters are treated as width 2, ASCII characters as width 1.
func MarkdownTableAlign(ctx context.Context, req *mcp.CallToolRequest, input MarkdownTableAlignInput) (
	*mcp.CallToolResult,
	MarkdownTableAlignOutput,
	error,
) {
	zap.L().Info("收到 Markdown 表格對齊請求", zap.Int("content_len", len(input.Markdown)))

	lines := strings.Split(input.Markdown, "\n")
	var out strings.Builder

	var tableStart int // -1 when not in a table

	flushTable := func(endIdx int) {
		if tableStart < 0 {
			return
		}
		tableLines := lines[tableStart:endIdx]
		aligned := alignTable(tableLines)
		for i, l := range aligned {
			if i > 0 {
				out.WriteString("\n")
			}
			out.WriteString(l)
		}
		tableStart = -1
	}

	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		isTable := strings.HasPrefix(trimmed, "|")
		if isTable {
			if tableStart < 0 {
				// entering a table block
				tableStart = i
			}
		} else {
			flushTable(i)
			if i > 0 {
				out.WriteString("\n")
			}
			out.WriteString(line)
		}
	}
	// flush any trailing table
	flushTable(len(lines))

	return nil, MarkdownTableAlignOutput{Markdown: out.String()}, nil
}

// alignTable takes a slice of table rows (each a full line starting with "|")
// and returns the rows with columns padded to equal display width.
func alignTable(rows []string) []string {
	// Parse each row into cells.
	var parsed [][]string
	for _, row := range rows {
		parsed = append(parsed, splitTableRow(row))
	}

	// Determine the max display width per column.
	numCols := 0
	for _, cells := range parsed {
		if len(cells) > numCols {
			numCols = len(cells)
		}
	}
	if numCols == 0 {
		return rows
	}
	colWidths := make([]int, numCols)
	for _, cells := range parsed {
		for i, cell := range cells {
			w := displayWidth(cell)
			if w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}

	// Pad each cell to its column width.
	result := make([]string, len(parsed))
	for r, cells := range parsed {
		padded := make([]string, numCols)
		for c := 0; c < numCols; c++ {
			var cell string
			if c < len(cells) {
				cell = cells[c]
			}
			padded[c] = padToWidth(cell, colWidths[c])
		}
		result[r] = "| " + strings.Join(padded, " | ") + " |"
	}
	return result
}

// splitTableRow splits a markdown table row into its cells.
// It handles rows like: | a | b | c |  →  ["a", "b", "c"]
func splitTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	// Remove leading and trailing pipe
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// displayWidth calculates the display width of a string where
// non-ASCII runes count as 2 and ASCII runes count as 1.
func displayWidth(s string) int {
	w := 0
	for _, ch := range s {
		if ch < 0x80 {
			w++
		} else {
			w += 2
		}
	}
	return w
}

// padToWidth pads s with trailing spaces so its display width equals target.
func padToWidth(s string, target int) string {
	w := displayWidth(s)
	if w >= target {
		return s
	}
	return s + strings.Repeat(" ", target-w)
}

// bibleBookNames maps Chinese book abbreviations to their full names.
var bibleBookNames = map[string]string{
	"創": "創世記", "出": "出埃及記", "利": "利未記", "民": "民數記", "申": "申命記",
	"書": "約書亞記", "士": "士師記", "得": "路得記", "撒上": "撒母耳記上", "撒下": "撒母耳記下",
	"王上": "列王紀上", "王下": "列王紀下", "代上": "歷代志上", "代下": "歷代志下",
	"拉": "以斯拉記", "尼": "尼希米記", "斯": "以斯帖記", "伯": "約伯記",
	"詩": "詩篇", "箴": "箴言", "傳": "傳道書", "歌": "雅歌",
	"賽": "以賽亞書", "耶": "耶利米書", "哀": "耶利米哀歌", "結": "以西結書", "但": "但以理書",
	"何": "何西阿書", "珥": "約珥書", "摩": "阿摩司書", "俄": "俄巴底亞書",
	"拿": "約拿書", "彌": "彌迦書", "鴻": "那鴻書", "哈": "哈巴谷書",
	"番": "西番雅書", "該": "哈該書", "亞": "撒迦利亞書", "瑪": "瑪拉基書",
	"太": "馬太福音", "可": "馬可福音", "路": "路加福音", "約": "約翰福音", "徒": "使徒行傳",
	"羅": "羅馬書", "林前": "哥林多前書", "林後": "哥林多後書", "加": "加拉太書", "弗": "以弗所書",
	"腓": "腓立比書", "西": "歌羅西書", "帖前": "帖撒羅尼迦前書", "帖後": "帖撒羅尼迦後書",
	"提前": "提摩太前書", "提後": "提摩太後書", "多": "提多書", "門": "腓利門書",
	"來": "希伯來書", "雅": "雅各書", "彼前": "彼得前書", "彼後": "彼得後書",
	"約一": "約翰壹書", "約二": "約翰貳書", "約三": "約翰參書", "猶": "猶大書", "啟": "啟示錄",
}

// bibleVerseRe matches verse entries like <b>1:1</b> <span class="bstw"> text </span>
var bibleVerseRe = regexp.MustCompile(`<b>(\d+):(\d+)</b>\s*<span class="bstw">(.*?)</span>`)

// bibleAPIBase is the base URL of the Chinese Bible online (cb.fhl.net).
// It is a var (rather than const) so tests can point it at a local test server.
var bibleAPIBase = "https://cb.fhl.net"

// BibleFetch retrieves a full chapter of Bible text (TCV 2019) from cb.fhl.net.
func BibleFetch(ctx context.Context, req *mcp.CallToolRequest, input BibleFetchInput) (
	*mcp.CallToolResult,
	BibleFetchOutput,
	error,
) {
	book := strings.TrimSpace(input.Book)
	if book == "" {
		return nil, BibleFetchOutput{}, fmt.Errorf("book abbreviation must not be empty")
	}
	fullName, ok := bibleBookNames[book]
	if !ok {
		return nil, BibleFetchOutput{}, fmt.Errorf("unknown book abbreviation %q; use abbreviations like 創, 詩, 箴, 太, 啟, etc.", book)
	}
	if input.Chapter < 1 {
		return nil, BibleFetchOutput{}, fmt.Errorf("chapter must be >= 1")
	}

	zap.L().Info("收到聖經經文請求", zap.String("book", fullName), zap.Int("chapter", input.Chapter))

	apiURL := fmt.Sprintf(bibleAPIBase+"/read1.php?chineses=%s&chap=%d&VERSION1=tcv2019&TABFLAG=0",
		url.PathEscape(book), input.Chapter)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, BibleFetchOutput{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("User-Agent", "swings/1.0 (+mcp)")
	httpReq.Header.Set("Accept", "text/html")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, BibleFetchOutput{}, fmt.Errorf("fetch bible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, BibleFetchOutput{}, fmt.Errorf("cb.fhl.net returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, BibleFetchOutput{}, fmt.Errorf("read response: %w", err)
	}

	// Extract chapter title (e.g. 創世記的價值)
	var chapterTitle string
	titleRe := regexp.MustCompile(`<b>([^<]+)</b></font><br`)
	if m := titleRe.FindSubmatch(body); m != nil {
		chapterTitle = strings.TrimSpace(string(m[1]))
	}

	// Parse all verses
	matches := bibleVerseRe.FindAllSubmatch(body, -1)
	verses := make([]BibleVerse, 0, len(matches))
	for _, m := range matches {
		vNum, _ := strconv.Atoi(string(m[2]))
		text := strings.TrimSpace(string(m[3]))
		verses = append(verses, BibleVerse{
			Verse:   vNum,
			Content: text,
		})
	}

	if len(verses) == 0 {
		return nil, BibleFetchOutput{}, fmt.Errorf("no verses found for %s chapter %d; the chapter may not exist in this translation", fullName, input.Chapter)
	}

	out := BibleFetchOutput{
		Book:    fullName,
		Chapter: input.Chapter,
		Verses:  verses,
	}
	_ = chapterTitle // reserved for future use

	return nil, out, nil
}
