// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package config

// Proxy tells LBTDS where to seek colors config
type Proxy struct {
	StorageType string `yaml:"storage_type"`
	ColorFile   string `yaml:"color_file"`
	PIDFile     string `yaml:"pid_file,omitempty"`
}
