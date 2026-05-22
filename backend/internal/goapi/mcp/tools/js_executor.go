package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"smlcloudplatform/internal/goapi/logger"
)

// jsExecTimeout — timeout สำหรับ JS script (รวม query ทุกตัวที่เรียกข้างใน)
const jsExecTimeout = 30 * time.Second

// jsForbiddenSQL — SQL keywords ที่ห้ามใช้ใน readonly mode (validate ก่อนยิง query)
var jsForbiddenSQL = []string{
	"INSERT ", "UPDATE ", "DELETE ", "DROP ", "TRUNCATE ", "ALTER ",
	"CREATE ", "GRANT ", "REVOKE ", "EXECUTE ", "CALL ", "MERGE ",
	"REPLACE ", "RENAME ", "VACUUM ", "REINDEX ", "CLUSTER ",
}

// JSExecResponse — ผลลัพธ์จากการรัน JS
type JSExecResponse struct {
	Success bool        `json:"success"`
	Result interface{} `json:"result,omitempty"`
	Logs []string    `json:"logs,omitempty"`
	Error string      `json:"error,omitempty"`
	ExecutionMs int64       `json:"execution_ms"`
	QueriesRun int         `json:"queries_run"`
}

// ExecuteJS รัน JavaScript code ใน Goja sandbox (readonly)
//
// Sandbox helpers ที่ AI ใช้ได้:
//
//	query_pg(sql, limit?)              → array of rows
//	query_mongo(collection, filter?, limit?) → array of documents
//	query_ch(sql, limit?)              → array of rows
//	log(...)                           → server log + return ใน .logs
//
// Script ต้อง return ค่าออกมา (ค่าเดียว — object/array/string/number)
// หรือใช้ statement สุดท้ายเป็นค่า (เพราะ wrap ใน IIFE)
func ExecuteJS(ctx context.Context, shopID, code string) (*JSExecResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("code is empty")
	}

	resp := &JSExecResponse{Logs: []string{}}
	started := time.Now()

	// Goja VM + timeout
	vm := goja.New()
	deadline := time.Now().Add(jsExecTimeout)
	timer := time.AfterFunc(jsExecTimeout, func() {
		vm.Interrupt("script timeout after 30s")
	})
	defer timer.Stop()

	queryCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	queryCount := 0

	// ==================== inject helpers ====================

	// log(...args) — เก็บใน resp.Logs ให้ AI เห็น + write server log
	_ = vm.Set("log", func(call goja.FunctionCall) goja.Value {
		var parts []string
		for _, arg := range call.Arguments {
			parts = append(parts, arg.String())
		}
		msg := strings.Join(parts, " ")
		logger.Info("[JS] %s", msg)
		resp.Logs = append(resp.Logs, msg)
		if len(resp.Logs) > 100 {
			resp.Logs = resp.Logs[len(resp.Logs)-100:]
		}
		return goja.Undefined()
	})

	// query_pg(sql, limit?) → []row
	_ = vm.Set("query_pg", func(call goja.FunctionCall) goja.Value {
		sql := call.Argument(0).String()
		limit := argInt(call, 1, 200)
		if err := validateReadonlySQL(sql); err != nil {
			panic(vm.NewGoError(err))
		}
		queryCount++
		result, err := ExecutePgCommand(queryCtx, shopID, sql, limit)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("query_pg: %w", err)))
		}
		return vm.ToValue(result.Rows)
	})

	// query_mongo(collection, filter?, limit?) → []document
	_ = vm.Set("query_mongo", func(call goja.FunctionCall) goja.Value {
		collection := call.Argument(0).String()
		filter := "{}"
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) {
			filter = call.Argument(1).String()
		}
		limit := argInt(call, 2, 200)
		queryCount++
		result, err := QueryMongoDB(queryCtx, shopID, "bcaiclouddb", collection, filter, limit)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("query_mongo: %w", err)))
		}
		return vm.ToValue(result.Documents)
	})

	// query_ch(sql, limit?) → []row
	_ = vm.Set("query_ch", func(call goja.FunctionCall) goja.Value {
		sql := call.Argument(0).String()
		limit := argInt(call, 1, 200)
		if err := validateReadonlySQL(sql); err != nil {
			panic(vm.NewGoError(err))
		}
		queryCount++
		result, err := QueryClickHouse(queryCtx, shopID, "", sql, limit)
		if err != nil {
			panic(vm.NewGoError(fmt.Errorf("query_ch: %w", err)))
		}
		return vm.ToValue(result.Rows)
	})

	// ==================== run script ====================
	logger.Info("[JS Executor] shop=%s code_len=%d", shopID, len(code))

	// wrap ใน IIFE — script ต้อง return ค่าออกมา
	wrapped := "(function(){\n" + code + "\n})()"

	val, err := vm.RunString(wrapped)
	resp.ExecutionMs = time.Since(started).Milliseconds()
	resp.QueriesRun = queryCount

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "timeout") {
			errMsg = "script timeout after 30s"
		}
		resp.Success = false
		resp.Error = errMsg
		logger.Error("[JS Executor] failed in %dms: %s", resp.ExecutionMs, errMsg)
		return resp, nil // คืน response error ให้ AI อ่านได้ ไม่ throw
	}

	resp.Success = true
	if val != nil && !goja.IsUndefined(val) && !goja.IsNull(val) {
		resp.Result = val.Export()
	}
	logger.Info("[JS Executor] ✓ done in %dms (queries=%d)", resp.ExecutionMs, queryCount)
	return resp, nil
}

// validateReadonlySQL ตรวจว่า SQL เป็น read-only
func validateReadonlySQL(sql string) error {
	upper := " " + strings.ToUpper(strings.TrimSpace(sql)) + " "
	for _, forbidden := range jsForbiddenSQL {
		if strings.Contains(upper, " "+forbidden) {
			return fmt.Errorf("forbidden SQL: readonly only (no %s)", strings.TrimSpace(forbidden))
		}
	}
	return nil
}

// argInt — ดึง int argument จาก JS function call (default ถ้าไม่มี)
func argInt(call goja.FunctionCall, idx int, defaultVal int) int {
	if len(call.Arguments) <= idx {
		return defaultVal
	}
	arg := call.Argument(idx)
	if goja.IsUndefined(arg) || goja.IsNull(arg) {
		return defaultVal
	}
	exported := arg.Export()
	switch v := exported.(type) {
	case int64:
		return int(v)
	case int:
		return v
	case float64:
		return int(v)
	default:
		return defaultVal
	}
}
