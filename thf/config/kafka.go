package config

type KafkaConfig struct {
	KafkaAdders []string `json:"kafka_adders"`
	Topics      []string `json:"topic"`
	GroupID     string   `json:"group_id"`
	ClientID    string   `json:"client_id"`
	LogPath     string   `json:"log_path"`
	EnableAuth  bool     `json:"enable_auth"`
	UserName    string   `json:"user_name"`
	Password    string   `json:"password"`
}
