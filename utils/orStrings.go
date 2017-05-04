package utils

// OrStrings returns the first non empty string in its arguments
func OrStrings(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}
