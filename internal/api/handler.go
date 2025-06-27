package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jairo/assetter/internal/models"
	"github.com/jairo/assetter/internal/storage"
	"github.com/jairo/assetter/internal/stability"
	"github.com/jairo/assetter/templates"
)

type Handler struct {
	storage     storage.Storage
	client      *stability.Client
	jobs        map[string]*Job
	modelStore  *models.ModelStore
}

type Job struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	ModelURL string `json:"model_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

func NewHandler(storage storage.Storage, client *stability.Client) *Handler {
	modelStore, err := models.NewModelStore("data")
	if err != nil {
		panic(fmt.Errorf("failed to initialize model store: %w", err))
	}

	return &Handler{
		storage:    storage,
		client:     client,
		jobs:       make(map[string]*Job),
		modelStore: modelStore,
	}
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	templates.HomePage().Render(w)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB max

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		http.Error(w, "Only JPG and PNG files are allowed", http.StatusBadRequest)
		return
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	jobID := generateJobID()
	job := &Job{
		ID:     jobID,
		Status: "processing",
	}
	h.jobs[jobID] = job

	go h.process3DGeneration(jobID, imageData)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": "processing",
	})
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/status/")
	job, exists := h.jobs[jobID]
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/download/")
	job, exists := h.jobs[jobID]
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	if job.Status != "completed" {
		http.Error(w, "Model not ready", http.StatusBadRequest)
		return
	}

	modelPath := filepath.Join("uploads", jobID+".glb")
	http.ServeFile(w, r, modelPath)
}

func (h *Handler) process3DGeneration(jobID string, imageData []byte) {
	job := h.jobs[jobID]
	
	result, err := h.client.Generate3D(imageData)
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		return
	}

	modelFilename := jobID + ".glb"
	if err := h.storage.SaveFile(modelFilename, result); err != nil {
		job.Status = "failed"
		job.Error = "Failed to save model"
		return
	}

	// Save model metadata
	if err := h.modelStore.AddModel(jobID, ""); err != nil {
		// Log error but don't fail the job
		fmt.Printf("Failed to save model metadata: %v\n", err)
	}

	job.Status = "completed"
	job.ModelURL = "/download/" + jobID
}

func (h *Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := h.modelStore.GetAll()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}