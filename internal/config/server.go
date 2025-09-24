package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type ServerConfig struct {
	Addr string
}

func (c *ServerConfig) Validate() error {
	if c.Addr == "" {
		return errors.New("server address can't be empty")
	}

	if !strings.Contains(c.Addr, ":") {
		return fmt.Errorf("server address must contain port")
	}

	return nil
}

func LoadServerConfig() *ServerConfig {
	var addr string

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	return &ServerConfig{
		Addr: addr,
	}
}
