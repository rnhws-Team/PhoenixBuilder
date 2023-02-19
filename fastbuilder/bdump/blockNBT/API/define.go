package blockNBT_API

import (
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
)

// ------------------------- sendwscmdwithresp.go -------------------------

// 存放等待命令反馈列的结构体
type CommandRequestWaiting struct {
	List map[string]bool // 储存还需要等待命令反馈的列(键为 uuid 的字符串形式)
	Lock sync.RWMutex    // 防止并发读写而上的锁
}

// 存放命令反馈池的结构体
type CommandOutput struct {
	Pool map[string]packet.CommandOutput // 储存命令反馈(键为 uuid 的字符串形式)
	Lock sync.RWMutex                    // 防止并发读写而上的锁
}

var CommandRequestWaitingList CommandRequestWaiting
var CommandOutputPool CommandOutput

// ------------------------- end -------------------------
