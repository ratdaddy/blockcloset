package loggrpc

type Options struct {
	Schema *Schema
}

type Schema struct {
	FullMethod      string
	Service         string
	Method          string
	System          string
	Code            string
	ProtocolName    string
	ProtocolVersion string
	Duration        string
	RemoteIP        string
	Scheme          string
	Host            string
	UserAgent       string
	RequestBytes    string
	ResponseBytes   string
	RequestID       string
}

var (
	SchemaOTEL = &Schema{
		FullMethod:      "rpc.full_method",
		Service:         "rpc.service",
		Method:          "rpc.method",
		System:          "rpc.system",
		Code:            "rpc.grpc.status_code",
		ProtocolName:    "network.protocol.name",
		ProtocolVersion: "network.protocol.version",
		Duration:        "server.duration",
		RemoteIP:        "network.peer.address",
		Scheme:          "url.scheme",
		Host:            "server.address",
		UserAgent:       "user_agent.original",
		RequestBytes:    "rpc.request.size",
		ResponseBytes:   "rpc.response.size",
		RequestID:       "rpc.request.header.x-request-id",
	}
)

func (s *Schema) Concise(concise bool) *Schema {
	if !concise {
		return s
	}

	return &Schema{}
}
