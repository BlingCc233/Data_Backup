// core/filters.go
package core

import (
	"log"
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

// ShouldInclude 判断一个文件/目录是否应该被包含在备份中
func (fc *FilterConfig) ShouldInclude(path string, info os.FileInfo) bool {
	// 规则: 任何一个 Exclude 规则匹配，则立即排除。
	//       如果定义了 Include 规则，则必须至少匹配一个 Include 规则。

	// 1. 路径筛选 (Exclude)
	for _, excludePath := range fc.ExcludePaths {
		if strings.HasPrefix(path, excludePath) {
			return false // 被排除路径匹配
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

	// 3. 时间筛选 (OlderThan / NewerThan)
	modTime := info.ModTime()
	if fc.OlderThan != nil && !modTime.Before(*fc.OlderThan) {
		return false // 不满足 "早于" 条件
	}
	if fc.NewerThan != nil && !modTime.After(*fc.NewerThan) {
		return false // 不满足 "晚于" 条件
	}

	// 4. 大小筛选 (仅对文件生效)
	if !info.IsDir() {
		size := info.Size()
		if fc.MinSize > 0 && size < fc.MinSize {
			return false // 小于最小尺寸
		}
		if fc.MaxSize != -1 && size > fc.MaxSize {
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
	// 如果定义了 IncludePaths，则路径必须匹配其中之一
	if len(fc.IncludePaths) > 0 {
		pathIncluded := false
		for _, includePath := range fc.IncludePaths {
			if strings.HasPrefix(path, includePath) {
				pathIncluded = true
				break
			}
		}
		if !pathIncluded {
			// 如果是目录，我们不能立即排除它，因为它的子文件可能匹配
			// 如果是文件，并且路径不匹配，则排除
			if !info.IsDir() {
				return false
			}
		}
	}

	// 6. 名称筛选 (Include)
	// 如果定义了 IncludeNames，则名称必须匹配其中之一
	if len(fc.IncludeNames) > 0 {
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

	// 通过所有检查
	log.Printf("Filter PASSED for: %s", path)
	return true
}
