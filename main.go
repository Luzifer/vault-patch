package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/env"
	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg struct {
		LogLevel       string `flag:"log-level" default:"info" description:"Verbosity of logs to use (debug, info, warning, error, ...)"`
		VaultAddress   string `flag:"vault-addr" env:"VAULT_ADDR" default:"https://127.0.0.1:8200" description:"Vault API address"`
		VaultToken     string `flag:"vault-token" env:"VAULT_TOKEN" vardefault:"vault-token" description:"Specify a token to use instead of app-id auth"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Print program version and exit"`
	}

	version = "dev"
)

func vaultTokenFromDisk() string {
	vf, err := homedir.Expand("~/.vault-token")
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(vf) //#nosec G304 // Intended to load file from disk
	if err != nil {
		return ""
	}

	return string(data)
}

func initApp() (err error) {
	rconfig.SetVariableDefaults(map[string]string{"vault-token": vaultTokenFromDisk()})
	rconfig.AutoEnv(true)

	if err = rconfig.Parse(&cfg); err != nil {
		return fmt.Errorf("parsing CLI options: %w", err)
	}

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parsing log level: %w", err)
	}

	logrus.SetLevel(logLevel)

	return nil
}

func main() {
	var err error

	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		fmt.Printf("vault-patch %s\n", version) //nolint:forbidigo
		os.Exit(0)
	}

	if len(rconfig.Args()) < 2 { //nolint:mnd
		logrus.Fatal("Usage: vault-patch [options] <path> <data>")
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.VaultAddress,
	})
	if err != nil {
		logrus.WithError(err).Fatal("creating Vault client")
	}

	client.SetToken(cfg.VaultToken)

	key := rconfig.Args()[1]

	s, err := client.Logical().Read(key)
	if err != nil || s == nil {
		logrus.WithError(err).Fatal("reading key from vault")
	}

	data := s.Data
	if data == nil {
		data = make(map[string]any)
	}

	for k, v := range env.ListToMap(rconfig.Args()[2:]) {
		data[k] = v
	}

	if _, err := client.Logical().Write(key, data); err != nil {
		logrus.WithError(err).Fatal("writing data to key")
	}

	logrus.Print("data successfully written to key")
}
