package resource

import (
	"database/sql"
	"github.com/microdevs/missy/log"
	"github.com/microdevs/missy/config"
	"net/http"
	"fmt"
	"github.com/gorilla/context"
)

type Mysql struct {
	Username string
	Password string
	Db string
	Host string
	Port string
	ActiveConnection *sql.DB
}

func (r *Mysql) Connection() (*sql.DB, error) {

	dsn := fmt.Sprintf("mysql://%s:%s@%s:%s/%s",r.Username,r.Password,r.Host,r.Port,r.Db)

	if r.ActiveConnection == nil {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Errorf("Connection to ")
			return nil, err
		}
		r.ActiveConnection = db
	}

	return r.ActiveConnection, nil
}

func (r *Mysql) Setup(c *config.Config) {
	user := config.EnvParameter{
		EnvName: "MYSQL_USER",
		Mandatory: true,
		InternalName: "mysql.user",
		Usage: "MySQL User Name",
		DefaultValue: "",
	}

	password := config.EnvParameter{
		EnvName: "MYSQL_PASSWORD",
		Mandatory: true,
		InternalName: "mysql.password",
		Usage: "MySQL Password",
		DefaultValue: "",
	}

	host := config.EnvParameter{
		EnvName: "MYSQL_HOST",
		Mandatory: false,
		InternalName: "mysql.host",
		Usage: "MySQL Host",
		DefaultValue: "localhost",
	}

	port := config.EnvParameter{
		EnvName: "MYSQL_PORT",
		Mandatory: false,
		InternalName: "mysql.port",
		Usage: "MySQL Host",
		DefaultValue: "3306",
	}

	db := config.EnvParameter{
		EnvName: "MYSQL_DB",
		Mandatory: false,
		InternalName: "mysql.db",
		Usage: "MySQL Host",
		DefaultValue: "mysql",
	}

	c.AddEnv(user, password, host, port, db)
}

func (r *Mysql) Initialize(req *http.Request) {
	resource := Mysql{
		Username: config.Get("mysql.user"),
		Password: config.Get("mysql.password"),
		Db: config.Get("mysql.db"),
		Host: config.Get("mysql.host"),
		Port: config.Get("mysql.port"),
	}

	context.Set(req, MysqlResourceKey, &resource)
}
