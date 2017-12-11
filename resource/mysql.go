package resource

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"github.com/microdevs/missy/config"
	"github.com/microdevs/missy/log"
	"sync"
)

const MysqlResourceName = "mysql"

var mysqlInstance *Mysql
var once sync.Once

type Mysql struct {
	Username         string
	Password         string
	Db               string
	Host             string
	Port             string
	ActiveConnection *sql.DB
}

func MysqlConnection() *sql.DB {

	if !config.GetInstance().ResourceAvailable(MysqlResourceName) {
		log.Panic("Resource MySQL is not configured. Please add a resource entry in .missy.yml")
		return nil
	}

	once.Do(func() {
		mysql := Mysql{
			Username: config.Get("mysql.user"),
			Password: config.Get("mysql.password"),
			Db: config.Get("mysql.db"),
			Host: config.Get("mysql.host"),
			Port: config.Get("mysql.port"),
		}
		connection, err := mysql.Connect()
		if err != nil {
			log.Fatal("Cannot connect to Resource MySQL")
		}
		mysql.ActiveConnection = connection
		mysqlInstance = &mysql
	})
	return mysqlInstance.ActiveConnection
}

func (r *Mysql) Connect() (*sql.DB, error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", r.Username, r.Password, r.Host, r.Port, r.Db)

	if r.ActiveConnection == nil {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Errorf("MySQL Connection to %s failed, %s", r.Host, err)
			return nil, err
		}
		r.ActiveConnection = db
	}

	return r.ActiveConnection, nil
}

func (r *Mysql) Setup(c *config.Config) {
	user := config.EnvParameter{
		EnvName:      "MYSQL_USER",
		Mandatory:    true,
		InternalName: "mysql.user",
		Usage:        "MySQL User Name",
		DefaultValue: "",
	}

	password := config.EnvParameter{
		EnvName:      "MYSQL_PASSWORD",
		Mandatory:    true,
		InternalName: "mysql.password",
		Usage:        "MySQL Password",
		DefaultValue: "",
	}

	host := config.EnvParameter{
		EnvName:      "MYSQL_HOST",
		Mandatory:    false,
		InternalName: "mysql.host",
		Usage:        "MySQL Host",
		DefaultValue: "localhost",
	}

	port := config.EnvParameter{
		EnvName:      "MYSQL_PORT",
		Mandatory:    false,
		InternalName: "mysql.port",
		Usage:        "MySQL Host",
		DefaultValue: "3306",
	}

	db := config.EnvParameter{
		EnvName:      "MYSQL_DB",
		Mandatory:    false,
		InternalName: "mysql.db",
		Usage:        "MySQL Host",
		DefaultValue: "mysql",
	}

	c.AddEnv(user, password, host, port, db)
}
