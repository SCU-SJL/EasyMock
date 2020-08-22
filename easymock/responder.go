package easymock

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"sync"
)

type EasyResponder struct {
	mu            sync.Mutex
	reqHandleFunc func(req *http.Request) (resp *http.Response, err error)
	available     bool
}

func NewStringEasyResponder(statusCode int, respBody string) *EasyResponder {
	resp := newStringResponse(statusCode, respBody)
	return NewEasyResponder(resp)
}

func NewBytesEasyResponder(statusCode int, respBody []byte) *EasyResponder {
	resp := newBytesResponse(statusCode, respBody)
	return NewEasyResponder(resp)
}

func NewJsonEasyResponder(statusCode int, respBody interface{}) (*EasyResponder, error) {
	jsonBody, err := json.Marshal(respBody)
	if err != nil {
		return nil, err
	}
	resp := newBytesResponse(statusCode, jsonBody)
	resp.Header.Set("Content-Type", "application/json")
	return NewEasyResponder(resp), nil
}

func NewXmlEasyResponder(statusCode int, respBody interface{}) (*EasyResponder, error) {
	xmlBody, err := xml.Marshal(respBody)
	if err != nil {
		return nil, err
	}
	resp := newBytesResponse(statusCode, xmlBody)
	resp.Header.Set("Content-Type", "application/xml")
	return NewEasyResponder(resp), nil
}

func NewEasyResponder(resp *http.Response) *EasyResponder {
	reqHandler := func(req *http.Request) (*http.Response, error) {
		res := *resp
		if body, ok := resp.Body.(*easyResponse); ok {
			res.Body = body.Clone()
		}
		res.Request = req
		return &res, nil
	}

	responder := &EasyResponder{
		mu:            sync.Mutex{},
		reqHandleFunc: reqHandler,
		available:     true,
	}
	return responder
}

func (eR *EasyResponder) Enable() {
	eR.mu.Lock()
	eR.available = true
	eR.mu.Unlock()
}

func (eR *EasyResponder) Disable() {
	eR.mu.Lock()
	eR.available = false
	eR.mu.Unlock()
}

func (eR *EasyResponder) IsAvailable() bool {
	eR.mu.Lock()
	defer eR.mu.Unlock()
	return eR.available
}
