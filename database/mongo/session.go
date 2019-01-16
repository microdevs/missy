package mongo

import (
	"time"
)

type Config struct {
	Host     string        `env:"MONGO_HOST" envDefault:"localhost:27017"`
	User     string        `env:"MONGO_USER"`
	Password string        `env:"MONGO_PASSWORD"`
	Timeout  time.Duration `env:"MONGO_TIMEOUT" envDefault:"10s"`
	SSL      bool          `env:"MONGO_SSL" envDefault:"false"`
}

// TODO prepare proper mongo interfaces that will be exposed

//func NewSession(c Config) (*mgo.Session, error) {
//	dialInfo := &mgo.DialInfo{
//		Addrs:    []string{c.Host},
//		Timeout:  c.Timeout,
//		Username: c.User,
//		Password: c.Password,
//	}
//	if c.SSL {
//		dialInfo = dialInfoWithSSL(dialInfo)
//	}
//	s, err := mgo.DialWithInfo(dialInfo)
//	if err != nil {
//		return nil, errors.Errorf("couldn't connect to '%s', err: %s", c.Host, err)
//	}
//	return s, nil
//}
//
//func dialInfoWithSSL(dialInfo *mgo.DialInfo) *mgo.DialInfo {
//	tlsConfig := &tls.Config{}
//	tlsConfig.InsecureSkipVerify = true
//	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
//		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
//		return conn, err
//	}
//	return dialInfo
//}
