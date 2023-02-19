package blockNBT_depends

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

func SendWSCommandWithResponce(env *environment.PBEnvironment, command string) (*packet.CommandOutput, error) {
	sender := env.CommandSender
	u_d, _ := uuid.NewUUID()
	chann := make(chan *packet.CommandOutput)
	(sender.GetUUIDMap()).Store(u_d.String(), chann)
	// prepare
	sender.SendWSCommand(command, u_d)
	// send command
	resp := <-chann
	close(chann)
	if resp != nil {
		return resp, nil
	}
	// get responce
	return &packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: unknown error occurred")
	// return
}
