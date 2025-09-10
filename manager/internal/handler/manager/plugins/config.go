package plugins

import (
	"context"
	"github.com/lzkking/edr/manager/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
)

const PlgConfigCollection = "plg_configs"

type PlgConfig struct {
	Name         string   `bson:"name" json:"name"`
	Type         string   `bson:"type" json:"type"`
	Version      string   `bson:"version" json:"version"`
	SHA256       string   `bson:"sha256" json:"sha256"`
	Signature    string   `bson:"signature" json:"signature"`
	DownloadUrls []string `bson:"download_urls" json:"download_urls"`
	Detail       string   `bson:"detail" json:"detail"`
	Key          string   `bson:"key"`
}

func GetPlgConfigs(ctx context.Context) ([]PlgConfig, error) {
	plgConfigsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(PlgConfigCollection)

	cur, err := plgConfigsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var plgConfigs []PlgConfig
	if err = cur.All(ctx, &plgConfigs); err != nil {
		return nil, err
	}
	return plgConfigs, nil
}
