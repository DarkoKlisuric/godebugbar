package engine

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const DefaultMaxBufferBytes = 4 << 20

type ResponseRecorder struct {
	writer      http.ResponseWriter
	injectHTML  bool
	headRequest bool
	maxBuffer   int
	statusCode  int
	wroteHeader bool
	buffering   bool
	hijacked    bool
	buf         bytes.Buffer
	bytesCount  int
}

func NewResponseRecorder(w http.ResponseWriter, r *http.Request, injectHTML bool, maxBuffer int) *ResponseRecorder {
	if maxBuffer <= 0 {
		maxBuffer = DefaultMaxBufferBytes
	}

	return &ResponseRecorder{
		writer:      w,
		injectHTML:  injectHTML,
		headRequest: r != nil && r.Method == http.MethodHead,
		maxBuffer:   maxBuffer,
		statusCode:  http.StatusOK,
	}
}

func (r *ResponseRecorder) Header() http.Header {
	return r.writer.Header()
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	if isInformational(statusCode) {
		r.writer.WriteHeader(statusCode)
		return
	}

	r.statusCode = statusCode
	r.wroteHeader = true
	r.buffering = r.shouldBuffer()
	if !r.buffering {
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

	if r.buffering && r.buf.Len()+len(b) > r.maxBuffer {
		if err := r.passthrough(); err != nil {
			return 0, err
		}
	}
	if r.buffering {
		return r.buf.Write(b)
	}

	return r.writer.Write(b)
}

func (r *ResponseRecorder) Flush() {
	_ = r.FlushError()
}

func (r *ResponseRecorder) FlushError() error {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	if err := r.passthrough(); err != nil {
		return err
	}

	return http.NewResponseController(r.writer).Flush()
}

func (r *ResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(r.writer).Hijack()
	if err == nil {
		r.hijacked = true
	}

	return conn, rw, err
}

func (r *ResponseRecorder) Unwrap() http.ResponseWriter {
	return r.writer
}

func (r *ResponseRecorder) Finish(snippet string) {
	if r.hijacked {
		return
	}
	if !r.wroteHeader {
		r.wroteHeader = true
		r.writer.WriteHeader(r.statusCode)
		return
	}
	if !r.buffering {
		return
	}

	out := InjectSnippet(r.buf.Bytes(), snippet)
	r.buffering = false
	r.buf.Reset()
	r.Header().Set("Content-Length", strconv.Itoa(len(out)))
	r.writer.WriteHeader(r.statusCode)
	_, _ = r.writer.Write(out)
}

func (r *ResponseRecorder) StatusCode() int {
	return r.statusCode
}

func (r *ResponseRecorder) Size() int {
	return r.bytesCount
}

func (r *ResponseRecorder) Hijacked() bool {
	return r.hijacked
}

func (r *ResponseRecorder) passthrough() error {
	if !r.buffering {
		return nil
	}

	r.buffering = false
	r.writer.WriteHeader(r.statusCode)
	_, err := r.writer.Write(r.buf.Bytes())
	r.buf.Reset()

	return err
}

func (r *ResponseRecorder) shouldBuffer() bool {
	if !r.injectHTML || r.headRequest || !allowsBody(r.statusCode) {
		return false
	}

	return IsInjectableHTML(r.Header())
}

func IsInjectableHTML(h http.Header) bool {
	ct := strings.ToLower(h.Get("Content-Type"))

	return strings.Contains(ct, "text/html") && h.Get("Content-Encoding") == ""
}

func isInformational(status int) bool {
	return status >= http.StatusContinue && status < http.StatusOK && status != http.StatusSwitchingProtocols
}

func allowsBody(status int) bool {
	return status != http.StatusNoContent && status != http.StatusNotModified
}
