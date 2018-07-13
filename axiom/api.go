package axiom

import "google.golang.org/grpc"

type AxiomPlan interface {
}

var serviceDesc = grpc.ServiceDesc{
	ServiceName: "Axiom",
	HandlerType: (*AxiomPlan)(nil),
	Methods:     []grpc.MethodDesc{
		/*{
			MethodName: "GetKillmailAttributes",
			Handler:    GetKillmailAttributes,
		},
		{
			MethodName: "GetFittingAttributes",
			Handler:    GetKillmailAttributes,
		},*/
	},
	Streams: []grpc.StreamDesc{},
}
