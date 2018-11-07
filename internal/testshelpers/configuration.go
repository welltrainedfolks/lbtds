// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package testshelpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// InitializeConfiguration creates temporary configuration with given name
// and returns true on success and false on failure.
func InitializeConfiguration(pathPrefix string, templateName string) bool {
	relativeFilePath := pathPrefix + "/internal/testshelpers/config_templates/" + templateName + ".yaml"

	absoluteTemplatePath, _ := filepath.Abs(relativeFilePath)

	exampleConfigData, err := ioutil.ReadFile(absoluteTemplatePath)
	if err != nil {
		fmt.Println("Failed to read template " + templateName + ": " + err.Error())
		return false
	}

	err = ioutil.WriteFile("/tmp/lbtds-test-"+templateName+".yaml", exampleConfigData, 0644)
	if err != nil {
		fmt.Println("Failed to write configuration file: " + err.Error())
		return false
	}

	os.Setenv("LBTDS_CONFIG", "/tmp/lbtds-test-"+templateName+".yaml")
	return true
}

// FlushConfiguration removes temporary configuration file from environment
// variable and disk
func FlushConfiguration(templateName string) bool {
	err := os.Remove("/tmp/lbtds-test-" + templateName + ".yaml")
	if err != nil {
		fmt.Println("Failed to remove configuration file /tmp/lbtds-test-" + templateName + ".yaml: " + err.Error())
		return false
	}

	os.Unsetenv("LBTDS_CONFIG")
	return true
}
