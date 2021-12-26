package restclient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const defaultTimeoutDuration = 60 * time.Second

/* HttpRequest is exported request object that contains all the necessary things to perform an HttpRequest,
can be created using HttpRequestBuilder  */
type HttpRequest struct {
	request        *http.Request // internal http.Request object
	auth           Authenticator // Custom Authentication Strategy to apply to the request
	respReference  interface{}   // Object reference to map the response of the request
	timeout        time.Duration // timeout value to be used for the request
	loggingEnabled bool          // log the result of the request if loggingEnabled
}

func newHttpClient(timeout time.Duration) *http.Client {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: tr,
		Timeout: func() time.Duration {
			if timeout < 0 {
				return defaultTimeoutDuration
			}
			return timeout
		}(),
	}
	return client
}

/* YieldRequest returns the underlying *http.Request object */
func (hr HttpRequest) YieldRequest() *http.Request {
	return hr.request
}

/* Get performs an HTTP GET request using the provided HttpRequest fields. Applies HttpRequest.auth directly to the resulting
request on it (nil auth means no auth). Decodes any response into HttpRequest.respReference. Also uses HttpRequest.timeout value
as the request timeout value, Zero (0) means no timeout. Returns a RequestError implying the result of the call */
func (hr HttpRequest) Get() RequestError {
	return doRequest(hr.request, http.MethodGet, hr.auth, hr.respReference, hr.loggingEnabled, hr.timeout)
}

/* Post performs an HTTP GET request using the provided HttpRequest fields. Applies HttpRequest.auth directly to the resulting
request on it (nil auth means no auth). Decodes any response into HttpRequest.respReference. Also uses HttpRequest.timeout value
as the request timeout value, Zero (0) means no timeout. Returns a RequestError implying the result of the call */
func (hr HttpRequest) Post() RequestError {
	return doRequest(hr.request, http.MethodPost, hr.auth, hr.respReference, hr.loggingEnabled, hr.timeout)
}

/* Put performs an HTTP GET request using the provided HttpRequest fields. Applies HttpRequest.auth directly to the resulting
request on it (nil auth means no auth). Decodes any response into HttpRequest.respReference. Also uses HttpRequest.timeout value
as the request timeout value, Zero (0) means no timeout. Returns a RequestError implying the result of the call */
func (hr HttpRequest) Put() RequestError {
	return doRequest(hr.request, http.MethodPut, hr.auth, hr.respReference, hr.loggingEnabled, hr.timeout)
}

/* Patch performs an HTTP GET request using the provided HttpRequest fields. Applies HttpRequest.auth directly to the resulting
request on it (nil auth means no auth). Decodes any response into HttpRequest.respReference. Also uses HttpRequest.timeout value
as the request timeout value, Zero (0) means no timeout. Returns a RequestError implying the result of the call */
func (hr HttpRequest) Patch() RequestError {
	return doRequest(hr.request, http.MethodPatch, hr.auth, hr.respReference, hr.loggingEnabled, hr.timeout)
}

/* Delete performs an HTTP GET request using the provided HttpRequest fields. Applies HttpRequest.auth directly to the resulting
request on it (nil auth means no auth). Decodes any response into HttpRequest.respReference. Also uses HttpRequest.timeout value
as the request timeout value, Zero (0) means no timeout. Returns a RequestError implying the result of the call */
func (hr HttpRequest) Delete() RequestError {
	return doRequest(hr.request, http.MethodDelete, hr.auth, hr.respReference, hr.loggingEnabled, hr.timeout)
}

func doRequest(req *http.Request, method string, auth Authenticator, respRef interface{}, loggingEnabled bool, timeout time.Duration) RequestError {

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
			return NewRequestBuildError(InvalidRequestErr, errors.Wrap(err, "cannot apply authentication information to request"))
		}
	}

	// Setup HttpClient
	httpClient := newHttpClient(timeout)
	doRequestAndTimeIfEnabled := func() (*http.Response, int64, error) {
		var err error
		var duration int64
		var resp *http.Response

		if loggingEnabled {
			startTime := time.Now()
			resp, err = httpClient.Do(req)
			duration = int64(time.Since(startTime) / time.Millisecond)
		} else {
			resp, err = httpClient.Do(req)
		}
		return resp, duration, err
	}

	logRequestIfEnabled := func(statusCode int, duration int64, err error) {
		if loggingEnabled {
			logMsg := fmt.Sprintf("[status]: %d [duration-ms]: %d [url]: %s", statusCode, duration, req.URL.String())
			if statusCode == 0 {
				errorLogger.Printf("Request failed, %s, [err]: %v", logMsg, err)
				return
			}
			infoLogger.Printf("Request finished %s", logMsg)
		}
	}

	// Do Request (Time and Log it if enabled)
	resp, duration, err := doRequestAndTimeIfEnabled()
	if err != nil {
		logRequestIfEnabled(0, duration, err)
		urlError := err.(*url.Error)
		if urlError.Timeout() {
			return NewRequestTimeoutError(HttpClientErr, errors.Wrap(err, "Connection Error, Request Timed out"))
		}
		return NewRequestConnectionError(HttpClientErr, errors.Wrap(err, "Connection Error"))
	}
	logRequestIfEnabled(resp.StatusCode, duration, nil)
	defer func() {
		errBodyClose := resp.Body.Close()
		if errBodyClose != nil {
			if err == nil {
				err = errors.Wrap(errBodyClose, "Failed to close response body")
			} else {
				errBodyClose = errors.Wrap(errBodyClose, "Failed to close response body")
				err = errors.Wrap(err, errBodyClose.Error())
			}
		}
	}()

	// Handle Response Status Code
	reqErr := prepareResponseError(resp)
	if reqErr != nil {
		return reqErr
	}

	// Read the body into respRef
	if respRef != nil {
		err = unmarshalResponseBody(resp, respRef)
		if err != nil {
			return NewRequestResponseParseError(InvalidRequestErr,
				errors.Wrapf(err, "Failed to decode response body into given responseRef %T variable", respRef))
		}
	}
	return nil
}

func prepareResponseError(response *http.Response) RequestError {
	if response.StatusCode < 400 {
		return nil
	}

	var topLevelErr error
	responseMessage, err := getFailedResponseBody(response)
	if err != nil {
		return NewRequestResponseParseError(InvalidResponseBodyErr, errors.Wrap(err, "Failed to read response body"))
	}
	switch response.StatusCode {
	case http.StatusUnauthorized:
		topLevelErr = UnauthorizedErr
	case http.StatusForbidden:
		topLevelErr = ForbiddenErr
	case http.StatusNotFound:
		topLevelErr = RecordNotFoundErr
	case http.StatusBadRequest:
		topLevelErr = ParseErr
	case http.StatusTooManyRequests:
		topLevelErr = TooManyRequestErr
	case http.StatusUnprocessableEntity:
		topLevelErr = UnprocessableEntityErr
	case http.StatusInternalServerError:
		topLevelErr = InternalServerErr
	case http.StatusServiceUnavailable:
		topLevelErr = ServiceUnavailableErr
	default:
		topLevelErr = UnexpectedResponseCodeErr
	}
	return NewRequestError(topLevelErr, errors.New(responseMessage), response.StatusCode)
}

func getFailedResponseBody(response *http.Response) (string, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "Failed to convert response body to error")
	}
	return string(body), nil
}

func readerToByte(reader io.Reader) ([]byte, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func unmarshalResponseBody(response *http.Response, v interface{}) error {
	return unmarshalReader(response.Body, v)
}

func unmarshalRequestBody(request *http.Request, v interface{}) error {
	return unmarshalReader(request.Body, v)
}

func unmarshalReader(r io.Reader, v interface{}) error {
	toByte, err := readerToByte(r)
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
