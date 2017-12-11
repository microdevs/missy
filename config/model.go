package config

// Holds the service config for the MiSSy service
type Config struct {
	Name string `yaml:"name"`
	Environment []EnvParameter `yaml:"environment,flow,omitempty"`
	Resources []string `yaml:"resources,omitempty"`
}
// Defines how a config value is passed through an environment variable. This struct as members for
// default values and usage description. It also can mark the variable non-mandatory. An external system
// environment variable always maps to an internal name. As a guideline the internal name should refer to the module
// it is used in and should have sections devided by dots, e.g. "datastore.mysql.host"
type EnvParameter struct {
	EnvName string `yaml:"envName"`
	DefaultValue string `yaml:"defaultValue,omitempty"`
	InternalName string `yaml:"internalName"`
	Mandatory bool `yaml:"mandatory"`
	Usage string`yaml:"usage"`
	Value string `yaml:"value,omitempty"`
}
// tests if resource is configured in config. We assume that this resource is then available
func (c *Config) ResourceAvailable(test string) bool {
	for _,v := range c.Resources {
		if v == test {
			return true
		}
	}
	return false
}