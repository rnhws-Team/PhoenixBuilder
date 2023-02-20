package blockNBT_CommandBlock

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 放置一个命令方块(可选)并写入命令方块数据
func (c *CommandBlock) WriteDatas(needToPlaceBlock bool) error {
	var mode uint32 = packet.CommandBlockImpulse
	// 初始化
	if c.BlockEntityDatas.Datas.Settings.ExcludeCommands {
		err := c.BlockEntityDatas.API.SetBlockFastly(c.BlockEntityDatas.Datas.Position, c.BlockEntityDatas.Block.Name, c.BlockEntityDatas.Datas.StatesString)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
		return nil
	}
	// 如果要求仅放置命令方块而不写入命令方块数据
	if needToPlaceBlock && !c.BlockEntityDatas.Datas.FastMode {
		err := c.BlockEntityDatas.API.SetBlock(c.BlockEntityDatas.Datas.Position, c.BlockEntityDatas.Block.Name, c.BlockEntityDatas.Datas.StatesString)
		if err != nil {
			return fmt.Errorf("WriteDatas: %v", err)
		}
	}
	// 放置命令方块
	if c.BlockEntityDatas.Block.Name == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else if c.BlockEntityDatas.Block.Name == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	}
	// 确定命令方块的类型
	err := c.BlockEntityDatas.API.WritePacket(&packet.CommandBlockUpdate{
		Block:              true,
		Position:           c.BlockEntityDatas.Datas.Position,
		Mode:               mode,
		NeedsRedstone:      !c.CommandBlockDatas.Auto,
		Conditional:        c.CommandBlockDatas.ConditionalMode,
		Command:            c.CommandBlockDatas.Command,
		LastOutput:         c.CommandBlockDatas.LastOutput,
		Name:               c.CommandBlockDatas.CustomName,
		ShouldTrackOutput:  c.CommandBlockDatas.TrackOutput,
		TickDelay:          c.CommandBlockDatas.TickDelay,
		ExecuteOnFirstTick: c.CommandBlockDatas.ExecuteOnFirstTick,
	})
	if err != nil {
		return fmt.Errorf("WriteDatas: %v", err)
	}
	// 写入命令方块数据
	return nil
	// 返回值
}
