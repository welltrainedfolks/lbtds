// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package config

// API represents LBTDS API configuration
type API struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}
