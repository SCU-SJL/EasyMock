package test

import (
	"easy-mock/easymock"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"testing"
)

var (
	mockGoogleUrl     = "https://www.google.com"
	mockGoogleStrResp = "this is google"
	mockBingUrl       = "https://www.bing.com"
	mockBingBytesResp = []byte("this is bing")
	mockByteDanceUrl  = "https://www.bytedance.com"
)

type MockerTestSuite struct {
	suite.Suite
}

func TestMocker(t *testing.T) {
	suite.Run(t, new(MockerTestSuite))
}

func (suite *MockerTestSuite) BeforeTest(suiteName, testName string) {
	fmt.Printf("[%s] - [%s] start\n", suiteName, testName)
	easymock.Start()
	easymock.RegisterResponder(http.MethodGet, mockGoogleUrl,
		easymock.NewStringEasyResponder(http.StatusOK, mockGoogleStrResp))
	easymock.RegisterResponder(http.MethodGet, mockBingUrl,
		easymock.NewBytesEasyResponder(http.StatusOK, mockBingBytesResp))
}

func (suite *MockerTestSuite) AfterTest(suiteName, testName string) {
	easymock.Shutdown()
	fmt.Printf("[%s] - [%s] ended\n", suiteName, testName)
}

func (suite *MockerTestSuite) TestEasyMockWithStringResp() {
	resp, err := http.Get(mockGoogleUrl)

	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.RespBodyEqual(mockGoogleStrResp, resp)
}

func (suite *MockerTestSuite) TestEasyMockWithBytesResp() {
	resp, err := http.Get(mockBingUrl)

	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.RespBodyEqual(mockBingBytesResp, resp)
}

func (suite *MockerTestSuite) TestEasyMockWithJsonResp() {
	company := struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		PostCode int    `json:"post_code"`
	}{
		Name:     "ByteDance",
		Address:  "成都市OCG国际中心",
		PostCode: 610041,
	}

	jsonResponder, err := easymock.NewJsonEasyResponder(http.StatusOK, company)
	suite.Nil(err)
	easymock.RegisterResponder(http.MethodGet, mockByteDanceUrl, jsonResponder)

	resp, err := http.Get(mockByteDanceUrl)
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	expectedJson, _ := json.Marshal(company)
	suite.RespBodyEqual(expectedJson, resp)
}

func (suite *MockerTestSuite) TestEasyMockWithNoResponder() {
	easymock.RemoveResponder(http.MethodGet, mockGoogleUrl)
	resp, err := http.Get(mockGoogleUrl)
	suite.Nil(resp)
	suite.NotNil(err)
	log.Println(err)
}

func (suite *MockerTestSuite) RespBodyEqual(expected interface{}, actual *http.Response) {
	switch data := expected.(type) {
	case string:
		body := make([]byte, len(data))
		n, err := actual.Body.Read(body)
		suite.Nil(err)
		suite.Equal(data, string(body[:n]))
	case []byte:
		body := make([]byte, len(data))
		n, err := actual.Body.Read(body)
		suite.Nil(err)
		suite.Equal(data, body[:n])
	}
}
