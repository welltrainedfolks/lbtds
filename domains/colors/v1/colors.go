// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package colorsv1

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

var (
	colorsModuleLog zerolog.Logger
)

func initColors() {
	colorsModuleLog = domainLog.With().Str("module", "colors").Logger()
	colorsModuleLog.Info().Msg("Initializing Colors storage...")

	ColorChanged = make(chan bool)
}

// GetCurrentColor gets current color for application
func GetCurrentColor() string {
	if currentColor == "" {
		normalizedColorsPath, err := filepath.Abs(c.Config.Proxy.ColorFile)
		if err != nil {
			c.Logger.Panic().Msgf("Failed to normalize current color file path. Path supplied: '%s'", c.Config.Proxy.ColorFile)
		}
		c.Logger.Debug().Msgf("Current color file path: %s", normalizedColorsPath)

		colorsData, err := ioutil.ReadFile(normalizedColorsPath)
		if err != nil {
			idx := 0
			for color := range c.Config.Colors {
				if idx == 0 {
					err = SetCurrentColor(color)
					if err != nil {
						colorsModuleLog.Warn().Err(err).Msgf("Failed to change color to %s", color)
					}
				}
				idx++
			}
		} else {
			currentColor = string(colorsData)
		}
	}
	return currentColor
}

// SetCurrentColor sets current color for application
func SetCurrentColor(color string) error {
	var err error
	currentColorMutex.Lock()
	defer currentColorMutex.Unlock()
	if c.Config.Colors[color] != nil {
		currentColor = color

		normalizedColorsPath, err := filepath.Abs(c.Config.Proxy.ColorFile)
		if err != nil {
			colorsModuleLog.Panic().Msgf("Failed to normalize current color file path. Path supplied: '%s'", c.Config.Proxy.ColorFile)
		}

		colorsFile, err := os.OpenFile(normalizedColorsPath, os.O_RDWR|os.O_CREATE, 0755)
		defer colorsFile.Close()
		if err != nil {
			colorsModuleLog.Panic().Err(err).Msg("Failed to open current color file or create one")
		}
		err = colorsFile.Truncate(0)
		if err != nil {
			colorsModuleLog.Panic().Err(err).Msg("Failed to truncate current color file")
		}
		_, err = colorsFile.Write([]byte(color))
		if err != nil {
			colorsModuleLog.Warn().Err(err).Msg("Failed to write current color to file")
		}

		colorsModuleLog.Info().Msgf("Current color changed to %s", currentColor)

		ColorChanged <- true
	} else {
		colorsModuleLog.Warn().Msgf("There is no such color in configuration: %s", color)
		err = errors.New("Invalid color name")
	}

	return err
}
