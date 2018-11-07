// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package config

// Color represents configuration for single color
type Color struct {
	Name     string          `yaml:"name"`
	Backends []BackendConfig `yaml:"backends"`
}

// BackendConfig represents configuration for single backend endpoint
type BackendConfig struct {
	// Type can be HTTP or TCP.
	Type string `yaml:"type"`
	// IP and port this proxy will listen on.
	ListenOn string `yaml:"listen_on"`
	// For HTTP source is a HTTP hostname for which request was received.
	Source string `yaml:"source"`
	// Backend servers.
	Destinations []string `yaml:"destinations"`
}
