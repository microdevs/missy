package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var yml = []byte(`
name: test
authorization:
  publicKeyFile: "/some/path"
environment:
  - envName: ENVVAR_A
    defaultValue: foo
    internalName: var.a
    mandatory: false
    usage: "This is the description for ENVVAR_A"
  - envName: ENVVAR_B
    defaultValue: "bar"
    internalName: var.b
    mandatory: false
    usage: "This is the description for ENVVAR_B"
`)

var ymlMandatory = []byte(`
name: test
environment:
  - envName: ENVVAR_A
    defaultValue: foo
    internalName: var.a
    mandatory: true
    usage: "This is the description for ENVVAR_A"
  - envName: ENVVAR_B
    defaultValue: "bar"
    internalName: var.b
    mandatory: false
    usage: "This is the description for ENVVAR_B"
`)

func TestParseValueConfig(t *testing.T) {
	err := parseConfigYAML(yml)
	if err != nil {
		t.Log("Error unmarshalling config yaml: ", err)
		t.Fail()
	}

	if config.Name != "test" {
		t.Log(fmt.Sprintf("Expected name to be test but is %s", config.Name))
		t.Fail()
	}

	if config.Authorization.PublicKeyFile != "/some/path" {
		t.Log(fmt.Sprintf("Expected publicKeyFile to be /some/path but is %s", config.Authorization.PublicKeyFile))
		t.Fail()
	}
}

func TestLoadDefaultConfigFile(t *testing.T) {
	ioutil.WriteFile(MissyConfigFile, yml, os.FileMode(0644))
	rb, readErr := readDefaultFile()
	if readErr != nil {
		t.Error("Cannot read file")
		return
	}
	if bytes.Compare(rb, yml) != 0 {
		t.Error("Contents of read file does not equal contents written.")
	}
	//cleanup
	os.Remove(MissyConfigFile)
}

func TestParseEnvironment(t *testing.T) {

	os.Setenv("ENVVAR_A", "testA")

	err := parseConfigYAML(ymlMandatory)
	if err != nil {
		t.Log("Error unmarshalling config yaml: ", err)
		t.Fail()
	}
	config.ParseEnv()

	if config.Get("var.a") != "testA" {
		t.Error("ENVVAR_A was not set correcly")
	}

	if config.Get("var.b") != "bar" {
		t.Error("ENVVAR_B was not set to default valule")
	}
}
