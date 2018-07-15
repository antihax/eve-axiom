package axiom

import (
	"context"
	"time"

	"github.com/antihax/goesi/esi"

	"github.com/antihax/eve-axiom/internal/msgpackcodec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type AxiomAPI struct {
	grpc *grpc.ClientConn
}

// NewAxiomAPIClient creates a new Axiom client
func NewAxiomAPIClient(server string) (*AxiomAPI, error) {
	r, err := grpc.Dial(server,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 5,
			Timeout:             time.Second * 5,
			PermitWithoutStream: true,
		}),
		grpc.WithInsecure(),
		grpc.WithCodec(&msgpackcodec.MsgPackCodec{}),
	)

	if err != nil {
		return nil, err

	}
	return &AxiomAPI{r}, nil
}

// GetKillmailAttributes takes an ESI killmail and returns attributes in json
func (s *AxiomAPI) GetKillmailAttributes(km *esi.GetKillmailsKillmailIdKillmailHashOk) (*string, error) {
	json := ""
	if err := s.grpc.Invoke(context.Background(), "/Axiom/GetKillmailAttributes",
		km, &json); err != nil {
		return nil, err
	}

	return &json, nil
}
