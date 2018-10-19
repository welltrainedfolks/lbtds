// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package config

// Struct is a main configuration structure that holds all other
// structs within.
type Struct struct {
	API    API                      `yaml:"api"`
	Proxy  Proxy                    `yaml:"proxy"`
	Colors map[string][]ProxyConfig `yaml:"colors"`
}
