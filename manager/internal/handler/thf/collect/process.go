package collect

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/manager/pkg/mongodb"
)

const ProcessInfoCollection = "processinfo"

type ProcessInfo struct {
	Pid        string `mapstructure:"pid" json:"pid" bson:"pid"`
	Cwd        string `mapstructure:"cwd" json:"cwd" bson:"cwd"`
	CmdLine    string `mapstructure:"cmdline" json:"cmdline" bson:"cmdline"`
	CheckSum   string `mapstructure:"checksum" json:"checksum" bson:"checksum"`
	ExeHash    string `mapstructure:"exe_hash" json:"exe_hash" bson:"exe_hash"`
	Exe        string `mapstructure:"exe" json:"exe" bson:"exe"`
	Comm       string `mapstructure:"comm" json:"comm" bson:"comm"` // Stat
	State      string `mapstructure:"state" json:"state" bson:"state"`
	Ppid       string `mapstructure:"ppid" json:"ppid" bson:"ppid"`
	Pgid       string `mapstructure:"pgid" json:"pgid" bson:"pgid"`
	Sid        string `mapstructure:"sid" json:"sid" bson:"sid"`
	StartTime  string `mapstructure:"start_time" json:"start_time" bson:"start_time"`
	Umask      string `mapstructure:"umask" json:"umask" bson:"umask"` // Status
	TracerPid  string `mapstructure:"tcpid" json:"tcpid" bson:"tcpid"`
	Ruid       string `mapstructure:"ruid" json:"ruid" bson:"ruid"`
	Euid       string `mapstructure:"euid" json:"euid" bson:"euid"`
	Suid       string `mapstructure:"suid" json:"suid" bson:"suid"`
	Fsuid      string `mapstructure:"fsuid" json:"fsuid" bson:"fsuid"`
	Rgid       string `mapstructure:"rgid" json:"rgid" bson:"rgid"`
	Egid       string `mapstructure:"egid" json:"egid" bson:"egid"`
	Sgid       string `mapstructure:"sgid" json:"sgid" bson:"sgid"`
	Fsgid      string `mapstructure:"fsgid" json:"fsgid" bson:"fsgid"`
	Rusername  string `mapstructure:"rusername" json:"rusername" bson:"rusername"`
	Eusername  string `mapstructure:"eusername" json:"eusername" bson:"eusername"`
	Susername  string `mapstructure:"susername" json:"susername" bson:"susername"`
	Fsusername string `mapstructure:"fsusername" json:"fsusername" bson:"fsusername"`
	NsPid      string `mapstructure:"nspid" json:"nspid" bson:"nspid"`
	NsPgid     string `mapstructure:"nspgid" json:"nspgid" bson:"nspgid"`
	NsSid      string `mapstructure:"nssid" json:"nssid" bson:"nssid"`
	Diff       string `mapstructure:"dns" json:"dns" bson:"dns"` // ProcessNamespace
	Cgroup     string `mapstructure:"cns" json:"cns" bson:"cns"`
	Ipc        string `mapstructure:"ins" json:"ins" bson:"ins"`
	Mnt        string `mapstructure:"mns" json:"mns" bson:"mns"`
	Net        string `mapstructure:"nns" json:"nns" bson:"nns"`
	Pns        string `mapstructure:"pns" json:"pns" bson:"pns"`
	Time       string `mapstructure:"tns" json:"tns" bson:"tns"`
	User       string `mapstructure:"uns" json:"uns" bson:"uns"`
	Uts        string `mapstructure:"utns" json:"utns" bson:"utns"`
	PackageSeq string `mapstructure:"package_seq" json:"package_seq" bson:"package_seq"`
}

type ProcessDataDB struct {
	AgentID     string `json:"agent_id" bson:"agent_id"`
	AgentTime   int64  `json:"agent_time" bson:"agent_time"`
	ProcessInfo `bson:",inline" mapstructure:",squash"`
}

func DealProcessData(c *gin.Context, collectData *CollectData) error {
	tmpData, err := json.Marshal(collectData.Fields)
	if err != nil {
		return err
	}
	var processInfo ProcessInfo
	err = json.Unmarshal(tmpData, &processInfo)
	if err != nil {
		return err
	}

	p := ProcessDataDB{
		AgentID:     collectData.AgentID,
		AgentTime:   collectData.AgentTime,
		ProcessInfo: processInfo,
	}

	//将数据保存到数据库
	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(ProcessInfoCollection)
	_, err = agentsCollection.InsertOne(c, p)
	if err != nil {
		return err
	}

	return nil
}
