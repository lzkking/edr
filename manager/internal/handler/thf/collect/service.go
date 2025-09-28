package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const ServiceInfoCollection = "service_info"

type Service struct {
	Name       string `json:"name" bson:"name"`
	Type       string `json:"type" bson:"type"`
	Command    string `json:"command" bson:"command"`
	Restart    string `json:"restart" bson:"restart"`
	WorkingDir string `json:"working_dir" bson:"working_dir"`
	Checksum   string `json:"checksum" bson:"checksum"`
	BusName    string `json:"bus_name" bson:"bus_name"`
	PackageSeq string `json:"package_seq" bson:"package_seq"`
}

type ServiceDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Service   `bson:",inline" mapstructure:",squash"`
}

func DealServiceData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}

	var service Service
	err = json.Unmarshal(tmpData, &service)
	if err != nil {
		return err
	}

	serviceDataDB := ServiceDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Service:   service,
	}

	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(ServiceInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, serviceDataDB)
	if err != nil {
		return err
	}
	return nil
}
