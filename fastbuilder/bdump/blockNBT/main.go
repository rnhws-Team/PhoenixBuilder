package blockNBT

import (
	"fmt"
	blockNBT_depends "phoenixbuilder/fastbuilder/bdump/blockNBT/depends"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/io/commands"
	"strings"
)

// convert types.Module into blockNBT_depends.GeneralBlock
func parseBlockModule(singleBlock *types.Module) (*blockNBT_depends.GeneralBlock, error) {
	// init var
	got, err := mcstructure.ParseStringNBT(singleBlock.Block.BlockStates, true)
	if err != nil {
		return &blockNBT_depends.GeneralBlock{}, fmt.Errorf("parseBlockModule: Could not parse block states; singleBlock.Block.BlockStates = %#v", singleBlock.Block.BlockStates)
	}
	blockStates, normal := got.(map[string]interface{})
	if !normal {
		return &blockNBT_depends.GeneralBlock{}, fmt.Errorf("parseBlockModule: Target block states is not map[string]interface{}; got = %#v", got)
	}
	// get block states
	return &blockNBT_depends.GeneralBlock{
		Name:   strings.Replace(strings.ToLower(strings.ReplaceAll(*singleBlock.Block.Name, " ", "")), "minecraft:", "", 1),
		States: blockStates,
		Pos:    [3]int32{int32(singleBlock.Point.X), int32(singleBlock.Point.Y), int32(singleBlock.Point.Z)},
		NBT:    singleBlock.NBTMap,
		Data:   blockNBT_depends.Datas{},
	}, nil
	// return
}

// 检查这个方块实体是否已被支持
func checkIfIsEffectiveNBTBlock(blockName string) string {
	value, ok := blockNBT_depends.SupportBlocksPool[blockName]
	if ok {
		return value
	}
	return ""
}

/*
带有 NBT 数据放置方块

如果你也想参与更多方块实体的支持，可以去看看这个库 https://github.com/df-mc/dragonfly

这个库也是用了 gophertunnel 的
*/
func placeBlockWithNBTData(pack *blockNBT_depends.Package) error {
	var err error
	// init var
	switch pack.BlockInfo.Data.Type {
	case "CommandBlock":
		err = CommandBlock(pack)
		if err != nil {
			return fmt.Errorf("placeBlockWithNBTData: %v", err)
		}
		// 命令方块
	case "Container":
		err = Container(pack)
		if err != nil {
			return fmt.Errorf("placeBlockWithNBTData: %v", err)
		}
		// 各类可被 replaceitem 生效的容器
	default:
		blockStates, err := mcstructure.ConvertCompoundToString(pack.BlockInfo.States, true)
		if err != nil {
			return fmt.Errorf("placeBlockWithNBTData: Could not parse block states; pack.BlockInfo.States = %#v", pack.BlockInfo.States)
		}
		// get string of block states
		pack.Environment.CommandSender.(*commands.CommandSender).SendDimensionalCommand(commands_generator.SetBlockRequest(&types.Module{
			Block: &types.Block{
				Name:        &pack.BlockInfo.Name,
				BlockStates: blockStates,
			},
			Point: types.Position{
				X: int(pack.BlockInfo.Pos[0]),
				Y: int(pack.BlockInfo.Pos[1]),
				Z: int(pack.BlockInfo.Pos[2]),
			},
		}, pack.Mainsettings))
		// send command
		return nil
		// 其他没有支持的方块实体
	}
	return nil
}

// 此函数是 package blockNBT 的主函数
func PlaceBlockWithNBTDataRun(
	env *environment.PBEnvironment,
	cfg *types.MainConfig,
	isFastMode bool,
	blockInfo *types.Module,
) error {
	defer blockNBT_depends.ApiIsUsing.Unlock()
	blockNBT_depends.ApiIsUsing.Lock()
	// lock(or unlock) api
	var err error
	// init var
	newRequest := blockNBT_depends.Package{
		Environment:  env,
		Mainsettings: cfg,
		IsFastMode:   isFastMode,
		BlockInfo:    &blockNBT_depends.GeneralBlock{},
	}
	newRequest.BlockInfo, err = parseBlockModule(blockInfo)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTDataRun: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", *blockInfo.Block.Name, blockInfo.Point.X, blockInfo.Point.Y, blockInfo.Point.Z, err)
	}
	newRequest.BlockInfo.Data.Type = checkIfIsEffectiveNBTBlock(newRequest.BlockInfo.Name)
	// get new request of place nbt block
	err = placeBlockWithNBTData(&newRequest)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithNBTDataRun: Failed to place the entity block named %v at (%d,%d,%d), and the error log is %v", newRequest.BlockInfo.Name, newRequest.BlockInfo.Pos[0], newRequest.BlockInfo.Pos[1], newRequest.BlockInfo.Pos[2], err)
	}
	// place block with nbt datas
	return nil
	// return
}
