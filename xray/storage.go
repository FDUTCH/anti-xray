package xray

import "github.com/df-mc/dragonfly/server/block/cube"

// Storage stores all masked blocks, with their nbt data.
type Storage interface {
	Store(pos cube.Pos, rid uint32, data map[string]any)
}

// NopStorage ...
type NopStorage struct{}

// Store ...
func (n NopStorage) Store(pos cube.Pos, rid uint32, data map[string]any) {}

// Block represent encoded block.
type Block struct {
	Rid  uint32
	Data map[string]any
}
