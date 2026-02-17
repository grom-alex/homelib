package service

import (
	"context"
	"testing"
	"time"

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
	assert.ErrorIs(t, err, ErrImportAlreadyRunning)
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

	// Wait for goroutine to finish
	time.Sleep(50 * time.Millisecond)

	status := svc.GetStatus()
	assert.Equal(t, "failed", status.Status)
	assert.NotNil(t, status.Error)
	assert.Contains(t, *status.Error, "open INPX file")
}

func TestImportService_CancelImport_Running(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{}, nil, nil, nil, nil, nil)

	cancelled := false
	svc.mu.Lock()
	svc.cancelFn = func() { cancelled = true }
	svc.mu.Unlock()

	svc.CancelImport()

	assert.True(t, cancelled)

	// cancelFn should be set to nil after calling
	svc.mu.Lock()
	assert.Nil(t, svc.cancelFn)
	svc.mu.Unlock()
}

func TestImportService_CancelImport_NotRunning(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{}, nil, nil, nil, nil, nil)

	// Should not panic when no cancelFn is set
	svc.CancelImport()
}

func TestImportService_SetAppContext(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{}, nil, nil, nil, nil, nil)

	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "test")
	svc.SetAppContext(ctx)

	assert.Equal(t, ctx, svc.appCtx)
}

func TestImportService_SetStatusForTest(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{}, config.LibraryConfig{}, nil, nil, nil, nil, nil)

	expected := models.ImportStatus{Status: "running", TotalRecords: 100}
	svc.SetStatusForTest(expected)

	status := svc.GetStatus()
	assert.Equal(t, "running", status.Status)
	assert.Equal(t, 100, status.TotalRecords)
}

func TestImportService_StartImport_WithParentContext(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{BatchSize: 100}, config.LibraryConfig{INPXPath: "/nonexistent/file.inpx"}, nil, nil, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := svc.StartImport(ctx)
	assert.NoError(t, err)

	// Wait for goroutine to finish
	time.Sleep(50 * time.Millisecond)

	status := svc.GetStatus()
	assert.Equal(t, "failed", status.Status)
}

func TestImportService_StartImport_DefaultBatchSize(t *testing.T) {
	svc := NewImportService(nil, config.ImportConfig{BatchSize: 0}, config.LibraryConfig{INPXPath: "/nonexistent/file.inpx"}, nil, nil, nil, nil, nil)

	err := svc.StartImport()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	status := svc.GetStatus()
	assert.Equal(t, "failed", status.Status)
}
