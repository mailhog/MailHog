package config

import (
	"flag"

	"github.com/ian-kent/envconf"
)

func DefaultConfig() *Config {
	return &Config{
		AuthFile: "",
	}
}

type Config struct {
	AuthFile string
}

var cfg = DefaultConfig()

func Configure() *Config {
	return cfg
}

func RegisterFlags() {
	flag.StringVar(&cfg.AuthFile, "auth-file", envconf.FromEnvP("MH_AUTH_FILE", "").(string), "A username:bcryptpw mapping file")
}
