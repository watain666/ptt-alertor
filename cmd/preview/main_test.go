package main

import (
	"strings"
	"testing"
)

func TestPageTemplates(t *testing.T) {
	for _, page := range []string{"telegram", "messenger", "docs", "top"} {
		t.Run(page, func(t *testing.T) {
			body, err := renderPage("../..", page+".html", previewData{URI: page})
			if err != nil {
				t.Fatal(err)
			}
			for _, required := range []string{"/assets/site.css", "/assets/site.js", "id=\"main-content\""} {
				if !strings.Contains(string(body), required) {
					t.Errorf("missing %s", required)
				}
			}
		})
	}
}

func TestLandingCounterAndEscaping(t *testing.T) {
	data := previewData{Count: []string{"9", ",", "9", "9", "9"}, WSHost: `wss://example.com/ws?x=" onmouseover="alert(1)`}
	body, err := renderPage("../..", "telegram.html", data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `id="counter">9,999</strong>`) {
		t.Fatal("counter must preserve its server-rendered value")
	}
	if strings.Contains(string(body), ` onmouseover="alert(1)`) {
		t.Fatal("websocket configuration must remain escaped")
	}
}

func TestDocsWithoutCounterFields(t *testing.T) {
	// The production docs handler deliberately has no Count or WSHost fields.
	_, err := renderPage("../..", "docs.html", struct{ URI, S3Domain string }{URI: "docs"})
	if err != nil {
		t.Fatal(err)
	}
}
