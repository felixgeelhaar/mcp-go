package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"go.klarlabs.de/mcp/protocol"
)

func taskReq(t *testing.T, method string, params any) *protocol.Request {
	t.Helper()
	return &protocol.Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  method,
		Params:  mustParams(t, params),
	}
}

// TestTaskAugmentation_Flow drives the full 2025-11-25 task lifecycle: an
// augmented tools/call returns a working CreateTaskResult, tasks/get reports
// working while the handler is gated, and tasks/result blocks until completion
// and returns the underlying result tagged with the related-task meta.
func TestTaskAugmentation_Flow(t *testing.T) {
	release := make(chan struct{})
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	type in struct {
		X string `json:"x"`
	}
	srv.Tool("slow").
		Description("gated tool").
		TaskSupport(TaskSupportOptional).
		Handler(func(_ context.Context, i in) (string, error) {
			<-release
			return "done:" + i.X, nil
		})
	handler := newRequestHandler(srv)

	// 1. Augmented call → CreateTaskResult (status working).
	resp, err := handler.HandleRequest(context.Background(),
		taskReq(t, protocol.MethodToolsCall, map[string]any{
			"name": "slow", "arguments": map[string]any{"x": "y"}, fieldTask: map[string]any{"ttl": 60000},
		}))
	if err != nil {
		t.Fatalf("augmented call: %v", err)
	}
	task, ok := resp.Result.(map[string]any)[fieldTask].(*AugTask)
	if !ok {
		t.Fatalf("expected CreateTaskResult with *AugTask, got %#v", resp.Result)
	}
	if task.Status != TaskSupportWorkingStatus {
		t.Fatalf("expected working, got %q", task.Status)
	}
	id := task.TaskID

	// 2. tasks/get while gated → still working.
	getResp, err := handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodTasksGet, map[string]any{fieldTaskID: id}))
	if err != nil {
		t.Fatalf("tasks/get: %v", err)
	}
	if got := getResp.Result.(*AugTask); got.Status != TaskSupportWorkingStatus {
		t.Fatalf("expected working before release, got %q", got.Status)
	}

	// 3. Release; tasks/result blocks until terminal and returns the result.
	close(release)
	resResp, err := handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodTasksResult, map[string]any{fieldTaskID: id}))
	if err != nil {
		t.Fatalf("tasks/result: %v", err)
	}
	res := resResp.Result.(map[string]any)
	raw, _ := json.Marshal(res)
	if !strings.Contains(string(raw), "done:y") {
		t.Errorf("expected tool output in result, got %s", raw)
	}
	meta, _ := res["_meta"].(map[string]any)
	rel, _ := meta[protocol.MetaKeyRelatedTask].(map[string]any)
	if rel == nil || rel["taskId"] != id {
		t.Errorf("expected related-task meta with taskId %s, got %v", id, meta)
	}

	// 4. tasks/get now completed.
	getResp2, _ := handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodTasksGet, map[string]any{fieldTaskID: id}))
	if got := getResp2.Result.(*AugTask); got.Status != "completed" {
		t.Errorf("expected completed, got %q", got.Status)
	}
}

// TaskSupportWorkingStatus is the initial task status literal, kept local to the
// test to avoid exporting an internal constant.
const TaskSupportWorkingStatus = "working"

func TestTaskAugmentation_Rejections(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	type in struct {
		X string `json:"x"`
	}
	srv.Tool("plain").Description("no tasks").
		Handler(func(_ in) (string, error) { return "ok", nil })
	srv.Tool("must").Description("task required").TaskSupport(TaskSupportRequired).
		Handler(func(_ in) (string, error) { return "ok", nil })
	handler := newRequestHandler(srv)

	// task on a forbidden tool → -32601.
	_, err := handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodToolsCall, map[string]any{
		"name": "plain", "arguments": map[string]any{"x": "y"}, fieldTask: map[string]any{},
	}))
	assertCode(t, err, protocol.CodeMethodNotFound, "task on forbidden tool")

	// plain call on a required-task tool → -32601.
	_, err = handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodToolsCall, map[string]any{
		"name": "must", "arguments": map[string]any{"x": "y"},
	}))
	assertCode(t, err, protocol.CodeMethodNotFound, "plain call on required-task tool")

	// tasks/get unknown id → -32602.
	_, err = handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodTasksGet, map[string]any{fieldTaskID: "nope"}))
	assertCode(t, err, protocol.CodeInvalidParams, "unknown task get")
}

func TestTaskAugmentation_CancelTerminalRejected(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	type in struct {
		X string `json:"x"`
	}
	done := make(chan struct{})
	srv.Tool("t").Description("").TaskSupport(TaskSupportOptional).
		Handler(func(_ context.Context, _ in) (string, error) { <-done; return "ok", nil })
	handler := newRequestHandler(srv)

	resp, _ := handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodToolsCall, map[string]any{
		"name": "t", "arguments": map[string]any{"x": "y"}, fieldTask: map[string]any{},
	}))
	id := resp.Result.(map[string]any)[fieldTask].(*AugTask).TaskID

	// Cancel the working task → cancelled.
	cResp, err := handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodTasksCancel, map[string]any{fieldTaskID: id}))
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cResp.Result.(*AugTask).Status != "cancelled" {
		t.Fatalf("expected cancelled, got %q", cResp.Result.(*AugTask).Status)
	}
	// Cancel again (terminal) → -32602.
	_, err = handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodTasksCancel, map[string]any{fieldTaskID: id}))
	assertCode(t, err, protocol.CodeInvalidParams, "cancel terminal task")
	close(done)
}

func TestTaskCapability_And_ExecutionTaskSupport(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	type in struct {
		X string `json:"x"`
	}
	srv.Tool("t").Description("").TaskSupport(TaskSupportOptional).
		Handler(func(_ in) (string, error) { return "ok", nil })
	res := initResult(t, srv, "2025-11-25")
	caps := res["capabilities"].(map[string]any)
	if _, ok := caps["tasks"]; !ok {
		t.Errorf("expected tasks capability advertised, got %v", caps)
	}

	handler := newRequestHandler(srv)
	lResp, _ := handler.HandleRequest(context.Background(), taskReq(t, protocol.MethodToolsList, map[string]any{}))
	tool := lResp.Result.(map[string]any)["tools"].([]map[string]any)[0]
	exec, _ := tool["execution"].(map[string]any)
	if exec == nil || exec["taskSupport"] != "optional" {
		t.Errorf("expected execution.taskSupport=optional in tools/list, got %v", tool["execution"])
	}
}

func TestTaskAugmentation_UnsolicitedRequiredModern(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	type in struct {
		X string `json:"x"`
	}
	srv.Tool("must").Description("task required").TaskSupport(TaskSupportRequired).
		Handler(func(_ in) (string, error) { return "ok", nil })
	handler := newRequestHandler(srv)

	resp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodToolsCall,
		Params: modernParams(t, protocol.ModernVersion, map[string]any{
			"name": "must", "arguments": map[string]any{"x": "y"},
		}),
	})
	if err != nil {
		t.Fatalf("modern required-task plain call: %v", err)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result type %T", resp.Result)
	}
	if result[fieldResultType] != protocol.ResultTypeTask {
		t.Fatalf("resultType = %v, want %q", result[fieldResultType], protocol.ResultTypeTask)
	}
	if result[fieldTaskID] == nil || result[fieldTaskID] == "" {
		t.Fatalf("expected unsolicited task handle, got %#v", result)
	}
}

func TestTasksResult_GatedOffForModern(t *testing.T) {
	handler, id, release := newWorkingTask(t)
	defer close(release)

	modern := &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`11`), Method: protocol.MethodTasksResult,
		Params: modernParams(t, protocol.ModernVersion, map[string]any{fieldTaskID: id}),
	}
	_, err := handler.HandleRequest(context.Background(), modern)
	var mcpErr *protocol.Error
	if !errors.As(err, &mcpErr) || mcpErr.Code != protocol.CodeMethodNotFound {
		t.Fatalf("got %v, want MethodNotFound for modern tasks/result", err)
	}
}

func TestModern_TasksGetInlinesCompletedResult(t *testing.T) {
	release := make(chan struct{})
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	srv.Tool("slow").Description("gated").TaskSupport(TaskSupportRequired).
		Handler(func(_ context.Context, _ struct{}) (string, error) {
			<-release
			return "done", nil
		})
	handler := newRequestHandler(srv)

	resp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodToolsCall,
		Params: modernParams(t, protocol.ModernVersion, map[string]any{
			"name": "slow", "arguments": map[string]any{},
		}),
	})
	if err != nil {
		t.Fatalf("tools/call: %v", err)
	}
	created := resp.Result.(map[string]any)
	id, _ := created[fieldTaskID].(string)
	if id == "" {
		t.Fatalf("missing taskId: %#v", created)
	}
	close(release)

	deadline := time.Now().Add(2 * time.Second)
	var got map[string]any
	for time.Now().Before(deadline) {
		getResp, getErr := handler.HandleRequest(context.Background(), &protocol.Request{
			JSONRPC: "2.0", ID: json.RawMessage(`2`), Method: protocol.MethodTasksGet,
			Params: modernParams(t, protocol.ModernVersion, map[string]any{fieldTaskID: id}),
		})
		if getErr != nil {
			t.Fatalf("tasks/get: %v", getErr)
		}
		got = getResp.Result.(map[string]any)
		if got[fieldStatus] == "completed" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got[fieldStatus] != "completed" {
		t.Fatalf("status = %v, want completed", got[fieldStatus])
	}
	if _, ok := got[fieldTTLMs]; !ok {
		t.Errorf("expected ttlMs on modern tasks/get")
	}
	raw, _ := json.Marshal(got[fieldResult])
	if !strings.Contains(string(raw), "done") {
		t.Errorf("expected inlined result, got %s", raw)
	}
}

func TestModern_IgnoresTaskOptInOnOptional(t *testing.T) {
	srv := NewServer(ServerInfo{Name: "s", Version: "1"})
	srv.Tool("opt").Description("optional").TaskSupport(TaskSupportOptional).
		Handler(func(_ struct{}) (string, error) { return "sync", nil })
	handler := newRequestHandler(srv)

	resp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: protocol.MethodToolsCall,
		Params: modernParams(t, protocol.ModernVersion, map[string]any{
			"name": "opt", "arguments": map[string]any{},
			fieldTask: map[string]any{"ttl": 60000},
		}),
	})
	if err != nil {
		t.Fatalf("tools/call: %v", err)
	}
	result := resp.Result.(map[string]any)
	if result[fieldResultType] == protocol.ResultTypeTask || result[fieldTaskID] != nil {
		t.Fatalf("optional tool must ignore retired task field, got %#v", result)
	}
	raw, _ := json.Marshal(result)
	if !strings.Contains(string(raw), "sync") {
		t.Errorf("expected synchronous result, got %s", raw)
	}
}

func TestModern_TasksUpdateEmptyAck(t *testing.T) {
	handler, id, release := newWorkingTask(t)
	defer close(release)

	resp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`3`), Method: protocol.MethodTasksUpdate,
		Params: modernParams(t, protocol.ModernVersion, map[string]any{fieldTaskID: id, "ttl": int64(120000)}),
	})
	if err != nil {
		t.Fatalf("tasks/update: %v", err)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %#v", resp.Result)
	}
	if _, hasTask := result[fieldTaskID]; hasTask {
		t.Errorf("modern tasks/update must be an empty ack, got %#v", result)
	}
}

func TestModern_TasksCancelEmptyAck(t *testing.T) {
	handler, id, release := newWorkingTask(t)
	defer close(release)

	resp, err := handler.HandleRequest(context.Background(), &protocol.Request{
		JSONRPC: "2.0", ID: json.RawMessage(`4`), Method: protocol.MethodTasksCancel,
		Params: modernParams(t, protocol.ModernVersion, map[string]any{fieldTaskID: id}),
	})
	if err != nil {
		t.Fatalf("tasks/cancel: %v", err)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %#v", resp.Result)
	}
	if _, hasTask := result[fieldTaskID]; hasTask {
		t.Errorf("modern tasks/cancel must be an empty ack, got %#v", result)
	}
}

func TestJSONRPCErrorMap(t *testing.T) {
	if got := jsonRPCErrorMap(nil); got["code"] != protocol.CodeInternalError {
		t.Errorf("nil err code = %v", got["code"])
	}
	inv := protocol.NewInvalidParams("nope")
	got := jsonRPCErrorMap(inv)
	if got["code"] != protocol.CodeInvalidParams || got["message"] != "nope" {
		t.Errorf("invalid params map = %#v", got)
	}
	got = jsonRPCErrorMap(inv.WithData("x"))
	if got["data"] != "x" {
		t.Errorf("expected data, got %#v", got)
	}
	got = jsonRPCErrorMap(errors.New("boom"))
	if got["code"] != protocol.CodeInternalError || got["message"] != "boom" {
		t.Errorf("plain err map = %#v", got)
	}
}

func assertCode(t *testing.T, err error, code int, ctx string) {
	t.Helper()
	var mcpErr *protocol.Error
	if !errors.As(err, &mcpErr) || mcpErr.Code != code {
		t.Fatalf("%s: expected error code %d, got %v", ctx, code, err)
	}
}
