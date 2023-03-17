package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 使用铁砧修改物品名称时会被使用的结构体
type AnvilChangeItemName struct {
	Slot uint8  // 被修改物品在背包所在的槽位
	Name string // 要修改的目标名称
}

// 在 pos 处放置一个方块状态为 blockStates 的铁砧，并依次发送 request 列表中的物品名称修改请求
func (g *GlobalAPI) ChangeItemNameByUsingAnvil(
	pos [3]int32,
	blockStates string,
	request []AnvilChangeItemName,
) error {
	err := g.SendSettingsCommand("gamemode 1", true)
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 更换游戏模式为创造
	uniqueId, correctPos, err := g.GenerateNewAnvil(pos, blockStates)
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 尝试生成一个铁砧并附带承重方块
	_, err = g.SendWSCommandWithResponce(fmt.Sprintf("tp %d %d %d", pos[0], pos[1], pos[2]))
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 传送机器人到铁砧处
	_, lockDown := g.PacketHandleResult.ContainerResources.Occupy(false)
	// 获取容器资源
	got, err := mcstructure.ParseStringNBT(blockStates, true)
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	blockStatesMap, normal := got.(map[string]interface{})
	if !normal {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: Could not convert got into map[string]interface{}; got = %#v", got)
	}
	// 获取要求放置的铁砧的方块状态
	err = g.ChangeSelectedHotbarSlot(0, true)
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 切换手持物品栏
	err = g.OpenContainer(pos, "minecraft:anvil", blockStatesMap, 0, false)
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 打开铁砧
	for _, value := range request {
		datas, err := g.PacketHandleResult.Inventory.GetItemStackInfo(0, value.Slot)
		if err != nil {
			return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
		}
		if datas.Stack.ItemType.NetworkID == 0 {
			continue
		}
		// 获取被改物品的相关信息
		resp, err := g.PlaceItemIntoContainer(value.Slot, 1, 0, uint8(datas.Stack.Count))
		if err != nil {
			return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
		}
		if resp[0].Status != 0 {
			return fmt.Errorf("ChangeItemNameByUsingAnvil: Operation %v have been canceled by error code %v; inventorySlot = %v, containerSlot = 1, moveCount = %v", resp[0].RequestID, resp[0].Status, value.Slot, datas.Stack.Count)
		}
		// 移动物品到铁砧
		err = g.WritePacket(&packet.AnvilDamage{
			Damage:        0,
			AnvilPosition: pos,
		})
		if err != nil {
			return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
		}
		// 请求损坏当前铁砧
		err = g.ChangeItemName(resp[0], value.Name, value.Slot)
		if err != nil {
			return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
		}
		// 发送改名请求
	}
	err = g.CloseContainer()
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 关闭铁砧
	lockDown.Unlock()
	// 释放容器公用资源
	err = g.RevertBlockUnderAnvil(uniqueId, correctPos)
	if err != nil {
		return fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 恢复铁砧下方的承重方块为原本方块
	return nil
	// 返回值
}
