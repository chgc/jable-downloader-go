package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jable-downloader-go/internal/downloader"
)

// DownloadRequest 下載請求結構
type DownloadRequest struct {
	URL     string `json:"url"`
	Convert bool   `json:"convert"`
}

// DownloadResponse 下載響應結構
type DownloadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitempty"`
}

// HealthResponse 健康檢查響應
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

// Server HTTP API 服務器
type Server struct {
	port       int
	mux        *http.ServeMux
	tasks      map[string]*DownloadTask
	tasksMutex sync.RWMutex
}

// DownloadTask 下載任務
type DownloadTask struct {
	ID        string
	URL       string
	Status    string
	CreatedAt time.Time
	Error     string
}

// NewServer 創建新的服務器實例
func NewServer(port int) *Server {
	s := &Server{
		port:  port,
		mux:   http.NewServeMux(),
		tasks: make(map[string]*DownloadTask),
	}
	s.setupRoutes()
	return s
}

// setupRoutes 設置路由
func (s *Server) setupRoutes() {
	// CORS 中間件
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next(w, r)
		}
	}

	// 路由
	s.mux.HandleFunc("/api/health", corsMiddleware(s.handleHealth))
	s.mux.HandleFunc("/api/download", corsMiddleware(s.handleDownload))
	s.mux.HandleFunc("/api/tasks", corsMiddleware(s.handleTasks))
}

// handleHealth 健康檢查
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
		Time:    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDownload 處理下載請求
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		s.sendError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// 創建任務
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	task := &DownloadTask{
		ID:        taskID,
		URL:       req.URL,
		Status:    "queued",
		CreatedAt: time.Now(),
	}

	s.tasksMutex.Lock()
	s.tasks[taskID] = task
	s.tasksMutex.Unlock()

	// 異步執行下載
	go s.executeDownload(task, req.Convert)

	// 返回響應
	response := DownloadResponse{
		Success: true,
		Message: "Download task created",
		TaskID:  taskID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// executeDownload 執行下載任務
func (s *Server) executeDownload(task *DownloadTask, convert bool) {
	s.updateTaskStatus(task.ID, "downloading")

	// 調用 downloader 包的下載函數
	d, err := downloader.NewDownloader(task.URL)
	if err != nil {
		s.updateTaskError(task.ID, err.Error())
		log.Printf("Failed to create downloader for %s: %v", task.URL, err)
		return
	}
	
	// 設置為自動模式（不詢問用戶）
	d.AutoMode = true
	
	// 設置轉檔模式
	if convert {
		d.EncodeMode = 1 // FastEncode - 僅轉換格式（推薦）
	} else {
		d.EncodeMode = 0 // NoEncode - 不轉檔
	}
	
	if err := d.Download(); err != nil {
		s.updateTaskError(task.ID, err.Error())
		log.Printf("Download failed for %s: %v", task.URL, err)
		return
	}

	s.updateTaskStatus(task.ID, "completed")
	log.Printf("Download completed for %s", task.URL)
}

// handleTasks 獲取任務列表
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.tasksMutex.RLock()
	tasks := make([]*DownloadTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	s.tasksMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// updateTaskStatus 更新任務狀態
func (s *Server) updateTaskStatus(taskID, status string) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()
	
	if task, ok := s.tasks[taskID]; ok {
		task.Status = status
	}
}

// updateTaskError 更新任務錯誤
func (s *Server) updateTaskError(taskID, errMsg string) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()
	
	if task, ok := s.tasks[taskID]; ok {
		task.Status = "failed"
		task.Error = errMsg
	}
}

// sendError 發送錯誤響應
func (s *Server) sendError(w http.ResponseWriter, message string, code int) {
	response := DownloadResponse{
		Success: false,
		Message: message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// Start 啟動服務器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🚀 API Server starting on http://localhost%s", addr)
	log.Printf("📝 Health check: http://localhost%s/api/health", addr)
	log.Printf("📥 Download API: http://localhost%s/api/download", addr)
	
	return http.ListenAndServe(addr, s.mux)
}
