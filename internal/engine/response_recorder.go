package engine

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
)

type ResponseRecorder struct {
	writer      http.ResponseWriter
	statusCode  int
	wroteHeader bool
	isHTML      bool
	decided     bool
	buf         bytes.Buffer
	bytesCount  int
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		writer:     w,
		statusCode: http.StatusOK,
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.writer.Header()
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.statusCode = statusCode
	r.wroteHeader = true
	r.decideMode()
	if !r.isHTML {
		r.writer.WriteHeader(statusCode)
	}
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		if r.Header().Get("Content-Type") == "" && len(b) > 0 {
			r.Header().Set("Content-Type", http.DetectContentType(b))
		}
		r.WriteHeader(http.StatusOK)
	}
	r.bytesCount += len(b)
	if r.isHTML {
		return r.buf.Write(b)
	}
	return r.writer.Write(b)
}

func (r *ResponseRecorder) FlushWithSnippet(snippet string) {
	if !r.wroteHeader {
		r.decideMode()
		if !r.isHTML {
			r.writer.WriteHeader(r.statusCode)
			return
		}
	}
	if !r.isHTML {
		return
	}

	body := r.buf.String()
	lower := strings.ToLower(body)
	idx := strings.LastIndex(lower, "</body>")
	var out string
	if idx != -1 {
		out = body[:idx] + snippet + body[idx:]
	} else {
		out = body + snippet
	}

	r.Header().Set("Content-Length", strconv.Itoa(len(out)))
	r.writer.WriteHeader(r.statusCode)
	_, _ = r.writer.Write([]byte(out))
}

func (r *ResponseRecorder) StatusCode() int {
	return r.statusCode
}

func (r *ResponseRecorder) Size() int {
	return r.bytesCount
}

func (r *ResponseRecorder) ContentType() string {
	return r.Header().Get("Content-Type")
}

func (r *ResponseRecorder) decideMode() {
	if r.decided {
		return
	}
	r.decided = true
	ct := strings.ToLower(r.Header().Get("Content-Type"))
	enc := r.Header().Get("Content-Encoding")
	r.isHTML = strings.Contains(ct, "text/html") && enc == ""
}
