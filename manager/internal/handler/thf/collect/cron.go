package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const CronInfoCollection = "cron_info"

type Cron struct {
	Path       string `json:"path" bson:"path"`
	Username   string `json:"username" bson:"username"`
	Schedule   string `json:"schedule" bson:"schedule"`
	Command    string `json:"command" bson:"command"`
	Checksum   string `json:"checksum" bson:"checksum"`
	PackageSeq string `json:"package_seq" bson:"package_seq"`
}

type CronDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Cron      `bson:",inline" mapstructure:",squash"`
}

func DealCronData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}

	var cron Cron
	err = json.Unmarshal(tmpData, &cron)
	if err != nil {
		return err
	}

	cronDataDb := CronDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Cron:      cron,
	}

	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(CronInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, cronDataDb)
	if err != nil {
		return err
	}

	return nil
}
