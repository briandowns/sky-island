package config

import "testing"

// TestLoad validates that the config.Load function reads
// the data in the given file
func TestLoad(t *testing.T) {
	conf, err := Load("testdata/config.toml")
	if err != nil {
		t.Error(err)
	}
	if conf == nil {
		t.Error("expected non nil")
	}
}

// TestLoad_Failure validates that the config.Load function
// will fail as expected when given an empty string
func TestLoad_EmptyFile_Failure(t *testing.T) {
	conf, err := Load("")
	if err == nil {
		t.Error(err)
	}
	if conf != nil {
		t.Error("expected nil")
	}
}

// TestLoad_BadJSON_Failure validates that the config.Load
// function will fail as expected when receiving bad JSON
func TestLoad_BadJSON_Failure(t *testing.T) {
	conf, err := Load("testdata/bad_config.json")
	if err == nil {
		t.Error(err)
	}
	if conf != nil {
		t.Error("expected nil")
	}
}
