package mongodb

import (
	"context"
	"fmt"
	"github.com/lzkking/edr/manager/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

var (
	once          sync.Once
	MongodbClient *mongo.Client
	DatabaseName  string
)

func NewMongoClient(host, user, password, authDB string, minPoolSize, maxPoolSize uint64, retryWrites, enableAuth bool) (client *mongo.Client, err error) {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		uri := fmt.Sprintf("mongodb://%s", host)

		var opts *options.ClientOptions
		if enableAuth {
			cred := options.Credential{
				Username:   user,
				Password:   password,
				AuthSource: authDB,
			}
			opts = options.Client().ApplyURI(uri).SetAuth(cred).SetMinPoolSize(minPoolSize).SetMaxPoolSize(maxPoolSize).SetRetryWrites(retryWrites)
		} else {
			opts = options.Client().ApplyURI(uri).SetMinPoolSize(minPoolSize).SetMaxPoolSize(maxPoolSize).SetRetryWrites(retryWrites)
		}

		client, err = mongo.Connect(ctx, opts)
		if err != nil {
			return
		}

		err = client.Ping(ctx, nil)
	})

	return
}

func init() {
	mongodbConfig := config.GetServerConfig().MongodbConfig
	var err error
	MongodbClient, err = NewMongoClient(
		mongodbConfig.Host,
		mongodbConfig.User,
		mongodbConfig.Password,
		mongodbConfig.AuthDB,
		mongodbConfig.MinPoolSize,
		mongodbConfig.MaxPoolSize,
		mongodbConfig.RetryWrites,
		mongodbConfig.EnableAuth,
	)
	DatabaseName = mongodbConfig.DB
	if err != nil {
		panic("连接mongodb失败")
	}
}
