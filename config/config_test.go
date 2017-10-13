package config

import (
	"testing"
	"syscall"
)

func TestConfig(t *testing.T) {
	//new config should be empty
	if len(parameters) > 0 {
		t.Errorf("New config is not empty, found %v", parameters)
	}

	envName := "TEST"
	defaultValue := "foo"
	internalName := "test"
	mandatory := true
	usage := "testusage"


	// register mandatory parameter
	RegisterParameter(envName, defaultValue, internalName, mandatory, usage)

	// make sure we have 1 registered parameter
	if len(parameters) != 1 {
		t.Errorf("Config should hold 1 parameter, found %d", len(parameters))
	}

	// make sure all was set propery
	p := parameters[internalName]
	if p.EnvName != envName {
		t.Errorf("Expeted to be envName %s got %s", envName, p.EnvName)
	}

	if p.Mandatory != mandatory {
		t.Errorf("Expeted mandatory to be %v got %v", mandatory, p.Mandatory)
	}

	if p.Usage != usage {
		t.Errorf("Expeted usage to be %s got %s", usage, p.Usage)
	}

	if p.Value != defaultValue {
		t.Errorf("Expeted value to be %s got %s", defaultValue, p.Value)
	}

	// get value, expect default value
	v0 := Get(internalName)
	if v0 != defaultValue {
		t.Errorf("Getting value before Parse() should return defaultValue %s, got %s", defaultValue, v0)
	}

	// set env variable and parse, test for right value
	testValue := "foobar"
	syscall.Setenv(envName, testValue)
	Parse()
	v1 := Get(internalName)
	if v1 != testValue {
		t.Errorf("Getting value after Parse() should return value %s, got %s", testValue, v1)
	}

}

