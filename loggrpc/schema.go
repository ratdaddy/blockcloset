package loggrpc

type Options struct {
	Schema *Schema
}

type Schema struct {
	FullMethod    string
	Service       string
	Method        string
	System        string
	Code          string
	Protocol      string
	Duration      string
	RemoteIP      string
	Scheme        string
	Host          string
	UserAgent     string
	RequestBytes  string
	ResponseBytes string
}

var (
	SchemaOTEL = &Schema{
		FullMethod:    "rpc.full_method",
		Service:       "rpc.service",
		Method:        "rpc.method",
		System:        "rpc.system",
		Code:          "grpc.code",
		Protocol:      "network.protocol.version",
		Duration:      "rpc.server.duration",
		RemoteIP:      "client.address",
		Scheme:        "url.scheme",
		Host:          "server.address",
		UserAgent:     "user_agent.original",
		RequestBytes:  "rpc.request.size",
		ResponseBytes: "rpc.response.size",
	}
)

func (s *Schema) Concise(concise bool) *Schema {
	if !concise {
		return s
	}

	return &Schema{}
}
