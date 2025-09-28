package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const ContainerInfoCollection = "container_info"

type Container struct {
	ContainerId   string `json:"container_id" bson:"container_id"`
	ContainerName string `json:"container_name" bson:"container_name"`
	State         string `json:"state" bson:"state"`
	ImageId       string `json:"image_id" bson:"image_id"`
	ImageName     string `json:"image_name" bson:"image_name"`
	Pid           string `json:"pid" bson:"pid"`
	Pns           string `json:"pns" bson:"pns"`
	RunTime       string `json:"run_time" bson:"run_time"`
	CreateTime    string `json:"create_time" bson:"create_time"`
	PackageSeq    string `json:"package_seq" bson:"package_seq"`
}

type ContainerDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Container `bson:",inline" mapstructure:",squash"`
}

func DealContainerData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}

	var container Container
	err = json.Unmarshal(tmpData, &container)
	if err != nil {
		return err
	}

	containerDataDb := ContainerDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Container: container,
	}

	//将数据保存到数据库
	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(ContainerInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, containerDataDb)
	if err != nil {
		return err
	}
	return nil
}
