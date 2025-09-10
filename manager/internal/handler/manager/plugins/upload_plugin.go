package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp/v3"
	"github.com/lzkking/edr/manager/pkg/mongodb"
	"github.com/lzkking/edr/manager/pkg/s3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"net/http"
)

func UploadPlugin(c *gin.Context) {
	name := c.PostForm("name")
	version := c.PostForm("version")
	if name == "" || version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name/version 不能为空"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//	将文件保存到S3上
	publicURL, key, sha256hex, signature, err := s3.S3Client.UploadStream(c, f, &file.Size, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plgConfig := PlgConfig{
		Name:      name,
		Type:      "",
		Version:   version,
		SHA256:    sha256hex,
		Signature: signature,
		DownloadUrls: []string{
			publicURL,
		},
		Key:    key,
		Detail: "",
	}

	filter := bson.M{
		"name": plgConfig.Name,
	}
	update := bson.M{
		"$set": plgConfig,
	}
	opts := options.Update().SetUpsert(true)
	plgsCollection := mongodb.MongodbClient.Database(mongodb.DatabaseName).Collection(PlgConfigCollection)
	_, err = plgsCollection.UpdateOne(c, filter, update, opts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "将插件数据保存到数据库失败"})
		zap.S().Warnf("将插件数据保存到数据库失败,失败原因:%v", err)
		return
	}
	zap.S().Debugf("%v", pp.Sprintf("%v", plgConfig))
	c.JSON(http.StatusOK, gin.H{
		"msg": "ok",
	})
}
