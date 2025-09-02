package buffer

import (
	pb "github.com/lzkking/edr/edrproto"
	"google.golang.org/protobuf/proto"
	"sync"
)

var (
	mu     = &sync.Mutex{}
	buf    = [2048]*pb.EncodedRecord{}
	offset = 0
	hook   func(any) any
)

func SetTransmissionHook(fn func(any) any) {
	hook = fn
}
func WriteEncodedRecord(rec *pb.EncodedRecord) {
	if hook != nil {
		rec = hook(rec).(*pb.EncodedRecord)
	}
	mu.Lock()
	if offset < len(buf) {
		buf[offset] = rec
		offset++
	} else {
		PutEncodedRecord(rec)
	}
	mu.Unlock()
}
func WriteRecord(rec *pb.Record) (err error) {
	erec := GetEncodedRecord(proto.Size(rec.Data))
	erec.DataType = rec.DataType
	erec.Timestamp = rec.Timestamp
	if cap(erec.Data) < proto.Size(rec.Data) {
		erec.Data = make([]byte, proto.Size(rec.Data))
	} else {
		erec.Data = erec.Data[:proto.Size(rec.Data)]
	}
	b, err := proto.Marshal(rec.Data)
	if err != nil {
		return
	}
	erec.Data = b

	mu.Lock()
	if offset < len(buf) {
		buf[offset] = erec
		offset++
	} else {
		// steal it
		buf[0] = erec
	}
	mu.Unlock()
	return
}

func ReadEncodedRecords() (ret []*pb.EncodedRecord) {
	mu.Lock()
	ret = make([]*pb.EncodedRecord, offset)
	copy(ret, buf[:offset])
	offset = 0
	mu.Unlock()
	return
}
