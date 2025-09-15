package collect

import (
	"fmt"
	"github.com/levigross/grequests"
	pb "github.com/lzkking/edr/edrproto"
	"github.com/lzkking/edr/thf/internal/manager"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"math/rand"
)

const PostCollectDataUrl = "http://%s/thf/collect"

type CollectData struct {
	AgentID   string            `json:"agent_id"`
	DataType  int32             `json:"data_type"`
	AgentTime int64             `json:"agent_time"`
	Fields    map[string]string `json:"fields"`
}

func DealCollectTaskData(m *pb.MQData) {
	datatype := m.DataType
	agentID := m.AgentID
	agentTime := m.AgentTime

	var p pb.Payload
	err := proto.Unmarshal(m.Body, &p)
	if err != nil {
		return
	}

	c := CollectData{
		AgentID:   agentID,
		DataType:  datatype,
		Fields:    p.Fields,
		AgentTime: agentTime,
	}

	data := &grequests.RequestOptions{
		JSON: c,
	}

	if len(manager.ManagerHosts) == 0 {
		err = manager.GetManagerHosts()
		if err != nil {
			return
		}
	}

	//	将收集到的用户数据同步到manager，manager负责入库
	_, err = grequests.Post(fmt.Sprintf(PostCollectDataUrl, manager.ManagerHosts[rand.Intn(len(manager.ManagerHosts))]), data)
	if err != nil {
		zap.S().Errorf("向manager同步收集的用户数据失败,失败的原因为:%v", err)
		return
	}
}
