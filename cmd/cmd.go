package main

import (
	"flag"
	"fmt"
	"github.com/microdevs/missy/cmd/missy"
	_ "github.com/microdevs/missy/cmd/missy"
	"github.com/microdevs/missy/log"
	"k8s.io/api/storage/v1beta1"
	"os"
	"path/filepath"

	"github.com/microdevs/missy/cmd/kubernetes"
	"k8s.io/client-go/util/homedir"

	"github.com/manifoldco/promptui"
)

var initCmd = flag.NewFlagSet("init", flag.ExitOnError)
var initForce bool
var kubeconfig *string

func main() {

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = initCmd.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = initCmd.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	initCmd.BoolVar(&initForce, "force", false, "Force recreate certificates")

	if len(os.Args) > 1 && os.Args[1] == "init" {

		fmt.Print(` __  __ _____  _____ _______     __
|  \/  |_   _|/ ____/ ____\ \   / /
| \  / | | | | (___| (___  \ \_/ / 
| |\/| | | |  \___ \\___ \  \   /  
| |  | |_| |_ ____) |___) |  | |   
|_|  |_|_____|_____/_____/   |_|
The  Microservice  Support  System

The Kubernetes initialization process assumes you have a kubernetes cluster running
and that you selected the right cluster with "kubectl config set-context cluster"

You will now install the following compagnion services in this cluster
- hashicorp/vault    - a secure secret store
- hashicorp/consul   - a key/value database and storage backend
- prometheus         - a monitoring system
- alert manager      - alerting for prometheus
- missy-controller   - service registry, controller and ui

`)

		var storageClass v1beta1.StorageClass

		storageClasses, err := kubernetes.GetStorageClasses(kubeconfig)
		if err != nil {
			log.Fatalf("Unable to communicate with the Kubernetes cluster %s", err)
		}

		// add the first item by default
		storageClass = storageClasses[0]

		// if there are more items let the user decide
		if len(storageClasses) > 1 {
			var choices []string
			storageClassChoices := func() []string {
				for _,s := range storageClasses {
					choices = append(choices, s.Name+" - "+s.Provisioner)
				}
				return choices
			}()

			prompt := promptui.Select{
				Label: "Which storage class would you like to use for persistant storages?",
				Items: storageClassChoices,
			}

			idx, _, err := prompt.Run()
			if err != nil {
				log.Fatal("Error during prompt")
			}

			storageClass = storageClasses[idx]
		}


		initCmd.Parse(os.Args[2:])
		if initForce {
			missy.CreateRootCA()
			// intermediate cert generation is not worked
			//missy.CreateCertFromCA( "intermediate", "root-cert.pem", "root-key.pem", "Missy", "Missy Subsystem", "Missy Subsystem Intermediate", "DE", "Berlin", true, []string{})
			missy.CreateCertFromCA( "vault", "root-cert.pem", "root-key.pem", "Missy", "Missy Subsystem", "vault", "DE", "Berlin", false, []string{"vault.missy.svc.cluster.local", "127.0.0.1"})
			missy.CreateCertFromCA( "consul", "root-cert.pem", "root-key.pem", "Missy", "Missy Subsystem", "server.dc1.cluster.local", "DE", "Berlin", false, []string{"server.dc1.cluster.local", "127.0.0.1"})
		}
		kubernetes.InitializeBaseSystem(kubeconfig)
		kubernetes.CreateConsulConfig()
		kubernetes.RunConsul(storageClass)

		//kubernetes.CreateVaultConfig()
		//kubernetes.RunVault()

		// kubernetes.ConfigureIntermediateRootCA()

	}
}
