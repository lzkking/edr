package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const KmodInfoCollection = "kmod_info"

type Kmod struct {
	Name       string `json:"name" bson:"name"`
	Size       string `json:"size" bson:"size"`
	Refcount   string `json:"refcount" bson:"refcount"`
	UsedBy     string `json:"used_by" bson:"used_by"`
	State      string `json:"state" bson:"state"`
	Addr       string `json:"addr" bson:"addr"`
	PackageSeq string `json:"package_seq" bson:"package_seq"`
}

type KmodDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Kmod      `bson:",inline" mapstructure:",squash"`
}

func DealKmodData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}

	var kmod Kmod
	err = json.Unmarshal(tmpData, &kmod)
	if err != nil {
		return err
	}

	kmodDataDb := KmodDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Kmod:      kmod,
	}

	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(KmodInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, kmodDataDb)
	if err != nil {
		return err
	}

	return nil
}
