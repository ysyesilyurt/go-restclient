package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RequestInfo struct {
	Scheme       string       // Scheme e.g. http
	Host         string       // Host e.g. ysyesilyurt.com
	PathElements []string     // PathElements represents each component in the path that is separated by a slash (/) e.g. ['posts', '1']
	Headers      *http.Header // Headers e.g {"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	Body         interface{}  // Body represents the to-be-marshalled RequestBody variable
	QueryParams  *url.Values  // QueryParams e.g {"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
}

func NewRequestInfo(scheme, host string, pathElements []string, queryParams *url.Values, headers *http.Header, body interface{}) RequestInfo {
	return RequestInfo{
		Scheme:       scheme,
		Host:         host,
		PathElements: pathElements,
		Headers:      headers,
		Body:         body,
		QueryParams:  queryParams,
	}
}

/* NewRequestInfoFromRawURL parses given rawURL and returns a RequestInfo out of it */
func NewRequestInfoFromRawURL(rawURL string, headers *http.Header, body interface{}) (RequestInfo, error) {
	parsedUrl, err := url.Parse(rawURL)
	if err != nil {
		return RequestInfo{}, err
	}
	return RequestInfo{
		Scheme:       parsedUrl.Scheme,
		Host:         parsedUrl.Host,
		PathElements: strings.Split(parsedUrl.Path, "/")[1:],
		Headers:      headers,
		Body:         body,
		QueryParams: func() *url.Values {
			v := parsedUrl.Query()
			return &v
		}(),
	}, nil
}

/* NewRequest builds a http.Request object using values provided inside NewRequestInfo
and returns a pointer to it. Note that it does not accept any request method. */
func NewRequest(ri RequestInfo) (*http.Request, error) {
	// Construct URL by escaping components
	escapedURLString := buildEndpoint(ri.Scheme, ri.Host, ri.PathElements)

	// Marshal RequestBody if exists
	var bodyReader io.Reader
	if ri.Body != nil {
		marshalled, err := json.Marshal(ri.Body)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to marshal request body")
		}
		bodyReader = bytes.NewReader(marshalled)
	}

	// Build the request object
	req, err := http.NewRequest("", escapedURLString, bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to construct http.Request")
	}
	req.Close = true

	// Validate that resulting URL path in request is valid
	if len(ri.PathElements) != 0 {
		if _, err = url.ParseRequestURI(req.URL.Path); err != nil {
			return nil, errors.Wrap(err, "Invalid Request URI")
		}
	}

	// Set queryParams if exists
	if ri.QueryParams != nil {
		req.URL.RawQuery = ri.QueryParams.Encode()
	}

	setHeaderIfExists := func(key string, value []string) {
		if key != "" && len(value) != 0 {
			for _, v := range value {
				req.Header.Add(key, v)
			}
		}
	}

	// Set custom headers
	if ri.Headers != nil {
		for hName, hValue := range *ri.Headers {
			setHeaderIfExists(hName, hValue)
		}
	}

	return req, err
}

/* buildEndpoint performs proper URL Escaping on path params and delivers the safe formatted endpoint string */
func buildEndpoint(scheme, host string, pathElements []string) string {
	urlFormat := strings.Builder{}
	sanitizedUrlElements := make([]interface{}, len(pathElements))
	escapedSchemeHost := fmt.Sprintf("%s://%s", url.PathEscape(scheme), url.PathEscape(host))
	urlFormat.WriteString(escapedSchemeHost)
	for i, ue := range pathElements {
		sanitizedUrlElements[i] = url.PathEscape(ue)
		urlFormat.WriteString("/%s")
	}
	return fmt.Sprintf(urlFormat.String(), sanitizedUrlElements...)
}
