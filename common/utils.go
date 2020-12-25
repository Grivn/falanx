package common

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"

	"github.com/ultramesh/flato-common/types"
	"github.com/ultramesh/flato-common/types/protos"
)

type Tools interface {
	TransactionHash(tx *protos.Transaction) string
	CalculateMD5Hash(list []string, timestamp int64) string
}

type toolsImpl struct {}

func NewTools() *toolsImpl {return &toolsImpl{}}

func (t *toolsImpl) TransactionHash(tx *protos.Transaction) string {
	return t.transactionHash(tx)
}

func (t *toolsImpl) CalculateMD5Hash(list []string, timestamp int64) string {
	return t.calculateMD5Hash(list, timestamp)
}

func (t *toolsImpl) transactionHash(tx *protos.Transaction) string {
	return types.GetHash(tx).Hex()
}

// calculateMD5Hash calculate hash by MD5
func (t *toolsImpl) calculateMD5Hash(list []string, timestamp int64) string {
	h := md5.New()
	for _, hash := range list {
		_, _ = h.Write([]byte(hash))
	}
	if timestamp > 0 {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(timestamp))
		_, _ = h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}
