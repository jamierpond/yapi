package domain

import (
	"strings"

	"yapi.run/cli/internal/constants"
)

// DetectTransport determines the transport type from URL scheme and GraphQL content.
func DetectTransport(url string, hasGraphQL bool) string {
	urlLower := strings.ToLower(url)

	if strings.HasPrefix(urlLower, "grpc://") || strings.HasPrefix(urlLower, "grpcs://") {
		return constants.TransportGRPC
	}
	if strings.HasPrefix(urlLower, "tcp://") {
		return constants.TransportTCP
	}
	if hasGraphQL {
		return constants.TransportGraphQL
	}
	return constants.TransportHTTP
}
