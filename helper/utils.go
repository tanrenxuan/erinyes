package helper

import "fmt"

// AddQuotation 返回带引号字符串
func AddQuotation(str string) string {
	return fmt.Sprintf("%q", str)
}
