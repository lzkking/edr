package config

type MongodbConfig struct {
	Host        string `json:"host"`
	User        string `json:"user"`
	Password    string `json:"password"`
	AuthDB      string `json:"auth_db"`
	DB          string `json:"db"`
	MinPoolSize uint64 `json:"min_pool_size"`
	MaxPoolSize uint64 `json:"max_pool_size"`
	RetryWrites bool   `json:"retry_writes"`
	EnableAuth  bool   `json:"enable_auth"`
}
