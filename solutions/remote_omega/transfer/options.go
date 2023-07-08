package transfer

const (
	DefaultPubSubAccessPoint      = "ipc://neomega_pub_sub.ipc" //"tcp://localhost:24016"
	DefaultCtrlAccessPoint        = "ipc://neomega_ctrl.ipc"    //"tcp://localhost:24015"
	DefaultDirectPubSubModeEnable = true
	DefaultDirectSendModeEnable   = true
)

type EndPointOption struct {
	PubAccessPoint  string
	CtrlAccessPoint string
	DirectSendMode  bool
	DirectSubMode   bool
}

func MakeDefaultEndPointOption() *EndPointOption {
	return &EndPointOption{
		PubAccessPoint:  DefaultPubSubAccessPoint,
		CtrlAccessPoint: DefaultCtrlAccessPoint,
		DirectSendMode:  DefaultDirectSendModeEnable,
		DirectSubMode:   DefaultDirectPubSubModeEnable,
	}
}
