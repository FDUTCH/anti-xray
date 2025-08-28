package xray

import (
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/dgraph-io/ristretto/v2"
)

// Protection is a handler for the SubChunk packet.
type Protection struct {
	blocksToHide      []bool
	transparentBlocks []bool
	cache             *ristretto.Cache[[]byte, []byte]
	storage           Storage
	airRid, stoneRid  uint32
}

// NewProtection created new instance.
// You may pass nil instead of cache but this may impact performance.
func NewProtection(cache *ristretto.Cache[[]byte, []byte], storage Storage) *Protection {
	world_finaliseBlockRegistry()

	if storage == nil {
		storage = NopStorage{}
	}

	stoneRid, _ := chunk.StateToRuntimeID("minecraft:stone", nil)
	airRid, _ := chunk.StateToRuntimeID("minecraft:air", nil)
	count := len(world.Blocks())
	return &Protection{
		blocksToHide:      make([]bool, count),
		transparentBlocks: make([]bool, count),
		cache:             cache,
		stoneRid:          stoneRid,
		airRid:            airRid,
		storage:           storage,
	}
}

// HideBlock sets the block as hidden.
func (p *Protection) HideBlock(block world.Block) {
	idx := world.BlockRuntimeID(block)
	if len(p.blocksToHide) < int(idx+1) {
		return
	}
	p.blocksToHide[idx] = true
}

// SetTransparent sets the block as transparent.
func (p *Protection) SetTransparent(block world.Block) {
	idx := world.BlockRuntimeID(block)
	if len(p.transparentBlocks) < int(idx+1) {
		return
	}
	p.transparentBlocks[idx] = true
}

// cacheEnabled ...
func (p *Protection) cacheEnabled() bool {
	return p.cache != nil
}

//go:linkname world_finaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func world_finaliseBlockRegistry()
