package main

import (
	"encoding/json"
	"os"
	"go/token"
	"go/ast"
	"go/parser"
	"text/template"
	"log"
	"fmt"
	"strings"
	"io"
	"reflect"
	"bytes"
)

type handlerTplArg struct {
	ApiTypeName string
	ApiMethodName string
	CheckAuthBlock string
	CheckHttpMethodBlock string
	ParamsTypeName string
}

type deserializeModelTplArg struct {
	ModelTypeName string
	DeserializeFieldsCode string
}

type handlerConfig struct {
	ApiMethodName string
	UrlPath string
}

var (
	genPackages = []string{
		"fmt",
		"net/http",
		"strconv",
		"encoding/json",
	}

	helperFuncsSrc = `

type finalResponse struct {
	Error string `+"`json:\"error\"`"+`
	Response interface{} `+"`json:\"response,omitempty\"`"+`
}

func checkAuth(w http.ResponseWriter, r *http.Request) error {
	if (r.Header.Get("X-Auth") != "100500") {
		return ApiError{http.StatusForbidden, fmt.Errorf("unauthorized")}
	}
	return nil
}

func handleError(err error, w http.ResponseWriter) {
	apiError, isApiError := err.(ApiError)
	var errorText string
	var httpStatus int
	if isApiError {
		errorText = apiError.Err.Error()
		httpStatus = apiError.HTTPStatus
	} else {
		errorText = err.Error()
		httpStatus = http.StatusInternalServerError
	}
	resp := finalResponse{
		Error: errorText,
		Response: nil,
	}
	respText, _ := json.Marshal(resp)
	w.WriteHeader(httpStatus)
	w.Write([]byte(respText))
}
`

	handlerTpl = template.Must(template.New("handler").Parse(`
func (api *{{.ApiTypeName}}) handler{{.ApiMethodName}}(w http.ResponseWriter, r *http.Request) {
	var err error
	{{.CheckHttpMethodBlock}}
	{{.CheckAuthBlock}}
	deserializedParams, err := deserialize{{.ParamsTypeName}}(r)
	if err != nil {
		handleError(err, w)
		return
	}
	result, err := api.{{.ApiMethodName}}(r.Context(), deserializedParams)
	if err != nil {
		handleError(err, w)
		return
	}
	resp := finalResponse{
		Error: "",
		Response: result,
	}
	respText, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respText))
}
`))

	checkAuthenticationStr = `
	// checking authentication
	err = checkAuth(w, r)
	if err != nil {
		handleError(err, w)
		return
	}
`

	deserializeModelTpl = template.Must(template.New("deserializeModel").Parse(`
func deserialize{{.ModelTypeName}}(r *http.Request) ({{.ModelTypeName}}, error) {
	model := {{.ModelTypeName}}{}
	{{.DeserializeFieldsCode}}
	return model, nil
}
`))
)

//{"url": "/user/profile", "auth": false}

type apigenConfig struct {
	URL string `json:"url"`
	Auth bool `json:"auth"`
	Method string `json:"method"`
}

func getApiMethodInfo(fDecl *ast.FuncDecl) (apiTypeName string, methodName string, paramsTypeName string) {
	apiTypeName = fDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	methodName = fDecl.Name.Name
	paramsTypeName = fDecl.Type.Params.List[1].Type.(*ast.Ident).Name
	return
}

func generateHandler(out io.Writer, fDecl *ast.FuncDecl, config apigenConfig) (
	apiTypeName string, apiMethodName string, paramsTypeName string) {

	apiTypeName, apiMethodName, paramsTypeName = getApiMethodInfo(fDecl)

	fmt.Printf("Generating handler for %s.%s(%s)\n", apiTypeName, apiMethodName, paramsTypeName)

	authCheck := ""
	if config.Auth {
		authCheck = checkAuthenticationStr
	}
	httpMethodCheck := ""
	if config.Method != "" {
		httpMethodCheck = `
	// checking http method
	if r.Method != "` + config.Method + `" {
		handleError(ApiError{http.StatusNotAcceptable,fmt.Errorf("bad method")}, w)
		return
	}
`
	}

	handlerTpl.Execute(out, handlerTplArg{
		ApiMethodName: apiMethodName,
		ApiTypeName: apiTypeName,
		CheckAuthBlock: authCheck,
		CheckHttpMethodBlock: httpMethodCheck,
		ParamsTypeName: paramsTypeName,
	})

	return apiTypeName, apiMethodName, paramsTypeName
}

func generateDeserializers(out io.Writer, node *ast.File, paramsTypesNames map[string]bool) {
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.GenDecl\n", f)
			continue
		}
		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
				continue
			}

			currTypeName := currType.Name.Name

			if !paramsTypesNames[currTypeName] {
				fmt.Printf("SKIP %s is not a type of handler's parameter\n", currTypeName)
				continue;
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
				continue
			}

			generateDeserializerForStruct(out, currStruct, currTypeName)
		}
	}
}

func generateDeserializerForStruct(out io.Writer, currStruct *ast.StructType, structName string) {
	fmt.Fprintf(out, "//Deserializer for struct %s\n", structName)

	var fieldsBuf bytes.Buffer
	fieldsBuf.WriteString(`
	var valStr string
`)
	for _, field := range currStruct.Fields.List {
		if field.Tag != nil {
			fieldName := field.Names[0].Name
			fieldType := field.Type.(*ast.Ident).Name
			tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
			fieldMetaStr := tag.Get("apivalidator")

			fmt.Printf("Generating deserialization and validation for field %s with tag %s\n", fieldName, fieldMetaStr)

			fieldsBuf.WriteString(generateFieldDeserializer(fieldName, fieldType, fieldMetaStr))
			fieldsBuf.WriteRune('\n')
		}
	}
	deserializeModelTpl.Execute(out, deserializeModelTplArg{
		ModelTypeName: structName,
		DeserializeFieldsCode: fieldsBuf.String(),
	})
}

/*
type OtherCreateParams struct {
	Username string `apivalidator:"required,min=3"`
	Name     string `apivalidator:"paramname=account_name"`
	Class    string `apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`
	Level    int    `apivalidator:"min=1,max=50"`
}
*/

func generateFieldDeserializer(fieldName string, fieldType string, fieldMetaStr string) (string) {
	var fieldBuf bytes.Buffer

	var whereFrom string
	tags := make(map[string]string)

	if fieldMetaStr != "" {
		tagsArr := strings.Split(fieldMetaStr, ",")
		for _, tag := range tagsArr {
			tagSplit := strings.Split(tag, "=")
			if len(tagSplit) == 2 {
				tags[tagSplit[0]] = tagSplit[1]
			} else if len(tagSplit) == 1 {
				tags[tagSplit[0]] = "true"
			} else {
				panic("Unsupported field tag")
			}
		}
	} else {
		tags = make(map[string]string)
	}

	paramnameTag, isSet := tags["paramname"]
	if isSet {
		whereFrom = paramnameTag
	} else {
		whereFrom = strings.ToLower(fieldName)
	}

	_, isRequired := tags["required"]

	defaultVal, defaultValIsSet := tags["default"]
	
	fieldBuf.WriteString(`
	if r.Method == "GET" {
		valStr = r.URL.Query().Get("` + whereFrom + `")
	} else {
		valStr = r.FormValue("` + whereFrom +`")
	}
`)

	if defaultValIsSet {
		fieldBuf.WriteString(`
	
	if valStr == "" {
		valStr = "` + defaultVal + `"
	}
`)
	}

	if isRequired {
		fieldBuf.WriteString(`
	if valStr == "" {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("` + whereFrom + ` must me not empty")}
	}
`)
	}

	enumVals, enumValsAreSet := tags["enum"]
	if enumValsAreSet {
		split := strings.Split(enumVals, "|")
		commaSeparated := strings.Join(split, ", ")
		splitWithQuotes := make([]string, 0, len(split))

		for _, s := range split {
			splitWithQuotes = append(splitWithQuotes, "\"" + s + "\"")
		}
		commaSeparatedWithQuotes := strings.Join(splitWithQuotes, ", ")

		fieldBuf.WriteString(`
	switch (valStr) {
	case ` + commaSeparatedWithQuotes + `:
		//ok!
	default:
		err := fmt.Errorf("` + whereFrom + ` must be one of [` + commaSeparated + `]")
		return model, ApiError{http.StatusBadRequest, err}
	}
`)
	}

	if (fieldType == "string") {
		fieldBuf.WriteString(`
	model.` + fieldName + ` = valStr
`)
	} else if (fieldType == "int") {
		fieldBuf.WriteString(`
	var convErr error
	model.` + fieldName + `, convErr = strconv.Atoi(valStr)
	if convErr != nil {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("` + whereFrom + ` must be int")}
	}
`)
	}

	minVal, minValIsSet := tags["min"]
	if minValIsSet && fieldType == "string" {
		fieldBuf.WriteString(`
	if len(model.` + fieldName + `) < ` + minVal + ` {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("` + whereFrom + ` len must be >= ` + minVal + `")}
	}
`)
	}
	if minValIsSet && fieldType == "int" {
		fieldBuf.WriteString(`
	if model.` + fieldName + ` < ` + minVal + ` {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("` + whereFrom + ` must be >= ` + minVal + `")}
	}
`)
	}

	maxVal, maxValIsSet := tags["max"]
	if maxValIsSet && fieldType == "string" {
		fieldBuf.WriteString(`
	if len(model.` + fieldName + `) > ` + maxVal + ` {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("` + whereFrom + ` len must be <= ` + maxVal + `")}
	}
`)
	}
	if maxValIsSet && fieldType == "int" {
		fieldBuf.WriteString(`
	if model.` + fieldName + ` > ` + maxVal + ` {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("` + whereFrom + ` must be <= ` + maxVal + `")}
	}
`)
	}

	return fieldBuf.String()
}

func generateServeHttp(out io.Writer, apiTypeName string, configs []handlerConfig) {
	fmt.Fprint(out, `
func (h *` + apiTypeName + `) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {`)
	for _, config := range configs {
		fmt.Fprint(out, `
	case "` + config.UrlPath + `":
		h.handler` + config.ApiMethodName + `(w, r)`)
	}

	fmt.Fprint(out, `
	default:
		handleError(ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}, w)`)

	fmt.Fprint(out, `
	}
}
`)
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	for _, pkg := range genPackages {
		fmt.Fprintln(out, `import "` + pkg + `"`)
	}
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, helperFuncsSrc)

	paramsTypesNames := make(map[string]bool)
	handlerConfigs := make(map[string][]handlerConfig)

	// generate entry-point handlers
	for _, f := range node.Decls {
		fDecl, isFunc := f.(*ast.FuncDecl); 
		if !isFunc {
			continue;
		}

		if fDecl.Doc == nil {
			continue;
		}

		needCodegen := false
		var config apigenConfig
		for _, comment := range fDecl.Doc.List {
			needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// apigen:api")
			configStr := strings.Replace(comment.Text, "// apigen:api ", "", 1)
			config = apigenConfig{}
			json.Unmarshal([]byte(configStr), &config)
		}

		if needCodegen {
			apiTypeName, apiMethodName, pTypeName := generateHandler(out, fDecl, config)
			paramsTypesNames[pTypeName] = true		
			handlerConfigs[apiTypeName] = append(handlerConfigs[apiTypeName], handlerConfig{
				ApiMethodName: apiMethodName,
				UrlPath: config.URL,
			})
		}
	}

	generateDeserializers(out, node, paramsTypesNames)

	for apiTypeName, apiHandlerConfigs := range handlerConfigs {
		generateServeHttp(out, apiTypeName, apiHandlerConfigs)
	}
}