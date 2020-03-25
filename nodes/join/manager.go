package join

import (
	"encoding/json"
	"github.com/ihaiker/gokit/errors"
	"github.com/ihaiker/gokit/remoting/rpc"
	"github.com/ihaiker/sudis/daemon"
	"math"
	"time"
)

type ToJoinManager struct {
	key, salt    string
	joined       map[string]*joinClient
	shutdown     bool
	OnRpcMessage rpc.OnMessage
}

func New(key, salt string) *ToJoinManager {
	return &ToJoinManager{
		key: key, salt: salt,
		joined: make(map[string]*joinClient),
	}
}

func (self *ToJoinManager) MustJoinIt(address string) {
	maxWaitSeconds := 5 * 60
	go func() {
		for i := 0; !self.shutdown; i++ {
			if err := self.Join(address); err == nil {
				return
			}
			seconds := int(math.Pow(2, float64(i)))
			if seconds > maxWaitSeconds {
				seconds = maxWaitSeconds
			}
			time.Sleep(time.Second * time.Duration(maxWaitSeconds))
			logger.Debug("重试连接主控节点：", address)
		}
	}()
}

func (self *ToJoinManager) Join(address string) (err error) {
	//已经连接成功了，这里的操作是为了客户端连接主控节点异常后，
	//使用命令主动再次连接的判断，因为客户端使用了指数递增方式等待，所以后面的等待是时间将会很长
	if _, has := self.joined[address]; has {
		return
	}

	client := newClient(address, self.salt, self.key, self.OnRpcMessage)
	err = client.Start()
	if err != nil {
		_ = errors.Safe(client.Stop)
		logger.Warn("连接主控异常：", err)
		return
	}
	logger.Info("连接主控: ", address)
	self.joined[address] = client
	return err
}

func (self *ToJoinManager) OnProgramStatusEvent(event daemon.FSMStatusEvent) {
	defer errors.Catch()
	request := &rpc.Request{URL: "program.status"}
	request.Body, _ = json.Marshal(&event)
	for _, client := range self.joined {
		client.Notify(request)
	}
}

func (self *ToJoinManager) Start() error {
	return nil
}

func (self *ToJoinManager) Stop() error {
	logger.Info("multi join stop")
	self.shutdown = true
	for _, client := range self.joined {
		_ = errors.Safe(client.Stop)
	}
	return nil
}