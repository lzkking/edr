package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const IntegrityInfoCollection = "integrity_info"

type Integrity struct {
	SoftwareName    string `json:"software_name" bson:"software_name"`
	Digest          string `json:"digest" bson:"digest"`
	OriginDigest    string `json:"origin_digest" bson:"origin_digest"`
	DigestAlgorithm string `json:"digest_algorithm" bson:"digest_algorithm"`
	Exe             string `json:"exe" bson:"exe"`
	ModifyTime      string `json:"modify_time" bson:"modify_time"`
	SoftwareVersion string `json:"software_version" bson:"software_version"`
	PackageSeq      string `json:"package_seq" bson:"package_seq"`
}

type IntegrityDataDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	Integrity `bson:",inline" mapstructure:",squash"`
}

func DealIntegrityData(ctx *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}

	var integrity Integrity
	err = json.Unmarshal(tmpData, &integrity)
	if err != nil {
		return err
	}

	integrityDataDb := IntegrityDataDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		Integrity: integrity,
	}

	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(IntegrityInfoCollection)
	_, err = agentsCollection.InsertOne(ctx, integrityDataDb)
	if err != nil {
		return err
	}

	return nil
}
