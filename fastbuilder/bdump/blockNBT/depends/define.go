package blockNBT_depends

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/types"
	"sync"
)

// -------------------------

var ApiIsUsing sync.Mutex

type Datas struct {
	Type   string      // 用于存放这种方块的类型，比如不同的告示牌都可以写成 sign
	Others interface{} // 存放其他一些必要数据
}

type GeneralBlock struct {
	Name   string
	States map[string]interface{}
	Pos    [3]int32
	NBT    map[string]interface{}
	Data   Datas
}

type Package struct {
	Environment  *environment.PBEnvironment // 运行环境（必须）
	Mainsettings *types.MainConfig          // 设置
	IsFastMode   bool                       // 是否是快速模式
	BlockInfo    *GeneralBlock              // 方块实体的各项信息
}

var SupportBlocksPool = map[string]string{
	"command_block":           "CommandBlock",
	"chain_command_block":     "CommandBlock",
	"repeating_command_block": "CommandBlock",
	// 命令方块
	"blast_furnace":      "Container",
	"lit_blast_furnace":  "Container",
	"smoker":             "Container",
	"lit_smoker":         "Container",
	"furnace":            "Container",
	"lit_furnace":        "Container",
	"chest":              "Container",
	"barrel":             "Container",
	"trapped_chest":      "Container",
	"hopper":             "Container",
	"dispenser":          "Container",
	"dropper":            "Container",
	"cauldron":           "Container",
	"lava_cauldron":      "Container",
	"jukebox":            "Container",
	"brewing_stand":      "Container",
	"undyed_shulker_box": "Container",
	"shulker_box":        "Container",
	"lectern":            "Container",
	// 容器
}

// -------------------------

type CommandBlock struct {
	Command            string
	CustomName         string
	LastOutput         string
	TickDelay          int32
	ExecuteOnFirstTick bool
	TrackOutput        bool
	ConditionalMode    bool
	Auto               bool
}

// -------------------------

type SingleItem struct {
	Name   string
	Count  uint8
	Damage uint16
	Slot   uint8
}

type Container []SingleItem

// 此表描述了可被 replaceitem 生效的容器，key 代表容器的方块名，而 value 代表此容器放置物品使用的复合标签或列表
var SupportContainerPool map[string]string = map[string]string{
	"blast_furnace":      "Items",
	"lit_blast_furnace":  "Items",
	"smoker":             "Items",
	"lit_smoker":         "Items",
	"furnace":            "Items",
	"lit_furnace":        "Items",
	"chest":              "Items",
	"barrel":             "Items",
	"trapped_chest":      "Items",
	"lectern":            "book",
	"hopper":             "Items",
	"dispenser":          "Items",
	"dropper":            "Items",
	"jukebox":            "RecordItem",
	"brewing_stand":      "Items",
	"undyed_shulker_box": "Items",
	"shulker_box":        "Items",
}

var ErrNotASupportContainer error = fmt.Errorf("checkIfIsEffectiveContainer: Not a supported container")

// -------------------------
