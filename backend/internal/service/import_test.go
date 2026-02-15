package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/models"
)

func TestImportService_StartImport_RejectsParallel(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{INPXPath: "/nonexistent"}, nil, nil, nil, nil, nil)

	// Simulate running state
	svc.mu.Lock()
	svc.status = models.ImportStatus{Status: "running"}
	svc.mu.Unlock()

	err := svc.StartImport()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestImportService_GetStatus_Idle(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{}, nil, nil, nil, nil, nil)

	status := svc.GetStatus()
	assert.Equal(t, "idle", status.Status)
	assert.Nil(t, status.StartedAt)
	assert.Nil(t, status.FinishedAt)
	assert.Nil(t, status.Stats)
	assert.Nil(t, status.Error)
}

func TestImportService_StartImport_InvalidFile(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{BatchSize: 100}, config.LibraryConfig{INPXPath: "/nonexistent/file.inpx"}, nil, nil, nil, nil, nil)

	err := svc.StartImport()
	// StartImport launches in background, so no immediate error
	assert.NoError(t, err)

	// Status should transition to running (or quickly to failed)
	status := svc.GetStatus()
	assert.Contains(t, []string{"running", "failed"}, status.Status)
}
