package aichat

// Format guard — แปลงคำตอบ markdown → HTML สำหรับ Flutter overlay (flutter_html)
//
// บาง model (devstral, llama) ไม่เคารพ "Output Format — HTML" instruction ใน system prompt
// แม้จะเขียนชัดเจนแค่ไหนก็ตาม. ถ้าปล่อย markdown ไปหา flutter_html มันจะ render เป็น text ดิบ
// ดังนั้นเรา detect + convert post-hoc โดยใช้ goldmark.

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// looksLikeMarkdown — heuristic ตรวจว่า text เป็น markdown หรือ HTML
//
// เกณฑ์: ถ้าเจอ pattern markdown อย่างน้อย 1 อย่าง และไม่มี HTML block tags ที่ชัดเจน
// → ถือว่าเป็น markdown
func looksLikeMarkdown(s string) bool {
	if s == "" {
		return false
	}
	// ตรวจ markdown patterns ก่อน — ถ้าเจอ pattern ที่ชัดเจน (heading, pipe table, bullet)
	// → ถือเป็น markdown แม้จะมี HTML fragment ผสม (goldmark unsafe mode handle ได้)
	mdPatterns := []string{
		"\n## ", "\n### ", "\n#### ",
		"\n| ", "\n|-", "\n|:",
		"\n- ", "\n* ", "\n1. ",
		"\n```",
	}
	first := s
	if i := strings.Index(s, "\n"); i >= 0 {
		first = s[:i]
	}
	hasMarkdownStart := strings.HasPrefix(first, "## ") || strings.HasPrefix(first, "### ") ||
		strings.HasPrefix(first, "# ") || strings.HasPrefix(first, "- ") ||
		strings.HasPrefix(first, "| ")
	hasMarkdownPattern := false
	for _, p := range mdPatterns {
		if strings.Contains(s, p) {
			hasMarkdownPattern = true
			break
		}
	}
	if hasMarkdownStart || hasMarkdownPattern {
		return true
	}

	// ไม่มี markdown pattern — ถ้าเป็น pure HTML แล้ว (มี block tag) ปล่อยผ่าน
	htmlMarkers := []string{"<h1", "<h2", "<h3", "<p>", "<table", "<ul>", "<ol>", "<div"}
	for _, m := range htmlMarkers {
		if strings.Contains(s, m) {
			return false
		}
	}
	return false
}

// markdownToHTML — แปลง markdown เป็น HTML โดยใช้ goldmark
//
// เปิด GFM extension เพื่อรองรับ pipe tables (สำคัญมากสำหรับ Result section)
// เปิด Unsafe เพื่อปล่อย raw HTML ผ่าน — เผื่อ model mix HTML + markdown
func markdownToHTML(md string) (string, error) {
	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := gm.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
