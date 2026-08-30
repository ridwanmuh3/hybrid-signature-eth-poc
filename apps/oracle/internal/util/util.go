package util

func BoolToString(b bool) string {
	if b {
		return "OK"
	}
	return "FAIL"
}
