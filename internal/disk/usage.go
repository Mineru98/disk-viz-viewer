package disk

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// DirInfo represents disk usage information for a directory
type DirInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	SizeStr  string    `json:"sizeStr"`
	IsDir    bool      `json:"isDir"`
	Children []DirInfo `json:"children,omitempty"`
}

// UsageResult represents the result of disk usage analysis
type UsageResult struct {
	RootPath   string    `json:"rootPath"`
	TotalSize  int64     `json:"totalSize"`
	TotalStr   string    `json:"totalStr"`
	Items      []DirInfo `json:"items"`
	Error      string    `json:"error,omitempty"`
}

// FormatSize converts bytes to human readable format
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return formatFloat(float64(bytes)/float64(TB)) + " TB"
	case bytes >= GB:
		return formatFloat(float64(bytes)/float64(GB)) + " GB"
	case bytes >= MB:
		return formatFloat(float64(bytes)/float64(MB)) + " MB"
	case bytes >= KB:
		return formatFloat(float64(bytes)/float64(KB)) + " KB"
	default:
		return formatFloat(float64(bytes)) + " B"
	}
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return string(rune('0'+int(f)%10)) + string([]byte{})
	}
	// Simple formatting without fmt package dependency in hot path
	return floatToString(f)
}

func floatToString(f float64) string {
	intPart := int64(f)
	fracPart := int64((f - float64(intPart)) * 100)
	if fracPart < 0 {
		fracPart = -fracPart
	}

	result := intToString(intPart)
	if fracPart > 0 {
		result += "." + padZero(fracPart)
	}
	return result
}

func intToString(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

func padZero(n int64) string {
	s := intToString(n)
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

// GetDirSize calculates the total size of a directory
func GetDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// AnalyzeDiskUsage analyzes disk usage for a given path
func AnalyzeDiskUsage(rootPath string, depth int) (*UsageResult, error) {
	// Clean and validate the path
	rootPath = filepath.Clean(rootPath)

	info, err := os.Stat(rootPath)
	if err != nil {
		return &UsageResult{
			RootPath: rootPath,
			Error:    err.Error(),
		}, err
	}

	if !info.IsDir() {
		return &UsageResult{
			RootPath:  rootPath,
			TotalSize: info.Size(),
			TotalStr:  FormatSize(info.Size()),
			Items: []DirInfo{
				{
					Name:    info.Name(),
					Path:    rootPath,
					Size:    info.Size(),
					SizeStr: FormatSize(info.Size()),
					IsDir:   false,
				},
			},
		}, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return &UsageResult{
			RootPath: rootPath,
			Error:    err.Error(),
		}, err
	}

	// Calculate sizes concurrently
	items := make([]DirInfo, 0, len(entries))
	var totalSize int64
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency to avoid too many open files
	semaphore := make(chan struct{}, 10)

	for _, entry := range entries {
		wg.Add(1)
		go func(e os.DirEntry) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			entryPath := filepath.Join(rootPath, e.Name())
			var size int64
			var children []DirInfo

			if e.IsDir() {
				size, _ = GetDirSize(entryPath)

				// If depth > 1, get children info
				if depth > 1 {
					children = getChildrenInfo(entryPath, depth-1)
				}
			} else {
				if info, err := e.Info(); err == nil {
					size = info.Size()
				}
			}

			dirInfo := DirInfo{
				Name:     e.Name(),
				Path:     entryPath,
				Size:     size,
				SizeStr:  FormatSize(size),
				IsDir:    e.IsDir(),
				Children: children,
			}

			mu.Lock()
			items = append(items, dirInfo)
			totalSize += size
			mu.Unlock()
		}(entry)
	}

	wg.Wait()

	// Sort by size descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Size > items[j].Size
	})

	return &UsageResult{
		RootPath:  rootPath,
		TotalSize: totalSize,
		TotalStr:  FormatSize(totalSize),
		Items:     items,
	}, nil
}

func getChildrenInfo(path string, depth int) []DirInfo {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	children := make([]DirInfo, 0, len(entries))

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		var size int64
		var subChildren []DirInfo

		if entry.IsDir() {
			size, _ = GetDirSize(entryPath)
			if depth > 1 {
				subChildren = getChildrenInfo(entryPath, depth-1)
			}
		} else {
			if info, err := entry.Info(); err == nil {
				size = info.Size()
			}
		}

		children = append(children, DirInfo{
			Name:     entry.Name(),
			Path:     entryPath,
			Size:     size,
			SizeStr:  FormatSize(size),
			IsDir:    entry.IsDir(),
			Children: subChildren,
		})
	}

	// Sort by size descending
	sort.Slice(children, func(i, j int) bool {
		return children[i].Size > children[j].Size
	})

	return children
}
