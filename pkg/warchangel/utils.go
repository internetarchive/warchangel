package warchangel

func intToBool(i int) bool {
	if i == 1 {
		return true
	}
	return false
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
