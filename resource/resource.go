package resource

import (
	"github.com/microdevs/missy/config"
	"net/http"
)

const MysqlResourceKey int = 100

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

func Initialize(req *http.Request) {
	c := config.GetInstance()
	for _, r := range c.Resources {
		switch r {
		case "mysql":
			my := Mysql{}
			my.Initialize(req)
		}
	}
}
