package helper

import (
	"fmt"
	"strings"
)

// AddQuotation 返回带引号字符串
func AddQuotation(str string) string {
	return fmt.Sprintf("%q", str)
}

// MyStringIf string 类型三元表达式
func MyStringIf(b bool, s1, s2 string) string {
	if b {
		return s1
	} else {
		return s2
	}
}

// JoinKeys 用于合并一个map的所有key
func JoinKeys(m map[string]bool, delimiter string) string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	return strings.Join(keys, delimiter)
}

func SliceContainsTarget(slice []string, target string) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}
