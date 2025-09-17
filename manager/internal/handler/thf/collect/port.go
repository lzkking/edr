package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const PortInfoCollection = "portinfo"

type Port struct {
	// from inet
	Family   string `mapstructure:"family" json:"family" bson:"family"`
	Protocol string `mapstructure:"protocol" json:"protocol" bson:"protocol"`
	State    string `mapstructure:"state" json:"state" bson:"state"`
	Sport    string `mapstructure:"sport" json:"sport" bson:"sport"`
	Dport    string `mapstructure:"dport" json:"dport" bson:"dport"`
	Sip      string `mapstructure:"sip" json:"sip" bson:"sip"`
	Dip      string `mapstructure:"dip" json:"dip" bson:"dip"`
	Uid      string `mapstructure:"uid" json:"uid" bson:"uid"`
	Inode    string `mapstructure:"inode" json:"inode" bson:"inode"`
	Username string `mapstructure:"username" json:"username" bson:"username"`

	// from process
	Pid        string `mapstructure:"pid" json:"pid" bson:"pid"`
	Exe        string `mapstructure:"exe" json:"exe" bson:"exe"`
	Comm       string `mapstructure:"comm" json:"comm" bson:"comm"`
	Cmdline    string `mapstructure:"cmdline" json:"cmdline" bson:"cmdline"`
	Psm        string `mapstructure:"psm" json:"psm" bson:"psm"`
	PodName    string `mapstructure:"pod_name" json:"pod_name" bson:"pod_name"`
	PackageSeq string `mapstructure:"package_seq" json:"package_seq" bson:"package_seq"`
}

type PortDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Port      `bson:",inline" mapstructure:",squash"`
}

func DealPortData(c *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}
	var portInfo Port
	err = json.Unmarshal(tmpData, &portInfo)
	if err != nil {
		return err
	}

	p := PortDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Port:      portInfo,
	}

	//将数据保存到数据库
	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(PortInfoCollection)
	_, err = agentsCollection.InsertOne(c, p)
	if err != nil {
		return err
	}

	return nil
}
