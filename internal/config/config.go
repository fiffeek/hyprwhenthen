package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

type Config struct {
	cfg  *RawConfig
	path string
	mu   sync.RWMutex
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{
		cfg:  nil,
		path: path,
		mu:   sync.RWMutex{},
	}
	logrus.WithFields(logrus.Fields{"path": path}).Debug("Creating config wrapper")
	if err := cfg.Reload(); err != nil {
		return nil, fmt.Errorf("cant initialize config: %w", err)
	}
	return cfg, nil
}

func (c *Config) Get() *RawConfig {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cfg
}

func (c *Config) Reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cfg, err := Load(c.path)
	if err != nil {
		return fmt.Errorf("cant reload config from %s: %w", c.path, err)
	}
	c.cfg = cfg
	return nil
}

type RawConfig struct {
	Events    []*Event            `toml:"handler"`
	OnEvents  map[string][]*Event `toml:",skip"`
	EventKeys []string            `toml:",skip"`
	General   *GeneralSection     `toml:"general"`
}

type GeneralSection struct {
	Timeout *time.Duration `toml:"timeout"`
}

type Event struct {
	On      string         `toml:"on"`
	When    string         `toml:"when"`
	Then    string         `toml:"then"`
	Timeout *time.Duration `toml:"timeout"`
}

func Load(configPath string) (*RawConfig, error) {
	configPath = os.ExpandEnv(configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file %s not found", configPath)
	}

	logrus.WithFields(logrus.Fields{"expanded": configPath}).Debug("Expanded config path")

	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("cant convert config path to absolute path %w", err)
	}

	logrus.WithFields(logrus.Fields{"abs": absConfig}).Debug("Found absolute config path")

	// nolint:gosec
	contents, err := os.ReadFile(absConfig)
	if err != nil {
		return nil, fmt.Errorf("cant read config file %s: %w", absConfig, err)
	}
	logrus.Debugf("Config contents: %s", contents)

	var config RawConfig
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode TOML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	logrus.Debug("Config is valid")

	return &config, nil
}

func (r *RawConfig) Validate() error {
	if len(r.Events) == 0 {
		return fmt.Errorf("at least one event must be configured")
	}

	for i, event := range r.Events {
		if err := event.Validate(); err != nil {
			return fmt.Errorf("event %d validation failed: %w", i, err)
		}
	}

	r.OnEvents = make(map[string][]*Event)
	for _, event := range r.Events {
		r.OnEvents[event.On] = append(r.OnEvents[event.On], event)
	}

	r.EventKeys = make([]string, 0, len(r.OnEvents))
	for key := range r.OnEvents {
		r.EventKeys = append(r.EventKeys, key)
	}

	if r.General == nil {
		r.General = &GeneralSection{}
	}

	if err := r.General.Validate(); err != nil {
		return fmt.Errorf("general section validation failed: %w", err)
	}

	return nil
}

func (r *GeneralSection) Validate() error {
	if r.Timeout == nil {
		return errors.New("timeout has to be set")
	}
	if r.Timeout != nil && *r.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	return nil
}

func (r *Event) Validate() error {
	if r.On == "" {
		return errors.New("'on' field is required")
	}
	if r.When == "" {
		return errors.New("'when' field is required")
	}
	if r.Then == "" {
		return errors.New("'then' field is required")
	}
	if r.Timeout != nil && *r.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}
	return nil
}
