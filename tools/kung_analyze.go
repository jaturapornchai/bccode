package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Response struct {
	Success bool `json:"success"`
	Data    struct {
		Answer     string `json:"answer"`
		Iterations int    `json:"iterations"`
		ToolsUsed  []struct {
			Tool   string `json:"tool"`
			Result any    `json:"result"`
		} `json:"tools_used"`
	} `json:"data"`
	TokenUsage struct {
		Model            string `json:"model"`
		PromptTokens     int    `json:"prompt_tokens"`
		CompletionTokens int    `json:"completion_tokens"`
	} `json:"token_usage"`
}

var questions = []string{
	"ROF002 ราคาเท่าไหร่",
	"หาสินค้า TOA",
	"กระเบื้องลอนคู่ มียี่ห้อไหนบ้าง",
	"ลูกค้าชื่อกระเบื้องทอง มีไหม",
	"ซัพพลายเออร์ปูนซีเมนต์มีใครบ้าง",
	"วิธีลางานพนักงาน",
	"กระเบื้องระเบิด สาเหตุและวิธีป้องกัน",
	"สินค้าราคาแพงที่สุด 5 อันดับ",
	"ปูนซีเมนต์ตราอินทรี ราคาตลาดเท่าไหร่",
	"มีลูกหนี้ค้างชำระกี่ราย",
}

func main() {
	dir := "/tmp/kung-test"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	fmt.Printf("%-3s %-42s %-8s %-6s %-8s %-10s %-40s\n", "#", "question", "time", "iter", "tools", "html", "tool-names")
	fmt.Println(strings.Repeat("-", 125))

	totalTools := 0
	badHTML := 0
	badNotFound := 0

	for i, q := range questions {
		metaB, _ := ioutil.ReadFile(fmt.Sprintf("%s/meta-%d.txt", dir, i))
		respB, err := ioutil.ReadFile(fmt.Sprintf("%s/resp-%d.json", dir, i))
		if err != nil {
			fmt.Printf("[%d] ERR: %v\n", i, err)
			continue
		}
		// extract time
		ms := string(metaB)
		time := "?"
		if idx := strings.Index(ms, "TIME="); idx >= 0 {
			rest := ms[idx+5:]
			if nl := strings.IndexAny(rest, "\n "); nl > 0 {
				time = rest[:nl]
			} else {
				time = strings.TrimSpace(rest)
			}
		}
		var r Response
		if err := json.Unmarshal(respB, &r); err != nil {
			fmt.Printf("[%d] JSON ERR: %v\n", i, err)
			continue
		}
		ans := r.Data.Answer
		hasHTML := strings.Contains(ans, "<h2") || strings.Contains(ans, "<table") || strings.Contains(ans, "<p>")
		mdLeak := strings.HasPrefix(strings.TrimSpace(ans), "##") || strings.Contains(ans, "\n## ") || strings.Contains(ans, "\n| ")
		htmlStr := "NO"
		if hasHTML {
			htmlStr = "OK"
		}
		if mdLeak {
			htmlStr += "!MD"
			badHTML++
		}
		tnames := []string{}
		for _, t := range r.Data.ToolsUsed {
			tnames = append(tnames, t.Tool)
		}
		totalTools += len(tnames)

		// Truncate displayed question to 40 runes
		qDisp := []rune(q)
		if len(qDisp) > 40 {
			qDisp = qDisp[:40]
		}
		fmt.Printf("[%d] %-42s %-8s %-6d %-8d %-10s %-40s\n",
			i, string(qDisp), time, r.Data.Iterations, len(tnames), htmlStr, strings.Join(tnames, ","))

		// Detect bad "not found" when tools returned nothing
		if len(tnames) == 0 && (strings.Contains(ans, "ไม่พบ") || strings.Contains(ans, "ไม่มี")) {
			badNotFound++
		}
	}

	fmt.Println()
	fmt.Printf("total tools called: %d (avg %.1f/q)\n", totalTools, float64(totalTools)/float64(len(questions)))
	fmt.Printf("HTML format violations: %d/%d\n", badHTML, len(questions))
	fmt.Printf("zero-tool 'not found' responses: %d/%d\n", badNotFound, len(questions))
}
