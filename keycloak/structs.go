package keycloak

//Client Values from TOML file
type Client struct {
	Realm       string `toml:"realm"`
	ID          string `toml:"resource"`
	Credentials Creds  `toml:"credentials"`
}

//Creds is a substruct of Keycloak
type Creds struct {
	Secret string `toml:"secret"`
}

type action int

const (
	actionLogin action = iota
	actionLogout
	actionPageAccess
	actionInvalid
)

type Action string

var (
	ActionLogin      Action = "Login"
	ActionLogout     Action = "Logout"
	ActionPageAccess Action = "Access"
)
