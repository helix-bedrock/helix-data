// Command joinpackets writes the registry packets BDS sends every joining
// player (item registry, creative content, crafting data, biome
// definitions, actor identifiers, and the voxel shapes and jigsaw
// structure data sent before StartGame) as their raw payloads, and StartGame's
// block definitions (the data-driven vanilla blocks), from a capture of
// a real join: a JSON-lines file with one {"p": packet class, "hex":
// header and payload} row per packet sent, in order.
//
// Each packet must decode with helix-bedrock/protocol, use every byte, and
// encode back to the same bytes, so the files are exactly what BDS sends
// and Helix can read them.
//
//	go run ./cmd/joinpackets -capture <join.jsonl> -out data/<version>/bds/packets
package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/helix-bedrock/protocol"
	"github.com/helix-bedrock/protocol/packet"
)

// files maps the BDS packet classes kept to their file names.
var files = map[string]string{
	"ItemRegistryPacket":              "item_registry.bin",
	"CreativeContentPacket":           "creative_content.bin",
	"CraftingDataPacket":              "crafting_data.bin",
	"BiomeDefinitionListPacket":       "biome_definition_list.bin",
	"AvailableActorIdentifiersPacket": "available_actor_identifiers.bin",
	"StartGamePacket":                 "start_game_blocks.bin",
	"VoxelShapesPacket":               "voxel_shapes.bin",
	"JigsawStructureDataPacket":       "jigsaw_structure_data.bin",
}

func main() {
	capture := flag.String("capture", "", "join capture (JSON lines)")
	out := flag.String("out", "", "output folder")
	flag.Parse()
	if *capture == "" || *out == "" {
		flag.Usage()
		os.Exit(2) //nolint:forbidigo // a command reports failure through its exit status
	}
	if err := run(*capture, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1) //nolint:forbidigo // a command reports failure through its exit status
	}
}

func run(capture, out string) error {
	f, err := os.Open(capture) //nolint:gosec // the path is one the person passed on the command line
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	raw := map[string][]byte{}
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 64<<20)
	for sc.Scan() {
		var row struct{ P, Hex string }
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			return fmt.Errorf("read capture: %w", err)
		}
		if _, keep := files[row.P]; !keep || row.Hex == "" {
			continue
		}
		if _, dup := raw[row.P]; dup {
			return fmt.Errorf("%s sent twice", row.P)
		}
		if raw[row.P], err = hex.DecodeString(row.Hex); err != nil {
			return fmt.Errorf("%s: %w", row.P, err)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	// Item stacks in the other packets name the shield by its ID, so the
	// item registry is read first.
	items, ok := raw["ItemRegistryPacket"]
	if !ok {
		return errors.New("no ItemRegistryPacket in the capture")
	}
	pk, _, err := decode(items, 0)
	if err != nil {
		return fmt.Errorf("ItemRegistryPacket: %w", err)
	}
	var shield int32
	for _, it := range pk.(*packet.ItemRegistry).Items {
		if it.Name == "minecraft:shield" {
			shield = int32(it.RuntimeID)
		}
	}
	if err := os.MkdirAll(out, 0o755); err != nil { //nolint:gosec // data folders are meant to be readable
		return err
	}
	for class, name := range files {
		b, ok := raw[class]
		if !ok {
			return fmt.Errorf("no %s in the capture", class)
		}
		pk, payload, err := decode(b, shield)
		if err != nil {
			return fmt.Errorf("%s: %w", class, err)
		}
		if sg, ok := pk.(*packet.StartGame); ok {
			// Only the block list: the rest is about the world and player
			// the capture was made with.
			var w bytes.Buffer
			protocol.Slice(protocol.NewWriter(&w, shield), &sg.Blocks)
			payload = w.Bytes()
		}
		if err := os.WriteFile(filepath.Join(out, name), payload, 0o644); err != nil { //nolint:gosec // data files are meant to be readable
			return err
		}
		fmt.Printf("%s: %d bytes\n", name, len(payload))
	}
	return nil
}

// decode reads one packet (header and payload), checks that it uses every
// byte and encodes back the same, and returns it with its payload.
func decode(b []byte, shield int32) (packet.Packet, []byte, error) {
	buf := bytes.NewBuffer(b)
	var h packet.Header
	if err := h.Read(buf); err != nil {
		return nil, nil, fmt.Errorf("header: %w", err)
	}
	payload := buf.Bytes()
	newPacket, ok := packet.NewServerPool()[h.PacketID]
	if !ok {
		return nil, nil, fmt.Errorf("unknown packet ID %d", h.PacketID)
	}
	pk := newPacket()
	r := bytes.NewBuffer(payload)
	if err := readPacket(pk, protocol.NewReader(r, shield, false)); err != nil {
		return nil, nil, err
	}
	if r.Len() != 0 {
		return nil, nil, fmt.Errorf("%d bytes left after decoding", r.Len())
	}
	var w bytes.Buffer
	pk.Marshal(protocol.NewWriter(&w, shield))
	if !bytes.Equal(w.Bytes(), payload) {
		return nil, nil, errors.New("does not encode back to the bytes BDS sent")
	}
	return pk, payload, nil
}

// readPacket decodes pk, turning a decoding panic of the reader into an
// error.
func readPacket(pk packet.Packet, r *protocol.Reader) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("decode: %v", v)
		}
	}()
	pk.Marshal(r)
	return nil
}
