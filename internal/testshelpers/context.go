// LBTDS — Load balancer that doesn't suck
// Copyright (c) 2018 Vladimir "fat0troll" Hodakov
// Copyright (c) 2018 Stanislav N. aka pztrn

package testshelpers

import (
	"lab.wtfteam.pro/wtfteam/lbtds/context"
)

// InitializeContext initializes context for tests.
func InitializeContext() *context.Context {
	c := context.NewContext()
	c.Init()
	c.InitConfiguration()
	c.InitAPIServer()

	return c
}
