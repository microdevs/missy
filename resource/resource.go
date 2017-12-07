package resource

import (
	"github.com/microdevs/missy/config"
)

type Instance interface {
	Connection() (interface{}, error)
	Setup(c *config.Config)
}

func Setup(c *config.Config) {
	for _,r := range c.Resources{
		switch r {
		case "mysql":
			my := Mysql{}
			my.Setup(c)
		}
	}
}