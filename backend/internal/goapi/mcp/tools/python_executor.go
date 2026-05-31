package tools

// Python executor — รัน Python 3 script ใน subprocess sandbox (readonly)
//
// ทำไมถึงมี Python แยกจาก JS:
//   - LLM เขียน Python เก่งที่สุด (Python มี training data มากที่สุด — HumanEval, MBPP ฯลฯ Python ชนะ)
//   - devstral/mistral/llama/claude ทุกตัวเขียน Python สะอาดกว่า JS
//   - Syntax ตรงกับวิธีคิดของ model → script ถูกครั้งแรกบ่อยขึ้น → iteration น้อย → เร็วขึ้น
//
// Architecture — JSON-RPC over stdin/stdout:
//   1. Go spawn python3 -I -S -c "<prelude + user_code>"
//   2. Prelude กำหนด query_pg/query_mongo/query_ch/log เป็น Python function
//      ที่เขียน JSON line ไปยัง stdout (RPC:), flush, แล้วอ่าน stdin (RES:) กลับมา
//   3. Main loop ฝั่ง Go อ่าน stdout ทีละบรรทัด:
//        RPC:<json>   → execute query ด้วย Go helper → เขียน RES:<json> กลับ stdin
//        LOG:<text>   → append logs
//        RESULT:<json> → final return value
//        DONE         → script เสร็จ
//   4. Timeout = 30s รวม query ทั้งหมด — ถ้าเลย interrupt process
//
// Security:
//   - python3 -I: ignore PYTHONPATH/PYTHONHOME และ user site-packages (clean env)
//   - python3 -S: ไม่โหลด site.py → ไม่มี pip packages ที่ user ติดตั้ง
//   - Subprocess runs as appuser (Dockerfile setting)
//   - Network: ไม่ควรมี — container ไม่เปิด outbound DB connection ตรงจาก Python
//   - SQL validation: ใช้ validateReadonlySQL เดียวกับ JS executor
//   - Timeout: 30s — process.Kill() ถ้าเลย

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

const (
	pyExecTimeout   = 30 * time.Second
	pyMaxLogLines   = 100
	pyMaxResultSize = 1 << 20 // 1 MB — เกินนี้ถือว่าผิดปกติ
)

// PythonPrelude — กำหนด helpers + RPC protocol ที่ user script ใช้ได้
//
// การทำงาน: แต่ละ helper เขียน RPC:<json>\n ลง stdout แล้ว flush
// จากนั้นอ่าน response 1 บรรทัดจาก stdin (RES:<json>)
//
// Return value: script ต้อง assign ค่าสุดท้ายใส่ตัวแปร __result__
// (เพราะ Python ไม่มี implicit return จาก top-level เหมือน JS IIFE)
const PythonPrelude = `
import sys, json

def __rpc(op, **kwargs):
    sys.stdout.write("RPC:" + json.dumps({"op": op, **kwargs}, ensure_ascii=False) + "\n")
    sys.stdout.flush()
    line = sys.stdin.readline()
    if not line:
        raise RuntimeError("RPC channel closed")
    if not line.startswith("RES:"):
        raise RuntimeError("bad RPC response: " + line[:80])
    resp = json.loads(line[4:])
    if not resp.get("ok"):
        raise RuntimeError(resp.get("error") or "query failed")
    return resp.get("data")

def query_pg(sql, limit=200):
    """Run readonly SELECT on PostgreSQL. Returns list[dict]."""
    return __rpc("query_pg", sql=sql, limit=limit)

def query_mongo(collection, filter=None, limit=200):
    """Query MongoDB. filter is a dict (or JSON string). Returns list[dict]."""
    if filter is None:
        filter = {}
    if isinstance(filter, dict):
        filter = json.dumps(filter, ensure_ascii=False)
    return __rpc("query_mongo", collection=collection, filter=filter, limit=limit)

def query_ch(sql, limit=200):
    """Run readonly SELECT on ClickHouse. Returns list[dict]."""
    return __rpc("query_ch", sql=sql, limit=limit)

def log(*args):
    """Print a debug line back to the agent."""
    msg = " ".join(str(a) for a in args)
    sys.stdout.write("LOG:" + msg + "\n")
    sys.stdout.flush()

__result__ = None
`

const pythonFooter = `
sys.stdout.write("RESULT:" + json.dumps(__result__, ensure_ascii=False, default=str) + "\n")
sys.stdout.flush()
sys.stdout.write("DONE\n")
sys.stdout.flush()
`

// PyExecResponse — ผลลัพธ์จากการรัน Python (มี shape เดียวกับ JS)
type PyExecResponse struct {
	Success     bool     `json:"success"`
	Result      any      `json:"result,omitempty"`
	Logs        []string `json:"logs,omitempty"`
	Error       string   `json:"error,omitempty"`
	ExecutionMs int64    `json:"execution_ms"`
	QueriesRun  int      `json:"queries_run"`
}

// ExecutePython รัน Python 3 script ใน subprocess sandbox
//
// User script เขียน Python ปกติ และ assign ค่าสุดท้ายให้ __result__
// เช่น:
//
//	rows = query_mongo("productBarcodes", {"names.name": {"$regex": "coffee", "$options": "i"}}, 10)
//	__result__ = {"count": len(rows), "items": rows}
//
// หรือแบบง่ายกว่า — ให้ AI คืนค่าจาก expression สุดท้ายอัตโนมัติไม่ได้
// เพราะ Python ไม่มี implicit return ที่ top level
func ExecutePython(ctx context.Context, shopID, code string) (*PyExecResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("code is empty")
	}

	resp := &PyExecResponse{Logs: []string{}}
	started := time.Now()

	// Prepare full script: prelude + user code + footer
	fullScript := PythonPrelude + "\n# ===== user code =====\n" + code + "\n" + pythonFooter

	// Subprocess with isolated environment
	// -I: isolated mode (ignore PYTHONPATH, PYTHONHOME, user site)
	// -S: don't run site.py (no user packages)
	// -u: unbuffered stdout (critical for line-by-line RPC)
	procCtx, cancel := context.WithTimeout(ctx, pyExecTimeout)
	defer cancel()

	cmd := exec.CommandContext(procCtx, "python3", "-I", "-S", "-u", "-c", fullScript)
	// Minimal env — no PYTHONPATH leakage
	cmd.Env = []string{"PYTHONIOENCODING=utf-8", "LANG=C.UTF-8", "LC_ALL=C.UTF-8"}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start python: %w", err)
	}

	logger.Info("[Python Executor] shop=%s code_len=%d", shopID, len(code))

	// Reader goroutine — handles the RPC loop
	reader := bufio.NewReaderSize(stdout, 64*1024)
	queryCount := 0
	doneCh := make(chan struct{})
	var readerErr error
	var mu sync.Mutex

	go func() {
		defer close(doneCh)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					mu.Lock()
					readerErr = err
					mu.Unlock()
				}
				return
			}
			line = strings.TrimRight(line, "\r\n")

			switch {
			case strings.HasPrefix(line, "RPC:"):
				queryCount++
				payload := line[4:]
				var req map[string]any
				if err := json.Unmarshal([]byte(payload), &req); err != nil {
					writeRPCError(stdin, fmt.Sprintf("bad RPC json: %v", err))
					continue
				}
				op, _ := req["op"].(string)
				data, rpcErr := handlePythonRPC(procCtx, shopID, op, req)
				if rpcErr != nil {
					writeRPCError(stdin, rpcErr.Error())
					continue
				}
				writeRPCSuccess(stdin, data)

			case strings.HasPrefix(line, "LOG:"):
				msg := line[4:]
				logger.Info("[PY] %s", msg)
				mu.Lock()
				resp.Logs = append(resp.Logs, msg)
				if len(resp.Logs) > pyMaxLogLines {
					resp.Logs = resp.Logs[len(resp.Logs)-pyMaxLogLines:]
				}
				mu.Unlock()

			case strings.HasPrefix(line, "RESULT:"):
				payload := line[7:]
				if len(payload) > pyMaxResultSize {
					mu.Lock()
					readerErr = fmt.Errorf("result too large (%d bytes, max %d)", len(payload), pyMaxResultSize)
					mu.Unlock()
					return
				}
				var result any
				if err := json.Unmarshal([]byte(payload), &result); err != nil {
					mu.Lock()
					readerErr = fmt.Errorf("parse result: %w", err)
					mu.Unlock()
					return
				}
				mu.Lock()
				resp.Result = result
				mu.Unlock()

			case line == "DONE":
				return

			default:
				// Unknown line — likely stray print, treat as log
				if line != "" {
					mu.Lock()
					resp.Logs = append(resp.Logs, "stray: "+line)
					mu.Unlock()
				}
			}
		}
	}()

	// Wait for reader to finish or timeout
	waitErr := cmd.Wait()
	<-doneCh // make sure reader drained
	_ = stdin.Close()

	resp.ExecutionMs = time.Since(started).Milliseconds()
	resp.QueriesRun = queryCount

	mu.Lock()
	rErr := readerErr
	mu.Unlock()

	if procCtx.Err() == context.DeadlineExceeded {
		resp.Success = false
		resp.Error = "script timeout after 30s"
		logger.Error("[Python Executor] timeout in %dms", resp.ExecutionMs)
		return resp, nil
	}

	if waitErr != nil {
		errMsg := waitErr.Error()
		if stderrBuf.Len() > 0 {
			// trim to last 500 chars of stderr — AI needs the traceback
			s := stderrBuf.String()
			if len(s) > 500 {
				s = "..." + s[len(s)-500:]
			}
			errMsg = s
		}
		resp.Success = false
		resp.Error = errMsg
		logger.Error("[Python Executor] failed in %dms: %s", resp.ExecutionMs, errMsg)
		return resp, nil
	}

	if rErr != nil {
		resp.Success = false
		resp.Error = rErr.Error()
		logger.Error("[Python Executor] reader err in %dms: %v", resp.ExecutionMs, rErr)
		return resp, nil
	}

	resp.Success = true
	logger.Info("[Python Executor] ✓ done in %dms (queries=%d)", resp.ExecutionMs, queryCount)
	return resp, nil
}

// handlePythonRPC ประมวลผล RPC request จาก Python subprocess
// คืนค่า (data, error) — data จะถูก wrap ใน {"ok":true,"data":...}
func handlePythonRPC(ctx context.Context, shopID, op string, req map[string]any) (any, error) {
	switch op {
	case "query_pg":
		sql, _ := req["sql"].(string)
		limit := pyArgInt(req, "limit", 200)
		if err := validateReadonlySQL(sql); err != nil {
			return nil, err
		}
		result, err := ExecutePgCommand(ctx, shopID, sql, limit)
		if err != nil {
			return nil, fmt.Errorf("query_pg: %w", err)
		}
		return result.Rows, nil

	case "query_mongo":
		collection, _ := req["collection"].(string)
		filter, _ := req["filter"].(string)
		if filter == "" {
			filter = "{}"
		}
		limit := pyArgInt(req, "limit", 200)
		result, err := QueryMongoDB(ctx, shopID, "bcaiclouddb", collection, filter, limit)
		if err != nil {
			return nil, fmt.Errorf("query_mongo: %w", err)
		}
		return result.Documents, nil

	case "query_ch":
		sql, _ := req["sql"].(string)
		limit := pyArgInt(req, "limit", 200)
		if err := validateReadonlySQL(sql); err != nil {
			return nil, err
		}
		result, err := QueryClickHouse(ctx, shopID, "", sql, limit)
		if err != nil {
			return nil, fmt.Errorf("query_ch: %w", err)
		}
		return result.Rows, nil

	default:
		return nil, fmt.Errorf("unknown RPC op: %s", op)
	}
}

func writeRPCSuccess(w io.Writer, data any) {
	payload, err := json.Marshal(map[string]any{"ok": true, "data": data})
	if err != nil {
		writeRPCError(w, fmt.Sprintf("marshal: %v", err))
		return
	}
	_, _ = w.Write([]byte("RES:"))
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte("\n"))
}

func writeRPCError(w io.Writer, msg string) {
	payload, _ := json.Marshal(map[string]any{"ok": false, "error": msg})
	_, _ = w.Write([]byte("RES:"))
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte("\n"))
}

func pyArgInt(req map[string]any, key string, def int) int {
	v, ok := req[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return def
}
