package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// 向租赁服发送 WS 命令且获取返回值
func (g *GlobalAPI) SendWSCommandWithResponce(command string) (packet.CommandOutput, error) {
	uniqueId, err := uuid.NewUUID()
	if err != nil || uniqueId == uuid.Nil {
		resp, err := g.SendWSCommandWithResponce(command)
		if err != nil {
			return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: %v", err)
		}
		return resp, nil
	}
	uniqueIdString := uniqueId.String()
	// 初始化
	CommandRequestWaitingList.Lock.Lock()
	CommandRequestWaitingList.List[uniqueIdString] = true
	CommandRequestWaitingList.Lock.Unlock()
	// 写入请求到等待队列
	err = g.SendWSCommand(command, uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: %v", err)
	}
	// 发送命令
	for {

		CommandRequestWaitingList.Lock.RLock()
		_, unsuccess := CommandRequestWaitingList.List[uniqueIdString]
		CommandRequestWaitingList.Lock.RUnlock()

		if !unsuccess {

			CommandOutputPool.Lock.RLock()
			got, normal := CommandOutputPool.Pool[uniqueIdString]
			if !normal {
				return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: Responce not found(THIS IS A BUG)")
			}
			CommandOutputPool.Lock.RUnlock()

			newMap := map[string]packet.CommandOutput{}

			CommandOutputPool.Lock.Lock()
			delete(CommandOutputPool.Pool, uniqueIdString)
			for key, value := range CommandOutputPool.Pool {
				newMap[key] = value
			}
			CommandOutputPool.Pool = newMap
			CommandOutputPool.Lock.Unlock()

			return got, nil
		}
	}
}
