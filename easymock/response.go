package easymock

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type easyResponse struct {
	body   interface{}
	seeker io.ReadSeeker
}

func (rb *easyResponse) init() {
	if rb.seeker == nil {
		switch d := rb.body.(type) {
		case string:
			rb.seeker = strings.NewReader(d)
		case []byte:
			rb.seeker = bytes.NewReader(d)
		}
	}
}

func (rb *easyResponse) Read(p []byte) (n int, err error) {
	rb.init()
	n, err = rb.seeker.Read(p)
	if err == io.EOF {
		_, _ = rb.seeker.Seek(io.SeekStart, io.SeekStart)
	}
	return
}

func (rb *easyResponse) Close() (err error) {
	rb.init()
	_, err = rb.seeker.Seek(io.SeekStart, io.SeekStart)
	return
}

func (rb *easyResponse) Clone() *easyResponse {
	return newEasyResponse(rb.body)
}

func newEasyResponse(data interface{}) *easyResponse {
	return &easyResponse{
		body: data,
	}
}

func newStringResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		Status:        strconv.Itoa(statusCode),
		StatusCode:    statusCode,
		Header:        http.Header{},
		Body:          newEasyResponse(body),
		ContentLength: int64(len([]byte(body))),
	}
}

func newBytesResponse(statusCode int, body []byte) *http.Response {
	return &http.Response{
		Status: strconv.Itoa(statusCode),
		StatusCode: statusCode,
		Header: http.Header{},
		Body: newEasyResponse(body),
		ContentLength: int64(len(body)),
	}
}
