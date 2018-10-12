package aux

import (
	"github.com/microdevs/missy/log"
	"github.com/microdevs/missy/messaging"
	"github.com/microdevs/missy/service"
)

const Messaging = "messaging"

// Use will enable auxiliary modules
func Use(auxiliaries ...string) {
	for _, a := range auxiliaries {
		switch a {
		case Messaging:
			messaging.InitConfig()
			break
		default:
			log.Warnf("Unknown auxiliary included: aux.Use(\"%s\")", a)
		}
	}
	service.Config().Parse()
}
