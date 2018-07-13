package axiom

import (
	"time"

	"github.com/antihax/eve-axiom/internal/msgpackcodec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Axiom struct {
	axiomAPI *grpc.Server
}

func NewAxiom() *Axiom {
	// Setup RPC server
	server := grpc.NewServer(grpc.CustomCodec(&msgpackcodec.MsgPackCodec{}),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:    time.Second * 5,
				Timeout: time.Second * 10,
			}),
	)

	return &Axiom{
		axiomAPI: server,
	}
}
