package core

import (
	"crypto/md5"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mosaan/mdatlas/pkg/types"
)

// Cache provides caching functionality for document structures
type Cache struct {
	mu         sync.RWMutex
	structures map[string]*CacheEntry
	maxSize    int
	ttl        time.Duration
}

// CacheEntry represents a cached document structure
type CacheEntry struct {
	Structure    *types.DocumentStructure
	LastAccessed time.Time
	FileModTime  time.Time
	FileHash     string
}

// NewCache creates a new cache instance
func NewCache(maxSize int, ttl time.Duration) *Cache {
	if maxSize <= 0 {
		maxSize = 100 // Default max size
	}

	if ttl <= 0 {
		ttl = 30 * time.Minute // Default TTL
	}

	cache := &Cache{
		structures: make(map[string]*CacheEntry),
		maxSize:    maxSize,
		ttl:        ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// GetStructure retrieves a cached document structure
func (c *Cache) GetStructure(filePath string) (*types.DocumentStructure, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.structures[filePath]
	if !exists {
		return nil, false
	}

	// Check if entry is expired
	if time.Since(entry.LastAccessed) > c.ttl {
		return nil, false
	}

	// Check if file has been modified
	if !c.isFileUnchanged(filePath, entry) {
		return nil, false
	}

	// Update access time
	entry.LastAccessed = time.Now()

	return entry.Structure, true
}

// SetStructure caches a document structure
func (c *Cache) SetStructure(filePath string, structure *types.DocumentStructure) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get file information
	stat, err := os.Stat(filePath)
	if err != nil {
		return // Skip caching if we can't stat the file
	}

	// Calculate file hash
	hash, err := c.calculateFileHash(filePath)
	if err != nil {
		return // Skip caching if we can't hash the file
	}

	// Check if we need to evict entries
	if len(c.structures) >= c.maxSize {
		c.evictLRU()
	}

	// Create cache entry
	entry := &CacheEntry{
		Structure:    structure,
		LastAccessed: time.Now(),
		FileModTime:  stat.ModTime(),
		FileHash:     hash,
	}

	c.structures[filePath] = entry
}

// InvalidateStructure removes a cached structure
func (c *Cache) InvalidateStructure(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.structures, filePath)
}

// Clear removes all cached structures
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.structures = make(map[string]*CacheEntry)
}

// Size returns the current number of cached structures
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.structures)
}

// Stats returns cache statistics
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		Size:    len(c.structures),
		MaxSize: c.maxSize,
		TTL:     c.ttl,
	}

	// Calculate oldest and newest entries
	var oldestAccess, newestAccess time.Time
	for _, entry := range c.structures {
		if oldestAccess.IsZero() || entry.LastAccessed.Before(oldestAccess) {
			oldestAccess = entry.LastAccessed
		}
		if newestAccess.IsZero() || entry.LastAccessed.After(newestAccess) {
			newestAccess = entry.LastAccessed
		}
	}

	stats.OldestEntry = oldestAccess
	stats.NewestEntry = newestAccess

	return stats
}

// isFileUnchanged checks if a file has been modified since caching
func (c *Cache) isFileUnchanged(filePath string, entry *CacheEntry) bool {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Check modification time
	if !stat.ModTime().Equal(entry.FileModTime) {
		return false
	}

	// Check file hash for additional verification
	hash, err := c.calculateFileHash(filePath)
	if err != nil {
		return false
	}

	return hash == entry.FileHash
}

// calculateFileHash calculates MD5 hash of a file
func (c *Cache) calculateFileHash(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := md5.Sum(content)
	return fmt.Sprintf("%x", hash), nil
}

// evictLRU evicts the least recently used entry
func (c *Cache) evictLRU() {
	if len(c.structures) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.structures {
		if oldestTime.IsZero() || entry.LastAccessed.Before(oldestTime) {
			oldestTime = entry.LastAccessed
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.structures, oldestKey)
	}
}

// cleanupExpired removes expired entries periodically
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		now := time.Now()
		for key, entry := range c.structures {
			if now.Sub(entry.LastAccessed) > c.ttl {
				delete(c.structures, key)
			}
		}

		c.mu.Unlock()
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size        int           `json:"size"`
	MaxSize     int           `json:"max_size"`
	TTL         time.Duration `json:"ttl"`
	OldestEntry time.Time     `json:"oldest_entry"`
	NewestEntry time.Time     `json:"newest_entry"`
}

// RefreshStructure forces a refresh of a cached structure
func (c *Cache) RefreshStructure(filePath string, parser *Parser) error {
	// Remove existing cache entry
	c.InvalidateStructure(filePath)

	// Read and parse the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	structure, err := parser.ParseStructure(content)
	if err != nil {
		return fmt.Errorf("failed to parse structure for %s: %w", filePath, err)
	}

	// Set file path and modification time
	structure.FilePath = filePath
	if stat, err := os.Stat(filePath); err == nil {
		structure.LastModified = stat.ModTime()
	}

	// Cache the new structure
	c.SetStructure(filePath, structure)

	return nil
}

// GetCachedFiles returns a list of currently cached file paths
func (c *Cache) GetCachedFiles() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	files := make([]string, 0, len(c.structures))
	for filePath := range c.structures {
		files = append(files, filePath)
	}

	return files
}

// WarmUpCache pre-loads structures for specified files
func (c *Cache) WarmUpCache(filePaths []string, parser *Parser) error {
	for _, filePath := range filePaths {
		if err := c.RefreshStructure(filePath, parser); err != nil {
			return fmt.Errorf("failed to warm up cache for %s: %w", filePath, err)
		}
	}

	return nil
}
