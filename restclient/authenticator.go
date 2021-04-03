package restclient

import "net/http"

/* Base Authenticator to authenticate http.Request objects.
 * Implement this interface to provide authentication method to your http.Request */
type Authenticator interface {
	/* Apply applies the underlying Authenticator's auth method to provided http.Request by setting the `Authorization` header */
	Apply(request *http.Request) error
}
