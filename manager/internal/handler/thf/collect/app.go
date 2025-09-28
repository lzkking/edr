package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const AppInfoCollection = "app_info"

type App struct {
	Name          string `json:"name" bson:"name"`
	Type          string `json:"type" bson:"type"`
	SVersion      string `json:"sversion" bson:"sversion"`
	Conf          string `json:"conf" bson:"conf"`
	ContainerId   string `json:"container_id" bson:"container_id"`
	ContainerName string `json:"container_name" bson:"container_name"`
	Pid           string `json:"pid" bson:"pid"`
	Exe           string `json:"exe" bson:"exe"`
	StartTime     string `json:"start_time" bson:"start_time"`
	PackageSeq    string `json:"package_seq" bson:"package_seq"`
}

type AppDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	App       `bson:",inline" mapstructure:",squash"`
}

func DealAppData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}

	var app App
	err = json.Unmarshal(tmpData, &app)
	if err != nil {
		return err
	}

	appDataDb := AppDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		App:       app,
	}

	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(AppInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, appDataDb)
	if err != nil {
		return err
	}
	return nil
}
