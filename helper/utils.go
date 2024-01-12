package helper

import "fmt"

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
