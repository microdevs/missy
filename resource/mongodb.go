package resource

import (
	"github.com/microdevs/missy/config"
	"gopkg.in/mgo.v2"
	"log"
)

var mgoInstance *MgoDb

const MgoResourceName = "mongodb"
const MgoEnvDSN = MgoResourceName + ".dataSourceName"

// Mysql is a type that holds the current active connection and its properties
type MgoDb struct {
	Dsn     string
	Session *mgo.Session
}

func MgoSession() *mgo.Session {
	if !config.GetInstance().ResourceAvailable(MgoResourceName) {
		log.Panic("Resource MySQL is not configured. Please add a resource entry in .missy.yml")
		return nil
	}

	once.Do(func() {
		dsn := config.Get(MgoEnvDSN)
		session, err := mgo.Dial(dsn)

		if err != nil {
			log.Fatal("Dialing into %s: %v", dsn, err)
		}

		log.Printf("Connected to DB %s", dsn)

		// Setup adds the MongoDb parameters to the service config
		mgoInstance = &MgoDb{Dsn: dsn, Session: session}
	})

	// Do not forget to close it.
	return mgoInstance.Session
}

func (r *MgoDb) Setup(c *config.Config) {
	dsn := config.EnvParameter{
		EnvName:      "MYSQL_DSN",
		Mandatory:    true,
		InternalName: MgoEnvDSN,
		Usage:        MgoResourceName + " Data Service Name",
		DefaultValue: "",
	}
	c.AddEnv(dsn)
}

// Leave the DB name empty so that we use the DB from the dialed url
