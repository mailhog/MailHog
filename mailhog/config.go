package mailhog

func DefaultConfig() *Config {
	return &Config{
		BindAddr: "0.0.0.0:1025",
		Hostname: "mailhog.example",
	}
}

type Config struct {
	BindAddr string
	Hostname string
}
