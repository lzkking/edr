package endpoint

import (
	"errors"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/lzkking/edr/service/config"
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

	defaultSyncTimeout = 2

	defaultMergeNum = 400
)

const (
	StatusGreen = iota
	StatusBlue
	StatusYellow
	StatusOrange
	StatusRed
)

var (
	defaultWeight = uint32(400)
)

const (
	syncUrl = "http://%s/service/sync"
)

var (
	EI *Endpoint
)

type RegisterInfo struct {
	Name     string `json:"name"`
	Ip       string `json:"ip"`
	Port     uint32 `json:"port"`
	Weight   uint32 `json:"weight"`
	Status   int    `json:"status"`    // 当前状态
	CreateAt int64  `json:"create_at"` // 创建的时间
	UpdateAt int64  `json:"update_at"` // 更新的时间
	Type     int    `json:"type"`
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

func (b *bucket) RefreshStatus() []RegisterInfo {
	b.mu.Lock()
	defer b.mu.Unlock()

	var redRegister []RegisterInfo

	for name, registerInfo := range b.data {
		d := time.Now().Unix() - registerInfo.UpdateAt
		if d <= 45 {
			registerInfo.Status = StatusGreen
		} else if d <= 60 {
			registerInfo.Status = StatusBlue
		} else if d <= 75 {
			registerInfo.Status = StatusYellow
		} else if d <= 90 {
			registerInfo.Status = StatusOrange
		} else {
			registerInfo.Status = StatusRed
			redRegister = append(redRegister, registerInfo)
		}
		b.data[name] = registerInfo
	}

	return redRegister
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
	if v, ok := b.data[name]; ok {
		v.Weight = registerInfo.Weight
		v.UpdateAt = time.Now().Unix()
		v.Status = StatusGreen
		b.data[name] = v
	} else {
		b.data[name] = RegisterInfo{
			Name:     registerInfo.Name,
			Ip:       registerInfo.Ip,
			Port:     registerInfo.Port,
			Weight:   registerInfo.Weight,
			CreateAt: time.Now().Unix(),
			UpdateAt: time.Now().Unix(),
			Type:     registerInfo.Type,
			Status:   StatusGreen,
		}
	}
}

func (b *bucket) GetGreenHosts() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	var hosts []string

	minWeight := uint32(10000)
	var host string
	for _, register := range b.data {
		if register.Status == StatusGreen {
			zap.S().Debugf("register info : %v, %v", register.Name, register.Weight)
			if minWeight > register.Weight {
				minWeight = register.Weight
				host = fmt.Sprintf("%v:%v", register.Ip, register.Port)
			}

			if register.Weight < defaultWeight {
				hosts = append(hosts, fmt.Sprintf("%v:%v", register.Ip, register.Port))
			}
		}
	}

	if len(hosts) == 0 && host != "" {
		hosts = append(hosts, host)
	}

	return hosts
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

	// 将信息同步给其他注册中心
	go endpoint.SyncSend()

	// 接收其他注册中心传递来的同步信息
	go endpoint.SyncRecv()

	// 刷新保存的注册者的状态信息
	go endpoint.Refresh()

	return endpoint
}

func (ei *Endpoint) Close() {
	close(ei.Stop)
}

func (ei *Endpoint) Refresh() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for _, b := range ei.buckets {
				redRegisters := b.RefreshStatus()
				if len(redRegisters) > 0 {
					for _, r := range redRegisters {
						zap.S().Debugf("redRegister info :%v", r.Name)
						ei.Delete(r)
					}
				}
			}
		case <-ei.Stop:
			zap.S().Warnf("endpoint refresh退出")
			return
		}
	}
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
				zap.S().Debugf("同步完成")
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
	zap.S().Debugf("将数据同步到其他注册中心")
	url := fmt.Sprintf(syncUrl, host)
	option := &grequests.RequestOptions{}
	option.JSON = transInfo
	option.RequestTimeout = defaultSyncTimeout * time.Second
	_, err := grequests.Post(url, option)
	if err != nil {
		zap.S().Warnf("同步信息到其他注册中心失败")
	}
	return
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

	zap.S().Debugf("注册者下线")

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

func (ei *Endpoint) GetGreenHosts(Type int) []string {
	if v, ok := ei.buckets[Type]; ok {
		return v.GetGreenHosts()
	}
	return []string{}
}

func init() {
	EI = NewEndpoint()
	defaultWeight = config.GetServerConfig().Weight
}
