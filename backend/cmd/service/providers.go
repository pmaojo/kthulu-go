package main

import (
	"go.uber.org/fx"
)

// RepositorySet gathers repository providers for the service command.
// Repositories are wired via ModuleSet's provider map for each active module.
var RepositorySet = fx.Options()

// ServiceSet gathers service providers based on use cases.
var ServiceSet = fx.Options()
