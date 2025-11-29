package disk

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// OSInfo represents operating system information
type OSInfo struct {
	OS      string   `json:"os"`      // "windows", "linux", "darwin", etc.
	IsWindows bool   `json:"isWindows"`
	DefaultPath string `json:"defaultPath"`
	Drives   []string `json:"drives,omitempty"` // Windows drive letters
}

// GetOSInfo returns information about the current operating system
func GetOSInfo() *OSInfo {
	osName := runtime.GOOS
	isWindows := osName == "windows"
	
	defaultPath := "/"
	if isWindows {
		// Windows의 경우 현재 작업 디렉토리의 드라이브를 기본값으로 사용
		if wd, err := os.Getwd(); err == nil {
			if len(wd) >= 2 && wd[1] == ':' {
				defaultPath = wd[:2] + "\\"
			} else {
				defaultPath = "C:\\"
			}
		} else {
			defaultPath = "C:\\"
		}
	}

	info := &OSInfo{
		OS:          osName,
		IsWindows:   isWindows,
		DefaultPath: defaultPath,
	}

	if isWindows {
		info.Drives = GetWindowsDrives()
	}

	return info
}

// GetWindowsDrives returns a list of available Windows drive letters
func GetWindowsDrives() []string {
	drives := []string{}
	
	// Windows에서 사용 가능한 드라이브 문자 확인 (A-Z)
	for letter := 'A'; letter <= 'Z'; letter++ {
		drive := string(letter) + ":\\"
		// 드라이브가 존재하는지 확인
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}

	return drives
}

// NormalizePath normalizes a path according to the OS
func NormalizePath(path string) string {
	path = strings.TrimSpace(path)
	
	if runtime.GOOS == "windows" {
		// Windows 경로 정규화
		// 슬래시를 백슬래시로 변환
		path = strings.ReplaceAll(path, "/", "\\")
		
		// 빈 경로나 "/", "\"인 경우 기본 드라이브로
		if path == "" || path == "/" || path == "\\" {
			if wd, err := os.Getwd(); err == nil && len(wd) >= 2 && wd[1] == ':' {
				return wd[:2] + "\\"
			}
			return "C:\\"
		}
		
		// 경로가 드라이브 문자로 시작하는지 확인
		hasDrive := len(path) >= 2 && path[1] == ':'
		
		if !hasDrive {
			// 상대 경로인 경우 현재 드라이브 추가
			if wd, err := os.Getwd(); err == nil && len(wd) >= 2 && wd[1] == ':' {
				drive := wd[:2]
				// 경로가 백슬래시로 시작하면 드라이브만 추가
				if strings.HasPrefix(path, "\\") {
					path = drive + path
				} else {
					path = drive + "\\" + path
				}
			} else {
				// 작업 디렉토리를 가져올 수 없으면 C: 드라이브 사용
				if strings.HasPrefix(path, "\\") {
					path = "C:" + path
				} else {
					path = "C:\\" + path
				}
			}
		}
		
		// filepath.Clean으로 정규화 (중복 슬래시 제거 등)
		return filepath.Clean(path)
	} else {
		// Linux/Unix 경로 정규화
		if path == "" {
			return "/"
		}
		// 절대 경로가 아니면 슬래시 추가
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return filepath.Clean(path)
	}
}

