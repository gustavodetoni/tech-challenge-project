package plate

import "strings"

func Normalize(value string) string {
	v := strings.ToUpper(strings.TrimSpace(value))
	v = strings.ReplaceAll(v, "-", "")
	v = strings.ReplaceAll(v, " ", "")
	return v
}
