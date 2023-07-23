package itemNBT

import (
	"fmt"
	env_interfaces "phoenixbuilder/fastbuilder/environment/interfaces"
	"sync"
)

var interfaceLock sync.Mutex

// 生成带有 NBT 数据的物品。
// 若你也想参与对于 NBT 物品的其他支持，
// 另见 https://github.com/df-mc/dragonfly
func GenerateItemWithNBTData(
	intf env_interfaces.GameInterface,
	singleItem ItemOrigin,
	additionalData *AdditionalData,
) error {
	defer interfaceLock.Unlock()
	interfaceLock.Lock()
	// lock(or unlock) api
	generalItem, err := ParseItemFromNBT(singleItem, additionalData.SupportBlocksPool)
	if err != nil {
		return fmt.Errorf("GenerateItemWithNBTData: Failed to generate the NBT item in hotbar %d, and the error log is %v", additionalData.HotBarSlot, err)
	}
	// get general item
	newRequest := ItemPackage{
		Interface:      intf,
		Item:           generalItem,
		AdditionalData: *additionalData,
	}
	newRequest.AdditionalData.Type = IsNBTItemSupported(newRequest.Item.Basic.Name)
	// get new request to generate new NBT item
	generateNBTItemMethod := getMethod(&newRequest)
	err = generateNBTItemMethod.Decode()
	if err != nil {
		return fmt.Errorf("GenerateItemWithNBTData: %v", err)
	}
	// get method and decode nbt data into golang struct
	err = generateNBTItemMethod.WriteData()
	if err != nil {
		return fmt.Errorf("GenerateItemWithNBTData: %v", err)
	}
	// assign nbt data
	return nil
	// return
}
