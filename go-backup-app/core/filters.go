// core/filters.go
package core

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FilterConfig 定义了所有可用的筛选条件
type FilterConfig struct {
	// 路径筛选 (前缀匹配)
	IncludePaths []string `json:"includePaths"` // 只包含这些路径下的文件/目录
	ExcludePaths []string `json:"excludePaths"` // 排除这些路径下的文件/目录

	// 名称/类型筛选 (Glob 模式匹配)
	IncludeNames []string `json:"includeNames"` // e.g., "*.log", "important_*"
	ExcludeNames []string `json:"excludeNames"` // e.g., "temp_*", "*.tmp"

	// 时间筛选
	NewerThan *time.Time `json:"newerThan"` // 文件修改时间晚于此
	OlderThan *time.Time `json:"olderThan"` // 文件修改时间早于此

	// 大小筛选 (单位: bytes)
	MinSize int64 `json:"minSize"` // 最小文件大小
	MaxSize int64 `json:"maxSize"` // 最大文件大小, -1 表示无上限
}

func isPathWithin(path, base string) bool {
	if base == "" {
		return false
	}
	cleanPath := filepath.Clean(path)
	cleanBase := filepath.Clean(base)

	if cleanPath == cleanBase {
		return true
	}

	rel, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil {
		return false
	}

	if rel == "." {
		return true
	}

	// Not within if rel starts with ".." (outside base).
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	return true
}

func (fc *FilterConfig) matchIncludePath(path string, info os.FileInfo) bool {
	if len(fc.IncludePaths) == 0 {
		return true
	}
	for _, includePath := range fc.IncludePaths {
		if includePath == "" {
			continue
		}
		// If the current path is inside an includePath, include it.
		if isPathWithin(path, includePath) {
			return true
		}
		// If the current path is a directory that contains an includePath, include it so that Walk can reach it.
		if info.IsDir() && isPathWithin(includePath, path) {
			return true
		}
	}
	return false
}

// ShouldInclude 判断一个文件/目录是否应该被包含在备份中
func (fc *FilterConfig) ShouldInclude(path string, info os.FileInfo) bool {
	// 规则: 任何一个 Exclude 规则匹配，则立即排除。
	//       如果定义了 Include 规则，则必须至少匹配一个 Include 规则。

	// 1. 路径筛选 (Exclude)
	for _, excludePath := range fc.ExcludePaths {
		if excludePath == "" {
			continue
		}
		if isPathWithin(path, excludePath) {
			return false
		}
	}

	// 2. 名称筛选 (Exclude)
	name := info.Name()
	for _, excludeName := range fc.ExcludeNames {
		matched, err := filepath.Match(excludeName, name)
		if err == nil && matched {
			return false // 被排除名称匹配
		}
	}

	// 3. 时间筛选 (OlderThan / NewerThan) - 仅对非目录生效
	if !info.IsDir() {
		modTime := info.ModTime()
		if fc.OlderThan != nil && !modTime.Before(*fc.OlderThan) {
			return false // 不满足 "早于" 条件
		}
		if fc.NewerThan != nil && !modTime.After(*fc.NewerThan) {
			return false // 不满足 "晚于" 条件
		}
	}

	// 4. 大小筛选 (仅对文件生效)
	if info.Mode().IsRegular() {
		size := info.Size()
		if fc.MinSize > 0 && size < fc.MinSize {
			return false // 小于最小尺寸
		}
		// MaxSize <= 0 (including 0 and -1) means "no upper limit".
		if fc.MaxSize > 0 && size > fc.MaxSize {
			return false // 大于最大尺寸
		}
	}

	// --- Include 规则检查 ---
	// 如果没有定义任何 Include 规则，那么到这里就默认包含
	hasIncludeRules := len(fc.IncludePaths) > 0 || len(fc.IncludeNames) > 0
	if !hasIncludeRules {
		return true
	}

	// 5. 路径筛选 (Include)
	if !fc.matchIncludePath(path, info) {
		return false
	}

	// 6. 名称筛选 (Include)
	if len(fc.IncludeNames) > 0 {
		// includeNames 用于筛选具体条目，不应阻止遍历目录。
		if !info.IsDir() {
			nameIncluded := false
			for _, includeName := range fc.IncludeNames {
				matched, err := filepath.Match(includeName, name)
				if err == nil && matched {
					nameIncluded = true
					break
				}
			}
			if !nameIncluded {
				return false
			}
		}
	}

	return true
}
