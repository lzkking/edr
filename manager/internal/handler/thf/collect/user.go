package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const UserInfoCollection = "userinfo"

type UserInfo struct {
	Gid                 string `json:"gid" bson:"gid"`
	Groupname           string `json:"groupname" bson:"groupname"`
	Home                string `json:"home" bson:"home"`
	Info                string `json:"info" bson:"info"`
	LastLoginIp         string `json:"last_login_ip" bson:"last_login_ip"`
	LastLoginTime       string `json:"last_login_time" bson:"last_login_time"`
	PackageSeq          string `json:"package_seq" bson:"package_seq"`
	Password            string `json:"password" bson:"password"`
	Shell               string `json:"shell" bson:"shell"`
	Sudoers             string `json:"sudoers" bson:"sudoers"`
	Uid                 string `json:"uid" bson:"uid"`
	Username            string `json:"username" bson:"username"`
	WeakPassword        string `json:"weak_password" bson:"weak_password"`
	WeakPasswordContent string `json:"weak_password_content" bson:"weak_password_content"`
}

type UserInfoDB struct {
	AgentID   string `json:"agent_id" bson:"agent_id"`
	AgentTime int64  `json:"agent_time" bson:"agent_time"`
	UserInfo  `bson:",inline" mapstructure:",squash"`
}

func DealUserInfoData(c *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}
	var userInfo UserInfo
	err = json.Unmarshal(tmpData, &userInfo)
	if err != nil {
		return err
	}

	u := UserInfoDB{
		AgentID:   collectData.AgentID,
		AgentTime: collectData.AgentTime,
		UserInfo:  userInfo,
	}

	//将数据保存到数据库
	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(UserInfoCollection)
	_, err = agentsCollection.InsertOne(c, u)
	if err != nil {
		return err
	}
	return nil
}
