package httpUtils

import "net/http"

type responseWriterWrapper struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

type PaginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Page   int `json:"page"`
}

type IndexResponse struct {
	Meta PaginationMeta `json:"meta"`
	Data interface{}    `json:"data"`
}

// NewWriterWrapper response writer wrapper
func NewWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w}
}

func (rw *responseWriterWrapper) Status() int {
	return rw.status
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}
