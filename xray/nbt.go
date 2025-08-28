package xray

import (
	"bytes"
	"io"
	"iter"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// decodeBlockEntities decodes blockEntities from buf.
func decodeBlockEntities(buf *bytes.Buffer) iter.Seq[chunk.BlockEntity] {
	dec := nbt.NewDecoderWithEncoding(io.LimitReader(buf, int64(buf.Len())), nbt.LittleEndian)

	return func(yield func(chunk.BlockEntity) bool) {
		for {
			be := chunk.BlockEntity{Data: make(map[string]any)}
			if err := dec.Decode(&be.Data); err != nil {
				return
			}
			be.Pos = blockPosFromNBT(be.Data)
			if !yield(be) {
				return
			}
		}
	}
}

// blockPosFromNBT returns block pos from nbt.
func blockPosFromNBT(data map[string]any) cube.Pos {
	x, _ := data["x"].(int32)
	y, _ := data["y"].(int32)
	z, _ := data["z"].(int32)
	return cube.Pos{int(x), int(y), int(z)}
}
