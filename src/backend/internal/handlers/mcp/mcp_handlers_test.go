package mcp_handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// swapURL points a package-level URL var at v for the duration of the test and
// restores the original value afterwards.
func swapURL(t *testing.T, p *string, v string) {
	t.Helper()
	orig := *p
	*p = v
	t.Cleanup(func() { *p = orig })
}

// --- Pure handlers ---------------------------------------------------------

func TestSayHi(t *testing.T) {
	res, out, err := SayHi(context.Background(), nil, SayHiInput{Name: "Raymond"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil CallToolResult, got %+v", res)
	}
	if out.Greeting != "Hi Raymond" {
		t.Errorf("Greeting = %q, want %q", out.Greeting, "Hi Raymond")
	}
}

func TestXmlFormat(t *testing.T) {
	res, out, err := XmlFormat(context.Background(), nil, XmlFormatterInput{XmlContent: "<a><b/></a>"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil CallToolResult, got %+v", res)
	}
	if !strings.Contains(out.XmlContent, "  <b/>") {
		t.Errorf("expected two-space indentation, got %q", out.XmlContent)
	}
}

func TestMarkdownTableAlign(t *testing.T) {
	t.Run("pure table", func(t *testing.T) {
		_, out, err := MarkdownTableAlign(context.Background(), nil,
			MarkdownTableAlignInput{Markdown: "| a | bb |\n| cc | d |"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "| a  | bb |\n| cc | d  |"
		if out.Markdown != want {
			t.Errorf("got %q, want %q", out.Markdown, want)
		}
	})

	t.Run("table mixed with prose", func(t *testing.T) {
		in := "intro\n| a | bb |\n| cc | d |\ntail"
		_, out, err := MarkdownTableAlign(context.Background(), nil,
			MarkdownTableAlignInput{Markdown: in})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "intro\n| a  | bb |\n| cc | d  |\ntail"
		if out.Markdown != want {
			t.Errorf("got %q, want %q", out.Markdown, want)
		}
	})
}

// --- Pure helper functions -------------------------------------------------

func TestStripHTML(t *testing.T) {
	got := stripHTML(`<span class="m">hit</span> ok`)
	if got != "hit ok" {
		t.Errorf("stripHTML = %q, want %q", got, "hit ok")
	}
}

func TestDisplayWidth(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"a", 1},
		{"abc", 3},
		{"中", 2},
		{"a中", 3},
	}
	for _, c := range cases {
		if got := displayWidth(c.in); got != c.want {
			t.Errorf("displayWidth(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestSplitTableRow(t *testing.T) {
	got := splitTableRow("| a | b |")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("splitTableRow = %v, want [a b]", got)
	}
}

func TestAlignTable(t *testing.T) {
	got := alignTable([]string{"| a | bb |", "| cc | d |"})
	want := []string{"| a  | bb |", "| cc | d  |"}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("alignTable = %v, want %v", got, want)
	}
}

// --- Network handlers (via httptest) --------------------------------------

func TestMermaidToImage(t *testing.T) {
	const svgBody = `<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`

	t.Run("svg returns TextContent", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/svg" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(svgBody))
		}))
		defer srv.Close()
		swapURL(t, &mermaidServiceURL, srv.URL)

		res, _, err := MermaidToImage(context.Background(), nil,
			MermaidToImageInput{Syntax: "graph TD; A-->B;", Format: "svg"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil CallToolResult")
		}
		if len(res.Content) != 1 {
			t.Fatalf("Content len = %d, want 1", len(res.Content))
		}
		tc, ok := res.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("expected *mcp.TextContent, got %T", res.Content[0])
		}
		if tc.Text != svgBody {
			t.Errorf("svg text = %q, want %q", tc.Text, svgBody)
		}
	})

	t.Run("png returns ImageContent", func(t *testing.T) {
		pngBytes := []byte{0x89, 'P', 'N', 'G', 1, 2}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/png" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write(pngBytes)
		}))
		defer srv.Close()
		swapURL(t, &mermaidServiceURL, srv.URL)

		res, _, err := MermaidToImage(context.Background(), nil,
			MermaidToImageInput{Syntax: "graph TD; A-->B;", Format: "png"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res == nil {
			t.Fatalf("expected non-nil CallToolResult")
		}
		ic, ok := res.Content[0].(*mcp.ImageContent)
		if !ok {
			t.Fatalf("expected *mcp.ImageContent, got %T", res.Content[0])
		}
		if ic.MIMEType != "image/png" {
			t.Errorf("MIMEType = %q, want image/png", ic.MIMEType)
		}
		if !bytes.Equal(ic.Data, pngBytes) {
			t.Errorf("image data = %v, want %v", ic.Data, pngBytes)
		}
	})

	t.Run("unsupported format", func(t *testing.T) {
		_, _, err := MermaidToImage(context.Background(), nil,
			MermaidToImageInput{Syntax: "graph TD", Format: "pdf"})
		if err == nil {
			t.Fatalf("expected error for unsupported format")
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()
		swapURL(t, &mermaidServiceURL, srv.URL)

		_, _, err := MermaidToImage(context.Background(), nil,
			MermaidToImageInput{Syntax: "graph TD", Format: "svg"})
		if err == nil {
			t.Fatalf("expected error for non-200 response")
		}
	})
}

func TestDuckDuckGoSearch(t *testing.T) {
	const body = `{"Heading":"Go language","AbstractText":"Go is a language.","RelatedTopics":[{"Text":"flat topic","URL":"https://example.com/a"},{"Name":"group","Topics":[{"Text":"sub1","URL":"https://example.com/b"},{"Text":"sub2","URL":"https://example.com/c"}]}]}`

	t.Run("success with flat and nested topics", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/i.js" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(body))
		}))
		defer srv.Close()
		swapURL(t, &duckduckgoAPIBase, srv.URL)

		_, out, err := DuckDuckGoSearch(context.Background(), nil, DuckDuckGoSearchInput{Query: "go"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Heading != "Go language" {
			t.Errorf("Heading = %q, want %q", out.Heading, "Go language")
		}
		if out.AbstractText != "Go is a language." {
			t.Errorf("AbstractText = %q", out.AbstractText)
		}
		if len(out.Results) != 3 {
			t.Fatalf("Results len = %d, want 3: %+v", len(out.Results), out.Results)
		}
		if out.Results[1].Name != "group" {
			t.Errorf("Results[1].Name = %q, want %q", out.Results[1].Name, "group")
		}
		if out.Results[2].Text != "sub2" {
			t.Errorf("Results[2].Text = %q, want %q", out.Results[2].Text, "sub2")
		}
	})

	t.Run("empty query", func(t *testing.T) {
		_, _, err := DuckDuckGoSearch(context.Background(), nil, DuckDuckGoSearchInput{Query: "   "})
		if err == nil {
			t.Fatalf("expected error for empty query")
		}
	})

	t.Run("non-200 response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()
		swapURL(t, &duckduckgoAPIBase, srv.URL)

		_, _, err := DuckDuckGoSearch(context.Background(), nil, DuckDuckGoSearchInput{Query: "go"})
		if err == nil {
			t.Fatalf("expected error for non-200 response")
		}
	})
}

func TestMediaWikiSearch(t *testing.T) {
	const body = `{"query":{"searchinfo":{"totalhits":2},"search":[` +
		`{"title":"Page One","snippet":"<span class=\"searchmatch\">hit</span> one"},` +
		`{"title":"Page Two","snippet":"plain two"}]}}`

	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api.php" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(body))
		}))
		defer srv.Close()
		swapURL(t, &mediawikiAPIBase, srv.URL)

		_, out, err := MediaWikiSearch(context.Background(), nil, MediaWikiSearchInput{Query: "hit"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.TotalHits != 2 {
			t.Errorf("TotalHits = %d, want 2", out.TotalHits)
		}
		if len(out.Results) != 2 {
			t.Fatalf("Results len = %d, want 2", len(out.Results))
		}
		if out.Results[0].Title != "Page One" {
			t.Errorf("Results[0].Title = %q", out.Results[0].Title)
		}
		if out.Results[0].Fragment != "hit one" {
			t.Errorf("Results[0].Fragment = %q, want %q", out.Results[0].Fragment, "hit one")
		}
		if !strings.HasSuffix(out.Results[0].Url, "/wiki/Page One") {
			t.Errorf("Results[0].Url = %q, want suffix /wiki/Page One", out.Results[0].Url)
		}
		if out.Results[1].Fragment != "plain two" {
			t.Errorf("Results[1].Fragment = %q, want %q", out.Results[1].Fragment, "plain two")
		}
	})

	t.Run("empty query", func(t *testing.T) {
		_, _, err := MediaWikiSearch(context.Background(), nil, MediaWikiSearchInput{Query: "  "})
		if err == nil {
			t.Fatalf("expected error for empty query")
		}
	})

	t.Run("limit clamped to max", func(t *testing.T) {
		var gotLimit string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotLimit = r.URL.Query().Get("srlimit")
			_, _ = w.Write([]byte(`{"query":{"searchinfo":{"totalhits":0},"search":[]}}`))
		}))
		defer srv.Close()
		swapURL(t, &mediawikiAPIBase, srv.URL)

		_, _, err := MediaWikiSearch(context.Background(), nil, MediaWikiSearchInput{Query: "x", Limit: 999})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotLimit != "50" {
			t.Errorf("srlimit = %q, want 50", gotLimit)
		}
	})

	t.Run("limit defaults when <=0", func(t *testing.T) {
		var gotLimit string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotLimit = r.URL.Query().Get("srlimit")
			_, _ = w.Write([]byte(`{"query":{"searchinfo":{"totalhits":0},"search":[]}}`))
		}))
		defer srv.Close()
		swapURL(t, &mediawikiAPIBase, srv.URL)

		_, _, err := MediaWikiSearch(context.Background(), nil, MediaWikiSearchInput{Query: "x", Limit: 0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotLimit != "10" {
			t.Errorf("srlimit = %q, want 10", gotLimit)
		}
	})
}

func TestBibleFetch(t *testing.T) {
	const chapterHTML = `<html><body>` +
		`<h2><font size="5"><b>創世記</b></font></h2>` +
		`<p><b>1:1</b> <span class="bstw">太初有上帝，</span>` +
		`<b>1:2</b> <span class="bstw">上帝創造了天地。</span></p>` +
		`</body></html>`

	newBibleServer := func(body string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/read1.php" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(body))
		}))
	}

	t.Run("success", func(t *testing.T) {
		srv := newBibleServer(chapterHTML)
		defer srv.Close()
		swapURL(t, &bibleAPIBase, srv.URL)

		_, out, err := BibleFetch(context.Background(), nil, BibleFetchInput{Book: "創", Chapter: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Book != "創世記" {
			t.Errorf("Book = %q, want %q", out.Book, "創世記")
		}
		if out.Chapter != 1 {
			t.Errorf("Chapter = %d, want 1", out.Chapter)
		}
		if len(out.Verses) != 2 {
			t.Fatalf("Verses len = %d, want 2: %+v", len(out.Verses), out.Verses)
		}
		if out.Verses[0].Verse != 1 {
			t.Errorf("Verses[0].Verse = %d, want 1", out.Verses[0].Verse)
		}
		if out.Verses[0].Content != "太初有上帝，" {
			t.Errorf("Verses[0].Content = %q", out.Verses[0].Content)
		}
	})

	t.Run("empty book", func(t *testing.T) {
		_, _, err := BibleFetch(context.Background(), nil, BibleFetchInput{Book: " ", Chapter: 1})
		if err == nil {
			t.Fatalf("expected error for empty book")
		}
	})

	t.Run("unknown book", func(t *testing.T) {
		_, _, err := BibleFetch(context.Background(), nil, BibleFetchInput{Book: "xyz", Chapter: 1})
		if err == nil {
			t.Fatalf("expected error for unknown book")
		}
	})

	t.Run("invalid chapter", func(t *testing.T) {
		_, _, err := BibleFetch(context.Background(), nil, BibleFetchInput{Book: "創", Chapter: 0})
		if err == nil {
			t.Fatalf("expected error for chapter < 1")
		}
	})

	t.Run("no verses in response", func(t *testing.T) {
		srv := newBibleServer(`<html><body><p>no verses here</p></body></html>`)
		defer srv.Close()
		swapURL(t, &bibleAPIBase, srv.URL)

		_, _, err := BibleFetch(context.Background(), nil, BibleFetchInput{Book: "創", Chapter: 1})
		if err == nil {
			t.Fatalf("expected error when no verses are present")
		}
	})
}

// --- PdfExtractText --------------------------------------------------------

// writeTestPDF writes a minimal, valid one-page PDF whose content stream draws
// the given text, so the test exercises a real success path without a fixture.
func writeTestPDF(t *testing.T, path, text string) {
	t.Helper()

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")

	var offsets []int

	offsets = append(offsets, buf.Len())
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	offsets = append(offsets, buf.Len())
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	offsets = append(offsets, buf.Len())
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 144] " +
		"/Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>\nendobj\n")

	content := "BT /F1 12 Tf 20 100 Td (" + text + ") Tj ET\n"
	offsets = append(offsets, buf.Len())
	fmt.Fprintf(&buf, "4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(content), content)

	offsets = append(offsets, buf.Len())
	buf.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	xrefPos := buf.Len()
	buf.WriteString("xref\n")
	fmt.Fprintf(&buf, "0 %d\n", len(offsets)+1)
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	buf.WriteString("trailer\n")
	fmt.Fprintf(&buf, "<< /Size %d /Root 1 0 R >>\n", len(offsets)+1)
	buf.WriteString("startxref\n")
	fmt.Fprintf(&buf, "%d\n", xrefPos)
	buf.WriteString("%%EOF\n")

	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}
}

func TestPdfExtractText(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path := func() string {
			p := os.TempDir() + "/swings_mcp_test.pdf"
			return p
		}()
		const text = "Hello World"
		writeTestPDF(t, path, text)

		_, out, err := PdfExtractText(context.Background(), nil, PdfExtractTextInput{Path: path})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Path != path {
			t.Errorf("Path = %q, want %q", out.Path, path)
		}
		if !strings.Contains(out.Content, text) {
			t.Errorf("Content = %q, want it to contain %q", out.Content, text)
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, _, err := PdfExtractText(context.Background(), nil, PdfExtractTextInput{Path: "  "})
		if err == nil {
			t.Fatalf("expected error for empty path")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, _, err := PdfExtractText(context.Background(), nil, PdfExtractTextInput{Path: "/nope/does-not-exist.pdf"})
		if err == nil {
			t.Fatalf("expected error for missing file")
		}
	})
}
