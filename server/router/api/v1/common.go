package v1

import (
	"fmt"
)

// extractIDFromResourceName 从资源名称中提取ID
func extractIDFromResourceName(name, expectedType string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("resource name is required")
	}

	// 分割资源名称
	parts := []string{}
	current := ""
	for _, r := range name {
		if r == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid resource name format: expected %s/{id}", expectedType)
	}

	resourceType := parts[0]
	if resourceType != expectedType {
		return 0, fmt.Errorf("invalid resource type: expected %s, got %s", expectedType, resourceType)
	}

	// 将ID解析为int64
	var id int64
	if _, err := fmt.Sscanf(parts[1], "%d", &id); err != nil {
		return 0, fmt.Errorf("invalid resource ID: %w", err)
	}

	if id <= 0 {
		return 0, fmt.Errorf("invalid resource ID: must be positive")
	}

	return id, nil
}
