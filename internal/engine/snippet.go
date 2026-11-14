package engine

const closingBodyTag = "</body>"

func InjectSnippet(body []byte, snippet string) []byte {
	if snippet == "" {
		return body
	}

	idx := lastIndexFoldASCII(body, closingBodyTag)
	if idx == -1 {
		idx = len(body)
	}

	out := make([]byte, 0, len(body)+len(snippet))
	out = append(out, body[:idx]...)
	out = append(out, snippet...)
	out = append(out, body[idx:]...)

	return out
}

func lastIndexFoldASCII(s []byte, sub string) int {
	for i := len(s) - len(sub); i >= 0; i-- {
		if equalFoldASCII(s[i:i+len(sub)], sub) {
			return i
		}
	}

	return -1
}

func equalFoldASCII(b []byte, s string) bool {
	for i := range len(s) {
		if lowerASCII(b[i]) != lowerASCII(s[i]) {
			return false
		}
	}

	return true
}

func lowerASCII(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}

	return c
}
