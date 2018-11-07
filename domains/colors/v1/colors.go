// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package colorsv1

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"lab.wtfteam.pro/wtfteam/lbtds/internal/config"
)

var (
	colorsModuleLog zerolog.Logger
)

func initColors() {
	colorsModuleLog = domainLog.With().Str("module", "colors").Logger()
	colorsModuleLog.Info().Msg("Initializing Colors storage...")

	ColorChanged = make(chan bool)
}

func fallbackToFirstColor() {
	if len(c.Config.Colors) > 0 {
		err := SetCurrentColor(c.Config.Colors[0].Name)
		if err != nil {
			colorsModuleLog.Warn().Err(err).Msgf("Failed to change color to %s", c.Config.Colors[0].Name)
		}
	}
}

func colorExists(color string) bool {
	for i := range c.Config.Colors {
		if c.Config.Colors[i].Name == color {
			return true
		}
	}

	return false
}

// GetCurrentColorConfiguration gets configuration for current color
func GetCurrentColorConfiguration() *config.Color {
	if currentColor != "" {
		for i := range c.Config.Colors {
			if c.Config.Colors[i].Name == currentColor {
				return &c.Config.Colors[i]
			}
		}
	}

	return nil
}

// GetCurrentColorName returns current color name
func GetCurrentColorName() string {
	currentColorMutex.Lock()
	defer currentColorMutex.Unlock()
	return currentColor
}

// GetCurrentColor gets current color for application
func GetCurrentColor() string {
	if currentColor == "" {
		normalizedColorsPath, _ := filepath.Abs(c.Config.Proxy.ColorFile)
		c.Logger.Debug().Msgf("Current color file path: %s", normalizedColorsPath)

		colorsData, err := ioutil.ReadFile(normalizedColorsPath)
		if err != nil {
			fallbackToFirstColor()
		} else {
			if colorExists(string(colorsData)) {
				currentColor = string(colorsData)
			} else {
				colorsModuleLog.Warn().Msgf("Unexpected color in current colors file: %s", string(colorsData))
				fallbackToFirstColor()
			}
		}
		ColorChanged <- true
	}
	return currentColor
}

// SetCurrentColor sets current color for application
func SetCurrentColor(color string) error {
	var err error
	currentColorMutex.Lock()
	defer currentColorMutex.Unlock()
	if colorExists(color) {
		currentColor = color

		normalizedColorsPath, _ := filepath.Abs(c.Config.Proxy.ColorFile)

		colorsFile, err := os.OpenFile(normalizedColorsPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			colorsModuleLog.Panic().Err(err).Msg("Failed to open current color file or create one")
			return err
		}
		defer colorsFile.Close()
		err = colorsFile.Truncate(0)
		if err != nil {
			colorsModuleLog.Panic().Err(err).Msg("Failed to truncate current color file")
			return err
		}
		_, err = colorsFile.Write([]byte(color))
		if err != nil {
			colorsModuleLog.Warn().Err(err).Msg("Failed to write current color to file")
			return err
		}

		colorsModuleLog.Info().Msgf("Current color changed to %s", currentColor)

		ColorChanged <- true
	} else {
		colorsModuleLog.Warn().Msgf("There is no such color in configuration: %s", color)
		err = errors.New("Invalid color name")
	}

	return err
}
