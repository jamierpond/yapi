// Package importer provides functionality to import external API collections into yapi format
package importer

// PostmanCollection represents a Postman Collection v2.1 schema
type PostmanCollection struct {
	Info PostmanInfo   `json:"info"`
	Item []PostmanItem `json:"item"`
}

// PostmanInfo contains metadata about the collection
type PostmanInfo struct {
	PostmanID   string `json:"_postman_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema"`
}

// PostmanItem represents either a folder or a request in the collection
type PostmanItem struct {
	Name    string         `json:"name"`
	Request *PostmanRequest `json:"request,omitempty"`
	Item    []PostmanItem  `json:"item,omitempty"` // For nested folders
}

// PostmanRequest represents an HTTP request in Postman format
type PostmanRequest struct {
	Method string             `json:"method"`
	Header []PostmanHeader    `json:"header,omitempty"`
	Body   *PostmanBody       `json:"body,omitempty"`
	URL    PostmanURL         `json:"url"`
}

// PostmanHeader represents a request header
type PostmanHeader struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled,omitempty"`
}

// PostmanBody represents a request body
type PostmanBody struct {
	Mode    string          `json:"mode"` // raw, urlencoded, formdata, etc.
	Raw     string          `json:"raw,omitempty"`
	Options *PostmanOptions `json:"options,omitempty"`
}

// PostmanOptions contains body options like language
type PostmanOptions struct {
	Raw *PostmanRawOptions `json:"raw,omitempty"`
}

// PostmanRawOptions contains raw body options
type PostmanRawOptions struct {
	Language string `json:"language,omitempty"` // json, xml, text, etc.
}

// PostmanURL can be either a string or an object
type PostmanURL struct {
	Raw      string   `json:"raw"`
	Protocol string   `json:"protocol,omitempty"`
	Host     []string `json:"host,omitempty"`
	Path     []string `json:"path,omitempty"`
	Query    []PostmanQueryParam `json:"query,omitempty"`
}

// PostmanQueryParam represents a URL query parameter
type PostmanQueryParam struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled,omitempty"`
}
