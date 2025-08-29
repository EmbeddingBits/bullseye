package models

import (
	"io/fs"
	"time"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Entry    fs.DirEntry
	Size     int64
	ModTime  time.Time
	IsHidden bool
}

// Model represents the main application model
type Model struct {
	CurrentDir     string
	BaseDir        string
	ParentDir      string
	Files          []FileInfo
	ParentFiles    []FileInfo
	Selected       int
	ParentSelected int
	ListOffset     int
	Preview        string
	PreviewOffset  int
	Width          int
	Height         int
	Err            error
	Config         interface{} // Will be properly typed when imported
	ShowHidden     bool
	SortBy         string // "name", "size", "modified"
	ReverseSort    bool
	SearchMode     bool
	SearchQuery    string
	ImagePreviewColored bool
}
