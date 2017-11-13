//+build !test

package config
import "github.com/microdevs/missy/log"

// Init function of the config package
func init() {
	log.Debug("Calling config init()")
	config := GetInstance()
	config.ParseEnv()
}