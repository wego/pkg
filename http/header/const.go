package header

// standard header names
const (
	AcceptLanguage = "Accept-Language"
	Authorization  = "Authorization"
	ContentType    = "Content-Type"
	UserAgent      = "User-Agent"
	ForwaredFor    = "X-Forwarded-For"
	RealIP         = "X-Real-Ip"
)

// custom header names
const (
	APIKey    = "ApiKey"
	RequestID = "X-Request-ID"
)

// header values
const (
	ApplicationJSON = "application/json"
	TextXML         = "text/xml"
)

// something esle
const (
	BearerPrefix = "Bearer "
)
