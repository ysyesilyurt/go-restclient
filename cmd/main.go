package main

import (
	"github.com/pkg/errors"
	"go-restclient/restclient"
	"log"
	"net/http"
	"net/url"
	"time"
)

type dummyHttpResponse struct {
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"data"`
}

type dummyBasicAuthenticator struct {
	Username, Password string
}

func NewTestBasicAuthenticator(username, password string) restclient.Authenticator {
	return &dummyBasicAuthenticator{
		Username: username,
		Password: password,
	}
}

func (b dummyBasicAuthenticator) Apply(request *http.Request) error {
	request.SetBasicAuth(b.Username, b.Password)
	return nil
}

func usage1() error {
	var response dummyHttpResponse
	headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	queryParams := url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
	ri := restclient.NewRequestInfo("https", "ysyesilyurt.com", []string{"tasks", "1"}, &queryParams, &headers, nil)
	req, err := restclient.NewRequest(ri)
	if err != nil {
		return errors.Wrap(err, "Failed to construct http.Request out of RequestInfo")
	}

	client := restclient.NewHttpClient(true, 30*time.Second)
	cri := restclient.NewDoRequestInfo(req, nil, &response)
	err = client.Get(cri)
	if err != nil {
		return errors.Wrap(err, "Failed to do HTTP GET request")
	}
	return nil
}

func usage2() error {
	var response dummyHttpResponse
	headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	ri, err := restclient.NewRequestInfoFromRawURL("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", &headers, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to construct RequestInfo out of Raw URL")
	}

	req, err := restclient.NewRequest(ri)
	if err != nil {
		return errors.Wrap(err, "Failed to construct http.Request out of RequestInfo")
	}

	client := restclient.NewHttpClient(true, 0)
	auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
	cri := restclient.NewDoRequestInfoWithTimeout(req, auth, &response, 15*time.Second)
	err = client.Get(cri)
	if err != nil {
		return errors.Wrap(err, "Failed to do HTTP GET request")
	}
	return nil
}

func usage3() error {
	var response dummyHttpResponse
	headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	ri, err := restclient.NewRequestInfoFromRawURL("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", &headers, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to construct RequestInfo out of Raw URL")
	}

	auth := NewTestBasicAuthenticator("ysyesilyurt", "0123")
	err = restclient.PerformGetRequest(ri, auth, &response, true, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "Failed to perform HTTP GET request")
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
