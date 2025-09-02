package httphandler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lzkking/edr/agent-center/internal/grpctrans/handler"
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

	msgCommand := pb.Command{
		AgentCtrl: taskModel.Command.AgentCtrl,
	}

	if taskModel.Command.Config != nil {
		msgCommand.Config = make([]*pb.ConfigItem, 0, len(taskModel.Command.Config))
		for _, v := range taskModel.Command.Config {
			tmp := &pb.ConfigItem{
				Name:        v.Name,
				Type:        v.Type,
				Version:     v.Version,
				SHA256:      v.SHA256,
				Signature:   v.Signature,
				DownloadURL: v.DownloadURL,
				Detail:      v.Detail,
			}

			msgCommand.Config = append(msgCommand.Config, tmp)
		}
	}

	if taskModel.Command.Task != nil {
		task := pb.PluginTask{
			DataType: taskModel.Command.Task.DataType,
			Name:     taskModel.Command.Task.Name,
			Data:     taskModel.Command.Task.Data,
			Token:    taskModel.Command.Task.Token,
		}
		msgCommand.Task = &task
	}

	err = handler.GlobalPool.PostCommand(taskModel.AgentID, &msgCommand)
	if err != nil {
		zap.S().Errorf("向%v发送命令失败,命令:%v,错误原因:%v", taskModel.AgentID, taskModel, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("向%v发送命令失败,命令:%v,错误原因:%v", taskModel.AgentID, taskModel, err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
	return
}
