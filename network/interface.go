package network

import (
	pb "github.com/Grivn/libfalanx/zcommon/protos"
)

type Network interface {
	Broadcast(msg *pb.ConsensusMessage)
}
