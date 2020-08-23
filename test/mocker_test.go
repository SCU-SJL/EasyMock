package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/SCU-SJL/easymock/easymock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

var (
	mockGoogleUrl       = "https://www.google.com"
	mockGoogleStrResp   = "this is google"
	mockGoogleBytesResp = []byte("this is bing")
	mockNoResponseUrl   = "https://www.notfound.com"
	company             = Company{
		Name:     "ByteDance",
		Address:  "成都市OCG国际中心",
		PostCode: 610041,
	}
)

type Company struct {
	Name     string `json:"name" xml:"name"`
	Address  string `json:"address" xml:"address"`
	PostCode int    `json:"post_code" xml:"post_code"`
}

type MockerTestSuite struct {
	suite.Suite
}

func TestMocker(t *testing.T) {
	suite.Run(t, new(MockerTestSuite))
}

func (suite *MockerTestSuite) BeforeTest(suiteName, testName string) {
	fmt.Printf("[%s] - [%s] start\n", suiteName, testName)
	easymock.Start()
}

func (suite *MockerTestSuite) AfterTest(suiteName, testName string) {
	easymock.Shutdown()
	fmt.Printf("[%s] - [%s] ended\n", suiteName, testName)
}

func (suite *MockerTestSuite) TestEasyMockWithStringResp() {
	suite.execTestCase(&httpGetCase{
		baseCase: baseCase{
			method:       http.MethodGet,
			url:          mockGoogleUrl,
			responder:    easymock.NewStringEasyResponder(http.StatusOK, mockGoogleStrResp),
			shouldRemove: true,
		},
		expectedRespBody: mockGoogleStrResp,
	})
}

func (suite *MockerTestSuite) TestEasyMockWithBytesResp() {
	suite.execTestCase(&httpGetCase{
		baseCase: baseCase{
			method:       http.MethodGet,
			url:          mockGoogleUrl,
			responder:    easymock.NewBytesEasyResponder(http.StatusOK, mockGoogleBytesResp),
			shouldRemove: true,
		},
		expectedRespBody: mockGoogleBytesResp,
	})
}

func (suite *MockerTestSuite) TestEasyMockWithJsonResp() {
	jsonResponder, err := easymock.NewJsonEasyResponder(http.StatusOK, company)
	suite.Nil(err)
	expectedRespBody, err := json.Marshal(company)
	suite.Nil(err)

	suite.execTestCase(&httpGetCase{
		baseCase: baseCase{
			method:       http.MethodGet,
			url:          mockGoogleUrl,
			responder:    jsonResponder,
			shouldRemove: true,
		},
		expectedRespBody: expectedRespBody,
	})
}

func (suite *MockerTestSuite) TestEasyMockWithJsonResp2() {
	m := map[string]interface{}{
		"name":   "sjl",
		"age":    20,
		"school": "SCU",
	}
	jsonResponder, err := easymock.NewJsonEasyResponder(http.StatusOK, m)
	suite.Nil(err)
	expectedRespBody, err := json.Marshal(m)
	suite.Nil(err)

	suite.execTestCase(&httpGetCase{
		baseCase: baseCase{
			method:       http.MethodGet,
			url:          mockGoogleUrl,
			responder:    jsonResponder,
			shouldRemove: true,
		},
		expectedRespBody: expectedRespBody,
	})
}

func (suite *MockerTestSuite) TestEasyMockWithXmlResp() {
	xmlResponder, err := easymock.NewXmlEasyResponder(http.StatusOK, company)
	suite.Nil(err)
	expectedRespBody, err := xml.Marshal(company)
	suite.Nil(err)

	suite.execTestCase(&httpGetCase{
		baseCase: baseCase{
			method:       http.MethodGet,
			url:          mockGoogleUrl,
			responder:    xmlResponder,
			shouldRemove: true,
		},
		expectedRespBody: expectedRespBody,
	})
}

func (suite *MockerTestSuite) TestEasyMockWithNoResponder() {
	suite.execTestCase(&httpGetCase{
		baseCase: baseCase{
			method:           http.MethodGet,
			url:              mockNoResponseUrl,
			responder:        nil,
			shouldNoResponse: true,
		},
		expectedRespBody: nil,
	})
}

func (suite *MockerTestSuite) execTestCase(testCase easyMockTestCase) {
	testCase.setupCase()
	suite.checkResult(testCase)
	testCase.tearDownCase()
}

func (suite *MockerTestSuite) checkResult(testCase easyMockTestCase) {
	switch typedCase := testCase.(type) {
	case *httpGetCase:
		suite.checkHttpGetResp(typedCase)
	}
}

func (suite *MockerTestSuite) checkHttpGetResp(testCase *httpGetCase) {
	resp, err := testCase.getResult()
	if testCase.shouldNoResponse {
		suite.NotNil(err)
		suite.Nil(resp)
		return
	}

	suite.Nil(err)
	actualBody := make([]byte, resp.ContentLength)

	n, err := resp.Body.Read(actualBody)
	suite.Nil(err)

	switch expectedResp := testCase.expectedRespBody.(type) {
	case string:
		suite.Equal([]byte(expectedResp), actualBody[:n])
	case []byte:
		suite.Equal(expectedResp, actualBody)
	}
}
