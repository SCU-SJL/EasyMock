package easymock

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
)

var (
	OriginTransport = http.DefaultTransport
	MockerTransport = NewEasyMockerTransport()
	OldClients      map[*http.Client]http.RoundTripper

	globalMu sync.Mutex

	routingFailedTmpl   = `routing failed, no responders were found for url '%s'`
	urlNotAvailableTmpl = `url '%s' is not available`
)

type EasyMocker struct {
	responderMu           sync.Mutex
	matchCntMu, missCntMu sync.Mutex
	responderMap          map[router]*EasyResponder
	regexResponderMap     map[router]*EasyRegexResponder
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
		responderMu:       sync.Mutex{},
		matchCntMu:        sync.Mutex{},
		missCntMu:         sync.Mutex{},
		responderMap:      make(map[router]*EasyResponder),
		regexResponderMap: make(map[router]*EasyRegexResponder),
		matchedCounter:    make(map[router]int),
		mismatchCounter:   make(map[router]int),
		totalCount:        0,
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
	if ok && responder.IsAvailable() {
		mocker.updateMatchCount(rt)
		return (*responder).reqHandler(req)
	}

	regexpResponder, regexOk := mocker.findRegexResponder(rt)
	if regexOk && regexpResponder.IsAvailable() {
		mocker.updateMatchCount(rt)
		return (*regexpResponder).reqHandler(req)
	}

	mocker.updateMismatchCount(rt)
	return mocker.connectFail(req,
		(ok && !responder.IsAvailable()) ||
		(regexOk && !regexpResponder.IsAvailable()))
}

func (mocker *EasyMocker) findRegexResponder(rt router) (*EasyRegexResponder, bool) {
	for _, regexResponder := range mocker.regexResponderMap {
		if regexResponder.isMatched(rt.Url) {
			return regexResponder, true
		}
	}
	return nil, false
}

func RegisterResponder(method, url string, responder *EasyResponder) {
	rt := router{
		Method: method,
		Url:    url,
	}

	MockerTransport.responderMu.Lock()
	defer MockerTransport.responderMu.Unlock()

	if _, ok := MockerTransport.responderMap[rt]; !ok {
		MockerTransport.responderMap[rt] = responder
		MockerTransport.matchedCounter[rt] = 0
	} else {
		registerFailed(rt)
	}
}

func RegisterRegexResponder(method, url string, regexResponder *EasyRegexResponder) {
	rt := router{
		Method: method,
		Url:    url,
	}
	MockerTransport.responderMu.Lock()
	defer MockerTransport.responderMu.Unlock()

	if _, ok := MockerTransport.regexResponderMap[rt]; !ok {
		regexResponder.oriUrl = url
		regexResponder.matcher = regexp.MustCompile(url)
		MockerTransport.regexResponderMap[rt] = regexResponder
		MockerTransport.matchedCounter[rt] = 0
	} else {
		registerFailed(rt)
	}
}

func registerFailed(rt router) {
	panic(fmt.Sprintf("responder of [%s - %s] already exists", rt.Method, rt.Url))
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

func (mocker *EasyMocker) connectFail(req *http.Request, available bool) (*http.Response, error) {
	if available {
		return nil, fmt.Errorf(routingFailedTmpl, req.URL.Scheme+`://`+req.URL.Host)
	}
	return nil, fmt.Errorf(urlNotAvailableTmpl, req.URL.Scheme+`://`+req.URL.Host)
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
