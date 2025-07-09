package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mosaan/mdatlas/pkg/types"
)

// AccessControl manages file access restrictions and security
type AccessControl struct {
	config *types.AccessConfig
}

// NewAccessControl creates a new AccessControl instance
func NewAccessControl(baseDir string) (*AccessControl, error) {
	// Resolve base directory to absolute path
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}
	
	// Check if base directory exists
	if _, err := os.Stat(absBaseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("base directory does not exist: %s", absBaseDir)
	}
	
	config := &types.AccessConfig{
		BaseDir:     absBaseDir,
		AllowedExts: []string{".md", ".markdown", ".txt"},
		MaxFileSize: 50 * 1024 * 1024, // 50MB
	}
	
	return &AccessControl{config: config}, nil
}

// IsAllowed checks if access to a file path is allowed
func (ac *AccessControl) IsAllowed(filePath string) bool {
	// Resolve to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	
	// Check if path is within base directory
	if !ac.isWithinBaseDir(absPath) {
		return false
	}
	
	// Check file extension
	if !ac.isAllowedExtension(absPath) {
		return false
	}
	
	// Check file size
	if !ac.isAllowedSize(absPath) {
		return false
	}
	
	// Check if file exists and is readable
	if !ac.isReadable(absPath) {
		return false
	}
	
	return true
}

// ValidatePath validates and normalizes a file path
func (ac *AccessControl) ValidatePath(filePath string) (string, error) {
	// Resolve relative to base directory
	var absPath string
	if filepath.IsAbs(filePath) {
		absPath = filePath
	} else {
		absPath = filepath.Join(ac.config.BaseDir, filePath)
	}
	
	// Clean the path to remove any path traversal attempts
	cleanPath := filepath.Clean(absPath)
	
	// Check if path is within base directory
	if !ac.isWithinBaseDir(cleanPath) {
		return "", fmt.Errorf("path outside base directory: %s", filePath)
	}
	
	// Check file extension
	if !ac.isAllowedExtension(cleanPath) {
		return "", fmt.Errorf("file extension not allowed: %s", filepath.Ext(cleanPath))
	}
	
	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}
	
	// Check file size
	if !ac.isAllowedSize(cleanPath) {
		return "", fmt.Errorf("file too large: %s", filePath)
	}
	
	return cleanPath, nil
}

// isWithinBaseDir checks if a path is within the base directory
func (ac *AccessControl) isWithinBaseDir(absPath string) bool {
	// Ensure both paths end with separator for proper comparison
	baseDir := ac.config.BaseDir
	if !strings.HasSuffix(baseDir, string(os.PathSeparator)) {
		baseDir += string(os.PathSeparator)
	}
	
	if !strings.HasSuffix(absPath, string(os.PathSeparator)) {
		// For files, check if the directory is within base
		dir := filepath.Dir(absPath) + string(os.PathSeparator)
		return strings.HasPrefix(dir, baseDir)
	}
	
	return strings.HasPrefix(absPath, baseDir)
}

// isAllowedExtension checks if the file extension is allowed
func (ac *AccessControl) isAllowedExtension(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	for _, allowedExt := range ac.config.AllowedExts {
		if ext == allowedExt {
			return true
		}
	}
	
	return false
}

// isAllowedSize checks if the file size is within limits
func (ac *AccessControl) isAllowedSize(filePath string) bool {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	
	return stat.Size() <= ac.config.MaxFileSize
}

// isReadable checks if the file is readable
func (ac *AccessControl) isReadable(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// GetConfig returns a copy of the access configuration
func (ac *AccessControl) GetConfig() types.AccessConfig {
	return *ac.config
}

// UpdateConfig updates the access configuration
func (ac *AccessControl) UpdateConfig(config types.AccessConfig) error {
	// Validate base directory
	absBaseDir, err := filepath.Abs(config.BaseDir)
	if err != nil {
		return fmt.Errorf("invalid base directory: %w", err)
	}
	
	if _, err := os.Stat(absBaseDir); os.IsNotExist(err) {
		return fmt.Errorf("base directory does not exist: %s", absBaseDir)
	}
	
	// Validate file size limit
	if config.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive")
	}
	
	// Validate allowed extensions
	if len(config.AllowedExts) == 0 {
		return fmt.Errorf("at least one allowed extension must be specified")
	}
	
	// Update configuration
	ac.config = &types.AccessConfig{
		BaseDir:     absBaseDir,
		AllowedExts: config.AllowedExts,
		MaxFileSize: config.MaxFileSize,
	}
	
	return nil
}

// ListAllowedFiles lists all files within the base directory that are allowed
func (ac *AccessControl) ListAllowedFiles() ([]string, error) {
	var allowedFiles []string
	
	err := filepath.Walk(ac.config.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Check if file is allowed
		if ac.IsAllowed(path) {
			// Convert to relative path from base directory
			relPath, err := filepath.Rel(ac.config.BaseDir, path)
			if err != nil {
				return err
			}
			allowedFiles = append(allowedFiles, relPath)
		}
		
		return nil
	})
	
	return allowedFiles, err
}

// GetFileInfo returns information about a file if access is allowed
func (ac *AccessControl) GetFileInfo(filePath string) (*FileInfo, error) {
	validPath, err := ac.ValidatePath(filePath)
	if err != nil {
		return nil, err
	}
	
	stat, err := os.Stat(validPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	
	// Calculate relative path from base directory
	relPath, err := filepath.Rel(ac.config.BaseDir, validPath)
	if err != nil {
		relPath = validPath
	}
	
	return &FileInfo{
		Path:         validPath,
		RelativePath: relPath,
		Size:         stat.Size(),
		ModTime:      stat.ModTime(),
		IsDir:        stat.IsDir(),
		Extension:    filepath.Ext(validPath),
	}, nil
}

// FileInfo represents information about a file
type FileInfo struct {
	Path         string    `json:"path"`
	RelativePath string    `json:"relative_path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	IsDir        bool      `json:"is_dir"`
	Extension    string    `json:"extension"`
}

// SecureFileReader provides secure file reading with access control
type SecureFileReader struct {
	accessControl *AccessControl
}

// NewSecureFileReader creates a new secure file reader
func NewSecureFileReader(accessControl *AccessControl) *SecureFileReader {
	return &SecureFileReader{
		accessControl: accessControl,
	}
}

// ReadFile securely reads a file with access control
func (sfr *SecureFileReader) ReadFile(filePath string) ([]byte, error) {
	validPath, err := sfr.accessControl.ValidatePath(filePath)
	if err != nil {
		return nil, err
	}
	
	content, err := os.ReadFile(validPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	return content, nil
}

// ReadFileLines securely reads file lines with access control
func (sfr *SecureFileReader) ReadFileLines(filePath string, startLine, endLine int) ([]string, error) {
	content, err := sfr.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Validate line ranges
	if startLine < 1 {
		startLine = 1
	}
	if endLine < 1 || endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > endLine {
		return nil, fmt.Errorf("invalid line range: start=%d, end=%d", startLine, endLine)
	}
	
	return lines[startLine-1:endLine], nil
}

