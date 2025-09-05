package agentcenter

import "github.com/gin-gonic/gin"

type AgentBasicInfo struct {
	AgentID      string   `bson:"agent_id"`
	ExtranetIPv4 []string `bson:"extranet_ipv4"`
	ExtranetIPv6 []string `bson:"extranet_ipv6"`
	IntranetIPv4 []string `bson:"intranet_ipv4"`
	IntranetIPv6 []string `bson:"intranet_ipv6"`
	Hostname     string   `bson:"hostname"`
}

type AgentDetailInfo struct {
	AgentBasicInfo `bson:",inline"`
}

func GetAgentConfig(c *gin.Context) {

}
