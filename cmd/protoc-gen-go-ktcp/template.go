package main

import (
	"bytes"
	"strings"
	"text/template"
)

var ktcpTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}
type {{.ServiceType}}KTCPServer interface {
{{- range .MethodSets}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}KTCPServer(s *ktcp.Server, srv {{.ServiceType}}KTCPServer) {
	{{- range .Methods}}
	s.AddRoute(uint32({{.ProtocolReqID}}), _{{$svrType}}_{{.Name}}{{.Num}}_KTCP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_KTCP_Handler(srv {{$svrType}}KTCPServer) func(ctx ktcp.Context) error {
	return func(ctx ktcp.Context) error {
		var in {{.Request}}
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		})
		out, err := h(ctx, &in)
		if err != nil {
			se := errors.FromError(err)
			return ctx.Send(uint32({{.ProtocolRespID}}), packing.ErrType, se)
		}
		reply := out.(*{{.Reply}})
		return ctx.Send(uint32({{.ProtocolRespID}}), packing.OKType, reply)
	}
}
{{end}}

`

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
}

type methodDesc struct {
	Name           string
	Num            int
	Request        string
	Reply          string
	ProtocolReqID  string
	ProtocolRespID string
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("ktcp").Parse(strings.TrimSpace(ktcpTemplate))
	if err != nil {
		panic(err)
	}
	if err = tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
