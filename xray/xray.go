package xray

import (
	"bytes"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

//go:linkname decodeSubChunk github.com/df-mc/dragonfly/server/world/chunk.decodeSubChunk
func decodeSubChunk(buf *bytes.Buffer, c *chunk.Chunk, index *byte, e chunk.Encoding) (*chunk.SubChunk, error)

func init() {
	rid, ok := chunk.StateToRuntimeID("minecraft:air", nil)
	if !ok {
		panic("cannot find air runtime ID")
	}
	airRid = rid
	rid, ok = chunk.StateToRuntimeID("minecraft:stone", nil)
	if !ok {
		panic("cannot find air runtime ID")
	}
	stoneRid = rid
}

var (
	airRid   uint32
	stoneRid uint32
)

func (p *Protection) meshSubchunk(sub *chunk.SubChunk, buf *bytes.Buffer, pos protocol.SubChunkPos) bool {
	if sub.Empty() {
		return false
	}

	var meshed bool

	enc := nbt.NewEncoderWithEncoding(buf, nbt.LittleEndian)
	// filtering block entities
	for ent := range decodeBlockEntities(buf) {
		x, y, z := byte(ent.Pos[0]), byte(ent.Pos[1]), byte(ent.Pos[2])
		bl := sub.Block(x, y, z, 0)
		if !p.blocksToHide[bl] || !p.mask(sub, x, y, z) {
			_ = enc.Encode(ent.Data)
		} else {
			blockPos := cube.Pos{
				int(pos.X()) + int(x),
				int(pos.Y()) + int(y),
				int(pos.Z()) + int(z),
			}
			p.storage.Store(blockPos, bl, ent.Data)
		}
	}

	// hiding blocks
	for x := range byte(16) {
		for y := range byte(16) {
			for z := range byte(16) {
				bl := sub.Block(x, y, z, 0)
				if p.blocksToHide[bl] && p.mask(sub, x, y, z) {
					sub.SetBlock(x, y, z, 0, stoneRid)
					blockPos := cube.Pos{
						int(pos.X()) + int(x),
						int(pos.Y()) + int(y),
						int(pos.Z()) + int(z),
					}
					p.storage.Store(blockPos, bl, nil)
					meshed = true
				}
			}
		}
	}

	return meshed
}

func (p *Protection) mask(sub *chunk.SubChunk, x, y, z byte) bool {
	return !p.transparentBlocks[sub.Block(x, y, z+1, 0)] &&
		!p.transparentBlocks[sub.Block(x, y, z-1, 0)] &&
		!p.transparentBlocks[sub.Block(x, y+1, z, 0)] &&
		!p.transparentBlocks[sub.Block(x, y-1, z, 0)] &&
		!p.transparentBlocks[sub.Block(x+1, y, z, 0)] &&
		!p.transparentBlocks[sub.Block(x-1, y, z, 0)]
}
