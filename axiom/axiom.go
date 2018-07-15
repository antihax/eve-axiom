package axiom

import (
	"net"
	"time"

	"github.com/antihax/eve-axiom/internal/msgpackcodec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Axiom provides API for parsing ship fits through dogma into attribute data
type Axiom struct {
	axiomAPI *grpc.Server
}

// NewAxiom creates a new Axiom microservice
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

// RunServer starts the server listening on port 3000
func (s *Axiom) RunServer() error {
	lis, err := net.Listen("tcp", ":3003")
	if err != nil {
		return err
	}

	s.axiomAPI.RegisterService(&serviceDesc, s)
	err = s.axiomAPI.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}
