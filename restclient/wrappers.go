package restclient

import (
	"github.com/pkg/errors"
	"net/http"
	"time"
)

/* One-liner wrappers around HttpClient its methods. */

/* PerformGetRequest creates a http.Request and a HttpClient with given timeout value,
then performs a HTTP GET request using provided Authenticator. Decodes any response to responseRef */
func PerformGetRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error {
	req, client, err := NewRequestAndClient(ri, loggingEnabled, timeout)
	if err != nil {
		return errors.Wrap(err, "Could not create request and client")
	}
	cri := NewDoRequestInfo(req, auth, &responseRef)
	return client.Get(cri)
}

/* PerformPostRequest creates a http.Request and a HttpClient with given timeout value,
then performs a HTTP POST request using provided Authenticator. Decodes any response to responseRef */
func PerformPostRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error {
	req, client, err := NewRequestAndClient(ri, loggingEnabled, timeout)
	if err != nil {
		return errors.Wrap(err, "Could not create request and client")
	}
	cri := NewDoRequestInfo(req, auth, &responseRef)
	return client.Post(cri)
}

/* PerformPutRequest creates a http.Request and a HttpClient with given timeout value,
then performs a HTTP PUT request using provided Authenticator. Decodes any response to responseRef */
func PerformPutRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error {
	req, client, err := NewRequestAndClient(ri, loggingEnabled, timeout)
	if err != nil {
		return errors.Wrap(err, "Could not create request and client")
	}
	cri := NewDoRequestInfo(req, auth, &responseRef)
	return client.Put(cri)
}

/* PerformPatchRequest creates a http.Request and a HttpClient with given timeout value,
then performs a HTTP PATCH request using provided Authenticator. Decodes any response to responseRef */
func PerformPatchRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error {
	req, client, err := NewRequestAndClient(ri, loggingEnabled, timeout)
	if err != nil {
		return errors.Wrap(err, "Could not create request and client")
	}
	cri := NewDoRequestInfo(req, auth, &responseRef)
	return client.Patch(cri)
}

/* PerformDeleteRequest creates a http.Request and a HttpClient with given timeout value,
then performs a HTTP DELETE request using provided Authenticator. Decodes any response to responseRef */
func PerformDeleteRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error {
	req, client, err := NewRequestAndClient(ri, loggingEnabled, timeout)
	if err != nil {
		return errors.Wrap(err, "Could not create request and client")
	}
	cri := NewDoRequestInfo(req, auth, &responseRef)
	return client.Delete(cri)
}

/* NewRequestAndClient creates a http.Request and a HttpClient using provided RequestInfo, loggingEnabled and timeout values */
func NewRequestAndClient(ri RequestInfo, loggingEnabled bool, timeout time.Duration) (*http.Request, HttpClient, error) {
	req, err := NewRequest(ri)
	if err != nil {
		return nil, HttpClient{}, errors.Wrap(err, "Could not create request")
	}
	client := NewHttpClient(loggingEnabled, timeout)
	return req, client, nil
}
