package axiom

import (
	"context"
	"encoding/json"

	"github.com/antihax/goesi/esi"
	"google.golang.org/grpc"
)

// AxiomPlan API interface
type AxiomPlan interface {
	GetKillmailAttributes(ctx context.Context, t *esi.GetKillmailsKillmailIdKillmailHashOk) (*string, error)
}

var serviceDesc = grpc.ServiceDesc{
	ServiceName: "Axiom",
	HandlerType: (*AxiomPlan)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetKillmailAttributes",
			Handler:    GetKillmailAttributesHandler,
		},
	},
}

func GetKillmailAttributesHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(esi.GetKillmailsKillmailIdKillmailHashOk)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(*Axiom).GetKillmailAttributes(ctx, in)
}

// GetKillmailAttributes converts ESI killmail into attributes via dogma parser
func (s Axiom) GetKillmailAttributes(ctx context.Context, km *esi.GetKillmailsKillmailIdKillmailHashOk) (*string, error) {
	attr, err := s.getAttributesFromKillmail(km)
	if err != nil {
		return nil, err
	}

	jsonb, err := json.Marshal(attr)
	json := string(jsonb)

	return &json, err
}
