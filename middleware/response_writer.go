package middleware

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

// WrapResponseWriter wraps http.ResponseWriter to capture status and bytes written.
type WrapResponseWriter interface {
	http.ResponseWriter
	Status() int
	BytesWritten() int
	WroteHeader() bool
	Unwrap() http.ResponseWriter
}

type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	wroteHeader  bool
}

func NewWrapResponseWriter(w http.ResponseWriter) WrapResponseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) BytesWritten() int {
	return rw.bytesWritten
}

func (rw *responseWriter) WroteHeader() bool {
	return rw.wroteHeader
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("webserver doesn't support hijacking")
}

func (rw *responseWriter) ReadFrom(src io.Reader) (int64, error) {
	if rf, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		if !rw.wroteHeader {
			rw.WriteHeader(http.StatusOK)
		}
		n, err := rf.ReadFrom(src)
		rw.bytesWritten += int(n)
		return n, err
	}
	return io.Copy(rw.ResponseWriter, src)
}
