// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov

package colorsv1

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"lab.wtfteam.pro/wtfteam/lbtds/internal/testshelpers"
)

func mockupDispatch() {
	go func() {
		for <-ColorChanged {
			fmt.Println("Got signal from ColorChanged channel")
		}
	}()
}

/* exported.go */

func TestInitialize(t *testing.T) {
	// Before running this bunch of tests we need to clear temporary files
	err := os.Remove("/tmp/lbtds-test-current")
	if err != nil {
		fmt.Println("Failed to erase files from previous test: " + err.Error())
	}

	testshelpers.InitializeConfiguration("../../../", "lbtds-other-color-path")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	testshelpers.FlushConfiguration("lbtds-other-color-path")
}

/* colors.go */

func TestColorExist(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	result := colorExists("green")
	require.True(t, result)
	result = colorExists("violet")
	require.False(t, result)

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestGetCurrentColorWhenColorFileAbsent(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	// In this configuration, when no color file exists, first color will be
	// chosen. It is "green"
	GetCurrentColor()
	require.Equal(t, "green", currentColor)
	require.Equal(t, currentColor, GetCurrentColorName())

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err := os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestGetCurrentColorConfiguration(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	// In this configuration, when no color file exists, first color will be
	// chosen. It is "green"
	GetCurrentColor()
	require.Equal(t, "green", currentColor)
	require.Equal(t, currentColor, GetCurrentColorName())

	result := GetCurrentColorConfiguration()
	require.Equal(t, currentColor, result.Name)

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err := os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestGetCurrentColorWhenColorFileExist(t *testing.T) {
	neededColor := "blue"
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	normalizedColorsPath, _ := filepath.Abs(c.Config.Proxy.ColorFile)
	colorsFile, err := os.OpenFile(normalizedColorsPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer colorsFile.Close()
	err = colorsFile.Truncate(0)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = colorsFile.Write([]byte(neededColor))
	if err != nil {
		t.Fatal(err.Error())
	}

	mockupDispatch()

	GetCurrentColor()
	require.Equal(t, neededColor, currentColor)

	// Clear cache for other tests
	err = os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestGetCurrentColorWhenColorFileExistWithWrongValue(t *testing.T) {
	neededColor := "violet"
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	normalizedColorsPath, _ := filepath.Abs(c.Config.Proxy.ColorFile)
	colorsFile, err := os.OpenFile(normalizedColorsPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer colorsFile.Close()
	err = colorsFile.Truncate(0)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = colorsFile.Write([]byte(neededColor))
	if err != nil {
		t.Fatal(err.Error())
	}

	mockupDispatch()

	GetCurrentColor()
	require.Equal(t, "green", currentColor)

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err = os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

/* api.go */

func TestReceiveColorChangeRequest(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	GetCurrentColor()
	require.Equal(t, "green", currentColor)

	newColorRequest := &colorRequestParams{
		Color: "blue",
	}
	newColorRequestData, _ := json.Marshal(&newColorRequest)

	replyBody, replyCode := testshelpers.HTTPTestRequest(t, c, newColorRequestData, nil, "POST", "v1", "/color", ChangeColor)
	assert.NotEmpty(t, replyBody)
	assert.Equal(t, "Color changed\n", string(replyBody))
	require.Equal(t, 200, replyCode)

	require.Equal(t, "blue", currentColor)

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err := os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestReceiveColorChangeRequestWithWrongBody(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	GetCurrentColor()
	require.Equal(t, "green", currentColor)

	newColorRequestData := []byte("abraquadabra")

	replyBody, replyCode := testshelpers.HTTPTestRequest(t, c, newColorRequestData, nil, "POST", "v1", "/color", ChangeColor)
	assert.NotEmpty(t, replyBody)
	assert.Equal(t, "Invalid request body\n", string(replyBody))
	require.Equal(t, 400, replyCode)

	require.Equal(t, "green", currentColor)

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err := os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestReceiveColorChangeRequestWithWrongColor(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	GetCurrentColor()
	require.Equal(t, "green", currentColor)

	newColorRequest := &colorRequestParams{
		Color: "velvet",
	}
	newColorRequestData, _ := json.Marshal(&newColorRequest)

	replyBody, replyCode := testshelpers.HTTPTestRequest(t, c, newColorRequestData, nil, "POST", "v1", "/color", ChangeColor)
	assert.NotEmpty(t, replyBody)
	assert.Equal(t, "Invalid color\n", string(replyBody))
	require.Equal(t, 404, replyCode)

	require.Equal(t, "green", currentColor)

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err := os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestReceiveColorChangeRequestWithWrongMethod(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	GetCurrentColor()
	require.Equal(t, "green", currentColor)

	newColorRequest := &colorRequestParams{
		Color: "blue",
	}
	newColorRequestData, _ := json.Marshal(&newColorRequest)

	replyBody, replyCode := testshelpers.HTTPTestRequest(t, c, newColorRequestData, nil, "PUT", "v1", "/color", ChangeColor)
	assert.NotEmpty(t, replyBody)
	assert.Equal(t, "404 page not found\n", string(replyBody))
	require.Equal(t, 404, replyCode)

	require.Equal(t, "green", currentColor)

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err := os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}

func TestReceiveColorChangeRequestWithEmptyBody(t *testing.T) {
	testshelpers.InitializeConfiguration("../../../", "lbtds-valid")
	c := testshelpers.InitializeContext()
	Initialize(c)

	require.NotNil(t, ColorChanged)
	require.Empty(t, currentColor)

	mockupDispatch()

	GetCurrentColor()
	require.Equal(t, "green", currentColor)

	replyBody, replyCode := testshelpers.HTTPTestRequest(t, c, nil, nil, "POST", "v1", "/color", ChangeColor)
	assert.NotEmpty(t, replyBody)
	assert.Equal(t, "Invalid request body\n", string(replyBody))
	require.Equal(t, 400, replyCode)

	require.Equal(t, "green", currentColor)

	c.SetShutdown()
	c.Shutdown()

	// Clear cache for other tests
	err := os.Remove(c.Config.Proxy.ColorFile)
	require.Nil(t, err)
	currentColor = ""

	testshelpers.FlushConfiguration("lbtds-valid")
}
