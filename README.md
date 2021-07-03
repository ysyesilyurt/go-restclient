# go-restclient

A Handy Go Package to Call Internal HTTP APIs with Builder Pattern

```
go get github.com/ysyesilyurt/go-restclient/restclient
```

## Write Once, Use Everywhere

Essentially `go-restclient` is a handy `builder` wrapper around `net/http` for all the [_pretty standard
stuff_](#execution-flow)
that needs to be repeated in all HTTP Request/Response cycles. The purpose is to save you from the pain of massive code
duplication those ordinary cycles may lead.

As the number of distinct Internal HTTP requests you are making grows, the need for such a wrapper package also
exponentially grows. In that sense  _handiness_ level of `go-restclient` is directly proportional to your need for such
request/responses 🙂

### Execution Flow

After a `Build()` call on `restclient.RequestBuilder()` upon one or many builder method calls:

* Constructs URL by escaping path components using provided components
* Sets queryParams if exists
* Validates that resulting URI is valid
* Sets custom and universal headers
* Sets authentication strategy to be used
* Sets Authorization header by applying provided authentication strategy
* Sets provided RequestBody if exists (or marshals then sets if given as JSON Body)
* Builds the request object
* Sets request context timeout and defer its cancellation
* Does Request (Times and Logs it if needed)
* Handles Response Status Code and any errors that can be returned from client's call (returns a
  detailed [`restclient.RequestError`](#error-handling))
* Reads the body and unmarshal it into given variable's reference

## Usage

* First build your HTTP Request using `restclient.RequestBuilder()` constructor and with all the builder methods you
  need.

```
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
```

* Then make your call...

```
reqErr = req.Get()
if reqErr != nil {
	return errors.Wrap(reqErr, "Failed to perform HTTP GET request")
}
```

### Available HTTP Calls

* `Get() RequestError`
* `Post() RequestError`
* `Put() RequestError`
* `Patch() RequestError`
* `Delete() RequestError`

### Available Builders

* `Scheme(scheme string)` -> Sets scheme field of your URL. Example:

```
req, reqErr := restclient.RequestBuilder().
                Scheme("https").
                Build()
```

* `Host(host string)` -> Sets host field of your URL. Example:

```
req, reqErr := restclient.RequestBuilder().
                Scheme("https").
                Host("ysyesilyurt.com").
                Build()
```

* `PathElements(pe []string)` -> Sets elements that reside in your URL's Path (not the query params) in the order given
  in the array. Example:

```
req, reqErr := restclient.RequestBuilder().
                Scheme("https").
                Host("ysyesilyurt.com").
                PathElements([]string{"tasks", "1"}).
                Build()
```

* `QueryParams(qp *url.Values)` -> Sets the Query Parameters that reside in your URL in the order given in the array.
  Example:

```
req, reqErr := restclient.RequestBuilder().
                Scheme("https").
                Host("ysyesilyurt.com").
                PathElements([]string{"tasks", "1"}).
                QueryParams(&url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}).
                Build()
```

* `RawUrl(rawUrl string)` -> Use this one if you do not want to parse your URL into several builder methods above and
  see your URL as a whole. This one will automatically parse your URL and set the fields accordingly. Example:

```
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Build()
```

* `Header(header *http.Header)` -> Sets the headers that you want to include in your request. Example:

```
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Header(&http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}).
                Build()
```

* `Body(body io.Reader)` -> Sets the request body in the form of `io.Reader` for your request, if you want to include a
  body for your request of course. Example:

```
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Header(&http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}).
                Body(bytes.NewBufferString(query)).
                Build()
```

* `BodyJson(bodyJson interface{})` -> Use this one instead of `Body(body io.Reader)` if your request body is in JSON
  form. You can pass your struct object directly to this builder method, it is going to marshal your object and set the
  body as `io.Reader`. Example:

```
requestBody := dummyBodyDto{
    Id:   123,
    Name: "1234",
}
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Header(&http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}).
                BodyJson(requestBody).
                Build()
```

* `Auth(auth Authenticator)` -> Sets the Authentication Strategy for your request. Implement `restclient.Authenticator`
  to create your own `Authenticator`, an example `Basic Auth` implementation can be found in `basic_authenticator.go`.
  Example:

```
basicAuth := newBasicAuthenticator("ysyesilyurt", "0123")
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Header(&http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}).
                Auth(basicAuth).
                Build()
```

* `ResponseReference(respRef interface{})` -> Sets the reference of the variable to map response object returned from
  request. Example:

```
var response dummyHttpResponse
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Header(&http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}).
                ResponseReference(&response).
                Build()
```

* `Request(req *http.Request)` -> If you happen to have a pre-prepared valid `http.Request` object and want to build a
  new request upon this request then you can simply use this builder method. Example:

```
var response dummyHttpResponse
req, reqErr := restclient.RequestBuilder().
                Request("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Header(&http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}).
                ResponseReference(&response).
                Build()
```

* `Timeout(timeout time.Duration)` -> Sets the timeout value that should be used for the request. _Default_ is 60
  seconds. Example:

```
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                Timeout(30 * time.Second)             
                Build()
```

* `LoggingEnabled(enabled bool)` -> Decides whether responses should be logged or not. _Default_ is false. Example:

```
req, reqErr := restclient.RequestBuilder().
                RawUrl("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1").
                LoggingEnabled(true)             
                Build()
```

## Error Handling

`go-restclient` defines `restclient.RequestError` interface to cover all the errors that can be returned
from `restclient.RequestBuilder.Build()` and result of the HTTP calls. The top level errors returned by the client can
be seen
in [errors.go](https://github.com/ysyesilyurt/go-restclient/blob/bc71e1bf147e9293635583edbca00e5db08dd7d5/restclient/errors.go#L9)
. The methods that can be used within the boundaries of `restclient.RequestError` are as follows:

```
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
```

## Legacy Version

You can also use the legacy version which is located
on [`legacy` branch](https://github.com/ysyesilyurt/go-restclient/tree/legacy) if you want to use `go-restclient` using
traditional ways (with constructing the helper objects and using those altogether blah blah...) or if you want to use a
version that allows reusing the HTTP clients that's being used internally for sending the requests (Thanks to separation
of these objects, this legacy version of `go-restclient` gives you the slightly extended feature set like _separation of
client timeouts and request-specific timeouts_ and etc.). You can find more information about the version itself and its
usage from [here](https://github.com/ysyesilyurt/go-restclient/tree/legacy#go-restclient-legacy-version). But all in all, I think the builder version is more handy and elegant so I just wanted to keep
builder version on the master branch.

## Contribution

Unit tests are implemented for the codebase but all the edge cases might not be covered, weird bugs may appear in such
cases. Feel free to open an issue if you spot a bug. In addition if you have an improvement idea you can request it in
the form of an issue (with providing a clear and brief description of the feature) or you can develop the improvement
and open a PR with again a brief description of the reasoning.
