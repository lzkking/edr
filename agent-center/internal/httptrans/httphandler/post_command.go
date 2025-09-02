package httphandler

import (
	"github.com/gin-gonic/gin"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"net/http"
)

type CommandRequest struct {
	AgentID string        `json:"agent_id" bson:"agent_id" binding:"required"`
	Command CommandDetail `json:"command" bson:"command" binding:"required"`
}

type CommandDetail struct {
	AgentCtrl int32       `json:"agent_ctrl,omitempty"`
	Task      *TaskMsg    `json:"task,omitempty"`
	Config    []ConfigMsg `json:"config,omitempty"`
}

type TaskMsg struct {
	DataType int32  `json:"data_type,omitempty"`
	Name     string `json:"name,omitempty"`
	Data     string `json:"data,omitempty"`
	Token    string `json:"token,omitempty"`
}

type ConfigMsg struct {
	Name        string   `json:"name,omitempty"`
	Type        string   `json:"type,omitempty"`
	Version     string   `json:"version,omitempty"`
	SHA256      string   `json:"sha256,omitempty"`
	Signature   string   `json:"signature,omitempty"`
	DownloadURL []string `json:"download_url,omitempty"`
	Detail      string   `json:"detail,omitempty"`
}

func PostCommand(c *gin.Context) {
	var taskModel CommandRequest
	err := c.BindJSON(&taskModel)
	if err != nil {
		zap.S().Errorf("接收到的json是错误的数据结构")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "接收到的数据结构是错误的",
		})
		return
	}
	pb.Command{}
}
