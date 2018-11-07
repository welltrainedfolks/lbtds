// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package context

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

/* exported.go */

func TestNewContext(t *testing.T) {
	c := NewContext()
	require.NotNil(t, c)
}

/* context.go */

func TestInit(t *testing.T) {
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)
}

func TestInitConfigurationDefaultFilePath(t *testing.T) {
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()

	// Why "False?"
	// Because on run time the default config path is "./lbtds.yaml". And there
	// is no such file in "context/" directory.
	// There is one proper configuration file, stored at examples/two-colors-two-servers/
	// and it will be used later
	require.False(t, result)
}

func TestInitConfigurationWithWrongFilePath(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "all your base are belong to us")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.False(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestInitConfigurationWithRightFilePath(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

/* pid_files.go */

func TestCheckPIDFile(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.CheckPIDFile()
	require.True(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestCheckPIDFileWhenAlreadyExists(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.CheckPIDFile()
	require.False(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestCheckPIDFileWhenWrongPIDFilePath(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-wrong-pid-file.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.CheckPIDFile()
	require.False(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestRemovePIDFile(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.RemovePIDFile()
	require.True(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestRemovePIDFileWhenPIDFileNotFound(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.RemovePIDFile()
	require.False(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestRemovePIDFileWhenWrongPIDFilePath(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-wrong-pid-file.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.RemovePIDFile()
	require.False(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestRemovePIDFileWhenPIDFileHaveWrondProcessID(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	brokenPIDFile, err := os.OpenFile(c.getPIDFilePath(), os.O_RDWR|os.O_CREATE, os.ModePerm)
	require.Nil(t, err)
	require.FileExists(t, c.getPIDFilePath())
	defer brokenPIDFile.Close()

	_, err = brokenPIDFile.Write([]byte("1"))
	require.Nil(t, err)

	result = c.RemovePIDFile()
	require.False(t, result)

	// Clean wrong file from filesystem
	err = os.Remove(c.getPIDFilePath())
	require.Nil(t, err)

	_, err = ioutil.ReadFile(c.getPIDFilePath())
	require.NotNil(t, err)
	os.Unsetenv("LBTDS_CONFIG")
}

/* api_server.go */

func TestInitAPIServer(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.CheckPIDFile()
	require.True(t, result)

	c.InitAPIServer()
	require.NotNil(t, c.APIServerMux)
	require.NotNil(t, c.APIServer)

	result = c.RemovePIDFile()
	require.True(t, result)
	os.Unsetenv("LBTDS_CONFIG")
}

func TestAPIServerStartAndShutdown(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.CheckPIDFile()
	require.True(t, result)

	c.InitAPIServer()
	require.NotNil(t, c.APIServerMux)
	require.NotNil(t, c.APIServer)

	c.StartAPIServer()
	require.True(t, c.APIServerUp)

	err := c.checkAPIHealth()
	require.Nil(t, err)

	c.ShutdownAPIServer()
	require.False(t, c.APIServerUp)

	c.RemovePIDFile()
	os.Unsetenv("LBTDS_CONFIG")
}

/* shutdown.go */

func TestIsShuttingDown(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()

	require.NotNil(t, c.Logger)
	require.NotNil(t, c.RandomSource)
	require.False(t, c.inShutdown)
	require.False(t, c.APIServerUp)

	result := c.InitConfiguration()
	require.True(t, result)

	result = c.IsShuttingDown()
	require.False(t, result)

	c.SetShutdown()
	result = c.IsShuttingDown()
	require.True(t, result)

	os.Unsetenv("LBTDS_CONFIG")
}

func TestShutdown(t *testing.T) {
	os.Setenv("LBTDS_CONFIG", "../internal/testshelpers/config_templates/lbtds-valid.yaml")
	c := NewContext()
	c.Init()
	c.InitConfiguration()
	c.InitAPIServer()
	c.StartAPIServer()
	c.SetShutdown()
	c.Shutdown()

	require.True(t, c.IsShuttingDown())
	require.False(t, c.APIServerUp)

	os.Unsetenv("LBTDS_CONFIG")
}
