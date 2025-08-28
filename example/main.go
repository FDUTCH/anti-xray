package main

import (
	"fmt"

	"github.com/FDUTCH/anti-xray/xray"
	"github.com/bedrock-gophers/intercept/intercept"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"log/slog"
	"os"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	chat.Global.Subscribe(chat.StdoutSubscriber{})
	conf, err := readConfig(slog.Default())
	if err != nil {
		panic(err)
	}

	conf.Listeners = intercept.WrapListeners(conf.Listeners)
	conf.Listeners = xray.WrapListeners(conf.Listeners)
	srv := conf.New()
	srv.CloseOnProgramEnd()

	protection := xray.NewProtection(nil, nil)

	// protecting blocks
	protection.HideBlock(block.AncientDebris{})
	protection.HideBlock(block.DiamondOre{})
	protection.HideBlock(block.GoldOre{})
	protection.HideBlock(block.IronOre{})
	protection.HideBlock(block.Chest{})

	// setting transparent blocks
	for _, bl := range world.Blocks() {
		if _, ok := bl.(block.LightDiffuser); ok {
			protection.SetTransparent(bl)
		}
	}

	intercept.Hook(packetHandler{protection})

	srv.Listen()
	for p := range srv.Accept() {
		_ = p
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log *slog.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
}

type packetHandler struct {
	x *xray.Protection
}

func (p packetHandler) HandleClientPacket(ctx *intercept.Context, pk packet.Packet) {}

func (p packetHandler) HandleServerPacket(ctx *intercept.Context, pk packet.Packet) {
	if sub, ok := pk.(*packet.SubChunk); ok {
		p.x.HandleSubchunk(sub)
	}
}
