# go-restclient (Legacy version)

A Handy Go Package to Call Internal HTTP APIs

```
go get github.com/ysyesilyurt/go-restclient/restclient
```

## Write Once, Use Everywhere

Essentially `go-restclient` is a handy wrapper around `net/http` for all the [_pretty standard stuff_](#feature-set) that needs to be repeated
in all HTTP Request/Response cycles. The purpose is to save you from the pain of massive code duplication those ordinary cycles may lead.

As the number of distinct Internal HTTP requests you are making grows, the need for such a wrapper package also
exponentially grows. In that sense  _handiness_ level of `go-restclient` is directly proportional to your need for such
request/responses 🙂

## Feature Set

#### After `restclient.NewRequest(...)`

* Constructs URL by escaping path components
* Marshals RequestBody if exists
* Builds the request object
* Validates that resulting URI is valid
* Sets queryParams if exists
* Sets custom headers

#### After calling `client.Get(...)` on `client := restclient.NewHttpClient()`

* Sets universal headers
* Sets Authorization header by applying specified authenticator's strategy
* Sets request context timeout and defer its cancellation
* Does Request (Times and Logs it if needed)
* Handles Response Status Code
* Reads the body and unmarshal it into given value

## Usage

### 1- Construct Request

* First you need to create a `restclient.RequestInfo` which is going to hold all the information you need when creating
  a new request object. Can be created using `restclient.NewRequestInfo` or `restclient.NewRequestInfoFromRawURL`.

```
/* Creating a RequestInfo Using NewRequestInfo 
 * NewRequestInfo(scheme, host string, pathElements []string, queryParams *url.Values, headers *http.Header, body interface{}) RequestInfo
 * Params:
 *  Scheme       -> e.g. http
 *  Host         -> e.g. ysyesilyurt.com
 *  PathElements -> represents each component in the path that is separated by a slash (/) e.g. ['posts', '1']
 *  Headers      -> e.g. http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
 *  Body         -> represents the to-be-marshalled RequestBody variable
 *  QueryParams  -> e.g. url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
 */
var requestBody dummyBodyDto
var response dummyHttpResponse
requestBody = dummyBodyDto{
    Id:   123,
    Name: "1234",
}
headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
queryParams := url.Values{"tenantId": []string{"d90c3101-53bc-4c54-94db-21582bab8e17"}, "vectorId": []string{"1"}}
ri := restclient.NewRequestInfo("https", "ysyesilyurt.com", []string{"tasks", "1"}, &queryParams, &headers, requestBody)

...

/* Creating a RequestInfo Using NewRequestInfoFromRawURL 
 * NewRequestInfoFromRawURL(rawURL string, headers *http.Header, body interface{}) (RequestInfo, error)
 * Params:
 *  rawURL  -> https://ysyesilyurt.com/assessments/scroll/d90c3101-53bc-4c54-94db-21582bab8e17/1?tenantId=ABC123&vectorId=1
 *  headers -> e.g. http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
 *  body    -> represents the to-be-marshalled RequestBody variable 
 */
headers := http.Header{"Content-Type": []string{"application/json"}, "Cookie": []string{"test-1234"}}
ri, err := restclient.NewRequestInfoFromRawURL("https://ysyesilyurt.com/tasks/1?tenantId=d90c3101-53bc-4c54-94db-21582bab8e17&vectorId=1", &headers, nil)
if err != nil {
    return errors.Wrap(err, "Failed to construct RequestInfo out of Raw URL")
}
```

* Then you can create a new `http.Request` using `restclient.NewRequest` which creates a new `*http.Request` and returns
  an `error` (with stack) using `restclient.RequestInfo`

```
...
/* NewRequest(ri RequestInfo) (*http.Request, error) */
req, err := restclient.NewRequest(ri)
if err != nil {
    return errors.Wrap(err, "Failed to construct http.Request out of RequestInfo")
}
```

### 2- Construct Client

* Create a `restclient.HttpClient` using `restclient.NewHttpClient`

```
/* NewHttpClient(loggingEnabled bool, timeout time.Duration) HttpClient 
 * Params:
 *  loggingEnabled -> Logs responses of the requests when enabled with request/response duration
 *  timeout -> Determines timeout duration to be used for the HttpClient. Valid for all the requests made using corresponding client.
 */
client := restclient.NewHttpClient(true, 30*time.Second)
```

### 3- Do Request

* First implement `restclient.Authenticator` if you want to authenticate your request by filling its `Authorization`
  header. You can find a sample implementation `testBasicAuthenticator` in `client_test.go`
* Create a `restclient.DoRequestInfo` using `restclient.NewDoRequestInfo` or `restclient.NewDoRequestInfoWithTimeout`

```
/* You can pass nil as Authenticator if you don't want to use the Authorization header.
 * NewDoRequestInfo(request *http.Request, auth Authenticator, responseReference interface{}) DoRequestInfo
 * Params:
 *  request           -> http.Request object (ideally the one that is created using restclient.NewRequest)
 *  auth              -> restclient.Authenticator (can be any authentication implementation based on restclient.Authenticator)
 *  responseReference -> variable that is going to be used for mapping the response of the request (Should be passed with its address)
 */
dri := restclient.NewDoRequestInfo(req, nil, &response)

...

/* NewDoRequestInfoWithTimeout(request *http.Request, auth Authenticator, responseReference interface{}, requestTimeout time.Duration) DoRequestInfo
 * Params:
 *  request           -> http.Request object (ideally the one that is created using restclient.NewRequest)
 *  auth              -> restclient.Authenticator (can be any authentication implementation based on restclient.Authenticator)
 *  responseReference -> variable that is going to be used for mapping the response of the request (Should be passed with its address)
 *  requestTimeout    -> timeout value that should be used for this specific request. Handy when you need a specific timeout for the request and don't want to use the HttpClient's timeout.
 */
dri := restclient.NewDoRequestInfo(req, nil, &response)
```

* Finally you can make your call! Use your `restclient.HttpClient` and `restclient.DoRequestInfo` to make all kinds of
  requests.

```
/* All methods below accepts a DoRequestInfo as parameter 
 * (hc HttpClient) Get(dri DoRequestInfo) error
 * (hc HttpClient) Post(dri DoRequestInfo) error
 * (hc HttpClient) Put(dri DoRequestInfo) error
 * (hc HttpClient) Patch(dri DoRequestInfo) error
 * (hc HttpClient) Delete(dri DoRequestInfo) error
 * Params:
 *  dri -> DoRequestInfo
 */
err = client.Get(dri)
err = client.Post(dri)
err = client.Put(dri)
err = client.Patch(dri)
err = client.Delete(dri)
```

### Better yet, use wrappers!

* Instead of going over all (well, most...) the steps above individually you can perform your request using wrappers below.

```
/*
 * PerformGetRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error
 * PerformPostRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error
 * PerformPutRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error
 * PerformPatchRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error
 * PerformDeleteRequest(ri RequestInfo, auth Authenticator, responseRef interface{}, loggingEnabled bool, timeout time.Duration) error
 * Params:
 *  ri             -> RequestInfo to construct the basics of the http.Request
 *  auth           -> restclient.Authenticator (can be any authentication implementation based on restclient.Authenticator)
 *  responseRef    -> variable that is going to be used for mapping the response of the request (Should be passed with its address)
 *  loggingEnabled -> Logs responses of the requests when enabled with request/response duration
 *  timeout        -> Determines timeout duration to be used for the HttpClient. Valid for all the requests made using corresponding client.
 */
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
```

