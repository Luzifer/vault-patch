package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/vault/api"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"

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

	data, err := ioutil.ReadFile(vf)
	if err != nil {
		return ""
	}

	return string(data)
}

func loadConfig() error {
	rconfig.SetVariableDefaults(map[string]string{
		"vault-token": vaultTokenFromDisk(),
	})
	if err := rconfig.Parse(&cfg); err != nil {
		return err
	}

	if cfg.VersionAndExit {
		fmt.Printf("vault-patch %s\n", version)
		os.Exit(0)
	}

	if logLevel, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(logLevel)
	} else {
		return fmt.Errorf("Unable to parse log level: %s", err)
	}

	return nil
}

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Unable to load CLI config: %s", err)
	}

	if len(rconfig.Args()) < 2 {
		log.Fatalf("Usage: vault-patch [options] path data")
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.VaultAddress,
	})
	if err != nil {
		log.Fatalf("Unable to create Vault client: %s", err)
	}

	client.SetToken(cfg.VaultToken)

	key := rconfig.Args()[1]

	s, err := client.Logical().Read(key)
	if err != nil || s == nil {
		log.Fatalf("Could not read key %q from vault: %s", key, err)
	}

	data := s.Data
	if data == nil {
		data = make(map[string]interface{})
	}

	for k, v := range env.ListToMap(rconfig.Args()[2:len(rconfig.Args())]) {
		data[k] = v
	}

	if _, err := client.Logical().Write(key, data); err != nil {
		log.Fatalf("Could not write data to key %q: %s", key, err)
	}

	log.Printf("Data successfully written to key %q", key)
}
