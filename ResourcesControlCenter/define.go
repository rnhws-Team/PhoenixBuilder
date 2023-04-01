package ResourcesControlCenter

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"

	"github.com/google/uuid"
)

/*
管理 PhoenixBuilder 的各类公用资源。

值得说明的是，此结构体的出现将会意味着 UQHolder 的弃用 [TODO]
*/
type Resources struct {
	// 如果当前结构体是在 PhoenixBuilder 启动时取得的，
	// 那么此认证结果为真，否则为假 。
	// 此参数有助于验证公用资源的唯一性，因为公用资源在内存中至多存在一个
	verified bool
	// 管理命令请求队列及命令返回值
	Command commandRequestWithResponce
	// 管理本地库存数据，如背包物品
	Inventory inventoryContents
	// 管理物品操作请求及结果
	ItemStackOperation itemStackReuqestWithResponce
	// 管理容器资源的占用状态，同时存储容器操作的结果
	Container container
}

// 存放命令请求及结果
type commandRequestWithResponce struct {
	// 命令请求队列
	commandRequest struct {
		// 防止并发读写而设置的读写锁
		lockDown sync.RWMutex
		// 存放命令请求的等待队列。
		// 每次写入请求后将会自动为此请求上锁以便于阻塞
		datas map[uuid.UUID]*sync.Mutex
	}
	// 命令请求的返回值
	commandResponce struct {
		// 防止并发读写而设置的读写锁
		lockDown sync.RWMutex
		// 存放命令返回值。
		// 每次写入返回值后将会自动为对应等待队列中的读写锁解锁
		datas map[uuid.UUID]packet.CommandOutput
	}
}

// 存放所有有效库存中的物品数据，例如背包和盔甲栏
type inventoryContents struct {
	// 防止并发读写而设置的读写锁
	lockDown sync.RWMutex
	// int32 代表打开的库存的窗口 ID ，即 WindowID ；
	// uint8 代表物品所在的槽位；
	// 最内层的 protocol.ItemInstance 存放物品数据
	datas map[uint32]map[uint8]protocol.ItemInstance
}

/*
存放物品操作请求及结果。

重要：
任何物品操作都应该通过此结构体下的有关实现来完成，否则可能会造成严重后果。
因此，为了绝对的安全，如果尝试绕过相关实现而直接发送物品操作数据包，则会造成程序惊慌。
*/
type itemStackReuqestWithResponce struct {
	// 物品操作请求队列
	itemStackRequest struct {
		// 防止并发读写而设置的读写锁
		lockDown sync.RWMutex
		// 存放物品操作请求的等待队列。
		// 每次写入请求后将会自动为此请求上锁以便于阻塞
		datas map[int32]singleItemStackRequest
	}
	// 物品操作的结果
	itemStackResponce struct {
		// 防止并发读写而设置的读写锁
		lockDown sync.RWMutex
		// 存放物品操作的结果。
		// 每次写入返回值后将会自动为对应等待队列中的读写锁解锁。
		datas map[int32]protocol.ItemStackResponse
	}
	/*
		记录已累计的 RequestID 。

		客户端在发送 ItemStackRequest 时需要发送一个 RequestID 。
		经过观察，这个值会随着请求发送的次数递减，且呈现为公差为 -2，
		首项为 -1 的递减型等差数列。

		特别地，如果你尝试在 RequestID 字段填写非负数或者偶数，
		那么客户端会被租赁服强制断开连接。

		尽管始终为 ItemStackRequest 的 RequestID 字段填写 -1 并不会造成造成断开连接的发生，
		但这样并不能保证物品操作的唯一性。

		因此，绝对地，请使用已提供的 API 发送物品操作请求，否则将导致程序惊慌
	*/
	currentRequestID int32
}

// 描述一个容器 ID
type ContainerID uint8

// 每个物品操作请求都会使用这样一个结构体，它用于描述单个的物品操作请求
type singleItemStackRequest struct {
	// 每个物品操作请求在发送前都应该上锁它以便于后续等待返回值时的阻塞
	lockDown *sync.Mutex
	// 描述多个库存(容器)中物品的变动结果。
	// 租赁服不会在返回 ItemStackResponce 时返回完整的物品数据，因此需要您提供对应
	// 槽位的更改结果以便于我们依此更新本地存储的库存数据
	datas map[ContainerID]StackRequestContainerInfo
}

// 描述单个库存(容器)中物品的变动结果
type StackRequestContainerInfo struct {
	// 其容器对应库存的窗口 ID
	WindowID uint32
	// 描述此容器中每个槽位的变动结果，键代表槽位编号，而值代表物品的新值。
	// 特别地，您无需设置物品数量和 NBT 中的物品名称以及物品的 StackNetworkID 信息，因为
	// 这些数据会在租赁服发回 ItemStackResponce 后被重新设置
	ChangeResult map[uint8]protocol.ItemInstance
}

/*
存储容器的 打开/关闭 状态，同时存储容器资源的占用状态。

重要：
容器由于是 PhoenixBuilder 的其中一个公用资源，因此为了公平性，
现在由我们(资源管理中心)负责完成对该公用资源的占用和释放之实现。

因此，为了绝对的安全，如果尝试绕过相关实现而直接 打开/关闭 容器，则会造成程序惊慌。

任何时刻，如果你需要打开或关闭容器，或者在某一段时间内使用某容器，则请提前占用此资源，
然后再发送相应数据包，完成后再释放此公用资源
*/
type container struct {
	// 容器被打开时的数据
	containerOpen struct {
		// 防止并发读写而设置的读写锁
		lockDown sync.RWMutex
		// 当客户端打开容器后，租赁服会以此数据包回应，届时此变量将被赋值。
		// 当容器被关闭或从未被打开，则此变量将会为 nil
		datas *packet.ContainerOpen
	}
	// 容器被关闭时的数据
	containerClose struct {
		// 防止并发读写而设置的读写锁
		lockDown sync.RWMutex
		/*
			客户端可以使用该数据包关闭已经打开的容器，
			而后，租赁服会以相同的数据包回应容器的关闭。

			当侦测到来自租赁服的响应，此变量将被赋值。
			当容器被打开或从未被关闭，则此变量将会为 nil
		*/
		datas *packet.ContainerClose
	}
	// 其他实现在打开或关闭容器后可能需要等待回应，此互斥锁便是为了完成这一实现
	awaitChanges sync.Mutex
	// PhoenixBuilder 同一时刻至多打开一个容器。此互斥锁是为了解决资源纠纷问题而设
	isUsing struct {
		// 当容器资源被占用时此互斥锁将会被锁定，否则反之
		lockDown sync.Mutex
		// 记录容器资源的占用者，用于确保资源释放者是占用者本身。
		// 此处应该记录一个 UUID
		holder string
	}
}
