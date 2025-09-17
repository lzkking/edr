package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const SoftwareInfoCollection = "software_info"

type Software struct {
	Name    string `mapstructure:"name" json:"name" bson:"name"`
	Version string `mapstructure:"sversion" json:"sversion" bson:"sversion"`
	// dpkg rpm pypi jar
	Type string `mapstructure:"type" json:"type" bson:"type"`
	// dpkg
	Source string `mapstructure:"source" json:"source" bson:"source"`
	Status string `mapstructure:"status" json:"status" bson:"status"`
	// rpm
	Vendor           string `mapstructure:"vendor" json:"vendor" bson:"vendor"`
	ComponentVersion string `mapstructure:"component_version" json:"component_version" bson:"component_version"`
	// jar
	Pid     string `mapstructure:"pid" json:"pid" bson:"pid"`
	PodName string `mapstructure:"pod_name" json:"pod_name" bson:"pod_name"`
	Psm     string `mapstructure:"psm" json:"psm" bson:"psm"`

	PackageSeq string `mapstructure:"package_seq" json:"package_seq" bson:"package_seq"`
}

type SoftwareDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Software  `bson:",inline" mapstructure:",squash"`
}

func DealSoftwareData(c *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}
	var softwareInfo Software
	err = json.Unmarshal(tmpData, &softwareInfo)
	if err != nil {
		return err
	}

	s := SoftwareDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Software:  softwareInfo,
	}

	//将数据保存到数据库
	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(SoftwareInfoCollection)
	_, err = agentsCollection.InsertOne(c, s)
	if err != nil {
		return err
	}
	return nil
}
