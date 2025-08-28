package xray

import (
	"bytes"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// Protection is a handler for the SubChunk packet.
type Protection struct {
	blocksToHide      []bool
	transparentBlocks []bool
	cache             *ristretto.Cache[[]byte, []byte]
	storage           Storage
}

// NewProtection created new instance.
// You may pass nil instead of cache but this may impact performance.
func NewProtection(cache *ristretto.Cache[[]byte, []byte], storage Storage) *Protection {
	world_finaliseBlockRegistry()

	if storage == nil {
		storage = NopStorage{}
	}

	count := len(world.Blocks())
	return &Protection{
		blocksToHide:      make([]bool, count),
		transparentBlocks: make([]bool, count),
		cache:             cache,
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

// HandleSubchunk hides the blocks inside subchunk.
func (p *Protection) HandleSubchunk(sub *packet.SubChunk) {
	var c *chunk.Chunk
	cacheEnabled := p.cacheEnabled()
	for i, entry := range sub.SubChunkEntries {
		if entry.Result != protocol.SubChunkResultSuccess || len(entry.RawPayload) == 0 {
			continue
		}

		if cacheEnabled {
			if encoded, ok := p.cache.Get(entry.RawPayload); ok {
				sub.SubChunkEntries[i].RawPayload = encoded
				continue
			}
		}

		if c == nil {
			dim, _ := world.DimensionByID(int(sub.Dimension))
			c = chunk.New(airRid, dim.Range())
		}

		buf := bytes.NewBuffer(entry.RawPayload)
		var index byte
		decodedSC, err := decodeSubChunk(buf, c, &index, chunk.NetworkEncoding)
		if err != nil {
			panic(err)
		}

		c.Sub()[index] = decodedSC

		pos := protocol.SubChunkPos{sub.Position[0] + int32(entry.Offset[0]), sub.Position[1] + int32(entry.Offset[1]), sub.Position[2] + int32(entry.Offset[2])}
		if p.meshSubchunk(decodedSC, buf, pos) {
			result := append(chunk.EncodeSubChunk(c, chunk.NetworkEncoding, int(index)), buf.Bytes()...)
			if cacheEnabled {
				p.cache.Set(sub.SubChunkEntries[i].RawPayload, result, 0)
			}
			sub.SubChunkEntries[i].RawPayload = result
		}
	}
}

// cacheEnabled ...
func (p *Protection) cacheEnabled() bool {
	return p.cache != nil
}

//go:linkname world_finaliseBlockRegistry github.com/df-mc/dragonfly/server/world.finaliseBlockRegistry
func world_finaliseBlockRegistry()
