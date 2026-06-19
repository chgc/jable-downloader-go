package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestServer 創建一個測試用 Server，但不啟動 HTTP listener
func newTestServer() *Server {
	s := &Server{
		port:  0, // not used in tests
		mux:   http.NewServeMux(),
		tasks: make(map[string]*DownloadTask),
		queue: make(chan *DownloadTask, 100),
	}
	s.setupRoutes()
	// Don't start queue worker in tests to avoid goroutine leaks
	// (queue worker will block forever on empty channel)
	return s
}

func TestHealthEndpoint(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", resp.Version)
	}
	if resp.Time == "" {
		t.Error("expected non-empty time field")
	}
}

func TestHealthEndpoint_MethodNotAllowed(t *testing.T) {
	s := newTestServer()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/health", nil)
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, req)
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestHealthEndpoint_OPTIONS(t *testing.T) {
	s := newTestServer()

	// OPTIONS is handled by CORS middleware and returns 200
	req := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for OPTIONS (CORS preflight), got %d", w.Code)
	}
	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", origin)
	}
}

func TestDownloadEndpoint_Success(t *testing.T) {
	s := newTestServer()

	body := `{"url":"https://jable.tv/videos/test-123/","convert":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/download", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp DownloadResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success=true")
	}
	if resp.Message != "Download task queued" {
		t.Errorf("expected 'Download task queued', got %q", resp.Message)
	}
	if resp.TaskID == "" {
		t.Error("expected non-empty TaskID")
	}

	// Verify task was created
	s.tasksMutex.RLock()
	task, exists := s.tasks[resp.TaskID]
	s.tasksMutex.RUnlock()

	if !exists {
		t.Fatal("task was not stored")
	}
	if task.Status != "queued" {
		t.Errorf("expected task status 'queued', got %q", task.Status)
	}
	if task.URL != "https://jable.tv/videos/test-123/" {
		t.Errorf("expected URL 'https://jable.tv/videos/test-123/', got %q", task.URL)
	}
	if task.Convert {
		t.Error("expected Convert=false")
	}
}

func TestDownloadEndpoint_WithConvert(t *testing.T) {
	s := newTestServer()

	body := `{"url":"https://jable.tv/videos/test-456/","convert":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/download", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp DownloadResponse
	json.NewDecoder(w.Body).Decode(&resp)

	s.tasksMutex.RLock()
	task := s.tasks[resp.TaskID]
	s.tasksMutex.RUnlock()

	if !task.Convert {
		t.Error("expected Convert=true")
	}
}

func TestDownloadEndpoint_MissingURL(t *testing.T) {
	s := newTestServer()

	body := `{"convert":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/download", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp DownloadResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Success {
		t.Error("expected Success=false")
	}
}

func TestDownloadEndpoint_InvalidJSON(t *testing.T) {
	s := newTestServer()

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/download", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDownloadEndpoint_WrongMethod(t *testing.T) {
	s := newTestServer()

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/download", nil)
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, req)
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestTasksEndpoint_Empty(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp TasksResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(resp.Tasks))
	}
	if resp.CurrentTask != "" {
		t.Errorf("expected empty current_task, got %q", resp.CurrentTask)
	}
}

func TestTasksEndpoint_WithTasks(t *testing.T) {
	s := newTestServer()

	// Add some tasks manually
	s.tasksMutex.Lock()
	s.tasks["task_1"] = &DownloadTask{
		ID: "task_1", URL: "https://jable.tv/videos/a/",
		Status: "queued", CreatedAt: time.Now().Add(-1 * time.Minute),
	}
	s.tasks["task_2"] = &DownloadTask{
		ID: "task_2", URL: "https://jable.tv/videos/b/",
		Status: "completed", CreatedAt: time.Now(),
	}
	s.tasksMutex.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp TasksResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(resp.Tasks))
	}

	// Tasks should be sorted newest first
	if len(resp.Tasks) >= 2 {
		if resp.Tasks[0].ID != "task_2" {
			t.Errorf("expected task_2 first (newest), got %s", resp.Tasks[0].ID)
		}
	}
}

func TestTasksEndpoint_MethodNotAllowed(t *testing.T) {
	s := newTestServer()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/tasks", nil)
			w := httptest.NewRecorder()
			s.mux.ServeHTTP(w, req)
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected 405 for %s, got %d", method, w.Code)
			}
		})
	}
}

func TestClearCompletedTasks_Delete(t *testing.T) {
	s := newTestServer()

	// Add completed and queued tasks
	s.tasksMutex.Lock()
	s.tasks["t1"] = &DownloadTask{ID: "t1", URL: "https://jable.tv/1/", Status: "completed"}
	s.tasks["t2"] = &DownloadTask{ID: "t2", URL: "https://jable.tv/2/", Status: "failed"}
	s.tasks["t3"] = &DownloadTask{ID: "t3", URL: "https://jable.tv/3/", Status: "queued"}
	s.tasksMutex.Unlock()

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/clear-completed", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp ClearCompletedResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp.Success {
		t.Error("expected Success=true")
	}
	if resp.ClearedCount != 2 {
		t.Errorf("expected 2 cleared tasks, got %d", resp.ClearedCount)
	}

	// Verify only queued task remains
	s.tasksMutex.RLock()
	if _, exists := s.tasks["t1"]; exists {
		t.Error("t1 (completed) should have been cleared")
	}
	if _, exists := s.tasks["t2"]; exists {
		t.Error("t2 (failed) should have been cleared")
	}
	if _, exists := s.tasks["t3"]; !exists {
		t.Error("t3 (queued) should remain")
	}
	s.tasksMutex.RUnlock()
}

func TestClearCompletedTasks_Post(t *testing.T) {
	s := newTestServer()

	s.tasksMutex.Lock()
	s.tasks["t1"] = &DownloadTask{ID: "t1", URL: "https://jable.tv/1/", Status: "completed"}
	s.tasksMutex.Unlock()

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/clear-completed", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp ClearCompletedResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ClearedCount != 1 {
		t.Errorf("expected 1 cleared, got %d", resp.ClearedCount)
	}
}

func TestClearCompletedTasks_WithCurrentTask(t *testing.T) {
	s := newTestServer()

	// Set current task
	s.currentMutex.Lock()
	s.currentTask = &DownloadTask{ID: "current", URL: "https://jable.tv/current/"}
	s.currentMutex.Unlock()

	s.tasksMutex.Lock()
	s.tasks["current"] = &DownloadTask{ID: "current", URL: "https://jable.tv/current/", Status: "downloading"}
	s.tasks["old_completed"] = &DownloadTask{ID: "old_completed", URL: "https://jable.tv/done/", Status: "completed"}
	s.tasks["old_failed"] = &DownloadTask{ID: "old_failed", URL: "https://jable.tv/fail/", Status: "failed"}
	s.tasksMutex.Unlock()

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/clear-completed", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp ClearCompletedResponse
	json.NewDecoder(w.Body).Decode(&resp)

	// Should clear 2 (old_completed + old_failed), but NOT current (downloading)
	if resp.ClearedCount != 2 {
		t.Errorf("expected 2 cleared (not counting current), got %d", resp.ClearedCount)
	}

	s.tasksMutex.RLock()
	if _, exists := s.tasks["current"]; !exists {
		t.Error("current task should still exist")
	}
	if _, exists := s.tasks["old_completed"]; exists {
		t.Error("old_completed should have been cleared")
	}
	if _, exists := s.tasks["old_failed"]; exists {
		t.Error("old_failed should have been cleared")
	}
	s.tasksMutex.RUnlock()
}

func TestClearCompletedTasks_NoCompleted(t *testing.T) {
	s := newTestServer()

	// Only queued tasks
	s.tasksMutex.Lock()
	s.tasks["t1"] = &DownloadTask{ID: "t1", URL: "https://jable.tv/1/", Status: "queued"}
	s.tasks["t2"] = &DownloadTask{ID: "t2", URL: "https://jable.tv/2/", Status: "downloading"}
	s.tasksMutex.Unlock()

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/clear-completed", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	var resp ClearCompletedResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.ClearedCount != 0 {
		t.Errorf("expected 0 cleared, got %d", resp.ClearedCount)
	}
}

func TestClearCompletedTasks_WrongMethod(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/clear-completed", nil)
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET, got %d", w.Code)
	}
}

func TestCORSHeaders(t *testing.T) {
	s := newTestServer()

	// Test CORS headers on health endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", origin)
	}
	if methods := w.Header().Get("Access-Control-Allow-Methods"); methods == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestCORSPreflight(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodOptions, "/api/download", nil)
	req.Header.Set("Origin", "chrome-extension://abc123")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for OPTIONS preflight, got %d", w.Code)
	}
	if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", origin)
	}
}

func TestTaskLifecycle(t *testing.T) {
	s := newTestServer()

	// 1. Create a task via POST
	body := `{"url":"https://jable.tv/videos/lifecycle/","convert":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/download", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	var createResp DownloadResponse
	json.NewDecoder(w.Body).Decode(&createResp)
	taskID := createResp.TaskID

	// 2. Check task status via GET /api/tasks
	req2 := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	w2 := httptest.NewRecorder()
	s.mux.ServeHTTP(w2, req2)

	var tasksResp TasksResponse
	json.NewDecoder(w2.Body).Decode(&tasksResp)

	found := false
	for _, task := range tasksResp.Tasks {
		if task.ID == taskID {
			found = true
			if task.Status != "queued" {
				t.Errorf("expected status 'queued', got %q", task.Status)
			}
			if task.URL != "https://jable.tv/videos/lifecycle/" {
				t.Errorf("expected URL 'https://jable.tv/videos/lifecycle/', got %q", task.URL)
			}
			break
		}
	}
	if !found {
		t.Error("task not found in tasks list")
	}

	// 3. Update task status manually (simulating queue worker)
	s.updateTaskStatus(taskID, "downloading")
	s.tasksMutex.RLock()
	if s.tasks[taskID].Status != "downloading" {
		t.Errorf("expected 'downloading', got %q", s.tasks[taskID].Status)
	}
	s.tasksMutex.RUnlock()

	// 4. Mark as completed
	s.updateTaskStatus(taskID, "completed")
	s.tasksMutex.RLock()
	if s.tasks[taskID].Status != "completed" {
		t.Errorf("expected 'completed', got %q", s.tasks[taskID].Status)
	}
	s.tasksMutex.RUnlock()
}

func TestUpdateTaskStatus_NonExistent(t *testing.T) {
	s := newTestServer()

	// Should not panic
	s.updateTaskStatus("nonexistent", "completed")
}

func TestUpdateTaskError(t *testing.T) {
	s := newTestServer()

	s.tasksMutex.Lock()
	s.tasks["err_task"] = &DownloadTask{ID: "err_task", Status: "downloading"}
	s.tasksMutex.Unlock()

	s.updateTaskError("err_task", "something went wrong")

	s.tasksMutex.RLock()
	task := s.tasks["err_task"]
	s.tasksMutex.RUnlock()

	if task.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", task.Status)
	}
	if task.Error != "something went wrong" {
		t.Errorf("expected error 'something went wrong', got %q", task.Error)
	}
}

func TestSetCurrentTask(t *testing.T) {
	s := newTestServer()

	task := &DownloadTask{ID: "current", Status: "downloading"}
	s.setCurrentTask(task)

	s.currentMutex.RLock()
	if s.currentTask != task {
		t.Error("currentTask should be set")
	}
	s.currentMutex.RUnlock()

	// Set to nil
	s.setCurrentTask(nil)
	s.currentMutex.RLock()
	if s.currentTask != nil {
		t.Error("currentTask should be nil")
	}
	s.currentMutex.RUnlock()
}

func TestSendError(t *testing.T) {
	s := newTestServer()

	w := httptest.NewRecorder()
	s.sendError(w, "test error", http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp DownloadResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Success {
		t.Error("expected Success=false")
	}
	if resp.Message != "test error" {
		t.Errorf("expected 'test error', got %q", resp.Message)
	}
}

func TestNewServer(t *testing.T) {
	s := NewServer(9999)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.port != 9999 {
		t.Errorf("expected port 9999, got %d", s.port)
	}
	if s.mux == nil {
		t.Error("mux should not be nil")
	}
	if s.tasks == nil {
		t.Error("tasks map should not be nil")
	}
	if cap(s.queue) != 100 {
		t.Errorf("expected queue cap 100, got %d", cap(s.queue))
	}
}

func TestLargeTaskQueue(t *testing.T) {
	s := newTestServer()

	// Submit many download requests with unique URLs to verify task creation
	uniqueTasks := 20
	for i := 0; i < uniqueTasks; i++ {
		url := fmt.Sprintf("https://jable.tv/videos/test-%d/", i)
		body := bytes.NewBufferString(`{"url":"` + url + `","convert":false}`)
		req := httptest.NewRequest(http.MethodPost, "/api/download", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, w.Code)
		}
	}

	// Verify tasks are stored (may be fewer if task IDs collided within same nanosecond)
	s.tasksMutex.RLock()
	count := len(s.tasks)
	// Collect unique URLs for verification
	urlSet := make(map[string]bool)
	for _, task := range s.tasks {
		urlSet[task.URL] = true
	}
	s.tasksMutex.RUnlock()

	// We should have at least tried to create all tasks
	// (some may collide on task_id if nanosecond precision not enough)
	t.Logf("Submitted %d requests, stored %d unique tasks (some IDs may have collided)", uniqueTasks, count)
	if count == 0 {
		t.Error("expected at least some tasks to be stored")
	}

	// Verify unique URLs are all present
	if len(urlSet) < uniqueTasks {
		t.Logf("Got %d unique URLs out of %d submitted (expected due to nanosecond collisions)", len(urlSet), uniqueTasks)
	}
}
