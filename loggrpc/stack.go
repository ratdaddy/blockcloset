package loggrpc

import (
	"path/filepath"
	"runtime"
	"strings"
)

type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

func CaptureStack(skip, maxFrames int) []StackFrame {
	const maxPC = 64
	var pcs [maxPC]uintptr

	n := runtime.Callers(skip, pcs[:])
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs[:n])
	out := make([]StackFrame, 0, maxFrames)

	isNoise := func(fn string) bool {
		return strings.HasPrefix(fn, "runtime.") ||
			strings.Contains(fn, "google.golang.org/grpc/") ||
			strings.Contains(fn, "go-grpc-middleware/") ||
			strings.Contains(fn, "interceptors/recovery") ||
			strings.Contains(fn, "logger.RecoverToStatus")
	}

	for i := 0; i < maxFrames; {
		fr, more := frames.Next()
		fn := fr.Function
		if fn == "" {
			fn = "unknown"
		}
		if !isNoise(fn) {
			out = append(out, StackFrame{
				Function: shortFunc(fn),
				File:     shortFile(fr.File),
				Line:     fr.Line,
			})
			i++
		}
		if !more {
			break
		}
	}

	return out
}

func shortFunc(fn string) string {
	if idx := strings.LastIndex(fn, "/"); idx >= 0 {
		return fn[idx+1:]
	}
	return fn
}

func shortFile(path string) string {
	dir := filepath.Dir(path)
	return filepath.Base(dir) + "/" + filepath.Base(path)
}
