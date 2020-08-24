package example

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SCU-SJL/easymock/easymock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

const (
	addBooksUrl  = "https://www.easymock.com/liabrary/add"
	listBooksUrl = "https://www.easymock.com/library/list"
)

type LibraryTestSuite struct {
	suite.Suite
	library []Book
}

type Book struct {
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
	Author string  `json:"author"`
}

func TestLibrary(t *testing.T) {
	suite.Run(t, new(LibraryTestSuite))
}

func (lib *LibraryTestSuite) SetupSuite() {
	easymock.Start()
	easymock.RegisterResponder(http.MethodGet, listBooksUrl, lib.MockListAllBooks())
	easymock.RegisterResponder(http.MethodPost, addBooksUrl, lib.MockAddBooks())
}

func (lib *LibraryTestSuite) BeforeTest(suiteName, testName string) {
	fmt.Printf("[%s] - [%s] start\n", suiteName, testName)
}

func (lib *LibraryTestSuite) AfterTest(suiteName, testName string) {
	fmt.Printf("[%s] - [%s] end\n", suiteName, testName)
}

func (lib *LibraryTestSuite) TearDownSuite() {
	easymock.Shutdown()
}

func (lib *LibraryTestSuite) MockListAllBooks() *easymock.EasyResponder {
	responder := easymock.NewEasyResponderWithReqHandler(func(req *http.Request) (resp *http.Response, err error) {
		resp, err = easymock.NewHttpResponseWithJson(http.StatusOK, lib.library)
		if err != nil {
			return easymock.NewHttpResponseWithString(http.StatusInternalServerError, ""), nil
		}
		return resp, nil
	})
	return responder
}

func (lib *LibraryTestSuite) MockAddBooks() *easymock.EasyResponder {
	responder := easymock.NewEasyResponderWithReqHandler(func(req *http.Request) (resp *http.Response, err error) {
		books := make([]Book, 0)
		if err := json.NewDecoder(req.Body).Decode(&books); err != nil {
			return easymock.NewHttpResponseWithString(http.StatusBadRequest, ""), nil
		}
		lib.library = append(lib.library, books...)
		resp, err = easymock.NewHttpResponseWithJson(http.StatusOK, lib.library)
		if err != nil {
			return easymock.NewHttpResponseWithString(http.StatusInternalServerError, ""), nil
		}
		return resp, nil
	})
	return responder
}

func (lib *LibraryTestSuite) TestAddBooks() {
	book, err := json.Marshal([]Book{lib.newBook()})
	lib.Nil(err)
	body := bytes.NewBuffer(book)
	req, err := http.NewRequest(http.MethodPost, addBooksUrl, body)
	lib.Nil(err)
	req.Header.Set("Content-Type", "application/json")
	cli := new(http.Client)
	resp, err := cli.Do(req)
	lib.Nil(err)
	lib.Equal(http.StatusOK, resp.StatusCode)
	lib.showBooks()
}

func (lib *LibraryTestSuite) TestListBooks() {
	resp, err := http.Get(listBooksUrl)
	lib.Nil(err)

	books := make([]Book, 0)
	err = json.NewDecoder(resp.Body).Decode(&books)
	lib.Nil(err)

	lib.Equal(1, len(books))
	lib.Equal(lib.newBook(), books[0])
}

func (lib *LibraryTestSuite) showBooks() {
	fmt.Println("----------------------------")
	for _, b := range lib.library {
		fmt.Printf("%#v\n", b)
	}
	fmt.Println("----------------------------")
}

func (lib *LibraryTestSuite) newBook() Book {
	return Book{
		Name:   "Romeo And Juliet",
		Price:  45.7,
		Author: "莎士比亚",
	}
}
