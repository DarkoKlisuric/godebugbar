package godebugbar

import (
	"bytes"
	"net/http"
	"strings"
)

const (
	contentTypeHeader = "Content-Type"
	contentLenHeader  = "Content-Length"
	htmlContentPrefix = "text/html"
	bodyCloseTag      = "</body>"
)

type responseRecorder struct {
	writer      http.ResponseWriter
	status      int
	size        int
	wroteHeader bool
	bufferHTML  bool
	body        bytes.Buffer
}

func newResponseRecorder(writer http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		writer: writer,
		status: http.StatusOK,
	}
}

func (r *responseRecorder) Header() http.Header {
	return r.writer.Header()
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}

	r.status = statusCode
	r.wroteHeader = true
	r.bufferHTML = strings.HasPrefix(strings.ToLower(r.writer.Header().Get(contentTypeHeader)), htmlContentPrefix)

	if !r.bufferHTML {
		r.writer.WriteHeader(statusCode)
	}
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		if r.writer.Header().Get(contentTypeHeader) == "" {
			r.writer.Header().Set(contentTypeHeader, http.DetectContentType(data))
		}

		r.WriteHeader(http.StatusOK)
	}

	if r.bufferHTML {
		written, err := r.body.Write(data)
		r.size += written

		return written, err
	}

	written, err := r.writer.Write(data)
	r.size += written

	return written, err
}

func (r *responseRecorder) FlushBuffered(snippet string) {
	if !r.bufferHTML {
		if !r.wroteHeader {
			r.writer.WriteHeader(r.status)
		}

		return
	}

	content := r.body.String()
	if snippet != "" {
		idx := strings.LastIndex(strings.ToLower(content), bodyCloseTag)
		if idx >= 0 {
			content = content[:idx] + snippet + content[idx:]
		} else {
			content += snippet
		}
	}

	r.writer.Header().Del(contentLenHeader)
	r.writer.WriteHeader(r.status)
	written, _ := r.writer.Write([]byte(content))
	r.size = written
}
