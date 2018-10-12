package service

// Configuration holds the service configuration for the MiSSy service
type Configuration struct {
	Name        string         `json:"name"`
	Environment []EnvParameter `json:"environment"`
}

// EnvParameter defines how a config value is passed through an environment variable. This struct as members for
// default values and usage description. It also can mark the variable non-mandatory. An external system
// environment variable always maps to an internal name. As a guideline the internal name should refer to the module
// it is used in and should have sections divided by dots, e.g. "datastore.mysql.host"
type EnvParameter struct {
	EnvName      string `json:"envName"`
	DefaultValue string `json:"defaultValue"`
	InternalName string `json:"internalName"`
	Mandatory    bool   `json:"mandatory"`
	Usage        string `json:"usage"`
	Value        string `json:"-"`
	Parsed       bool   `json:"-"`
}
