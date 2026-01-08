package db

// Storage implements Fosite storage using a database backend.
type Storage struct{}

// NewFositeStorage creates a new Storage instance.
func NewFositeStorage() *Storage { return &Storage{} }
