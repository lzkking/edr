package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const NetInterfaceInfoCollection = "net_interface_info"

type NetInterface struct {
	Name         string `json:"name" bson:"name"`
	HardwareAddr string `json:"hardware_addr" bson:"hardware_addr"`
	Addrs        string `json:"addrs" bson:"addrs"`
	Index        string `json:"index" bson:"index"`
	Mtu          string `json:"mtu" bson:"mtu"`
	PackageSeq   string `json:"package_seq" bson:"package_seq"`
}

type NetInterfaceDataDB struct {
	AgentID      string `json:"agent_id" bson:"agent_id"`
	AgentTime    int64  `json:"agent_time" bson:"agent_time"`
	NetInterface `bson:",inline" mapstructure:",squash"`
}

func DealNetInterfaceData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}

	var netInterface NetInterface
	err = json.Unmarshal(tmpData, &netInterface)
	if err != nil {
		return err
	}

	netInterfaceDataDb := NetInterfaceDataDB{
		AgentID:      collectData.AgentID,
		AgentTime:    collectData.AgentTime,
		NetInterface: netInterface,
	}

	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(NetInterfaceInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, netInterfaceDataDb)
	if err != nil {
		return err
	}

	return nil
}
