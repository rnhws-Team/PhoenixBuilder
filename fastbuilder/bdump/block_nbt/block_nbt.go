package blockNBT

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment/interfaces"
	"phoenixbuilder/fastbuilder/types"
)

// ------------------------- interface -------------------------

/*
GeneralBlockNBT 提供了一个通用的接口，
以便于您可以方便的解析对应的方块实体，
然后放置它并以最大的可能性注入 NBT 数据。

该接口实际与下方的 BlockEntity 结构体绑定
*/
type GeneralBlockNBT interface {
	Decode() error
	WriteData() error
}

// ------------------------- general -------------------------

// GeneralBlock 结构体用于描述通用型方块的数据
type GeneralBlock struct {
	// 方块名称(不含命名空间且应该全部小写)
	Name string
	// 方块状态
	States map[string]interface{}
	// 当前方块所携带的 NBT 数据
	NBT map[string]interface{}
}

// AdditionalData 结构体用于描述一个方块实体的其他附加数据，例如方块的绝对坐标
type AdditionalData struct {
	// 字符串形式的方块状态，用于在放置方块时使用
	BlockStates string
	// 方块坐标(绝对坐标)
	Position [3]int32
	// 该方块的类型，例如各式各样的告示牌可以写作 Sign
	// TODO: USE ENUM INSTEAD
	Type string
	// 此参数应当只被 PhoenixBuilder 使用，除非 Omega 也需要设置一些功能
	Settings *types.MainConfig
	// 是否是快速模式放置；若为真，则大多数方块实体的 NBT 数据将不会被 assign
	FastMode bool
	// 部分情况下可能会携带的不定数据，通常情况下应该为空
	Others interface{}
}

// BlockEntity 是用于包装每个方块实体的结构体
type BlockEntity struct {
	// 储存执行该方块状态放置所需的 API ，例如发包需要用到的函数等
	// 此参数需要外部实现主动赋值，
	// 主要是为了兼容 Omega 和 PhoenixBuilder 对功能的同时使用
	Interface interfaces.GameInterface
	// 一个通用型方块的数据，例如名称、方块状态和所携带的 NBT 数据
	Block GeneralBlock
	// 此方块的其他附加数据，例如方块的绝对坐标
	AdditionalData AdditionalData
}

// ------------------------- command_block -------------------------

// 描述单个命令方块中已解码的部分
type CommandBlockData struct {
	Command            string // Command(TAG_String) = ""
	CustomName         string // CustomName(TAG_String) = ""
	LastOutput         string // LastOutput(TAG_String) = ""
	TickDelay          int32  // TickDelay(TAG_Int) = 0
	ExecuteOnFirstTick bool   // ExecuteOnFirstTick(TAG_Byte) = 1
	TrackOutput        bool   // TrackOutput(TAG_Byte) = 1
	ConditionalMode    bool   // conditionalMode(TAG_Byte) = 0
	Auto               bool   // auto(TAG_Byte) = 1
}

// CommandBlock 结构体用于描述一个完整的命令方块数据
type CommandBlock struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 存放已解码的命令方块数据
	CommandBlockData
	// 为向下兼容而设，因为旧方法下不需要放置命令方块
	ShouldPlaceBlock bool
}

// ------------------------- container -------------------------

// 描述单个物品在解码前的 NBT 表达形式
type ItemOrigin map[string]interface{}

// 描述物品的单个附魔属性
type Enchantment struct {
	ID    uint8 // 该附魔属性的 ID
	Level int16 // 该附魔属性的等级
}

// 描述单个物品的物品组件数据
type ItemComponents struct {
	CanPlaceOn  []string
	CanBreak    []string
	ItemLock    string
	KeepOnDeath bool
}

// 描述单个物品的自定义数据。
// 这些数据实际上并不存在，
// 只是我们为了区分一些特殊的物品而设
type ItemCustomData struct {
	/*
		假设该物品本身就是一个带有 NBT 的方块，
		那么如果我们已经在 PhoenixBuilder 实现了这些方块的 NBT 注入，
		那么对于箱子内的这些物品来说，
		我们也仍然可以通过 PickBlock 的方法来实现对它们的兼容。

		因此，如果该物品带有 NBT 且是一个方块，
		那么此字段不为空指针。
	*/
	Tag *GeneralBlock
	/*
		这个物品可能是一本写了字或者签过名的书，
		而此字段描述的书上的具体内容及签名相关的数据。

		因此，如果该物品是书且写了字或者签过名，
		那么此值不为空指针。

		TODO: 兼容此特性
	*/
	// BookData *...
}

// 描述单个物品的基本数据
type ItemBasicData struct {
	Name   string // Name(TAG_String) = ""
	Count  uint8  // Count(TAG_Byte) = 0
	Damage uint16 // TAG_Short = 0
	Slot   uint8  // Slot(TAG_Byte) = 0
}

// 描述单个物品的附加数据
type ItemAdditionalData struct {
	DisplayName    string         // 该物品的显示名称
	Enchantments   []Enchantment  // 该物品的附魔属性
	ItemComponents ItemComponents // 该物品的物品组件
}

// 描述一个单个的物品
type Item struct {
	Basic      ItemBasicData      // 该物品的基本数据
	Additional ItemAdditionalData // 该物品的附加数据
	Custom     ItemCustomData     // 由 PhoenixBuilder 定义的自定义数据
}

// 描述一个容器
type Container struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 容器内的物品数据
	ItemContents []Item
}

// 未被支持的容器会被应用此错误信息。
// 用于 Container.go 中的 ReplaceNBTMapToContainerList 等函数
var ErrNotASupportedContainer error = fmt.Errorf("replaceNBTMapToContainerList: Not a supported container")

// 用于 Container.go 中的 ReplaceNBTMapToContainerList 等函数
var KeyName string = "data"

// ------------------------- sign -------------------------

// 描述一个告示牌
type Sign struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
}
