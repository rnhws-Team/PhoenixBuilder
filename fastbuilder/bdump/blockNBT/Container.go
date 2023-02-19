package blockNBT

import (
	"fmt"
	blockNBT_depends "phoenixbuilder/fastbuilder/bdump/blockNBT/depends"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/io/commands"
	"phoenixbuilder/mirror/chunk"
	"strings"
)

// 检查一个方块是否是有效的容器；这里的有效指的是可以被 replaceitem 命令生效的容器
func checkIfIsEffectiveContainer(name string) (string, error) {
	value, ok := blockNBT_depends.SupportContainerPool[name]
	if ok {
		return value, nil
	}
	return "", blockNBT_depends.ErrNotASupportContainer
}

// convert interface{} into struct(blockNBT_depends.Container)
func parseContainer(container interface{}) (blockNBT_depends.Container, error) {
	var correct []interface{} = []interface{}{}
	var ans blockNBT_depends.Container = blockNBT_depends.Container{}
	// 初始化
	got, normal := container.([]interface{})
	if !normal {
		got, normal := container.(map[string]interface{})
		if !normal {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container; container = %#v", container)
		}
		correct = append(correct, got)
	} else {
		correct = got
	}
	// 把物品放入 correct 中
	// 如果这个物品是一个唱片机或者讲台，那么传入的 container 是一个 map[string]interface{} 而非 []interface{}
	// 为了更好的兼容性(更加方便)，这里都会把 map[string]interface{} 处理成通常情况下的 []interface{}
	// correct 就是处理结果
	for key, value := range correct {
		var count uint8 = uint8(0)
		var itemData uint16 = uint16(0)
		var name string = ""
		var slot uint8 = uint8(0)
		// 初始化
		containerData, normal := value.(map[string]interface{})
		if !normal {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v]; container[%v] = %#v", key, key, value)
		}
		// correct 这个列表中的每一项都必须是一个复合标签，也就得是 map[string]interface{} 才行
		_, ok := containerData["Count"]
		if !ok {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Count\"]; container[%v] = %#v", key, key, containerData)
		}
		count_got, normal := containerData["Count"].(byte)
		if !normal {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Count\"]; container[%v] = %#v", key, key, containerData)
		}
		count = uint8(count_got)
		// 拿一下物品数量(物品数量是一定存在的)
		_, ok = containerData["Name"]
		if !ok {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Name\"]; container[%v] = %#v", key, key, containerData)
		}
		got, normal := containerData["Name"].(string)
		if !normal {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Name\"]; container[%v] = %#v", key, key, containerData)
		}
		name = strings.Replace(strings.ToLower(got), "minecraft:", "", 1)
		// 拿一下这个物品的物品名称(命名空间 minecraft 已移除; 此数据必定存在)
		_, ok = containerData["Damage"]
		if !ok {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Damage\"]; container[%v] = %#v", key, key, containerData)
		}
		damage_got, normal := containerData["Damage"].(int16)
		if !normal {
			return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Damage\"]; container[%v] = %#v", key, key, containerData)
		}
		itemData = uint16(damage_got)
		// 拿一下物品的 Damage 值; Damage 值不一定就是物品的数据值(附加值); 此数据必定存在 [需要验证]
		_, ok = containerData["tag"]
		if ok {
			tag, normal := containerData["tag"].(map[string]interface{})
			if !normal {
				return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"tag\"]; container[%v] = %#v", key, key, containerData)
			}
			// 这个 container["tag"] 一定是一个复合标签 [需要验证]
			_, ok = tag["Damage"]
			if ok {
				got, normal := tag["Damage"].(int32)
				if !normal {
					return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"tag\"]; container[%v] = %#v", key, key, containerData)
				}
				itemData = uint16(got)
			}
		}
		// 拿一下这个工具的耐久值（当然也可能是别的，甚至它都不是个工具）
		// 这个 tag 里的 Damage 实际上也不一定就是物品的数据值(附加值)
		// 需要说明的是，tag 不一定存在，且 tag 存在，Damage 也不一定存在
		_, ok = containerData["Block"]
		if ok {
			Block, normal := containerData["Block"].(map[string]interface{})
			if !normal {
				return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Block\"]; container[%v] = %#v", key, key, containerData)
			}
			// 这个 container["Block"] 一定是一个复合标签；如果 Block 找得到则说明这个物品是一个方块
			_, ok = Block["val"]
			if ok {
				got, normal := Block["val"].(int16)
				if !normal {
					return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Block\"][\"val\"]; container[%v][\"Block\"] = %#v", key, key, Block)
				}
				itemData = uint16(got)
				// 如果这个物品是个方块，也就是 Block 找得到的话
				// 那在 Block 里面一定有一个 val 去声明这个方块的方块数据值(附加值) [仅限 Netease MC]
			} else {
				_, ok = Block["states"]
				if !ok {
					itemData = 0
				} else {
					got, normal := Block["states"].(map[string]interface{})
					if !normal {
						itemData = 0
					} else {
						runtimeId, found := chunk.StateToRuntimeID(name, got)
						if !found {
							return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Could not convert legacy block to standard runtime id; got = %#v", got)
						}
						legacyBlock, found := chunk.RuntimeIDToLegacyBlock(runtimeId)
						if !found {
							return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Could not convert standard runtime id to block states; got = %#v", got)
						}
						itemData = legacyBlock.Val
					}
				}
			}
		}
		// 拿一下这个方块的方块数据值(附加值)
		// 这个 Block 里的 val 一定是这个物品对应的方块的方块数据值(附加值)
		// 需要说明的是，Block 不一定存在，但如果 Block 存在，则 val 一定存在 [仅 Netease MC]
		/*
			以上三个都在拿物品数据值(附加值)
			需要说明的是，数据值的获取优先级是这样的
			Damage < tag["Damage"] < Block["val"]
			需要说明的是，以上列举的三个情况不能涵盖所有的物品数据值(附加值)的情况，所以我希望可以有个人看一下普世情况是长什么样的，请帮帮我！
		*/
		_, ok = containerData["Slot"]
		if ok {
			got, normal := containerData["Slot"].(byte)
			if !normal {
				return blockNBT_depends.Container{}, fmt.Errorf("parseContainer: Crashed in container[%v][\"Slot\"]; container[%v] = %#v", key, key, containerData)
			}
			slot = uint8(got)
		}
		// 拿一下这个物品所在的栏位(槽位)
		// 这个栏位(槽位)不一定存在，例如唱片机和讲台这种就不存在了(这种方块就一个物品，就不需要这个数据了)
		ans = append(ans, blockNBT_depends.SingleItem{
			Name:   name,
			Count:  count,
			Damage: itemData,
			Slot:   slot,
		})
		// 提交数据
	}
	// get datas
	return ans, nil
	// return
}

// convert map[string]interface{} into struct(blockNBT_depends.Container)
func getContainerData(blockNBT map[string]interface{}, blockName string) (blockNBT_depends.Container, error) {
	key, err := checkIfIsEffectiveContainer(blockName)
	if err != nil {
		return blockNBT_depends.Container{}, err
	}
	_, ok := blockNBT[key]
	// 这里是确定一下这个容器是否是我们支持了的容器
	if ok {
		ans, err := parseContainer(blockNBT[key])
		if err != nil {
			return blockNBT_depends.Container{}, fmt.Errorf("getContainerData: %v", err)
		}
		return ans, nil
	}
	// 如果这是个容器且对应的 key 可以找到，那么就去解析容器数据
	return blockNBT_depends.Container{}, nil
	// 对于唱片机和讲台这种容器，如果它们没有被放物品的话，那么对应的 key 是找不到的
	// 但是这并非是错误
}

// place container and write item into it
func Container(pack *blockNBT_depends.Package) error {
	blockStates, err := mcstructure.ConvertCompoundToString(pack.BlockInfo.States, true)
	if err != nil {
		return fmt.Errorf("Container: Could not parse block states; pack.BlockInfo.States = %#v", pack.BlockInfo.States)
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
	// place a container
	containerData, err := getContainerData(pack.BlockInfo.NBT, pack.BlockInfo.Name)
	if err != nil {
		return fmt.Errorf("Container: %v", err)
	}
	// get container data
	for _, value := range containerData {
		pack.Environment.CommandSender.(*commands.CommandSender).SendDimensionalCommand(commands_generator.ReplaceItemRequest(
			&types.Module{
				ChestSlot: &types.ChestSlot{
					Name:   value.Name,
					Count:  value.Count,
					Damage: value.Damage,
					Slot:   value.Slot,
				},
				Point: types.Position{
					X: int(pack.BlockInfo.Pos[0]),
					Y: int(pack.BlockInfo.Pos[1]),
					Z: int(pack.BlockInfo.Pos[2]),
				},
			},
			pack.Mainsettings,
		))
	}
	// get replaceitem request list
	return nil
	// return
}
