package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/glst"
)

// --- Mocks ---

type mockMetadataStore struct {
	store map[string]string
}

func newMockMetadataStore() *mockMetadataStore {
	return &mockMetadataStore{store: make(map[string]string)}
}

func (m *mockMetadataStore) Get(ctx context.Context, key string) (string, error) {
	return m.store[key], nil
}

func (m *mockMetadataStore) Set(ctx context.Context, key, value string) error {
	m.store[key] = value
	return nil
}

type mockGenreTreeRepo struct {
	loadTreeCalls   int
	loadTreeEntries []glst.GenreEntry
	loadTreeResult  int

	getIDsByCodesResult map[string][]int
	getIDsCallCount     int

	unsortedGenreID int
}

func (m *mockGenreTreeRepo) LoadTree(ctx context.Context, entries []glst.GenreEntry) (int, error) {
	m.loadTreeCalls++
	m.loadTreeEntries = entries
	if m.loadTreeResult > 0 {
		return m.loadTreeResult, nil
	}
	return len(entries), nil
}

func (m *mockGenreTreeRepo) GetIDsByCodes(ctx context.Context, codes []string) (map[string][]int, error) {
	m.getIDsCallCount++
	if m.getIDsByCodesResult != nil {
		return m.getIDsByCodesResult, nil
	}
	return make(map[string][]int), nil
}

func (m *mockGenreTreeRepo) GetUnsortedGenreID(ctx context.Context) (int, error) {
	if m.unsortedGenreID > 0 {
		return m.unsortedGenreID, nil
	}
	return 1, nil
}

type mockBookRemapper struct {
	remapCalls int
	booksCount int
}

func (m *mockBookRemapper) RemapBookGenres(ctx context.Context, codeToIDs map[string][]int, fallbackID int, batchSize int) (int, error) {
	m.remapCalls++
	return m.booksCount, nil
}

// --- Tests ---

func TestGenreTreeService_LoadIfNeeded_SkipUnchanged(t *testing.T) {
	genreData := []byte("0.1 sf_all;Фантастика\n")
	hash := computeSHA256(genreData)

	meta := newMockMetadataStore()
	meta.store["genre_tree_hash"] = hash

	repo := &mockGenreTreeRepo{}
	books := &mockBookRemapper{}

	svc := NewGenreTreeService(genreData, meta, repo, books)

	result, err := svc.LoadIfNeeded(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, result.GenresLoaded)
	assert.True(t, result.Skipped)
	assert.Equal(t, 0, repo.loadTreeCalls, "LoadTree should not be called when hash matches")
}

func TestGenreTreeService_LoadIfNeeded_LoadsWhenChanged(t *testing.T) {
	genreData := []byte("0.1 sf_all;Фантастика\n0.2 det_all;Детективы\n")

	meta := newMockMetadataStore()
	// Different hash or empty — will trigger load
	meta.store["genre_tree_hash"] = "old-hash"

	repo := &mockGenreTreeRepo{unsortedGenreID: 10}
	books := &mockBookRemapper{booksCount: 100}

	svc := NewGenreTreeService(genreData, meta, repo, books)

	result, err := svc.LoadIfNeeded(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Skipped)
	assert.Equal(t, 2, result.GenresLoaded)
	assert.Equal(t, 100, result.BooksRemapped)
	assert.Equal(t, 1, repo.loadTreeCalls)
	assert.Equal(t, 1, books.remapCalls)
}

func TestGenreTreeService_ForceReload_IgnoresHash(t *testing.T) {
	genreData := []byte("0.1 sf_all;Фантастика\n")
	hash := computeSHA256(genreData)

	meta := newMockMetadataStore()
	meta.store["genre_tree_hash"] = hash // Same hash

	repo := &mockGenreTreeRepo{unsortedGenreID: 10}
	books := &mockBookRemapper{}

	svc := NewGenreTreeService(genreData, meta, repo, books)

	result, err := svc.ForceReload(context.Background())
	require.NoError(t, err)
	assert.False(t, result.Skipped)
	assert.Equal(t, 1, result.GenresLoaded)
	assert.Equal(t, 1, repo.loadTreeCalls, "ForceReload must call LoadTree even if hash matches")
}

func TestGenreTreeService_ForceReload_UpdatesHash(t *testing.T) {
	genreData := []byte("0.1 sf_all;Фантастика\n")

	meta := newMockMetadataStore()
	repo := &mockGenreTreeRepo{unsortedGenreID: 10}
	books := &mockBookRemapper{}

	svc := NewGenreTreeService(genreData, meta, repo, books)

	_, err := svc.ForceReload(context.Background())
	require.NoError(t, err)

	storedHash := meta.store["genre_tree_hash"]
	assert.Equal(t, computeSHA256(genreData), storedHash)
}

func TestGenreTreeService_Idempotency(t *testing.T) {
	genreData := []byte("0.1 sf_all;Фантастика\n0.2 det_all;Детективы\n")

	meta := newMockMetadataStore()
	repo := &mockGenreTreeRepo{unsortedGenreID: 10}
	books := &mockBookRemapper{booksCount: 50}

	svc := NewGenreTreeService(genreData, meta, repo, books)

	// First call: loads
	result1, err := svc.LoadIfNeeded(context.Background())
	require.NoError(t, err)
	assert.False(t, result1.Skipped)
	assert.Equal(t, 1, repo.loadTreeCalls)

	// Second call: skips (hash now matches)
	result2, err := svc.LoadIfNeeded(context.Background())
	require.NoError(t, err)
	assert.True(t, result2.Skipped)
	assert.Equal(t, 1, repo.loadTreeCalls, "should not load again")
}

func TestGenreTreeService_ParseWarnings(t *testing.T) {
	genreData := []byte("0.1 sf_all;Фантастика\nbad line\n0.2 det_all;Детективы\n")

	meta := newMockMetadataStore()
	repo := &mockGenreTreeRepo{unsortedGenreID: 10}
	books := &mockBookRemapper{}

	svc := NewGenreTreeService(genreData, meta, repo, books)

	result, err := svc.ForceReload(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, result.GenresLoaded)
	assert.NotEmpty(t, result.Warnings)
}
