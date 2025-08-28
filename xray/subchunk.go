package xray

import (
	"bytes"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

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
