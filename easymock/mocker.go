package easymock

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	OriginTransport = http.DefaultTransport
	MockerTransport = NewEasyMockerTransport()
	OldClients      map[*http.Client]http.RoundTripper

	globalMu sync.Mutex

	routingFailedTmpl = `routing failed, no responders were found for url '%s'`
)

// TODO add no responder to EasyMocker
type EasyMocker struct {
	responderMu           sync.Mutex
	matchCntMu, missCntMu sync.Mutex
	responderMap          map[router]*EasyResponder
	matchedCounter        map[router]int
	mismatchCounter       map[router]int
	totalCount            int
}

type router struct {
	Method string
	Url    string
}

func NewEasyMockerTransport() *EasyMocker {
	return &EasyMocker{
		responderMu:     sync.Mutex{},
		matchCntMu:      sync.Mutex{},
		responderMap:    make(map[router]*EasyResponder),
		matchedCounter:  make(map[router]int),
		mismatchCounter: make(map[router]int),
		totalCount:      0,
	}
}

func Start() {
	globalMu.Lock()
	if http.DefaultTransport != MockerTransport {
		OriginTransport = http.DefaultTransport
	}
	http.DefaultTransport = MockerTransport
	globalMu.Unlock()
}

func StartWithClient(client *http.Client) {
	globalMu.Lock()
	if _, exist := OldClients[client]; exist {
		OldClients[client] = client.Transport
	}
	client.Transport = MockerTransport
	globalMu.Unlock()
}

func Reset() {
	globalMu.Lock()
	MockerTransport.matchedCounter = make(map[router]int)
	MockerTransport.responderMap = make(map[router]*EasyResponder)
	MockerTransport.mismatchCounter = make(map[router]int)
	MockerTransport.totalCount = 0
	globalMu.Unlock()
}

func Shutdown() {
	globalMu.Lock()
	http.DefaultTransport = OriginTransport
	for client, transport := range OldClients {
		client.Transport = transport
		delete(OldClients, client)
	}
	globalMu.Unlock()
}

func (mocker *EasyMocker) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}
	rt := router{
		Method: method,
		Url:    url,
	}

	responder, ok := mocker.responderMap[rt]

	if !ok {
		mocker.updateMismatchCount(rt)
		return mocker.connectFail(req)
	}

	mocker.updateMatchCount(rt)
	return (*responder).reqHandleFunc(req)
}

func RegisterResponder(method, url string, responder *EasyResponder) {
	rt := router{
		Method: method,
		Url:    url,
	}
	MockerTransport.responderMu.Lock()
	MockerTransport.responderMap[rt] = responder
	MockerTransport.matchedCounter[rt] = 0
	MockerTransport.responderMu.Unlock()
}

func RemoveResponder(method, url string) {
	rt := router{
		Method: method,
		Url:    url,
	}
	MockerTransport.responderMu.Lock()
	delete(MockerTransport.responderMap, rt)
	MockerTransport.responderMu.Unlock()
}

func (mocker *EasyMocker) connectFail(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf(routingFailedTmpl, req.URL.Scheme+`://`+req.URL.Host)
}

func (mocker *EasyMocker) updateMatchCount(rt router) {
	mocker.matchCntMu.Lock()
	if _, exist := mocker.matchedCounter[rt]; exist {
		mocker.matchedCounter[rt]++
	} else {
		mocker.matchedCounter[rt] = 1
	}
	mocker.matchCntMu.Unlock()
}

func (mocker *EasyMocker) updateMismatchCount(rt router) {
	mocker.missCntMu.Lock()
	if _, exist := mocker.mismatchCounter[rt]; exist {
		mocker.mismatchCounter[rt]++
	} else {
		mocker.mismatchCounter[rt] = 1
	}
	mocker.missCntMu.Unlock()
}
