package agentcenter

import (
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp/v3"
	"github.com/lzkking/edr/manager/internal/handler/manager/plugins"
	"github.com/lzkking/edr/manager/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"net/http"
)

const AgentsCollection = "agents"

type AgentBasicInfo struct {
	AgentID      string   `bson:"agent_id" json:"agent_id"`
	ExtranetIPv4 []string `bson:"extranet_ipv4" json:"extranet_ipv4"`
	ExtranetIPv6 []string `bson:"extranet_ipv6" json:"extranet_ipv6"`
	IntranetIPv4 []string `bson:"intranet_ipv4" json:"intranet_ipv4"`
	IntranetIPv6 []string `bson:"intranet_ipv6" json:"intranet_ipv6"`
	Hostname     string   `bson:"hostname" json:"hostname"`
}

type AgentDetailInfo struct {
	AgentBasicInfo    `bson:",inline"`
	Version           string `bson:"version" json:"version"`
	Product           string `bson:"product" json:"product"`
	SourceAddr        string `bson:"source_addr" json:"source_addr"`
	CreateAt          int64  `bson:"create_at" json:"create_at"`
	LastHeartbeatTime int64  `bson:"last_heartbeat_time" json:"last_heartbeat_time"`

	//系统信息
	Arch            string `bson:"arch" json:"arch"`
	KernelVersion   string `bson:"kernel_version" json:"kernel_version"`
	Platform        string `bson:"platform" json:"platform"`
	PlatformFamily  string `bson:"platform_family" json:"platform_family"`
	PlatformVersion string `bson:"platform_version" json:"platform_version"`

	//进程信息
	CpuPercent string `bson:"cpu_percent" json:"cpu_percent"`
	Rss        string `bson:"rss" json:"rss"`
	ReadSpeed  string `bson:"read_speed" json:"read_speed"`
	WriteSpeed string `bson:"write_speed" json:"write_speed"`
	Fds        string `bson:"fds" json:"fds"`
	StartAt    string `bson:"start_at" json:"start_at"`

	//linux平均负载
	Load1        string `bson:"load_1" json:"load_1"`
	Load5        string `bson:"load_5" json:"load_5"`
	Load15       string `bson:"load_15" json:"load_15"`
	RunningProcs string `bson:"running_procs" json:"running_procs"`
	TotalProcs   string `bson:"total_procs" json:"total_procs"`

	//硬件信息
	HostSerial string `bson:"host_serial" json:"host_serial"`
	HostID     string `bson:"host_id" json:"host_id"`
	HostModel  string `bson:"host_model" json:"host_model"`
	HostVendor string `bson:"host_vendor" json:"host_vendor"`

	//DNS信息
	Dnss string `bson:"dnss" json:"dnss"`

	//gateway信息
	Gateway string `bson:"gateway" json:"gateway"`

	// 工作路径大小
	WorkDirSize string `bson:"work_dir_size" json:"work_dir_size"`

	// 协程数量,可使用cpu核心数
	NumGoroutine string `bson:"num_goroutine" json:"num_goroutine"`
	NumMaxProcs  string `bson:"num_max_procs" json:"num_max_procs"`

	// CPU信息
	CpuName       string `bson:"cpu_name" json:"cpu_name"`
	BootTime      string `bson:"boot_time" json:"boot_time"`
	SysCpuPercent string `bson:"sys_cpu_percent" json:"sys_cpu_percent"`
	SysMemPercent string `bson:"sys_men_percent" json:"sys_mem_percent"`
}

// GetAgentConfig - 获取传递来的心跳包,并返回插件数据
func GetAgentConfig(c *gin.Context) {
	var agentDetail AgentDetailInfo
	err := c.ShouldBindJSON(&agentDetail)
	if err != nil {
		zap.S().Warnf("传递来的json数据有问题,shouldbindjson失败,失败原因:%v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "传递来的结构有问题",
		})
		return
	}

	zap.S().Debugf("传递来的心跳包数据:%v", pp.Sprintf("%v", agentDetail))

	//	将传递来的心跳包数据保存到数据库中
	filter := bson.M{
		"agent_id": agentDetail.AgentID,
	}
	update := bson.M{
		"$set": agentDetail,
	}
	opts := options.Update().SetUpsert(true)
	agentsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(AgentsCollection)
	_, err = agentsCollection.UpdateOne(c, filter, update, opts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "将心跳包数据保存到数据库失败"})
		zap.S().Warnf("将心跳包数据保存到数据库失败,失败原因:%v", err)
		return
	}

	//从数据库中读取插件配置信息,并将数据返回
	plgConfigs, err := plugins.GetPlgConfigs(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "从数据库中获取插件配置数据失败"})
		return
	}

	if len(plgConfigs) > 0 {
		c.JSON(http.StatusOK, []plugins.PlgConfig{})
	} else {
		c.JSON(http.StatusOK, plgConfigs)
	}
}
