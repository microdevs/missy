package config

import (
	"testing"
	_ "syscall"
	"io/ioutil"
	"os"
	"bytes"
)

var yml = []byte(`
name: test
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
	parseConfigYAML(yml)
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

	parseConfigYAML(ymlMandatory)
	config.ParseEnv()

	if config.Get("var.a") != "testA" {
		t.Error("ENVVAR_A was not set correcly")
	}

	if config.Get("var.b") != "bar" {
		t.Error("ENVVAR_B was not set to default valule")
	}
}
