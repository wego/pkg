package header

// standard header names
const (
	AcceptLanguage = "Accept-Language"
	Authorization  = "Authorization"
	ContentType    = "Content-Type"
	UserAgent      = "User-Agent"
	ForwaredFor    = "X-Forwarded-For"
	ForwardedProto = "X-Forwarded-Proto"
	RealIP         = "X-Real-Ip"
)

// custom header names
const (
	APIKey    = "ApiKey"
	RequestID = "X-Request-ID"
	WegoAuth  = "Wego-Authorization"
)

// header values
const (
	ApplicationJSON            = "application/json"
	ApplicationPDF             = "application/pdf"
	ApplicationXML             = "application/xml"
	ApplicationXFormURLEncoded = "application/x-www-form-urlencoded"
	TextXML                    = "text/xml"
)

// something esle
const (
	BearerPrefix = "Bearer "
)
