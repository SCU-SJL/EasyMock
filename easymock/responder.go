package easymock

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"regexp"
	"sync"
)

type RequestHandler func(req *http.Request) (resp *http.Response, err error)

type EasyResponder struct {
	mu         sync.Mutex
	reqHandler RequestHandler
	available  bool
}

func NewStringEasyResponder(statusCode int, respBody string) *EasyResponder {
	resp := NewHttpResponseWithString(statusCode, respBody)
	return NewEasyResponderWithResp(resp)
}

func NewBytesEasyResponder(statusCode int, respBody []byte) *EasyResponder {
	resp := NewHttpResponseWithBytes(statusCode, respBody)
	return NewEasyResponderWithResp(resp)
}

func NewJsonEasyResponder(statusCode int, respBody interface{}) (*EasyResponder, error) {
	jsonBody, err := json.Marshal(respBody)
	if err != nil {
		return nil, err
	}
	resp := NewHttpResponseWithBytes(statusCode, jsonBody)
	resp.Header.Set("Content-Type", "application/json")
	return NewEasyResponderWithResp(resp), nil
}

func NewXmlEasyResponder(statusCode int, respBody interface{}) (*EasyResponder, error) {
	xmlBody, err := xml.Marshal(respBody)
	if err != nil {
		return nil, err
	}
	resp := NewHttpResponseWithBytes(statusCode, xmlBody)
	resp.Header.Set("Content-Type", "application/xml")
	return NewEasyResponderWithResp(resp), nil
}

func NewEasyResponderWithResp(resp *http.Response) *EasyResponder {
	reqHandler := func(req *http.Request) (*http.Response, error) {
		res := *resp
		if body, ok := resp.Body.(*easyResponse); ok {
			res.Body = body.Clone()
		}
		res.Request = req
		return &res, nil
	}

	responder := &EasyResponder{
		mu:         sync.Mutex{},
		reqHandler: reqHandler,
		available:  true,
	}
	return responder
}

func NewEasyResponderWithReqHandler(reqHandler RequestHandler) *EasyResponder {
	return &EasyResponder{
		mu:         sync.Mutex{},
		reqHandler: reqHandler,
		available:  true,
	}
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

type EasyRegexResponder struct {
	*EasyResponder
	oriUrl  string
	matcher *regexp.Regexp
}

func (eRR *EasyRegexResponder) isMatched(url string) bool {
	return eRR.matcher.Match([]byte(url))
}

func NewEasyRegexResponderWithReqHandler(reqHandler RequestHandler) *EasyRegexResponder {
	return &EasyRegexResponder{
		EasyResponder: NewEasyResponderWithReqHandler(reqHandler),
	}
}