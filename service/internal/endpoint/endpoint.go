package endpoint

import (
	"errors"
	"github.com/lzkking/edr/service/internal/cluster"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	TypeAgentCenterHttp = 0
	TypeAgentCenterGrpc = 1
	TypeManager         = 2

	ActionRegister = "Register"
	ActionDelete   = "Delete"

	defaultSendChannelLen = 1024 * 8
	defaultRecvChannelLen = 1024 * 8
	defaultSendTime       = 2
	defaultRecvTime       = 10

	defaultMergeNum = 400
)

type RegisterInfo struct {
	Name   string `json:"name"`
	Ip     string `json:"ip"`
	Port   uint32 `json:"port"`
	Weight uint32 `json:"weight"`
	Type   int    `json:"type"`
}

type SyncInfo struct {
	Action   string       `json:"action"`
	Registry RegisterInfo `json:"registry"`
}

type TransInfo struct {
	Source string     `json:"source"`
	Data   []SyncInfo `json:"data"`
}

type bucket struct {
	mu   sync.RWMutex
	data map[string]RegisterInfo
}

func (b *bucket) GetRegisterInfo(name string) RegisterInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if v, ok := b.data[name]; ok {
		return v
	}
	return RegisterInfo{}
}

func (b *bucket) SetRegisterInfo(name string, registerInfo RegisterInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[name] = registerInfo
}

func (b *bucket) DeleteRegisterInfo(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.data, name)
}

type Endpoint struct {
	buckets  map[int]*bucket
	sendChan chan SyncInfo
	recvChan chan TransInfo
	Stop     chan bool
}

func NewEndpoint() *Endpoint {
	endpoint := &Endpoint{
		buckets: map[int]*bucket{
			TypeAgentCenterHttp: {data: make(map[string]RegisterInfo)},
			TypeAgentCenterGrpc: {data: make(map[string]RegisterInfo)},
			TypeManager:         {data: make(map[string]RegisterInfo)},
		},
		sendChan: make(chan SyncInfo, defaultSendChannelLen),
		recvChan: make(chan TransInfo, defaultRecvChannelLen),
		Stop:     make(chan bool),
	}

	return endpoint
}

func (ei *Endpoint) SyncSend() {
	ticker := time.NewTicker(time.Second * defaultSendTime)
	defer ticker.Stop()
	syncInfoList := make([]SyncInfo, 0)
	for {
		select {
		case syncInfo := <-ei.sendChan:
			syncInfoList = append(syncInfoList, syncInfo)
			if len(syncInfoList) > defaultMergeNum {
				transInfo := TransInfo{
					Source: "",
					Data:   syncInfoList,
				}

				wg := &sync.WaitGroup{}
				hosts := cluster.GlobalCluster.GetOtherHost()
				//	将数据同步到所有的注册中心
				for _, host := range hosts {
					wg.Add(1)
					send(wg, host, transInfo)
				}

				wg.Wait()
				syncInfoList = make([]SyncInfo, 0)
			}
		case <-ticker.C:
			if len(syncInfoList) > 0 {
				transInfo := TransInfo{
					Source: "",
					Data:   syncInfoList,
				}

				wg := &sync.WaitGroup{}
				hosts := cluster.GlobalCluster.GetOtherHost()
				//	将数据同步到所有的注册中心
				for _, host := range hosts {
					wg.Add(1)
					send(wg, host, transInfo)
				}

				wg.Wait()
				syncInfoList = make([]SyncInfo, 0)
			}
		case <-ei.Stop:
			zap.S().Warnf("endpoint sync send 退出")
			return
		}
	}
}

func send(wg *sync.WaitGroup, host string, transInfo TransInfo) {
	defer wg.Done()

	//将数据同步到其他注册中心
}

func (ei *Endpoint) SyncRecv() {
	ticker := time.NewTicker(time.Second * defaultRecvTime)
	defer ticker.Stop()
	for {
		select {
		case transInfo := <-ei.recvChan:
			for _, syncInfo := range transInfo.Data {
				switch syncInfo.Action {
				case ActionRegister:
					// 保存到map表中
					if v, ok := ei.buckets[syncInfo.Registry.Type]; ok {
						v.SetRegisterInfo(syncInfo.Registry.Name, syncInfo.Registry)
					}
				case ActionDelete:
					//删除
					if v, ok := ei.buckets[syncInfo.Registry.Type]; ok {
						v.DeleteRegisterInfo(syncInfo.Registry.Name)
					}
				}
			}
		case <-ticker.C:
		case <-ei.Stop:
			zap.S().Warnf("endpoint sync recv退出")
			return

		}
	}
}

func (ei *Endpoint) Recv(transInfo TransInfo) error {
	select {
	case ei.recvChan <- transInfo:
	default:
		zap.S().Warnf("endpoint的recv chan 满了")
		return errors.New("endpoint的recv chan 满了")
	}
	return nil
}

func (ei *Endpoint) Register(registerInfo RegisterInfo) {
	if _, ok := ei.buckets[registerInfo.Type]; !ok {
		return
	}

	tmpRegisterInfo := ei.buckets[registerInfo.Type].GetRegisterInfo(registerInfo.Name)
	if tmpRegisterInfo.Name == registerInfo.Name &&
		tmpRegisterInfo.Ip == registerInfo.Ip &&
		tmpRegisterInfo.Port == registerInfo.Port &&
		tmpRegisterInfo.Weight == registerInfo.Weight {
		return
	}

	ei.buckets[registerInfo.Type].SetRegisterInfo(registerInfo.Name, registerInfo)
	// 将信息同步给其他

	syncInfo := SyncInfo{
		Action:   ActionRegister,
		Registry: registerInfo,
	}

	select {
	case ei.sendChan <- syncInfo:
	default:
		zap.S().Warnf("endpoint未能成功将数据进行同步")
	}
}

func (ei *Endpoint) Delete(registerInfo RegisterInfo) {
	if _, ok := ei.buckets[registerInfo.Type]; !ok {
		return
	}

	ei.buckets[registerInfo.Type].DeleteRegisterInfo(registerInfo.Name)

	syncInfo := SyncInfo{
		Action:   ActionDelete,
		Registry: registerInfo,
	}

	select {
	case ei.sendChan <- syncInfo:
	default:
		zap.S().Warnf("endpoint未能成功将数据进行同步")
	}
}
