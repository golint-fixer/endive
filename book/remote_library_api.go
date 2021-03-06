package book

import (
	"encoding/xml"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RemoteLibraryAPI is the interface for accessing remote library information.
type RemoteLibraryAPI interface {
	GetBook(id, key string) (Metadata, error)
	GetBookIDByQuery(author, title, key string) (id string, err error)
	GetBookIDByISBN(isbn, key string) (id string, err error)
}

// GetXMLData retrieves XML responses from online APIs.
func getXMLData(uri string, i interface{}) (err error) {
	currentPass := 0
	const maxTries = 5
	var data []byte
	for currentPass < maxTries {
		data, err = getRequest(uri)
		if err != nil {
			currentPass++
			// wait a little
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	// test if the last pass was successful
	if err != nil {
		return
	}
	return xml.Unmarshal(data, i)
}

func getRequest(uri string) (body []byte, err error) {
	// 10s timeout
	timeout := time.Duration(10 * time.Second)
	client := http.Client{Timeout: timeout}
	res, err := client.Get(uri)
	if err != nil {
		return body, err
	}

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
	}
	return
}

// GoodReads implements RemoteLibraryAPI and retrieves information from goodreads.com.
type GoodReads struct {
}

const apiRoot = "https://www.goodreads.com/"

// response is the top xml element in goodreads response.
type response struct {
	Book   Metadata      `xml:"book"`
	Search searchResults `xml:"search"`
}

// searchResults is the main xml element in goodreads search.
type searchResults struct {
	ResultsNumber string `xml:"total-results"`
	Works         []work `xml:"results>work"`
}

// works holds the work information in the xml response.
type work struct {
	ID     string `xml:"best_book>id"`
	Author string `xml:"best_book>author>name"`
	Title  string `xml:"best_book>title"`
}

// GetBook returns a GoodreadsBook from its Goodreads ID
func (g GoodReads) GetBook(id, key string) (Metadata, error) {
	uri := apiRoot + "book/show/" + id + ".xml?key=" + key
	r := response{}
	err := getXMLData(uri, &r)
	return r.Book, err
}

func makeSearchQuery(parts ...string) (query string) {
	query = strings.Join(parts, "+")
	r := strings.NewReplacer(" ", "+")
	return html.EscapeString(r.Replace(query))
}

// GetBookIDByQuery gets a Goodreads ID from a query
func (g GoodReads) GetBookIDByQuery(author, title, key string) (id string, err error) {
	uri := apiRoot + "search/index.xml?key=" + key + "&q=" + makeSearchQuery(author, title)
	r := response{}
	err = getXMLData(uri, &r)
	if err != nil {
		return
	}
	// parsing results
	numberOfHits, err := strconv.Atoi(r.Search.ResultsNumber)
	if err != nil {
		return
	}
	if numberOfHits != 0 {
		// TODO: if more than 1 hit, give the user a choice, as in beets import.
		for _, work := range r.Search.Works {
			if work.Author == author && work.Title == title {
				return work.ID, nil
			}
		}
		fmt.Println("Could not find exact match, returning first hit.")
		return r.Search.Works[0].ID, nil
	}
	return
}

// GetBookIDByISBN gets a Goodreads ID from an ISBN
func (g GoodReads) GetBookIDByISBN(isbn, key string) (id string, err error) {
	uri := apiRoot + "search/index.xml?key=" + key + "&q=" + isbn
	r := response{}
	err = getXMLData(uri, &r)
	if err != nil {
		return
	}
	// parsing results
	numberOfHits, err := strconv.Atoi(r.Search.ResultsNumber)
	if err != nil {
		return
	}
	if numberOfHits != 0 {
		id = r.Search.Works[0].ID
		if numberOfHits > 1 {
			fmt.Println("Got more than 1 hit while searching by ISBN! Returned first hit.")
		}
	}
	return
}
