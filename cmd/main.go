package main

import (
	"flag"
	"github.com/microdevs/missy/cmd/missy"
	_ "github.com/microdevs/missy/cmd/missy"
	"os"
	"path/filepath"

	"github.com/microdevs/missy/cmd/kubernetes"
	"k8s.io/client-go/util/homedir"
)

var initCmd = flag.NewFlagSet("init", flag.ExitOnError)
var initForce bool

func main() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = initCmd.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = initCmd.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	initCmd.BoolVar(&initForce, "force", false, "Force recreate certificates")

	if len(os.Args) > 1 && os.Args[1] == "init" {
		initCmd.Parse(os.Args[2:])
		if initForce {
			missy.CreateRootCA()
			missy.CreateCertFromCA( "intermediate", "root-cert.pem", "root-key.pem", "Missy", "Missy Subsystem", "Missy Subsystem Intermediate", "DE", "Berlin", true)
			missy.CreateCertFromCA( "vault", "intermediate-cert.pem", "intermediate-key.pem", "Missy", "Missy Subsystem", "vault", "DE", "Berlin", false)
		}
		kubernetes.InitializeBaseSystem(kubeconfig)
		kubernetes.CreateVaultConfig()
		kubernetes.UploadCertificateToKubernetes()
		kubernetes.RunVault()

		// kubernetes.ConfigureIntermediateRootCA()

	}
}
