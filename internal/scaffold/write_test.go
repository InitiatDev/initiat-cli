package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/InitiatDev/initiat-cli/internal/setup"
)

func TestWrite_FirstTime(t *testing.T) {
	dir := t.TempDir()
	cfg := &setup.SetupConfig{Version: 1, Name: "test", Defaults: &setup.Defaults{Timeout: "5m"}}
	wroteSetup, wroteConfig, err := Write(cfg, WriteOptions{Dir: dir, ProjectName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !wroteSetup || !wroteConfig {
		t.Errorf("first write: wroteSetup=%v wroteConfig=%v", wroteSetup, wroteConfig)
	}
	setupPath := filepath.Join(dir, ".initiat", "setup.yml")
	if _, err := os.Stat(setupPath); os.IsNotExist(err) {
		t.Error("setup.yml was not created")
	}
	configPath := filepath.Join(dir, ".initiat", "config.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.yml was not created")
	}
}

func TestWrite_Idempotent_NoForce(t *testing.T) {
	dir := t.TempDir()
	cfg := &setup.SetupConfig{Version: 1, Name: "test", Defaults: &setup.Defaults{Timeout: "5m"}}
	_, _, err := Write(cfg, WriteOptions{Dir: dir, ProjectName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	wroteSetup, wroteConfig, err := Write(cfg, WriteOptions{Dir: dir, ProjectName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if wroteSetup || wroteConfig {
		t.Errorf("second write without force: wroteSetup=%v wroteConfig=%v", wroteSetup, wroteConfig)
	}
}

func TestWrite_ForceOverwritesSetup(t *testing.T) {
	dir := t.TempDir()
	cfg := &setup.SetupConfig{Version: 1, Name: "test", Defaults: &setup.Defaults{Timeout: "5m"}}
	_, _, err := Write(cfg, WriteOptions{Dir: dir, ProjectName: "test"})
	if err != nil {
		t.Fatal(err)
	}
	wroteSetup, _, err := Write(cfg, WriteOptions{Dir: dir, ProjectName: "test", ForceSetup: true})
	if err != nil {
		t.Fatal(err)
	}
	if !wroteSetup {
		t.Error("ForceSetup: expected wroteSetup true")
	}
}

func TestSetupExists(t *testing.T) {
	dir := t.TempDir()
	exists, err := SetupExists(dir)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("SetupExists should be false when no file")
	}
	if err := os.MkdirAll(filepath.Join(dir, ".initiat"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".initiat", "setup.yml"), []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	exists, err = SetupExists(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("SetupExists should be true when file exists")
	}
}
