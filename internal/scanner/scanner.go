package scanner

import (
	"io/fs"
	"path/filepath"
)

// FileInfo holds path and size for a scanned file
type FileInfo struct {
	Path string
	Size int64
}

// ScanError records a file that couldn't be scanned
type ScanError struct {
	Path  string
	Error string
}

// ScanResult holds all results from scanning a directory
type ScanResult struct {
	Files  []FileInfo
	Errors []ScanError
	Total  int64 // total bytes scanned
}

// Scanner walks directories and collects file information
type Scanner struct {
	minSize int64
}

// New creates a Scanner with the given minimum file size
func New(minSize int64) *Scanner {
	return &Scanner{minSize: minSize}
}

// Scan walks the directory tree and returns all files meeting the size criteria
func (s *Scanner) Scan(root string) (*ScanResult, error) {
	result := &ScanResult{
		Files:  []FileInfo{},
		Errors: []ScanError{},
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Handle errors accessing the path
		if err != nil {
			// If error is on root path, fail immediately
			if path == root {
				return err
			}
			result.Errors = append(result.Errors, ScanError{Path: path, Error: err.Error()})
			return nil
		}

		// Skip directories and non-regular files (symlinks, devices, etc.)
		if !d.Type().IsRegular() {
			return nil
		}

		// Get file info for size
		info, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, ScanError{Error: err.Error(), Path: path})
			return nil
		}
		fileSize := info.Size()

		// Skip files below minimum size
		if fileSize < s.minSize {
			return nil
		}

		// Add to results
		result.Files = append(result.Files, FileInfo{Path: path, Size: fileSize})
		result.Total += fileSize

		return nil
	})

	if err != nil {
		return nil, err // root path doesn't exist or similar fatal error
	}

	return result, nil

}
