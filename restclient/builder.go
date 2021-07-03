package restclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpRequestBuilder struct {
	hr HttpRequest
	ri requestInfo
}

/* RequestBuilder builds a HttpRequest using the methods define to fill the fields for HttpRequest.
There exists default values for some fields:
- Timeout: 60 Seconds (defaultTimeoutDuration)
- LoggingEnabled: false
*/
func RequestBuilder() HttpRequestBuilder {
	return HttpRequestBuilder{
		hr: HttpRequest{timeout: defaultTimeoutDuration, loggingEnabled: false},
		ri: requestInfo{},
	}
}

/* requestInfo is an internal type to ease things with HttpRequestBuilder */
type requestInfo struct {
	scheme       string       // scheme e.g. http
	host         string       // host e.g. example.com
	pathElements []string     // pathElements represents each component in the path that is separated by a slash (/) e.g. ['posts', '1']
	header       *http.Header // header e.g {"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
	body         io.Reader    // body represents RequestBody
	queryParams  *url.Values  // queryParams e.g {"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
}

func (hrb HttpRequestBuilder) Scheme(scheme string) HttpRequestBuilder {
	hrb.ri.scheme = scheme
	return hrb
}

func (hrb HttpRequestBuilder) Host(host string) HttpRequestBuilder {
	hrb.ri.host = host
	return hrb
}

func (hrb HttpRequestBuilder) PathElements(pe []string) HttpRequestBuilder {
	hrb.ri.pathElements = pe
	return hrb
}

func (hrb HttpRequestBuilder) QueryParams(qp *url.Values) HttpRequestBuilder {
	hrb.ri.queryParams = qp
	return hrb
}

func (hrb HttpRequestBuilder) RawUrl(rawUrl string) HttpRequestBuilder {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		errorLogger.Printf("Failed to Parse Given Raw URL! Leaving fields that is filled with Raw URL empty.. %v", err)
		return hrb
	}
	hrb.ri.scheme = parsedUrl.Scheme
	hrb.ri.host = parsedUrl.Host
	hrb.ri.pathElements = strings.Split(parsedUrl.Path, "/")[1:]
	queryParams := parsedUrl.Query()
	hrb.ri.queryParams = &queryParams
	return hrb
}

func (hrb HttpRequestBuilder) Header(header *http.Header) HttpRequestBuilder {
	hrb.ri.header = header
	return hrb
}

/* HttpRequestBuilder.Body is the direct io.Reader RequestBody to be applied to the request */
func (hrb HttpRequestBuilder) Body(body io.Reader) HttpRequestBuilder {
	hrb.ri.body = body
	return hrb
}

/* HttpRequestBuilder.BodyJson represents the unprocessed (to-be-marshalled) RequestBody which is going to be sent in Json form.
Use this builder method if RequestBody will be sent in Json form. This method accepts RequestBody directly in its raw form
(no need for any marshalling or io.Reader conversion ops) */
func (hrb HttpRequestBuilder) BodyJson(bodyJson interface{}) HttpRequestBuilder {
	if bodyJson != nil {
		marshalled, err := json.Marshal(bodyJson)
		if err != nil {
			errorLogger.Printf("Failed to marshal request body! Leaving request body empty.. %v", err)
			return hrb
		}
		hrb.ri.body = bytes.NewReader(marshalled)
	}
	return hrb
}

/* HttpRequestBuilder.Auth sets the restclient.Authenticator for the request. Implement restclient.Authenticator
to use custom authentication strategies */
func (hrb HttpRequestBuilder) Auth(auth Authenticator) HttpRequestBuilder {
	hrb.hr.auth = auth
	return hrb
}

func (hrb HttpRequestBuilder) ResponseReference(respRef interface{}) HttpRequestBuilder {
	hrb.hr.respReference = respRef
	return hrb
}

/* HttpRequestBuilder.Request provides directly sets internal http.Request with the provided one. Use this if you
consider using this builder with a pre-prepared http.Request object */
func (hrb HttpRequestBuilder) Request(req *http.Request) HttpRequestBuilder {
	hrb.hr.request = req
	return hrb
}

/* HttpRequestBuilder.Timeout sets timeout value to be used for the response. Default is 60 (defaultTimeoutDuration) seconds. */
func (hrb HttpRequestBuilder) Timeout(timeout time.Duration) HttpRequestBuilder {
	hrb.hr.timeout = timeout
	return hrb
}

/* HttpRequestBuilder.LoggingEnabled decides whether request result should be logged or not. Default is true. */
func (hrb HttpRequestBuilder) LoggingEnabled(enabled bool) HttpRequestBuilder {
	hrb.hr.loggingEnabled = enabled
	return hrb
}

func (hrb HttpRequestBuilder) Build() (*HttpRequest, RequestError) {
	var err error

	if hrb.hr.request == nil {
		err = validateRequiredRequestFields(hrb.ri)
		if err != nil {
			return nil, NewRequestBuildError(InvalidRequestErr, errors.Wrap(err, "Invalid request fields"))
		}

		// Build the request object if request is nil
		// Construct URL by escaping components
		escapedURLString := buildEndpoint(hrb.ri.scheme, hrb.ri.host, hrb.ri.pathElements)
		hrb.hr.request, err = http.NewRequest("", escapedURLString, hrb.ri.body)
		if err != nil {
			return nil, NewRequestBuildError(InvalidRequestErr, errors.Wrap(err, "Failed to construct http.Request"))
		}
	} else if hrb.ri.body != nil {
		// If request is not nil, set body
		hrb.hr.request.Body = ioutil.NopCloser(hrb.ri.body)
	}

	hrb.hr.request.Close = true

	// Validate that resulting URL path in request is valid
	if len(hrb.ri.pathElements) != 0 {
		if _, err = url.ParseRequestURI(hrb.hr.request.URL.Path); err != nil {
			return nil, NewRequestBuildError(InvalidRequestErr, errors.Wrap(err, "Invalid Request URI"))
		}
	}

	// Set queryParams if exists
	if hrb.ri.queryParams != nil {
		hrb.hr.request.URL.RawQuery = hrb.ri.queryParams.Encode()
	}

	setHeaderIfExists := func(key string, value []string) {
		if key != "" && len(value) != 0 {
			for _, v := range value {
				hrb.hr.request.Header.Add(key, v)
			}
		}
	}

	// Set custom headers
	if hrb.ri.header != nil {
		for hName, hValue := range *hrb.ri.header {
			setHeaderIfExists(hName, hValue)
		}
	}

	return &hrb.hr, nil
}

/* validateRequiredRequestFields ensures that built HttpRequest object has its required fields set to a nonzero value */
func validateRequiredRequestFields(ri requestInfo) error {
	if ri.scheme == "" {
		return errors.New("Empty request scheme")
	}

	if ri.host == "" {
		return errors.New("Empty request host")
	}

	return nil
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
