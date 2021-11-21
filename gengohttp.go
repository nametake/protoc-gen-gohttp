package main

import (
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	bytesPackage   = protogen.GoImportPath("bytes")
	contextPackage = protogen.GoImportPath("context")
	base64Package  = protogen.GoImportPath("encoding/base64")
	fmtPackage     = protogen.GoImportPath("fmt")
	ioPackage      = protogen.GoImportPath("io")
	ioutilPackage  = protogen.GoImportPath("io/ioutil")
	mimePackage    = protogen.GoImportPath("mime")
	httpPackage    = protogen.GoImportPath("net/http")
	strconvPackage = protogen.GoImportPath("strconv")
	stringsPackage = protogen.GoImportPath("strings")
	reflectPackage = protogen.GoImportPath("reflect")
)

var (
	protoPackage     = protogen.GoImportPath("google.golang.org/protobuf/proto")
	protojsonPackage = protogen.GoImportPath("google.golang.org/protobuf/encoding/protojson")
	grpcPackage      = protogen.GoImportPath("google.golang.org/grpc")
	codesPackage     = protogen.GoImportPath("google.golang.org/grpc/codes")
	statusPackage    = protogen.GoImportPath("google.golang.org/grpc/status")
	anypbPackage     = protogen.GoImportPath("google.golang.org/protobuf/types/known/anypb")
)

func GenerateFile(gen *protogen.Plugin, file *protogen.File) (*protogen.GeneratedFile, error) {
	isGenerated := false
	for _, srv := range file.Services {
		for _, method := range srv.Methods {
			if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
				continue
			}
			isGenerated = true
		}
	}

	if !isGenerated {
		return nil, nil
	}

	filename := file.GeneratedFilenamePrefix + ".http.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	g.P("// Code generated by protoc-gen-gohttp. DO NOT EDIT.")
	g.P("// source: ", file.Desc.Path())
	g.P()
	g.P("package ", file.GoPackageName)

	for _, srv := range file.Services {
		if err := genService(g, srv); err != nil {
			return nil, err
		}
	}

	return g, nil
}

func genService(g *protogen.GeneratedFile, srv *protogen.Service) error {
	genServiceInterface(g, srv)
	genStruct(g, srv)
	genConstructor(g, srv)

	for _, method := range srv.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}

		genMethod(g, method)
		genMethodWithName(g, method)
		if err := genMethodHTTPRule(g, method); err != nil {
			return err
		}
	}

	return nil
}

func callbackSignature(g *protogen.GeneratedFile) string {
	return "func(ctx " +
		g.QualifiedGoIdent(contextPackage.Ident("Context")) +
		", w " +
		g.QualifiedGoIdent(httpPackage.Ident("ResponseWriter")) +
		", r *" +
		g.QualifiedGoIdent(httpPackage.Ident("Request")) +
		", arg, ret " +
		g.QualifiedGoIdent(protoPackage.Ident("Message")) +
		", err error)"
}

func methodSignature(g *protogen.GeneratedFile, method *protogen.Method, prefix string) string {
	return "func (h *" + method.Parent.GoName + "HTTPConverter) " +
		method.GoName + prefix + "(cb " + callbackSignature(g) +
		", interceptors ..." + g.QualifiedGoIdent(grpcPackage.Ident("UnaryServerInterceptor")) + ") "
}

func genDefaultCallback(g *protogen.GeneratedFile) {
	g.P("if cb == nil {")
	g.P("	cb = ", callbackSignature(g), " {")
	g.P("		if err != nil {")
	g.P("			w.WriteHeader(", httpPackage.Ident("StatusInternalServerError"), ")")
	g.P("			p := ", statusPackage.Ident("New"), "(", codesPackage.Ident("Unknown"), ", err.Error()).Proto()")
	g.P("			switch contentType, _, _ := ", mimePackage.Ident("ParseMediaType"), "(r.Header.Get(\"Content-Type\")); contentType {")
	g.P("				case \"application/protobuf\", \"application/x-protobuf\":")
	g.P("					buf, err := ", protoPackage.Ident("Marshal"), "(p)")
	g.P("					if err != nil {")
	g.P("						return")
	g.P("					}")
	g.P("					if _, err := ", ioPackage.Ident("Copy"), "(w, ", bytesPackage.Ident("NewBuffer"), "(buf)); err != nil {")
	g.P("						return")
	g.P("					}")
	g.P("				case \"application/json\":")
	g.P("					buf, err := ", protojsonPackage.Ident("Marshal"), "(p)")
	g.P("					if err != nil {")
	g.P("						return")
	g.P("					}")
	g.P("					if _, err := ", ioPackage.Ident("Copy"), "(w, ", bytesPackage.Ident("NewBuffer"), "(buf)); err != nil {")
	g.P("						return")
	g.P("					}")
	g.P("				default:")
	g.P("			}")
	g.P("		}")
	g.P("	}")
	g.P("}")
}

func genServiceInterface(g *protogen.GeneratedFile, srv *protogen.Service) {
	g.P("// ", srv.GoName, "HTTPService is the server API for ", srv.GoName, " service.")
	g.P("type ", srv.GoName, "HTTPService interface {")

	for _, method := range srv.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		g.P(method.Comments.Leading, method.GoName, "(", contextPackage.Ident("Context"), ", *", genMessageName(method.Input), ") (*", genMessageName(method.Output), ", error)")
	}
	g.P("}")
}

func genStruct(g *protogen.GeneratedFile, srv *protogen.Service) {
	g.P("// ", srv.GoName, "HTTPConverter has a function to convert ", srv.GoName, "HTTPService interface to http.HandlerFunc.")
	g.P("type ", srv.GoName, "HTTPConverter struct {")
	g.P("srv ", srv.GoName, "HTTPService")
	g.P("}")
}

func genConstructor(g *protogen.GeneratedFile, srv *protogen.Service) {
	g.P("// New", srv.GoName, "HTTPConverter returns ", srv.GoName, "HTTPConverter.")
	g.P("func New", srv.GoName, "HTTPConverter(srv ", srv.GoName, "HTTPService) *", srv.GoName, "HTTPConverter {")
	g.P("	return &", srv.GoName, "HTTPConverter{")
	g.P("		srv: srv,")
	g.P("	}")
	g.P("}")
}

func genMethod(g *protogen.GeneratedFile, method *protogen.Method) {
	g.P("// ", method.GoName, " returns ", method.Parent.GoName, "HTTPService interface's ", method.GoName, " converted to http.HandlerFunc.")
	if method.Comments.Leading.String() != "" {
		g.P("//")
	}
	g.P(method.Comments.Leading, methodSignature(g, method, ""), httpPackage.Ident("HandlerFunc"), " {")
	genDefaultCallback(g)
	g.P("	return ", httpPackage.Ident("HandlerFunc"), "(func(w ", httpPackage.Ident("ResponseWriter"), ", r *", httpPackage.Ident("Request"), ") {")
	g.P("		ctx := r.Context()")
	g.P("")
	g.P("		contentType, _, _ := ", mimePackage.Ident("ParseMediaType"), "(r.Header.Get(\"Content-Type\"))")
	g.P("")
	g.P("		accepts := ", stringsPackage.Ident("Split"), "(r.Header.Get(\"Accept\"), \",\")")
	g.P("		accept := accepts[0]")
	g.P("		if accept == \"*/*\" || accept == \"\" {")
	g.P("			if contentType != \"\" {")
	g.P("				accept = contentType")
	g.P("			} else {")
	g.P("				accept = \"application/json\"")
	g.P("			}")
	g.P("		}")
	g.P("")
	g.P("		w.Header().Set(\"Content-Type\", accept)")
	g.P("")
	g.P("		arg := &", genMessageName(method.Input), "{}")
	g.P("		if r.Method != ", httpPackage.Ident("MethodGet"), " {")
	g.P("			body, err := ", ioutilPackage.Ident("ReadAll"), "(r.Body)")
	g.P("			if err != nil {")
	g.P("				cb(ctx, w, r, nil, nil, err)")
	g.P("				return")
	g.P("			}")
	g.P("")
	g.P("			switch contentType {")
	g.P("			case \"application/protobuf\", \"application/x-protobuf\":")
	g.P("				if err := ", protoPackage.Ident("Unmarshal"), "(body, arg); err != nil {")
	g.P("					cb(ctx, w, r, nil, nil, err)")
	g.P("					return")
	g.P("				}")
	g.P("			case \"application/json\":")
	g.P("				if err := ", protojsonPackage.Ident("Unmarshal"), "(body, arg); err != nil {")
	g.P("					cb(ctx, w, r, nil, nil, err)")
	g.P("					return")
	g.P("				}")
	g.P("			default:")
	g.P("				w.WriteHeader(", httpPackage.Ident("StatusUnsupportedMediaType"), ")")
	g.P("				_, err := ", fmtPackage.Ident("Fprintf"), "(w, \"Unsupported Content-Type: %s\", contentType)")
	g.P("				cb(ctx, w, r, nil, nil, err)")
	g.P("				return")
	g.P("			}")
	g.P("		}")
	g.P("")
	g.P("		n := len(interceptors)")
	g.P("		chained := func(ctx ", contextPackage.Ident("Context"), ", arg interface{}, info *", grpcPackage.Ident("UnaryServerInfo"), ", handler ", grpcPackage.Ident("UnaryHandler"), ") (interface{}, error) {")
	g.P("			chainer := func(currentInter ", grpcPackage.Ident("UnaryServerInterceptor"), ", currentHandler ", grpcPackage.Ident("UnaryHandler"), ") ", grpcPackage.Ident("UnaryHandler"), " {")
	g.P("				return func(currentCtx ", contextPackage.Ident("Context"), ", currentReq interface{}) (interface{}, error) {")
	g.P("					return currentInter(currentCtx, currentReq, info, currentHandler)")
	g.P("				}")
	g.P("			}")
	g.P("")
	g.P("			chainedHandler := handler")
	g.P("			for i := n - 1; i >= 0; i-- {")
	g.P("				chainedHandler = chainer(interceptors[i], chainedHandler)")
	g.P("			}")
	g.P("			return chainedHandler(ctx, arg)")
	g.P("		}")
	g.P("")
	g.P("		info := &", grpcPackage.Ident("UnaryServerInfo"), "{")
	g.P("			Server:     h.srv,")
	g.P("			FullMethod: \"/", method.Desc.ParentFile().Package(), ".", method.Parent.GoName, "/", method.GoName, "\",")
	g.P("		}")
	g.P("")
	g.P("		handler := func(c ", contextPackage.Ident("Context"), ", req interface{}) (interface{}, error) {")
	g.P("			return h.srv.", method.GoName, "(c, req.(*", genMessageName(method.Input), "))")
	g.P("		}")
	g.P("")
	g.P("		iret, err := chained(ctx, arg, info, handler)")
	g.P("		if err != nil {")
	g.P("			cb(ctx, w, r, arg, nil, err)")
	g.P("			return")
	g.P("		}")
	g.P("")
	g.P("		ret, ok := iret.(*", genMessageName(method.Output), ")")
	g.P("		if !ok {")
	g.P("			cb(ctx, w, r, arg, nil, fmt.Errorf(\"/", method.Desc.ParentFile().Package(), ".", method.Parent.GoName, "/", method.GoName, ": interceptors have not return ", genMessageName(method.Output), "\"))")
	g.P("			return")
	g.P("		}")
	g.P("")
	g.P("		switch accept {")
	g.P("		case \"application/protobuf\", \"application/x-protobuf\":")
	g.P("			buf, err := ", protoPackage.Ident("Marshal"), "(ret)")
	g.P("			if err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("			if _, err := ", ioPackage.Ident("Copy"), "(w, ", bytesPackage.Ident("NewBuffer"), "(buf)); err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("		case \"application/json\":")
	g.P("			buf, err := ", protojsonPackage.Ident("Marshal"), "(ret)")
	g.P("			if err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("			if _, err := ", ioPackage.Ident("Copy"), "(w, ", bytesPackage.Ident("NewBuffer"), "(buf)); err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("		default:")
	g.P("			w.WriteHeader(", httpPackage.Ident("StatusUnsupportedMediaType"), ")")
	g.P("			_, err := fmt.Fprintf(w, \"Unsupported Accept: %s\", accept)")
	g.P("			cb(ctx, w, r, arg, ret, err)")
	g.P("			return")
	g.P("		}")
	g.P("		cb(ctx, w, r, arg, ret, nil)")
	g.P("	})")
	g.P("}")
}

func genMethodWithName(g *protogen.GeneratedFile, method *protogen.Method) {
	g.P("// ", method.GoName, "WithName returns Service name, Method name and ", method.Parent.GoName, "HTTPService interface's ", method.GoName, " converted to http.HandlerFunc.")
	if method.Comments.Leading.String() != "" {
		g.P("//")
	}
	g.P(method.Comments.Leading, methodSignature(g, method, "WithName"), " (string, string, ", httpPackage.Ident("HandlerFunc"), ") {")
	g.P("	return \"", method.Parent.GoName, "\", \"", method.GoName, "\", h.", method.GoName, "(cb, interceptors...)")
	g.P("}")
}

func genMethodHTTPRule(g *protogen.GeneratedFile, method *protogen.Method) error {
	options, ok := method.Desc.Options().(*descriptorpb.MethodOptions)
	if !ok {
		return nil
	}

	httpRule, ok := proto.GetExtension(options, annotations.E_Http).(*annotations.HttpRule)
	if !ok {
		return nil
	}

	var (
		httpMethod string
		pattern    string
	)

	switch httpRule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		httpMethod = "http.MethodGet"
		pattern = httpRule.GetGet()
	case *annotations.HttpRule_Put:
		httpMethod = "http.MethodPut"
		pattern = httpRule.GetPut()
	case *annotations.HttpRule_Post:
		httpMethod = "http.MethodPost"
		pattern = httpRule.GetPost()
	case *annotations.HttpRule_Delete:
		httpMethod = "http.MethodDelete"
		pattern = httpRule.GetDelete()
	case *annotations.HttpRule_Patch:
		httpMethod = "http.MethodPatch"
		pattern = httpRule.GetPatch()
	default:
		return nil
	}

	pathParams, err := parsePathParam(pattern)
	if err != nil {
		return err
	}

	queryParams := createQueryParams(method)

	g.P("// ", method.GoName, "HTTPRule returns HTTP method, path and ", method.Parent.GoName, "HTTPService interface's ", method.GoName, " converted to http.HandlerFunc.")
	if method.Comments.Leading.String() != "" {
		g.P("//")
	}
	g.P(method.Comments.Leading, methodSignature(g, method, "HTTPRule"), " (string, string, ", httpPackage.Ident("HandlerFunc"), ") {")
	genDefaultCallback(g)
	g.P("	return ", httpMethod, ", \"", pattern, "\", ", httpPackage.Ident("HandlerFunc"), "(func(w ", httpPackage.Ident("ResponseWriter"), ", r *", httpPackage.Ident("Request"), ") {")
	g.P("		ctx := r.Context()")
	g.P("")
	g.P("		contentType, _, _ := ", mimePackage.Ident("ParseMediaType"), "(r.Header.Get(\"Content-Type\"))")
	g.P("")
	g.P("		accepts := ", stringsPackage.Ident("Split"), "(r.Header.Get(\"Accept\"), \",\")")
	g.P("		accept := accepts[0]")
	g.P("		if accept == \"*/*\" || accept == \"\" {")
	g.P("			if contentType != \"\" {")
	g.P("				accept = contentType")
	g.P("			} else {")
	g.P("				accept = \"application/json\"")
	g.P("			}")
	g.P("		}")
	g.P("")
	g.P("		w.Header().Set(\"Content-Type\", accept)")
	g.P("")
	g.P("		arg := &", genMessageName(method.Input), "{}")
	if _, ok := httpRule.GetPattern().(*annotations.HttpRule_Get); ok {
		g.P("if r.Method == http.MethodGet {")
		for _, p := range queryParams {
			for _, pattern := range pathParams {
				if p.GoName == pattern.GoName {
					goto Pass
				}
			}
			genQueryString(g, p)
		Pass:
		}
		g.P("}")
	} else {
		g.P("		if r.Method != ", httpPackage.Ident("MethodGet"), " {")
		g.P("			body, err := ", ioutilPackage.Ident("ReadAll"), "(r.Body)")
		g.P("			if err != nil {")
		g.P("				cb(ctx, w, r, nil, nil, err)")
		g.P("				return")
		g.P("			}")
		g.P("")
		g.P("			switch contentType {")
		g.P("			case \"application/protobuf\", \"application/x-protobuf\":")
		g.P("				if err := ", protoPackage.Ident("Unmarshal"), "(body, arg); err != nil {")
		g.P("					cb(ctx, w, r, nil, nil, err)")
		g.P("					return")
		g.P("				}")
		g.P("			case \"application/json\":")
		g.P("				if err := ", protojsonPackage.Ident("Unmarshal"), "(body, arg); err != nil {")
		g.P("					cb(ctx, w, r, nil, nil, err)")
		g.P("					return")
		g.P("				}")
		g.P("			default:")
		g.P("				w.WriteHeader(", httpPackage.Ident("StatusUnsupportedMediaType"), ")")
		g.P("				_, err := ", fmtPackage.Ident("Fprintf"), "(w, \"Unsupported Content-Type: %s\", contentType)")
		g.P("				cb(ctx, w, r, nil, nil, err)")
		g.P("				return")
		g.P("			}")
		g.P("		}")
	}
	g.P("")

	if len(pathParams) != 0 {
		g.P("p := strings.Split(r.URL.Path, \"/\")")
	}

	for _, t := range pathParams {
		for _, p := range t.GetSplitedGoNames() {
			g.P(reflectPackage.Ident("ValueOf"), "(&arg.", p, ").Elem().Set(", reflectPackage.Ident("ValueOf"), "(", reflectPackage.Ident("New"), "(", reflectPackage.Ident("TypeOf"), "(arg.", p, ").Elem()).Interface()))")
		}

		g.P("arg.", t.GoName, " = p[", t.Index, "]")
	}

	g.P("")
	g.P("		n := len(interceptors)")
	g.P("		chained := func(ctx ", contextPackage.Ident("Context"), ", arg interface{}, info *", grpcPackage.Ident("UnaryServerInfo"), ", handler ", grpcPackage.Ident("UnaryHandler"), ") (interface{}, error) {")
	g.P("			chainer := func(currentInter ", grpcPackage.Ident("UnaryServerInterceptor"), ", currentHandler ", grpcPackage.Ident("UnaryHandler"), ") ", grpcPackage.Ident("UnaryHandler"), " {")
	g.P("				return func(currentCtx ", contextPackage.Ident("Context"), ", currentReq interface{}) (interface{}, error) {")
	g.P("					return currentInter(currentCtx, currentReq, info, currentHandler)")
	g.P("				}")
	g.P("			}")
	g.P("")
	g.P("			chainedHandler := handler")
	g.P("			for i := n - 1; i >= 0; i-- {")
	g.P("				chainedHandler = chainer(interceptors[i], chainedHandler)")
	g.P("			}")
	g.P("			return chainedHandler(ctx, arg)")
	g.P("		}")
	g.P("")
	g.P("		info := &", grpcPackage.Ident("UnaryServerInfo"), "{")
	g.P("			Server:     h.srv,")
	g.P("			FullMethod: \"/", method.Desc.ParentFile().Package(), ".", method.Parent.GoName, "/", method.GoName, "\",")
	g.P("		}")
	g.P("")
	g.P("		handler := func(c ", contextPackage.Ident("Context"), ", req interface{}) (interface{}, error) {")
	g.P("			return h.srv.", method.GoName, "(c, req.(*", genMessageName(method.Input), "))")
	g.P("		}")
	g.P("")
	g.P("		iret, err := chained(ctx, arg, info, handler)")
	g.P("		if err != nil {")
	g.P("			cb(ctx, w, r, arg, nil, err)")
	g.P("			return")
	g.P("		}")
	g.P("")
	g.P("		ret, ok := iret.(*", genMessageName(method.Output), ")")
	g.P("		if !ok {")
	g.P("			cb(ctx, w, r, arg, nil, fmt.Errorf(\"/", method.Desc.ParentFile().Package(), ".", method.Parent.GoName, "/", method.GoName, ": interceptors have not return ", genMessageName(method.Output), "\"))")
	g.P("			return")
	g.P("		}")
	g.P("")
	g.P("		switch accept {")
	g.P("		case \"application/protobuf\", \"application/x-protobuf\":")
	g.P("			buf, err := ", protoPackage.Ident("Marshal"), "(ret)")
	g.P("			if err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("			if _, err := ", ioPackage.Ident("Copy"), "(w, ", bytesPackage.Ident("NewBuffer"), "(buf)); err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("		case \"application/json\":")
	g.P("			buf, err := ", protojsonPackage.Ident("Marshal"), "(ret)")
	g.P("			if err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("			if _, err := ", ioPackage.Ident("Copy"), "(w, ", bytesPackage.Ident("NewBuffer"), "(buf)); err != nil {")
	g.P("				cb(ctx, w, r, arg, ret, err)")
	g.P("				return")
	g.P("			}")
	g.P("		default:")
	g.P("			w.WriteHeader(", httpPackage.Ident("StatusUnsupportedMediaType"), ")")
	g.P("			_, err := fmt.Fprintf(w, \"Unsupported Accept: %s\", accept)")
	g.P("			cb(ctx, w, r, arg, ret, err)")
	g.P("			return")
	g.P("		}")
	g.P("		cb(ctx, w, r, arg, ret, nil)")
	g.P("	})")
	g.P("}")

	return nil
}

func genQueryString(g *protogen.GeneratedFile, queryParam *queryParam) {
	switch queryParam.Desc.Kind() {
	case protoreflect.BoolKind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]bool, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", strconvPackage.Ident("ParseBool"), "(v)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, c)")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", strconvPackage.Ident("ParseBool"), "(v)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = c")
			g.P("}")
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]int32, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", strconvPackage.Ident("ParseInt"), "(v, 10, 32)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, int32(c))")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", strconvPackage.Ident("ParseInt"), "(v, 10, 32)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = int32(c)")
			g.P("}")
		}
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]uint32, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", strconvPackage.Ident("ParseUint"), "(v, 10, 32)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, uint32(c))")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", strconvPackage.Ident("ParseUint"), "(v, 10, 32)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = uint32(c)")
			g.P("}")
		}
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]int64, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", strconvPackage.Ident("ParseInt"), "(v, 10, 64)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, c)")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", strconvPackage.Ident("ParseInt"), "(v, 10, 64)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = c")
			g.P("}")
		}
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]uint64, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", strconvPackage.Ident("ParseUint"), "(v, 10, 64)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, c)")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", strconvPackage.Ident("ParseUint"), "(v, 10, 64)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = c")
			g.P("}")
		}
	case protoreflect.FloatKind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]float32, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", strconvPackage.Ident("ParseFloat"), "(v, 32)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, float32(c))")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", strconvPackage.Ident("ParseFloat"), "(v, 32)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = float32(c)")
			g.P("}")
		}
	case protoreflect.DoubleKind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]float64, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", strconvPackage.Ident("ParseFloat"), "(v, 64)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, c)")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", strconvPackage.Ident("ParseFloat"), "(v, 64)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = c")
			g.P("}")
		}
	case protoreflect.StringKind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([]string, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		arr = append(arr, v)")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	arg.", queryParam.GoName, " = v")
			g.P("}")
		}
	case protoreflect.BytesKind:
		if queryParam.Desc.IsList() {
			g.P("if repeated := r.URL.Query()[\"", queryParam.Name, "\"]; len(repeated) != 0 {")
			g.P("	arr := make([][]byte, 0, len(repeated))")
			g.P("	for _, v := range repeated {")
			g.P("		c, err := ", base64Package.Ident("StdEncoding.DecodeString"), "(v)")
			g.P("		if err != nil {")
			g.P("			cb(ctx, w, r, nil, nil, err)")
			g.P("			return")
			g.P("		}")
			g.P("		arr = append(arr, c)")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = arr")
			g.P("}")
		} else {
			g.P("if v := r.URL.Query().Get(\"", queryParam.Name, "\"); v != \"\" {")
			g.P("	c, err := ", base64Package.Ident("StdEncoding.DecodeString"), "(v)")
			g.P("	if err != nil {")
			g.P("		cb(ctx, w, r, nil, nil, err)")
			g.P("		return")
			g.P("	}")
			g.P("	arg.", queryParam.GoName, " = c")
			g.P("}")
		}
	}
}

func genMessageName(msg *protogen.Message) protogen.GoIdent {
	switch msg.Location.SourceFile {
	case "google/protobuf/any.proto":
		return anypbPackage.Ident(msg.GoIdent.GoName)
	default:
		return msg.GoIdent
	}
}
