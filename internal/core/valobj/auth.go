package valobj

type AuthType string

const (
	AuthNone   AuthType = "none"
	AuthBasic  AuthType = "basic"
	AuthBearer AuthType = "bearer"
	AuthAPIKey AuthType = "apikey"
)

func (s AuthType) String() string {
	switch s {
	case AuthNone:
		return "None"
	case AuthBasic:
		return "Basic"
	case AuthBearer:
		return "Bearer"
	case AuthAPIKey:
		return "API Key"
	default:
		return "None"
	}
}

type APIKeyLocation string

const (
	APIKeyInHeader APIKeyLocation = "header"
	APIKeyInQuery  APIKeyLocation = "query"
)

func (s APIKeyLocation) String() string {
	switch s {
	case APIKeyInQuery:
		return string(APIKeyInQuery)
	case APIKeyInHeader:
		return string(APIKeyInHeader)
	default:
		return ""
	}
}

type Auth struct {
	Type        AuthType       `json:"type,omitempty"`
	Username    string         `json:"username,omitempty"`
	Password    string         `json:"password,omitempty"`
	Token       string         `json:"token,omitempty"`
	APIKeyName  string         `json:"api_key_name,omitempty"`
	APIKeyValue string         `json:"api_key_value,omitempty"`
	APIKeyIn    APIKeyLocation `json:"api_key_in,omitempty"`
}
