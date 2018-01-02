package resource

import (
	"gopkg.in/mgo.v2"
)

var mgoInstance *mgo.Database

// Mysql is a type that holds the current active connection and its properties
type Mgo struct {
	Username string
	Password string
	Db       string
	Host     string
	Port     string
	Session  *mgo.Session
}


func MgoConnection() *mgo.Database {

}