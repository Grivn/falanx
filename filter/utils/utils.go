package utils

// StringToHex converts string to hex format
func StringToHex(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s
	}
	return "0x" + s
}

// HexToString converts hex string to string without '0x' prefix
func HexToString(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	}
	return s
}
