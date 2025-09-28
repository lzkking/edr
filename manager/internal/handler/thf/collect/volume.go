package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const VolumeInfoCollection = "volume_info"

type Volume struct {
	Name       string `json:"name" bson:"name"`
	FsType     string `json:"fstype" bson:"fstype"`
	MountPoint string `json:"mount_point" bson:"mount_point"`
	Total      string `json:"total" bson:"total"`
	Used       string `json:"used" bson:"used"`
	Free       string `json:"free" bson:"free"`
	Usage      string `json:"usage" bson:"usage"`
	PackageSeq string `json:"package_seq" bson:"package_seq"`
}

type VolumeDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Volume    `bson:",inline" mapstructure:",squash"`
}

func DealVolumeData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}
	var volume Volume
	err = json.Unmarshal(tmpData, &volume)
	if err != nil {
		return err
	}

	volumeDataDb := VolumeDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Volume:    volume,
	}

	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(VolumeInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, volumeDataDb)
	if err != nil {
		return err
	}
	return nil
}
