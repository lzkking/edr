package collect

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CollectData struct {
	AgentID   string            `json:"agent_id"`
	DataType  int32             `json:"data_type"`
	AgentTime int64             `json:"agent_time"`
	Fields    map[string]string `json:"fields"`
}

func DealCollectData(c *gin.Context) {
	var collectData CollectData
	var err error
	err = c.ShouldBindJSON(&collectData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "解析传递来的数据失败",
		})
		return
	}

	switch collectData.DataType {
	case 7310:
		err = DealUserInfoData(c, &collectData)
	case 7311:
		err = DealPortData(c, &collectData)
	case 7312:
		err = DealProcessData(c, &collectData)
	case 7313:
		err = DealSoftwareData(c, &collectData)
	default:
		err = errors.New("未知data type")
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "ok"})

}
