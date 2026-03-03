package backend

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// FileDialogFilter 文件对话框过滤器
type FileDialogFilter struct {
	DisplayName string `json:"displayName"`
	Pattern     string `json:"pattern"`
}

// OpenFileDialog 打开文件选择对话框
func (a *App) OpenFileDialog(title string, filters []FileDialogFilter) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}

	// 转换过滤器格式
	runtimeFilters := make([]runtime.FileFilter, len(filters))
	for i, f := range filters {
		runtimeFilters[i] = runtime.FileFilter{
			DisplayName: f.DisplayName,
			Pattern:     f.Pattern,
		}
	}

	// 打开文件选择对话框
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: runtimeFilters,
	})

	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}

	return filePath, nil
}

// SaveFileDialog 打开文件保存对话框
func (a *App) SaveFileDialog(title string, defaultFilename string, filters []FileDialogFilter) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}

	// 转换过滤器格式
	runtimeFilters := make([]runtime.FileFilter, len(filters))
	for i, f := range filters {
		runtimeFilters[i] = runtime.FileFilter{
			DisplayName: f.DisplayName,
			Pattern:     f.Pattern,
		}
	}

	// 打开文件保存对话框
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
		Filters:         runtimeFilters,
	})

	if err != nil {
		return "", fmt.Errorf("failed to open save dialog: %w", err)
	}

	return filePath, nil
}

// OpenDirectoryDialog 打开目录选择对话框
func (a *App) OpenDirectoryDialog(title string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}

	// 打开目录选择对话框
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})

	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}

	return dirPath, nil
}
