package resource

import (
	"github.com/go-playground/locales/mg"
	"github.com/microdevs/missy/config"
	"net/http"
)

// Instance interface
type Instance interface {
	Connection() (interface{}, error)
	Setup(c *config.Config)
	Initialize(r *http.Request)
}

// Setup calls Setup() functions on all registered resources
func Setup(c *config.Config) {
	for _, r := range c.Resources {
		switch r {
		case MysqlResourceName:
			my := Mysql{}
			my.Setup(c)
		case MgoResourceName:
			mg := MgoDb{}
			mg.Setup(c)
		}
	}
}
