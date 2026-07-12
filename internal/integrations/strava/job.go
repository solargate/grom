package strava

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/workouts"
)

const (
	PhaseIdle       = "idle"
	PhaseUploading  = "uploading"
	PhaseExtracting = "extracting"
	PhaseImporting  = "importing"
	PhaseCompleted  = "completed"
	PhaseFailed     = "failed"
)

type JobStatus struct {
	Active          bool         `json:"active"`
	Phase           string       `json:"phase"`
	UploadProgress  float64      `json:"upload_progress"`
	ImportProgress  float64      `json:"import_progress"`
	ImportCurrent   int          `json:"import_current"`
	ImportTotal     int          `json:"import_total"`
	Message         string       `json:"message,omitempty"`
	Result          *ImportResult `json:"result,omitempty"`
}

type JobManager struct {
	tempDir        string
	workoutStore   *workouts.Store
	equipmentStore *equipment.Store
	onPublish      PublishWorkoutFunc

	mu   sync.Mutex
	jobs map[string]*userJob
}

type userJob struct {
	UserID   string
	Nickname string
	JobID    string
	WorkDir  string
	Status   JobStatus
}

func NewJobManager(tempDir string, workoutStore *workouts.Store, equipmentStore *equipment.Store, onPublish PublishWorkoutFunc) *JobManager {
	return &JobManager{
		tempDir:        tempDir,
		workoutStore:   workoutStore,
		equipmentStore: equipmentStore,
		onPublish:      onPublish,
		jobs:           make(map[string]*userJob),
	}
}

func (m *JobManager) UserWorkDir(nickname string) string {
	return filepath.Join(m.tempDir, nickname)
}

func (m *JobManager) BeginUpload(userID, nickname string) (*userJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.jobs[userID]; ok && existing.Status.Active {
		return nil, fmt.Errorf("import already in progress")
	}

	jobID := uuid.NewString()
	workDir := filepath.Join(m.tempDir, nickname, "strava-"+jobID)
	if err := os.MkdirAll(workDir, 0700); err != nil {
		return nil, err
	}

	job := &userJob{
		UserID:   userID,
		Nickname: nickname,
		JobID:    jobID,
		WorkDir:  workDir,
		Status: JobStatus{
			Active:         true,
			Phase:          PhaseUploading,
			UploadProgress: 0,
		},
	}
	m.jobs[userID] = job
	m.persist(job)
	return job, nil
}

func (m *JobManager) SaveArchive(job *userJob, reader io.Reader, expectedSize int64) (int64, error) {
	archivePath := filepath.Join(job.WorkDir, "archive.zip")
	file, err := os.OpenFile(archivePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return 0, err
	}

	written, err := io.Copy(file, reader)
	if err != nil {
		file.Close()
		_ = os.Remove(archivePath)
		return 0, err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		_ = os.Remove(archivePath)
		return 0, err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(archivePath)
		return 0, err
	}
	if written == 0 {
		_ = os.Remove(archivePath)
		return 0, fmt.Errorf("archive file is empty")
	}
	if expectedSize > 0 && written != expectedSize {
		_ = os.Remove(archivePath)
		return 0, fmt.Errorf("archive upload incomplete: received %d bytes, expected %d", written, expectedSize)
	}
	if err := ValidateSavedArchive(archivePath); err != nil {
		_ = os.Remove(archivePath)
		return 0, err
	}

	job.Status.UploadProgress = 1
	job.Status.Phase = PhaseExtracting
	m.persist(job)
	return written, nil
}

func (m *JobManager) StartImport(userID string) {
	m.mu.Lock()
	job, ok := m.jobs[userID]
	m.mu.Unlock()
	if !ok {
		return
	}

	go m.runImport(job)
}

func (m *JobManager) runImport(job *userJob) {
	archivePath := filepath.Join(job.WorkDir, "archive.zip")

	archive, err := OpenArchive(archivePath)
	if err != nil {
		m.fail(job, fmt.Sprintf("open archive failed: %v", err))
		return
	}
	defer archive.Close()

	job.Status.Phase = PhaseImporting
	job.Status.ImportProgress = 0
	m.persist(job)

	importer := NewImporter(m.workoutStore, m.equipmentStore, archive, m.onPublish)
	result, err := importer.ImportAll(job.Nickname, func(current, total int) {
		job.Status.ImportCurrent = current
		job.Status.ImportTotal = total
		if total > 0 {
			job.Status.ImportProgress = float64(current) / float64(total)
		}
		m.persist(job)
	})
	if err != nil {
		m.fail(job, fmt.Sprintf("import failed: %v", err))
		return
	}

	job.Status.Active = false
	job.Status.Phase = PhaseCompleted
	job.Status.ImportProgress = 1
	job.Status.Result = &result
	m.persist(job)
	m.cleanup(job)
	m.remove(job.UserID)
}

func (m *JobManager) fail(job *userJob, message string) {
	job.Status.Active = false
	job.Status.Phase = PhaseFailed
	job.Status.Message = message
	m.persist(job)
	m.cleanup(job)
	m.remove(job.UserID)
}

func (m *JobManager) cleanup(job *userJob) {
	userDir := filepath.Join(m.tempDir, job.Nickname)
	_ = os.RemoveAll(userDir)
}

func (m *JobManager) remove(userID string) {
	m.mu.Lock()
	delete(m.jobs, userID)
	m.mu.Unlock()
}

func (m *JobManager) Status(userID string) JobStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, ok := m.jobs[userID]; ok {
		return job.Status
	}

	path := m.statusFilePath(userID)
	data, err := os.ReadFile(path)
	if err != nil {
		return JobStatus{Active: false, Phase: PhaseIdle}
	}

	var status JobStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return JobStatus{Active: false, Phase: PhaseIdle}
	}
	if status.Phase == PhaseCompleted || status.Phase == PhaseFailed {
		_ = os.Remove(path)
	}
	return status
}

func (m *JobManager) persist(job *userJob) {
	data, err := json.Marshal(job.Status)
	if err != nil {
		return
	}
	path := m.statusFilePath(job.UserID)
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

func (m *JobManager) statusFilePath(userID string) string {
	return filepath.Join(m.tempDir, ".jobs", userID+".json")
}

func (m *JobManager) SetUploadProgress(userID string, progress float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job, ok := m.jobs[userID]; ok {
		job.Status.UploadProgress = progress
		m.persist(job)
	}
}

func (m *JobManager) WaitForCompletion(userID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status := m.Status(userID)
		if !status.Active {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
