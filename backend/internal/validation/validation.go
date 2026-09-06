package validation

import "strings"

func Required(value string) bool {
	return strings.TrimSpace(value) != ""
}

func Email(value string) bool {
	value = strings.TrimSpace(value)
	return strings.Count(value, "@") == 1 && !strings.HasPrefix(value, "@") && !strings.HasSuffix(value, "@")
}
