package resource

import (
	"github.com/microdevs/missy/config"
	"net/http"
)

type Instance interface {
	Connection() (interface{}, error)
	Setup(c *config.Config)
	Initialize(r *http.Request)
}

func Setup(c *config.Config) {
	for _, r := range c.Resources {
		switch r {
		case "mysql":
			my := Mysql{}
			my.Setup(c)
		}
	}
}
