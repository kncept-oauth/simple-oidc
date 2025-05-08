package client

type Client interface {
	GetClientId() string

	// can return an empty array
	GetAllowedScopes() []string

	// treat the response from GetAllowedRedirectUris as a regex?
	// RECOMMENDED TO BE FALSE
	IsAllowedRedirectUriRegex() bool

	// regex scripts for redirect uris
	GetAllowedRedirectUris() []string
}

type ClientStruct struct {
	ClientId                string
	AllowedScopes           []string
	RegexAllowedRedirectUri bool
	AllowedRedirectUris     []string
}

type ClientStore interface {
	GetClient(clientId string) (Client, error)
	SaveClient(client ClientStruct) error
	ListClients() ([]Client, error)
}

func (obj ClientStruct) GetClientId() string {
	return obj.ClientId
}

func (obj ClientStruct) GetAllowedScopes() []string {
	return obj.AllowedScopes
}

func (obj ClientStruct) IsAllowedRedirectUriRegex() bool {
	return obj.RegexAllowedRedirectUri
}

func (obj ClientStruct) GetAllowedRedirectUris() []string {
	return obj.AllowedRedirectUris
}
