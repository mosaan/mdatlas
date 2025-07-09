package types

import "time"

// DocumentStructure represents the structure information of a document
type DocumentStructure struct {
	FilePath     string    `json:"file_path"`
	TotalChars   int       `json:"total_chars"`
	TotalLines   int       `json:"total_lines"`
	Structure    []Section `json:"structure"`
	LastModified time.Time `json:"last_modified"`
}

// Section represents section information in the document
type Section struct {
	ID        string    `json:"id"`
	Level     int       `json:"level"`
	Title     string    `json:"title"`
	CharCount int       `json:"char_count"`
	LineCount int       `json:"line_count"`
	StartLine int       `json:"start_line"`
	EndLine   int       `json:"end_line"`
	Children  []Section `json:"children"`
}

// SectionContent represents the content of a section
type SectionContent struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	Format          string `json:"format"`
	IncludeChildren bool   `json:"include_children"`
}

// AccessConfig represents file access control settings
type AccessConfig struct {
	BaseDir     string   `json:"base_dir"`
	AllowedExts []string `json:"allowed_extensions"`
	MaxFileSize int64    `json:"max_file_size"`
}
