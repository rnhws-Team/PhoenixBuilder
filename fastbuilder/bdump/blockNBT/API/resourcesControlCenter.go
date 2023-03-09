package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"

	"github.com/google/uuid"
)

// 测定 key 是否在 c.commandRequest.datas 中。如果在，则返回真，否则返回假
func (c *CommandRequest) testRequest(key uuid.UUID) bool {
	c.commandRequest.lockDown.RLock()
	defer c.commandRequest.lockDown.RUnlock()
	// init
	_, ok := c.commandRequest.datas[key]
	return ok
	// return
}

// 当发送命令请求(packet.CommandRequest)后，您可能需要获取它的返回值，那么此函数用于
// 将带有特定 uuid 的命令请求保存在 c.commandRequest.datas 中并锁定当前资源以等待返回值
func (c *CommandRequest) writeRequest(key uuid.UUID) error {
	if c.testRequest(key) {
		return fmt.Errorf("writeRequest: %v is already exist in c.commandRequest.datas", key.String())
	}
	// if key is already exist
	c.commandRequest.lockDown.Lock()
	// lock down resources
	c.commandRequest.datas[key] = &sync.Mutex{}
	c.commandRequest.datas[key].Lock()
	// lock down command request
	c.commandRequest.lockDown.Unlock()
	// unlock resources
	return nil
	// return
}

// 将带有特定 uuid 的命令请求从 c.commandRequest.datas 中释放并移除
func (c *CommandRequest) removeRequest(key uuid.UUID) error {
	if !c.testRequest(key) {
		return fmt.Errorf("removeRequest: %v is not recorded in c.commandRequest.datas", key.String())
	}
	// if key is not exist
	c.commandRequest.lockDown.Lock()
	// lock down resources
	c.commandRequest.datas[key].Unlock()
	// unlock command request
	delete(c.commandRequest.datas, key)
	newMap := map[uuid.UUID]*sync.Mutex{}
	for k, value := range c.commandRequest.datas {
		newMap[k] = value
	}
	c.commandRequest.datas = newMap
	// remove key and value from c.commandRequest.datas
	c.commandRequest.lockDown.Unlock()
	// unlock resources
	return nil
	// return
}

func (c *CommandRequest) writeResponce(key uuid.UUID, resp packet.CommandOutput) error {
	c.commandResponce.lockDown.Lock()
	defer c.commandResponce.lockDown.Unlock()
	// init
	c.commandResponce.datas[key] = resp
	// send command responce
	err := c.removeRequest(key)
	if err != nil {
		return fmt.Errorf("WriteResponce: %v", err)
	}
	// remove command reuqest from c.commandRequest.datas
	return nil
	// return
}

func (c *CommandRequest) readResponceAndDelete(key uuid.UUID) packet.CommandOutput
