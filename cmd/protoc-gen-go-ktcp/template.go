package main

import (
	"bytes"
	"strings"
	"text/template"
)

var ktcpTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

type handlerFunc func(ctx ktcp.Context, srv {{.ServiceType}}KTCPServer) error

type {{.ServiceType}}KTCPServer interface {
{{- range .MethodSets}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

var handleFunctions = map[uint32]handlerFunc{
{{- range .Methods}}
	uint32({{.ProtocolRespID}}):       _{{$svrType}}_{{.Name}}{{.Num}}_KTCP_Handler,
{{- end}}
}

func Router(ctx ktcp.Context, srv {{.ServiceType}}KTCPServer) (err error) {
	if f, exist := handleFunctions[uint32(ctx.GetReqMsg().ID)]; !exist {
		return fmt.Errorf("not found handler func for %v", ctx.GetReqMsg().ID)
	} else {
		return f(ctx, srv)
	}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_KTCP_Handler(ctx ktcp.Context, srv {{$svrType}}KTCPServer) error {
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
