package service

import (
	"os"
	"testing"
)

func TestParseEnvironment(t *testing.T) {

	os.Setenv("ENVVAR_A", "foo")
	Config().RegisterParameter(
		"ENVVAR_A", "", "var.a", true, "This tests the env parser")
	Config().RegisterParameter(
		"ENVVAR_B", "bar", "var.b", false, "This tests the env parser")

	Config().Parse()

	if config.Get("var.a") != "foo" {
		t.Error("ENVVAR_A was not set correcly")
	}

	if config.Get("var.b") != "bar" {
		t.Error("ENVVAR_B was not set to default valule")
	}
}
