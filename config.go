package jump

import (
	"github.com/BurntSushi/toml"
)

var defaultConfig = Config{
	DataDir: "./data",
	TLSDir:  "./tls",
	Port:    8080,
	TLSPort: 10443,
}

// NewConfig will create a new configuration from a TOML file location
func NewConfig(filename string) (cp *Config, err error) {
	var c Config
	if _, err = toml.DecodeFile(filename, &c); err != nil {
		return
	}

	cp = &c
	return
}

// Config represents a jump configuration
type Config struct {
	DataDir string `toml:"dataDir"`
	TLSDir  string `toml:"tlsDir"`

	Port    uint16 `toml:"port"`
	TLSPort uint16 `toml:"tlsPort"`
}
