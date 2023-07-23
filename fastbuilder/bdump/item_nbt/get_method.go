package itemNBT

// 取得用于生成目标 NBT 物品的 接口/方法
func getMethod(item *ItemPackage) GeneralItemNBT {
	switch item.AdditionalData.Type {
	default:
		return &Default{ItemPackage: item}
		// 其他尚且未被支持的方块实体
	}
}
