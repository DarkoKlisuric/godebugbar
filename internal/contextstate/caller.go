package contextstate

import (
	"runtime"
	"strconv"
	"strings"
)

const callerDepth = 48

var defaultSkippedFrames = []string{
	"github.com/DarkoKlisuric/godebugbar/internal/",
	"github.com/DarkoKlisuric/godebugbar/adapter/",
	"github.com/DarkoKlisuric/godebugbar.",
	"gorm.io/",
	"github.com/jackc/pgx/",
	"database/sql.",
	"runtime.",
}

func Caller(skip []string) string {
	pcs := make([]uintptr, callerDepth)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		if frame.File != "" && !skippedFrame(frame, skip) {
			return frame.File + ":" + strconv.Itoa(frame.Line)
		}
		if !more {
			return ""
		}
	}
}

func skippedFrame(frame runtime.Frame, skip []string) bool {
	for _, s := range skip {
		if s != "" && (strings.Contains(frame.Function, s) || strings.Contains(frame.File, s)) {
			return true
		}
	}
	if strings.HasSuffix(frame.File, "_test.go") {
		return false
	}
	for _, prefix := range defaultSkippedFrames {
		if strings.HasPrefix(frame.Function, prefix) {
			return true
		}
	}

	return false
}
