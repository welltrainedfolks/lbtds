// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package config

// ProxyConfig represents configuration for single proxy endpoint
type ProxyConfig struct {
	Source       string   `yaml:"source"`
	Destinations []string `yaml:"destinations"`
}
