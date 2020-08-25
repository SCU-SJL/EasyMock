package example

import (
	"fmt"
	"github.com/SCU-SJL/easymock/easymock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strconv"
	"testing"
)

const addBookUrlPrefix = `https://www.easymock.com/bookstore/add?[0-9a-z&]+`

type BookStoreTestSuite struct {
	suite.Suite
	books []Book
}

func TestBookStore(t *testing.T) {
	suite.Run(t, new(BookStoreTestSuite))
}

func (bs *BookStoreTestSuite) SetupSuite() {
	easymock.Start()
	easymock.RegisterResponder(http.MethodGet, listBooksUrl, bs.MockListAllBooks())
	easymock.RegisterRegexResponder(http.MethodPost, addBookUrlPrefix, bs.MockAddBook())
}

func (bs *BookStoreTestSuite) BeforeTest(suiteName, testName string) {
	fmt.Printf("[%s] - [%s] start\n", suiteName, testName)
}

func (bs *BookStoreTestSuite) AfterTest(suiteName, testName string) {
	fmt.Printf("[%s] - [%s] end\n", suiteName, testName)
}

func (bs *BookStoreTestSuite) TearDownSuite() {
	easymock.Shutdown()
}

func (bs *BookStoreTestSuite) MockListAllBooks() *easymock.EasyResponder {
	responder := easymock.NewEasyResponderWithReqHandler(func(req *http.Request) (resp *http.Response, err error) {
		resp, err = easymock.NewHttpResponseWithJson(http.StatusOK, bs.books)
		if err != nil {
			return easymock.NewHttpResponseWithString(http.StatusInternalServerError, ""), nil
		}
		return resp, nil
	})
	return responder
}

func (bs *BookStoreTestSuite) MockAddBook() *easymock.EasyRegexResponder {
	regexResponder := easymock.NewEasyRegexResponderWithReqHandler(func(req *http.Request) (resp *http.Response, err error) {
		price, err := strconv.ParseFloat(req.URL.Query().Get("price"), 64)
		if err != nil {
			return easymock.NewHttpResponseWithString(http.StatusBadRequest, "invalid price"), nil
		}
		query := req.URL.Query()
		book := Book{
			Name:   query.Get("name"),
			Price:  price,
			Author: query.Get("author"),
		}
		bs.books = append(bs.books, book)
		return easymock.NewHttpResponseWithJson(http.StatusOK, bs.books)
	})
	return regexResponder
}

func (bs *BookStoreTestSuite) TestAddBook() {
	req, err := http.NewRequest(http.MethodPost, "https://www.easymock.com/bookstore/add?name=Hello&price=34.2&author=sjl", nil)
	bs.Nil(err)
	cli := new(http.Client)
	resp, err := cli.Do(req)
	bs.Nil(err)
	buf := make([]byte, 128)
	n, err := resp.Body.Read(buf)
	bs.Nil(err)
	fmt.Println(string(buf[:n]))
}

func (bs *BookStoreTestSuite) TestAddBook2() {
	req, err := http.NewRequest(http.MethodPost, "https://www.easymock.com/bookstore/add?name=World&price=24.2&author=sjl", nil)
	bs.Nil(err)
	cli := new(http.Client)
	resp, err := cli.Do(req)
	bs.Nil(err)
	buf := make([]byte, 128)
	n, err := resp.Body.Read(buf)
	bs.Nil(err)
	fmt.Println(string(buf[:n]))
}
