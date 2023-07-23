package itemNBT

import (
	"encoding/json"
	"fmt"
)

// 从 SupportItemsPool 检查这个 NBT 物品是否已被支持。
// 如果尚未被支持，则返回空字符串，否则返回这种物品的类型。
// 以告示牌为例，所有的告示牌都可以写作为 Sign
func IsNBTItemSupported(blockName string) string {
	value, ok := SupportItemsPool[blockName]
	if ok {
		return value
	}
	return ""
}

// 将 itemComponents 编码为游戏支持的 JSON 格式。
// 如果传入的 itemComponents 为空指针，则返回空字符串
func MarshalItemComponents(itemComponents *ItemComponents) string {
	type can_place_on_or_can_destroy struct {
		Blocks []string `json:"blocks"`
	}
	type item_lock struct {
		Mode string `json:"mode"`
	}
	res := map[string]interface{}{}
	// 初始化
	if itemComponents == nil {
		return ""
	}
	// 如果物品组件不存在，那么应该返回空字符串而非 {}
	if len(itemComponents.CanPlaceOn) > 0 {
		res["can_place_on"] = can_place_on_or_can_destroy{Blocks: itemComponents.CanPlaceOn}
	}
	if len(itemComponents.CanDestroy) > 0 {
		res["can_destroy"] = can_place_on_or_can_destroy{Blocks: itemComponents.CanDestroy}
	}
	if itemComponents.KeepOnDeath {
		res["keep_on_death"] = struct{}{}
	}
	if len(itemComponents.ItemLock) != 0 {
		res["item_lock"] = item_lock{Mode: itemComponents.ItemLock}
	}
	// 赋值
	bytes, _ := json.Marshal(res)
	return string(bytes)
	// 返回值
}

/*
将 singleItem 解析为 GeneralItem ；
supportBlocksPool 指代此位置所对应的表格：
"phoenixbuilder/fastbuilder/bdump/block_nbt/pool.go:SupportBlocksPool"。
这么做为了避免循环导入 Package ，因此我们要求由相关的外层调用者提供此字段。

特别地，如果此物品存在 item_lock 物品组件，
则只会解析物品组件的相关数据，
因为存在 item_lock 的物品无法跨容器移动；

如果此物品是一个 NBT 方块，
则附魔属性将被丢弃，因为无法为方块附魔
*/
func ParseItemFromNBT(
	singleItem ItemOrigin,
	supportBlocksPool map[string]string,
) (GeneralItem, error) {
	itemBasicData, err := DecodeItemBasicData(singleItem)
	if err != nil {
		return GeneralItem{}, fmt.Errorf("ParseItemFromNBT: %v", err)
	}
	// basic
	itemAdditionalData, err := DecodeItemAdditionalData(singleItem)
	if err != nil {
		return GeneralItem{}, fmt.Errorf("ParseItemFromNBT: %v", err)
	}
	// additional
	if itemAdditionalData != nil && itemAdditionalData.ItemComponents != nil && len(itemAdditionalData.ItemComponents.ItemLock) != 0 {
		return GeneralItem{
			Basic:      itemBasicData,
			Additional: itemAdditionalData,
			Custom:     nil,
		}, nil
	}
	// 如果此物品使用了物品组件 item_lock ，
	// 则后续数据将不被解析。
	// 因为存在 item_lock 的物品无法跨容器移动
	itemCustomData, err := DecodeItemCustomData(itemBasicData, singleItem, supportBlocksPool)
	if err != nil {
		return GeneralItem{}, fmt.Errorf("ParseItemFromNBT: %v", err)
	}
	// custom
	if itemCustomData != nil && itemCustomData.SubBlockData != nil && itemAdditionalData != nil {
		itemAdditionalData.Enchantments = nil
	}
	// 如果此物品是一个 NBT 方块，
	// 则附魔属性将被丢弃，因为无法为方块附魔
	return GeneralItem{
		Basic:      itemBasicData,
		Additional: itemAdditionalData,
		Custom:     itemCustomData,
	}, nil
	// return
}
