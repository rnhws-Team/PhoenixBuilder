package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
)

/*
打开 pos 处名为 blockName 且方块状态为 blockStates 的容器，且只有当打开完成后才会返回值。
当 needOccupyContainerResources 为真时，此函数会主动占用容器资源，但一般情况下我建议此参数填 false ，
因为打开容器仅仅是一系列容器操作的一个步骤，因此此函数中不应该贸然修改容器资源，否则可能会造成潜在的问题
*/
func (g *GlobalAPI) OpenContainer(
	pos [3]int32,
	blockName string,
	blockStates map[string]interface{},
	needOccupyContainerResources bool,
) error {
	var lock *sync.Mutex = &sync.Mutex{}
	if needOccupyContainerResources {
		_, lock = g.PacketHandleResult.ContainerResources.Occupy(false)
	}
	// lock down resources
	err := g.UseItemOnBlocks(0, pos, blockName, blockStates, false)
	if err != nil {
		return fmt.Errorf("OpenContainer: %v", err)
	}
	// open container
	g.PacketHandleResult.ContainerResources.AwaitResponce()
	// wait changes
	if needOccupyContainerResources {
		lock.Unlock()
	}
	// unlock resources
	return nil
	// return
}

/*
关闭已经打开的容器；如果容器已被关闭，则返回的布尔值为 false；只有当容器被关闭后才会返回值
*/
func (g *GlobalAPI) CloseContainer() (bool, error) {
	defer g.PacketHandleResult.ContainerResources.release()
	// release sharing resources
	err := g.WritePacket(&packet.ContainerClose{
		WindowID:   g.PacketHandleResult.ContainerResources.GetContainerOpenDatas().WindowID,
		ServerSide: false,
	})
	if err != nil {
		return false, fmt.Errorf("CloseContainer: %v", err)
	}
	// close container
	g.PacketHandleResult.ContainerResources.AwaitResponce()
	// wait changes
	return true, nil
	// return
}
