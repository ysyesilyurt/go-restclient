package main

import (
	"github.com/pkg/errors"
	"github.com/ysyesilyurt/go-restclient/restclient"
	"log"
	"net/http"
	"net/url"
	"time"
)

type dummyBodyDto struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type dummyHttpResponse struct {
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
}

type dummyBasicAuthenticator struct {
	Username, Password string
}

func newTestBasicAuthenticator(username, password string) restclient.Authenticator {
	return &dummyBasicAuthenticator{
		Username: username,
		Password: password,
	}
}

func (b dummyBasicAuthenticator) Apply(request *http.Request) error {
	request.SetBasicAuth(b.Username, b.Password)
	return nil
}

/* Example HTTP GET request with providing individual Scheme, Host, PathElements etc. fields */
func usage1() error {
	var response dummyHttpResponse
	headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	queryParams := url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
	testAuth := newTestBasicAuthenticator("ysyesilyurt", "0123")

	// https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1
	req, reqErr := restclient.RequestBuilder().
		Scheme("https").
		Host("ysyesilyurt.com").
		PathElements([]string{"tasks", "1"}).
		QueryParams(&queryParams).
		Header(&headers).
		Auth(testAuth).
		ResponseReference(&response).
		LoggingEnabled(true).
		Timeout(30 * time.Second).
		Build()

	if reqErr != nil {
		return errors.Wrap(reqErr, "Failed to construct HTTP request")
	}

	reqErr = req.Get()
	if reqErr != nil {
		return errors.Wrap(reqErr, "Failed to perform HTTP GET request")
	}
	return nil
}

/* Example HTTP POST request with requestBody and using defaults of some fields */
func usage2() error {
	var response dummyHttpResponse
	requestBody := dummyBodyDto{
		Id:   123,
		Name: "1234",
	}
	headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	queryParams := url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
	testAuth := newTestBasicAuthenticator("ysyesilyurt", "0123")

	// Using following fields' defaults: LoggingEnabled (false), Timeout (60 secs)
	// https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1
	req, reqErr := restclient.RequestBuilder().
		Scheme("https").
		Host("ysyesilyurt.com").
		PathElements([]string{"tasks", "1"}).
		QueryParams(&queryParams).
		BodyJson(requestBody).
		Header(&headers).
		Auth(testAuth).
		ResponseReference(&response).
		Build()

	if reqErr != nil {
		return errors.Wrap(reqErr, "Failed to construct HTTP request")
	}

	reqErr = req.Post()
	if reqErr != nil {
		return errors.Wrap(reqErr, "Failed to perform HTTP POST request")
	}
	return nil
}

/* Example HTTP GET request with RawUrl */
func usage3() error {
	var response dummyHttpResponse
	rawUrl := "https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1"
	headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	testAuth := newTestBasicAuthenticator("ysyesilyurt", "0123")

	req, reqErr := restclient.RequestBuilder().
		RawUrl(rawUrl).
		Header(&headers).
		Auth(testAuth).
		ResponseReference(&response).
		LoggingEnabled(true).
		Timeout(30 * time.Second).
		Build()

	if reqErr != nil {
		return errors.Wrap(reqErr, "Failed to construct HTTP request")
	}

	reqErr = req.Get()
	if reqErr != nil {
		return errors.Wrap(reqErr, "Failed to perform HTTP GET request")
	}
	return nil
}

func main() {
	// Below usages will fail with `no such host` when run
	err := usage1()
	if err != nil {
		log.Printf("Example usage1 failed, reason: %v", err)
	}

	err = usage2()
	if err != nil {
		log.Printf("Example usage2 failed, reason: %v", err)
	}

	err = usage3()
	if err != nil {
		log.Printf("Example usage3 failed, reason: %v", err)
	}
}
