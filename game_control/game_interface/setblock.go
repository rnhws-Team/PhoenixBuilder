package GameInterface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
)

// 在 pos 处以 setblock 命令放置名为 name 且方块状态为 states 的方块。只有请求被返回时此函数再返回值
func (g *GameInterface) SetBlock(pos [3]int32, name string, states string) error {
	request := commands_generator.SetBlockRequest(&types.Module{
		Block: &types.Block{
			Name:        &name,
			BlockStates: states,
		},
		Point: types.Position{
			X: int(pos[0]),
			Y: int(pos[1]),
			Z: int(pos[2]),
		},
	}, &types.MainConfig{})
	// get setblock command
	resp := g.SendWSCommandWithResponse(request)
	// send setblock request
	if resp.ErrorType == ResourcesControl.ErrCommandRequestTimeOut {
		err := g.SendSettingsCommand(request, true)
		if err != nil {
			return fmt.Errorf("SetBlock: %v", err)
		}
		resp = g.SendWSCommandWithResponse("list")
		if resp.Error != nil && resp.ErrorType != ResourcesControl.ErrCommandRequestTimeOut {
			return fmt.Errorf("SetBlock: %v", resp.Error)
		}
	}
	// special solutions for timeouts
	return nil
	// return
}

// 在 pos 处以 setblock 命令放置名为 name 且方块状态为 states 的方块。
// 特别地，此方法使用 settings command 来发送命令，因此该函数在被调用后不会等待返回值
func (g *GameInterface) SetBlockAsync(pos [3]int32, name string, states string) error {
	request := commands_generator.SetBlockRequest(&types.Module{
		Block: &types.Block{
			Name:        &name,
			BlockStates: states,
		},
		Point: types.Position{
			X: int(pos[0]),
			Y: int(pos[1]),
			Z: int(pos[2]),
		},
	}, &types.MainConfig{})
	err := g.SendSettingsCommand(request, true)
	if err != nil {
		return fmt.Errorf("SetBlockForgetfully: %v", err)
	}
	return nil
}
