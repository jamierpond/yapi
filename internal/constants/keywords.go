package constants

import "strings"

// HTTP methods
const (
	MethodGET     = "GET"
	MethodPOST    = "POST"
	MethodPUT     = "PUT"
	MethodDELETE  = "DELETE"
	MethodPATCH   = "PATCH"
	MethodHEAD    = "HEAD"
	MethodOPTIONS = "OPTIONS"
)

// Transport types
const (
	TransportHTTP    = "http"
	TransportGRPC    = "grpc"
	TransportTCP     = "tcp"
	TransportGraphQL = "graphql"
)

// ValidHTTPMethods contains all valid HTTP verbs for validation
var ValidHTTPMethods = map[string]bool{
	MethodGET:     true,
	MethodPOST:    true,
	MethodPUT:     true,
	MethodDELETE:  true,
	MethodPATCH:   true,
	MethodHEAD:    true,
	MethodOPTIONS: true,
}

// CanonicalizeMethod returns canonical uppercase method name.
func CanonicalizeMethod(m string) string {
	return strings.ToUpper(strings.TrimSpace(m))
}
