package blockNBT

import (
	"fmt"
	blockNBT_depends "phoenixbuilder/fastbuilder/bdump/blockNBT/depends"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/io/commands"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/chunk"
	"strings"
)

// convert map[string]interface{} into struct named commandBlock
func parseCommandBlock(cb map[string]interface{}) (blockNBT_depends.CommandBlock, error) {
	var normal bool = false
	var command string = ""
	var customName string = ""
	var lastOutput string = ""
	var tickDelay int32 = int32(0)
	var executeOnFirstTick bool = true
	var trackOutput bool = true
	var conditionalMode bool = false
	var auto bool = true
	// 初始化
	_, ok := cb["Command"]
	if ok {
		command, normal = cb["Command"].(string)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"Command\"]; cb = %#v", cb)
		}
	}
	// Command
	_, ok = cb["CustomName"]
	if ok {
		customName, normal = cb["CustomName"].(string)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"CustomName\"]; cb = %#v", cb)
		}
	}
	// CustomName
	_, ok = cb["LastOutput"]
	if ok {
		lastOutput, normal = cb["LastOutput"].(string)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"LastOutput\"]; cb = %#v", cb)
		}
	}
	// LastOutput
	_, ok = cb["TickDelay"]
	if ok {
		tickDelay, normal = cb["TickDelay"].(int32)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"TickDelay\"]; cb = %#v", cb)
		}
	}
	// TickDelay
	_, ok = cb["ExecuteOnFirstTick"]
	if ok {
		got, normal := cb["ExecuteOnFirstTick"].(byte)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"ExecuteOnFirstTick\"]; cb = %#v", cb)
		}
		if got == byte(0) {
			executeOnFirstTick = false
		} else {
			executeOnFirstTick = true
		}
	}
	// ExecuteOnFirstTick
	_, ok = cb["TrackOutput"]
	if ok {
		got, normal := cb["TrackOutput"].(byte)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"TrackOutput\"]; cb = %#v", cb)
		}
		if got == byte(0) {
			trackOutput = false
		} else {
			trackOutput = true
		}
	}
	// TrackOutput
	_, ok = cb["conditionalMode"]
	if ok {
		got, normal := cb["conditionalMode"].(byte)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"conditionalMode\"]; cb = %#v", cb)
		}
		if got == byte(0) {
			conditionalMode = false
		} else {
			conditionalMode = true
		}
	}
	// conditionalMode
	_, ok = cb["auto"]
	if ok {
		got, normal := cb["auto"].(byte)
		if !normal {
			return blockNBT_depends.CommandBlock{}, fmt.Errorf("parseCommandBlock: Crashed in cb[\"auto\"]; cb = %#v", cb)
		}
		if got == byte(0) {
			auto = false
		} else {
			auto = true
		}
	}
	// auto
	return blockNBT_depends.CommandBlock{
		Command:            command,
		CustomName:         customName,
		LastOutput:         lastOutput,
		TickDelay:          tickDelay,
		ExecuteOnFirstTick: executeOnFirstTick,
		TrackOutput:        trackOutput,
		ConditionalMode:    conditionalMode,
		Auto:               auto,
	}, nil
	// return
}

// place command block and write datas to it
func CommandBlock(pack *blockNBT_depends.Package) error {
	commandBlockData, err := parseCommandBlock(pack.BlockInfo.NBT)
	if err != nil {
		return fmt.Errorf("CommandBlock: %v", err)
	}
	// get command block data
	var mode uint32 = packet.CommandBlockImpulse
	if strings.Contains(strings.ToLower(pack.BlockInfo.Name), "chain_command_block") {
		mode = packet.CommandBlockChain
	}
	if strings.Contains(strings.ToLower(pack.BlockInfo.Name), "repeating_command_block") {
		mode = packet.CommandBlockRepeating
	}
	// get mode of command block
	blockStates, err := mcstructure.ConvertCompoundToString(pack.BlockInfo.States, true)
	if err != nil {
		return fmt.Errorf("CommandBlock: Could not parse block states; pack.BlockInfo.States = %#v", pack.BlockInfo.States)
	}
	// get string of block states
	reqeust := commands_generator.SetBlockRequest(&types.Module{
		Block: &types.Block{
			Name:        &pack.BlockInfo.Name,
			BlockStates: blockStates,
		},
		Point: types.Position{
			X: int(pack.BlockInfo.Pos[0]),
			Y: int(pack.BlockInfo.Pos[1]),
			Z: int(pack.BlockInfo.Pos[2]),
		},
	}, pack.Mainsettings)
	// get setblock command
	if blockStates != blockNBT_depends.NoNeedToPlaceCommandBlockStatesString {
		if !pack.IsFastMode {
			pack.Environment.CommandSender.(*commands.CommandSender).SendSizukanaCommand(fmt.Sprintf("tp %d %d %d", pack.BlockInfo.Pos[0], pack.BlockInfo.Pos[1], pack.BlockInfo.Pos[2]))
			blockNBT_depends.SendWSCommandWithResponce(pack.Environment, reqeust)
		} else {
			pack.Environment.CommandSender.(*commands.CommandSender).SendDimensionalCommand(reqeust)
		}
	}
	// setblock
	if pack.Mainsettings.InvalidateCommands {
		commandBlockData.Command = fmt.Sprintf("|%v", commandBlockData.Command)
	}
	// invalidate command
	pack.Environment.Connection.(*minecraft.Conn).WritePacket(
		&packet.CommandBlockUpdate{
			Block:              true,
			Position:           pack.BlockInfo.Pos,
			Mode:               mode,
			NeedsRedstone:      !commandBlockData.Auto,
			Conditional:        commandBlockData.ConditionalMode,
			Command:            commandBlockData.Command,
			LastOutput:         commandBlockData.LastOutput,
			Name:               commandBlockData.CustomName,
			ShouldTrackOutput:  commandBlockData.TrackOutput,
			TickDelay:          commandBlockData.TickDelay,
			ExecuteOnFirstTick: commandBlockData.ExecuteOnFirstTick,
		},
	)
	return nil
	// return
}

func convertBoolToByte(input bool) byte {
	if input {
		return 1
	}
	return 0
}

// for operation 36 and more
func PlaceCommandBlockWithLegacyMethod(
	env *environment.PBEnvironment,
	cfg *types.MainConfig,
	isFastMode bool,
	block *types.Module,
) error {
	var blockStates map[string]interface{}
	var normal bool
	// init var
	if block.Block.Name != nil {
		if len(block.Block.BlockStates) > 0 {
			got, err := mcstructure.ParseStringNBT(block.Block.BlockStates, true)
			if err != nil {
				return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: Could not parse block states; block.Block.BlockStates = %#v", block.Block.BlockStates)
			}
			blockStates, normal = got.(map[string]interface{})
			if !normal {
				return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: the target block states is not a map[string]interface{}; got = %#v", got)
			}
		} else {
			runtimeId, found := chunk.LegacyBlockToRuntimeID(*block.Block.Name, block.Block.Data)
			if !found {
				return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: Could not convert legacy block to standard runtime id; *block.Block.Name = %#v, block.Block.Data = %#v", *block.Block.Name, block.Block.Data)
			}
			_, blockStates, found = chunk.RuntimeIDToState(runtimeId)
			if !found {
				return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: Could not convert standard runtime id to block states; *block.Block.Name = %#v, block.Block.Data = %#v", *block.Block.Name, block.Block.Data)
			}
		}
	}
	// get block states
	pack := blockNBT_depends.Package{
		Environment:  env,
		Mainsettings: cfg,
		IsFastMode:   isFastMode,
		BlockInfo: &blockNBT_depends.GeneralBlock{
			Pos: [3]int32{int32(block.Point.X), int32(block.Point.Y), int32(block.Point.Z)},
			NBT: map[string]interface{}{
				"Command":            block.CommandBlockData.Command,
				"CustomName":         block.CommandBlockData.CustomName,
				"LastOutput":         block.CommandBlockData.LastOutput,
				"TickDelay":          block.CommandBlockData.TickDelay,
				"ExecuteOnFirstTick": convertBoolToByte(block.CommandBlockData.ExecuteOnFirstTick),
				"TrackOutput":        convertBoolToByte(block.CommandBlockData.TrackOutput),
				"conditionalMode":    convertBoolToByte(block.CommandBlockData.Conditional),
				"auto":               convertBoolToByte(!block.CommandBlockData.NeedsRedstone),
			},
		},
	}
	// get struct datas
	pack.BlockInfo.Name = "command_block"
	if block.CommandBlockData.Mode == packet.CommandBlockRepeating {
		pack.BlockInfo.Name = "repeating_command_block"
	}
	if block.CommandBlockData.Mode == packet.CommandBlockChain {
		pack.BlockInfo.Name = "chain_command_block"
	}
	// set name for command block
	if block.Block.Name != nil {
		pack.BlockInfo.States = blockStates
	} else {
		pack.BlockInfo.States = blockNBT_depends.NoNeedToPlaceCommandBlockStatesMap
	}
	// for operation 26 and more(?)
	err := CommandBlock(&pack)
	if err != nil {
		return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
	}
	// place command block and write datas
	return nil
	// return
}
