package handler

import (
	"github.com/lzkking/edr/agent-center/internal/grpctrans/pool"
	"github.com/lzkking/edr/agent-center/pkg/kafka"
	pb "github.com/lzkking/edr/edrproto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
)

func handleRawData(event *pb.PackagedData, conn *pool.Connection) {
	zap.S().Debugf("接收到agent传递来的数据")
	agentId := event.AgentId
	intranetIpv4 := event.IntranetIpv4
	extranetIpv4 := event.ExtranetIpv4
	intranetIpv6 := event.IntranetIpv6
	extranetIpv6 := event.ExtranetIpv6
	hostname := event.Hostname
	version := event.Version
	product := event.Product
	svrTime := time.Now().Unix()

	recs := event.Records
	for i, _ := range recs {
		mqMsg := kafka.MQMsgPool.Get().(*pb.MQData)
		mqMsg.DataType = recs[i].DataType
		mqMsg.AgentTime = recs[i].Timestamp
		mqMsg.Body = recs[i].Data
		mqMsg.AgentID = agentId
		mqMsg.IntranetIPv4 = strings.Join(intranetIpv4, ",")
		mqMsg.ExtranetIPv4 = strings.Join(extranetIpv4, ",")
		mqMsg.IntranetIPv6 = strings.Join(intranetIpv6, ",")
		mqMsg.ExtranetIPv6 = strings.Join(extranetIpv6, ",")
		mqMsg.Hostname = hostname
		mqMsg.Version = version
		mqMsg.Product = product
		mqMsg.SvrTime = svrTime
		switch recs[i].DataType {
		case 2025:
			//agent心跳包,分析心跳包,并且包心跳包保存到detail中,并根据是否是首次传递的agent判断是否需要下发插件数据
			zap.S().Infof("接收到agent心跳包")
			parseAgentHeartBeat(event, recs[i], conn)
		case 2026:
			zap.S().Infof("接收agent插件的心跳包")
		case 2027:
			zap.S().Infof("接收到收集任务")
		}
		kafka.KafkaProducer.SendPBWithKey(agentId, mqMsg)

	}
}

func parseAgentHeartBeat(req *pb.PackagedData, rec *pb.EncodedRecord, conn *pool.Connection) map[string]interface{} {
	zap.S().Infof("分析agent传递来的心跳包")
	var detail map[string]interface{}
	detail = make(map[string]interface{})

	//	编码pb.PackagedData的数据
	detail["intranet_ipv4"] = strings.Join(req.IntranetIpv4, ",")
	detail["extranet_ipv4"] = strings.Join(req.ExtranetIpv4, ",")
	detail["intranet_ipv6"] = strings.Join(req.IntranetIpv6, ",")
	detail["extranet_ipv6"] = strings.Join(req.ExtranetIpv6, ",")
	detail["hostname"] = req.Hostname
	detail["version"] = req.Version
	detail["product"] = req.Product

	// 编码conn的数据
	detail["agent_id"] = conn.AgentId
	detail["source_addr"] = conn.SourceAddr
	detail["create_at"] = conn.CreateAt
	detail["last_heartbeat_time"] = time.Now().Unix()

	// 编码rec中的数据
	var payload pb.Payload
	err := proto.Unmarshal(rec.Data, &payload)
	if err == nil {
		for k, v := range payload.Fields {
			zap.S().Debugf("key: %v, value: %v", k, v)
			detail[k] = v
		}
	} else {
		zap.S().Warnf("agent 编码的heartbeat的数据有问题")
	}

	// 确定是否有心跳包的记录,没有的话,需要传递到manager

	return detail
}
