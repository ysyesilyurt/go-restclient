package restclient

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

var (
	InvalidRequestErr         = errors.New("Invalid request error")
	InvalidResponseBodyErr    = errors.New("Invalid response body error")
	UnexpectedResponseCodeErr = errors.New("Unexpected HTTP response code")
	HttpClientErr             = errors.New("Http client error")
	UnauthorizedErr           = errors.New("Unauthorized - Authentication failed")
	ForbiddenErr              = errors.New("Resource is forbidden, check your authentication token and permissions")
	RecordNotFoundErr         = errors.New("Resource is not found")
	ParseErr                  = errors.New("Not well-formatted request or missing fields")
	TooManyRequestErr         = errors.New("Too many requests - Resource unavailable")
	UnprocessableEntityErr    = errors.New("Syntactically correct but semantically incorrect request")
	InternalServerErr         = errors.New("Internal server error")
	ServiceUnavailableErr     = errors.New("Service unavailable")
)

type RequestError interface {
	Error() string
	GetTopLevelError() error   // GetTopLevelError returns top level error that originates the title
	GetUnderlyingError() error // GetUnderlyingError returns underlying error that originates the message
	GetTitle() string          // GetTitle returns top level error 'title' referring to message
	GetMessage() string        // GetMessage returns detailed error message for the request
	GetStatusCode() int        // GetStatusCode returns status code for the request. '0' means request failed for some reason and can be checked using methods below
	Timeout() bool             // Timeout returns if request failed due to a timeout
	ConnectionError() bool     // ConnectionError returns if request failed due to a connection error (Failed to get response for some reason)
	ResponseParseError() bool  // ResponseParseError returns if response of the request could not be parsed into given response reference variable
	RequestBuildError() bool   // RequestBuildError returns if request could not be built due to some reason
}

type requestErrorImpl struct {
	topLevelErr, err                                                  error
	statusCode                                                        int
	isTimeout, isConnectionErr, isResponseParseErr, isRequestBuildErr bool
}

func (r requestErrorImpl) GetTopLevelError() error {
	return r.topLevelErr
}

func (r requestErrorImpl) GetUnderlyingError() error {
	return r.err
}

func (r requestErrorImpl) GetTitle() string {
	return r.topLevelErr.Error()
}

func (r requestErrorImpl) GetMessage() string {
	return r.err.Error()
}

func (r requestErrorImpl) GetStatusCode() int {
	return r.statusCode
}

func (r requestErrorImpl) Timeout() bool {
	return r.isTimeout
}

func (r requestErrorImpl) ConnectionError() bool {
	return r.isConnectionErr
}

func (r requestErrorImpl) ResponseParseError() bool {
	return r.isResponseParseErr
}

func (r requestErrorImpl) RequestBuildError() bool {
	return r.isRequestBuildErr
}

func (r requestErrorImpl) Error() string {
	return fmt.Sprintf("%s - %s - Status Code: %d", r.GetTitle(), r.GetMessage(), r.GetStatusCode())
}

func NewRequestError(topLevelErr, err error, statusCode int) RequestError {
	return &requestErrorImpl{
		topLevelErr: topLevelErr,
		err:         err,
		statusCode:  statusCode,
	}
}

func NewRequestTimeoutError(topLevelErr, err error) RequestError {
	return &requestErrorImpl{
		topLevelErr:     topLevelErr,
		err:             err,
		statusCode:      http.StatusRequestTimeout,
		isTimeout:       true,
		isConnectionErr: true,
	}
}

func NewRequestConnectionError(topLevelErr, err error) RequestError {
	return &requestErrorImpl{
		topLevelErr:     topLevelErr,
		err:             err,
		isConnectionErr: true,
	}
}

func NewRequestBuildError(topLevelErr, err error) RequestError {
	return &requestErrorImpl{
		topLevelErr:       topLevelErr,
		err:               err,
		isRequestBuildErr: true,
	}
}

func NewRequestResponseParseError(topLevelErr, err error) RequestError {
	return &requestErrorImpl{
		topLevelErr:        topLevelErr,
		err:                err,
		isResponseParseErr: true,
	}
}
