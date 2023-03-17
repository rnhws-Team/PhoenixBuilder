package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

/*
将背包中槽位为 inventorySlot 的物品移动到库存 containerID 的第 containerSlot 槽位，
且只移动 moveCount 个物品。
此函数将 containerSlot 处的物品当作空气处理。如果涉及到交换物品等操作，或许您需要使用其他函数。
当且仅当物品操作得到租赁服的响应后，此函数才会返回物品操作结果。
*/
func (g *GlobalAPI) PlaceItemIntoContainer(
	inventorySlot uint8,
	containerSlot uint8,
	containerID uint8,
	moveCount uint8,
) ([]protocol.ItemStackResponse, error) {
	datas, err := g.PacketHandleResult.Inventory.GetItemStackInfo(0, inventorySlot)
	if err != nil {
		return []protocol.ItemStackResponse{}, fmt.Errorf("PlaceItemIntoContainer: %v", err)
	}
	// 取得背包中指定物品栏的物品数据
	placeStackRequestAction := protocol.PlaceStackRequestAction{}
	if moveCount <= uint8(datas.Stack.Count) {
		placeStackRequestAction.Count = moveCount
	} else {
		placeStackRequestAction.Count = uint8(datas.Stack.Count)
	}
	// 得到欲移动的物品数量
	placeStackRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    12,
		Slot:           inventorySlot,
		StackNetworkID: datas.StackNetworkID,
	}
	placeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    containerID,
		Slot:           containerSlot,
		StackNetworkID: 0,
	}
	// 前置准备
	ans, err := g.SendItemStackRequestWithResponce(&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				Actions: []protocol.StackRequestAction{
					&placeStackRequestAction,
				},
				FilterStrings: []string{},
			},
		},
	})
	if err != nil {
		return []protocol.ItemStackResponse{}, fmt.Errorf("PlaceItemIntoContainer: %v", err)
	}
	// 发送物品操作请求
	return ans, nil
	// 返回值
}

// 将已放入铁砧第一格(注意是第一格)的物品的物品名称修改为 name 并返还到背包中的 slot 处。
// 当且仅当租赁服回应操作结果后再返回值。
func (g *GlobalAPI) ChangeItemName(resp protocol.ItemStackResponse, name string, slot uint8) error {
	var stackNetworkID int32
	var count uint8
	for _, value := range resp.ContainerInfo {
		if value.ContainerID == 0 {
			stackNetworkID = value.SlotInfo[0].StackNetworkID
			count = value.SlotInfo[0].Count
			break
		}
	}
	// 初始化
	newRequestID := g.PacketHandleResult.ItemStackOperation.GetNewRequestID()
	// 请求一个新的 RequestID 用于 ItemStackRequest
	placeStackRequestAction := protocol.PlaceStackRequestAction{}
	placeStackRequestAction.Count = count
	placeStackRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    0x3c,
		Slot:           0x32,
		StackNetworkID: newRequestID,
	}
	placeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    0xc,
		Slot:           slot,
		StackNetworkID: 0,
	}
	// 构造一个新的 PlaceStackRequestAction 结构体
	newItemStackRequest := packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				RequestID: newRequestID,
				Actions: []protocol.StackRequestAction{
					&protocol.CraftRecipeOptionalStackRequestAction{
						RecipeNetworkID:   0,
						FilterStringIndex: 0,
					},
					&protocol.ConsumeStackRequestAction{
						DestroyStackRequestAction: protocol.DestroyStackRequestAction{
							Count: count,
							Source: protocol.StackRequestSlotInfo{
								ContainerID:    0,
								Slot:           1,
								StackNetworkID: stackNetworkID,
							},
						},
					},
					&placeStackRequestAction,
				},
				FilterStrings: []string{name},
			},
		},
	}
	// 构造一个新的 ItemStackRequest 结构体
	g.PacketHandleResult.ItemStackOperation.WriteRequest(newRequestID)
	// 写入请求到等待队列
	err := g.WritePacket(&newItemStackRequest)
	if err != nil {
		return fmt.Errorf("ChangeItemName: %v", err)
	}
	// 发送物品操作请求
	g.PacketHandleResult.ItemStackOperation.AwaitResponce(newRequestID)
	ans, err := g.PacketHandleResult.ItemStackOperation.LoadResponceAndDelete(newRequestID)
	if err != nil {
		return fmt.Errorf("ChangeItemName: %v", err)
	}
	// 取得物品操作请求的结果
	if ans.Status != 0 {
		return fmt.Errorf("ChangeItemName: Operation %v have been canceled by error code %v; ans = %#v", ans.RequestID, ans.Status, ans)
	}
	// 如果物品操作请求被拒绝
	return nil
	// 返回值
}
