// Package config loads the application's runtime configuration from a YAML
// file on disk.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the server's runtime settings.
type Config struct {
	Addr         string `yaml:"addr"`
	ContentDir   string `yaml:"content_dir"`
	TemplatesDir string `yaml:"templates_dir"`
	StaticDir    string `yaml:"static_dir"`
	LogFile      string `yaml:"log_file"`
}

// Default returns the configuration used when no config file is present.
func Default() Config {
	return Config{
		Addr:         "127.0.0.1:8080",
		ContentDir:   "content",
		TemplatesDir: "templates",
		StaticDir:    "static",
		LogFile:      "medisite.log",
	}
}

// Load reads and parses the YAML config file at path, starting from
// Default() so any field the file omits keeps its default value. A missing
// file is not an error — Default() is returned as-is, so the app runs with
// sane defaults on a fresh checkout. A file that exists but fails to parse
// is a startup error.
func Load(path string) (Config, error) {
	cfg := Default()

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("reading config file %q: %w", path, err)
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config file %q: %w", path, err)
	}
	return cfg, nil
}
