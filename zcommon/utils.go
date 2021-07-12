package zcommon

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type Tools interface {
	TransactionHash(tx *pb.Transaction) string
	CalculateMD5Hash(list []string, timestamp int64) string
}

type toolsImpl struct {}

func NewTools() *toolsImpl {return &toolsImpl{}}

func (t *toolsImpl) TransactionHash(tx *pb.Transaction) string {
	return t.transactionHash(tx)
}

func (t *toolsImpl) CalculateMD5Hash(list []string, timestamp int64) string {
	return hex.EncodeToString(t.calculateMD5Hash(list, timestamp))
}

func (t *toolsImpl) transactionHash(tx *pb.Transaction) string {
	payload, _ := tx.Marshal()
	return CalculatePayloadHash(payload, 0)
}

// calculateMD5Hash calculate hash by MD5
func (t *toolsImpl) calculateMD5Hash(list []string, timestamp int64) []byte {
	h := md5.New()
	for _, hash := range list {
		_, _ = h.Write([]byte(hash))
	}
	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		_, _ = h.Write(b)
	}
	return h.Sum(nil)
}

func CalculatePayloadHash(payload []byte, timestamp int64) string {
	h := md5.New()
	_, _ = h.Write(payload)

	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		_, _ = h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}