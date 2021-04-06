package restclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const DefaultTimeoutDuration = 60 * time.Second

var (
	unauthorizedErr        = errors.New("Unauthorized - Authentication failed")
	forbiddenErr           = errors.New("Resource is forbidden, check your authentication token and permissions")
	recordNotFoundErr      = errors.New("Resource is not found")
	badRequestErr          = errors.New("Not well-formatted request or missing fields")
	tooManyRequestErr      = errors.New("Too many requests - Resource unavailable")
	unprocessableEntityErr = errors.New("Syntactically correct but semantically incorrect request")
	internalServerErr      = errors.New("Internal server error")
	serviceUnavailableErr  = errors.New("Service unavailable")
)

type HttpClient struct {
	client         *http.Client
	loggingEnabled bool
	timeout        time.Duration
}

func NewHttpClient(loggingEnabled bool, timeout time.Duration) HttpClient {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: tr,
		Timeout: func() time.Duration {
			if timeout <= 0 {
				return DefaultTimeoutDuration
			}
			return timeout
		}(),
	}
	return HttpClient{client, loggingEnabled, timeout}
}

type DoRequestInfo struct {
	request        *http.Request
	auth           Authenticator
	respRef        interface{}
	requestTimeout time.Duration
}

func NewDoRequestInfo(request *http.Request, auth Authenticator, responseReference interface{}) DoRequestInfo {
	return DoRequestInfo{
		request: request,
		auth:    auth,
		respRef: responseReference,
	}
}

/* NewDoRequestInfoWithTimeout creates a DoRequestInfo with given requestTimeout. Should be used whenever a specific
timeout value is required for individual request. Will override HttpClient timeout value when smaller, otherwise will have no effect. */
func NewDoRequestInfoWithTimeout(request *http.Request, auth Authenticator, responseReference interface{}, requestTimeout time.Duration) DoRequestInfo {
	return DoRequestInfo{
		request:        request,
		auth:           auth,
		respRef:        responseReference,
		requestTimeout: requestTimeout,
	}
}

/* Get performs an HTTP GET request using the provided dri.request after applying dri.auth on it (nil auth means no auth). Decodes any response into dri.respRef.
Request specific timeout can be set using dri.requestTimeout and will be used if HttpClient timeout value is longer than given timeout value. Zero (0) means no timeout. */
func (hc HttpClient) Get(dri DoRequestInfo) error {
	return hc.do(dri.request, http.MethodGet, dri.auth, dri.respRef, dri.requestTimeout)
}

/* Post performs an HTTP POST request using the provided dri.request after applying dri.auth on it (nil auth means no auth). Decodes any response into dri.respRef.
Request specific timeout can be set using dri.requestTimeout and will be used if HttpClient timeout value is longer than given timeout value. Zero (0) means no timeout. */
func (hc HttpClient) Post(dri DoRequestInfo) error {
	return hc.do(dri.request, http.MethodPost, dri.auth, dri.respRef, dri.requestTimeout)
}

/* Put performs an HTTP PUT request using the provided dri.request after applying dri.auth on it (nil auth means no auth). Decodes any response into dri.respRef.
Request specific timeout can be set using dri.requestTimeout and will be used if HttpClient timeout value is longer than given timeout value. Zero (0) means no timeout. */
func (hc HttpClient) Put(dri DoRequestInfo) error {
	return hc.do(dri.request, http.MethodPut, dri.auth, dri.respRef, dri.requestTimeout)
}

/* Patch performs an HTTP PATCH request using the provided dri.request after applying dri.auth on it (nil auth means no auth). Decodes any response into dri.respRef.
Request specific timeout can be set using dri.requestTimeout and will be used if HttpClient timeout value is longer than given timeout value. Zero (0) means no timeout. */
func (hc HttpClient) Patch(dri DoRequestInfo) error {
	return hc.do(dri.request, http.MethodPatch, dri.auth, dri.respRef, dri.requestTimeout)
}

/* Delete performs an HTTP DELETE request using the provided dri.request after applying dri.auth on it (nil auth means no auth). Decodes any response into dri.respRef.
Request specific timeout can be set using dri.requestTimeout and will be used if HttpClient timeout value is longer than given timeout value. Zero (0) means no timeout. */
func (hc HttpClient) Delete(dri DoRequestInfo) error {
	return hc.do(dri.request, http.MethodDelete, dri.auth, dri.respRef, dri.requestTimeout)
}

func (hc HttpClient) do(req *http.Request, method string, auth Authenticator, respRef interface{}, timeout time.Duration) error {

	setHeaderIfNotSetAlready := func(key, value string) {
		if req.Header.Get(key) == "" && value != "" {
			req.Header.Set(key, value)
		}
	}

	// Set universal headers
	setHeaderIfNotSetAlready("Accept", "application/json")
	req.Method = method
	switch method {
	case http.MethodPut, http.MethodPatch, http.MethodPost:
		setHeaderIfNotSetAlready("Content-Type", "application/json")
	}

	// Set Authorization header by applying specified authenticator's strategy if exists
	if auth != nil {
		err := auth.Apply(req)
		if err != nil {
			return errors.Wrap(err, "cannot apply authentication information to request")
		}
	}

	// Set context timeout and defer its cancellation if a proper timeout value is specified
	if timeout > 0 {
		if hc.timeout <= 0 || timeout < hc.timeout {
			ctx := req.Context()
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			req = req.Clone(ctx)
		} else {
			infoLogger.Printf("Given request timeout duration (%s) is longer than client timeout (%s)."+
				" It will have no effect, client timeout will be used.", timeout, hc.timeout)
		}
	}

	doRequestAndTimeIfEnabled := func() (*http.Response, int64, error) {
		var err error
		var duration int64
		var resp *http.Response

		if hc.loggingEnabled {
			startTime := time.Now()
			resp, err = hc.client.Do(req)
			duration = int64(time.Since(startTime) / time.Millisecond)
		} else {
			resp, err = hc.client.Do(req)
		}
		return resp, duration, err
	}

	logRequestIfEnabled := func(statusCode int, duration int64, err error) {
		if hc.loggingEnabled {
			if statusCode == 0 {
				errorLogger.Printf("Request failed, [duration_ms]: %d [reason]: %s", duration, err.Error())
			} else {
				infoLogger.Printf("Request completed, [status_code]: %d [duration_ms]: %d", statusCode, duration)
			}
		}
	}

	// Do Request (Time and Log it if enabled)
	resp, duration, err := doRequestAndTimeIfEnabled()
	if err != nil {
		logRequestIfEnabled(0, duration, err)
		return errors.Wrap(err, "Connection Error")
	}
	logRequestIfEnabled(resp.StatusCode, duration, nil)
	defer func() {
		errBodyClose := resp.Body.Close()
		if errBodyClose != nil {
			if err == nil {
				err = errors.Wrap(errBodyClose, "Failed to close response body")
			} else {
				errorLogger.Printf("Failed to close response body, Reason: %s", errBodyClose.Error())
			}
		}
	}()

	// Handle Response Status Code
	err = PrepareResponseError(resp)
	if err != nil {
		return err
	}


	// Read the body into respRef
	if respRef != nil {
		err = UnmarshalResponseBody(resp, respRef)
		if err != nil {
			return errors.Wrap(err, "Failed to decode response body into responseRef")
		}
	}
	return nil
}

func ReaderToByte(reader io.Reader) ([]byte, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func UnmarshalResponseBody(response *http.Response, v interface{}) error {
	return UnmarshalReader(response.Body, v)
}

func UnmarshalRequestBody(request *http.Request, v interface{}) error {
	return UnmarshalReader(request.Body, v)
}

func UnmarshalReader(r io.Reader, v interface{}) error {
	toByte, err := ReaderToByte(r)
	if err != nil {
		return errors.Wrap(err, "Failed to read body")
	}
	// Unmarshal into v if v is not a []byte o/w directly assign v to []byte
	if _, ok := v.([]byte); !ok {
		err = json.Unmarshal(toByte, v)
		if err != nil {
			return errors.Wrapf(err, "Failed unmarshal body")
		}
	} else {
		v = toByte
	}
	return nil
}

func PrepareResponseError(response *http.Response) error {
	if response.StatusCode < 400 {
		return nil
	}
	responseMessage, err := getFailedResponseBody(response)
	if err != nil {
		return errors.Wrapf(err, "could not read failed response's body, response code: %d", response.StatusCode)
	}
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return errors.Wrap(unauthorizedErr, responseMessage)
	case http.StatusForbidden:
		return errors.Wrap(forbiddenErr, responseMessage)
	case http.StatusNotFound:
		return errors.Wrap(recordNotFoundErr, responseMessage)
	case http.StatusBadRequest:
		return errors.Wrap(badRequestErr, responseMessage)
	case http.StatusTooManyRequests:
		return errors.Wrap(tooManyRequestErr, responseMessage)
	case http.StatusUnprocessableEntity:
		return errors.Wrap(unprocessableEntityErr, responseMessage)
	case http.StatusInternalServerError:
		return errors.Wrap(internalServerErr, responseMessage)
	case http.StatusServiceUnavailable:
		return errors.Wrap(serviceUnavailableErr, responseMessage)
	}
	return errors.Wrap(errors.Errorf("Unhandled HTTP response code %d", response.StatusCode), responseMessage)
}

func getFailedResponseBody(response *http.Response) (string, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "Failed to convert response body to error")
	}
	return string(body), nil
}
