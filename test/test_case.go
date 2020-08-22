package test

import (
	"easy-mock/easymock"
	"io"
	"net/http"
)

type easyMockTestCase interface {
	setupCase()
	tearDownCase()
	getResult() (*http.Response, error)
}

type baseCase struct {
	method           string
	url              string
	responder        *easymock.EasyResponder
	shouldNoResponse bool
	shouldRemove     bool
}

func (base *baseCase) setupCase() {
	if base.responder != nil {
		easymock.RegisterResponder(base.method, base.url, base.responder)
	}
}

func (base *baseCase) tearDownCase() {
	if base.shouldRemove {
		easymock.RemoveResponder(base.method, base.url)
	}
}

type httpGetCase struct {
	baseCase
	expectedRespBody interface{}
}

func (gc *httpGetCase) getResult() (*http.Response, error) {
	return http.Get(gc.url)
}

type httpPostCase struct {
	baseCase
	postContentType  string
	postBody         io.Reader
	expectedRespBody interface{}
}

func (gc *httpPostCase) getResult() (*http.Response, error) {
	return http.Post(gc.url, gc.postContentType, gc.postBody)
}
