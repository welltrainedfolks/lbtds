// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package context

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// CheckPIDFile checks for existing PID file and creates one if possible
func (c *Context) CheckPIDFile() bool {
	normalizedPIDFilePath := c.getPIDFilePath()

	c.Logger.Debug().Msgf("PID file path: %s", normalizedPIDFilePath)

	// We need to fail here, not pass
	currentPID, err := ioutil.ReadFile(normalizedPIDFilePath)
	if err != nil {
		// Write PID to new file
		newPIDfile, err := os.OpenFile(normalizedPIDFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			c.Logger.Error().Err(err).Msg("Failed to create PID file")
			return false
		}
		defer newPIDfile.Close()

		_, err = newPIDfile.Write([]byte(strconv.Itoa(os.Getppid())))
		if err != nil {
			c.Logger.Error().Err(err).Msg("Failed to write PID file")
			return false
		}
	} else {
		// PID file exists
		c.Logger.Error().Msgf("There is already LBTDS instance with the same configuration running at PID %s. Stop it or remove PID file if instance already stopped.", string(currentPID))
		return false
	}

	return true
}

// RemovePIDFile removes PID file on stop
func (c *Context) RemovePIDFile() bool {
	normalizedPIDFilePath := c.getPIDFilePath()

	parentPID, err := ioutil.ReadFile(normalizedPIDFilePath)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to read PID file")
		return false
	}

	if string(parentPID) != strconv.Itoa(os.Getppid()) {
		c.Logger.Error().Err(err).Msgf("PID file contains wrong PID: expected %d, but got %s", os.Getppid(), parentPID)
		return false
	}

	err = os.Remove(normalizedPIDFilePath)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Failed to remove PID file")
		return false
	}

	return true
}

// getPIDFilePath returns PID file path based on config and OS
func (c *Context) getPIDFilePath() string {
	var pidFile string
	if c.Config.Proxy.PIDFile != "" {
		pidFile = c.Config.Proxy.PIDFile
	} else {
		switch runtime.GOOS {
		case "windows":
			c.Logger.Panic().Msg("LBTDS doesn't support Windows at this time. Please, read CONTRIBUTING.md for adding Windows support if you're interested in it.")
		case "darwin":
			pidFile = "/usr/local/var/run/lbtds.lock"
		case "linux":
			pidFile = "/var/run/lbtds.lock"
		default:
			pidFile = "/var/run/lbtds.lock"
		}
	}

	normalizedPIDFilePath, _ := filepath.Abs(pidFile)

	return normalizedPIDFilePath
}
