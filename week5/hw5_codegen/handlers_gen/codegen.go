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
)

type handlerTplArg struct {
	ApiTypeName string
	ApiMethodName string
	CheckAuthBlock string
	ParamsTypeName string
}

var (
	genPackages = []string{
		"fmt",
		"net/http",
		"strings",
		"strconv",
	}

	helperFuncsSrc = `
	type ApiValidator struct {
	}
	
	func (validator *ApiValidator) ValidateRequiredStr(fieldName string, fieldValue string) (*ApiError) {
		if (fieldValue == "") {
			return &ApiError{http.StatusBadRequest, fmt.Errorf("%s must me not empty", fieldName)}
		}
		return nil
	}
	
	func (validator *ApiValidator) ValidateRequiredInt(fieldName string, fieldValue int) (*ApiError) {
		if (fieldValue == 0) {
			return &ApiError{http.StatusBadRequest, fmt.Errorf("%s must me not empty", fieldName)}
		}
		return nil
	}
	
	func (validator *ApiValidator) ValidateMinStr(fieldName string, fieldValue string, min int) (*ApiError) {
		if (len(fieldValue) < min) {
			return &ApiError{http.StatusBadRequest, fmt.Errorf("%s len must be >= %v", fieldName, min)}
		}
		return nil
	}
	
	func (validator *ApiValidator) ValidateMinInt(fieldName string, fieldValue int, min int) (*ApiError) {
		if (fieldValue < min) {
			return &ApiError{http.StatusBadRequest, fmt.Errorf("%s must be >= %v", fieldName, min)}
		}
		return nil
	}
	
	func (validator *ApiValidator) ValidateMaxStr(fieldName string, fieldValue string, max int) (*ApiError) {
		if (len(fieldValue) < max) {
			return &ApiError{http.StatusBadRequest, fmt.Errorf("%s len must be <= %v", fieldName, max)}
		}
		return nil
	}
	
	func (validator *ApiValidator) ValidateMaxInt(fieldName string, fieldValue int, max int) (*ApiError) {
		if (fieldValue < max) {
			return &ApiError{http.StatusBadRequest, fmt.Errorf("%s must be <= %v", fieldName, max)}
		}
		return nil
	}
	
	func (validator *ApiValidator) ValidateEnum(fieldName string, 
		fieldValue string, allowedValues []string, defaultValue string) (*ApiError) {
	
		for _, val := range allowedValues {
			if val == fieldValue {
				return nil
			}
		}
	
		err := fmt.Errorf("%s must be one of [%s]", fieldName, strings.Join(allowedValues, ", "))
		return &ApiError{http.StatusBadRequest, err}
	}

	func checkAuth(w http.ResponseWriter, r *http.Request) *ApiError {
		if (r.Header.Get("X-Auth") != "100500") {
			return &ApiError{http.StatusUnauthorized, fmt.Errorf("unauthorized")}
		}
		return nil
	}

	func handleError(err *ApiError, w http.ResponseWriter) {
		w.WriteHeader(err.HTTPStatus)
		w.Write([]byte(err.Err.Error()))
	}
`

	checkRequiredIntTpl = template.Must(template.New("checkRequired_int").Parse(`
		
	`))

	parseAndValidateIntTpl = template.Must(template.New("parseAndValidate_int").Parse(`
		// {{.FieldName}}
		var {{.FieldName}}Raw string
		{{.FieldName}}Raw = {{.HttpRequestVar}}.FormValue("{{.WhereFrom}}")
		// Check required validator
		{{.CheckRequired}}
		// Substitute empty value with default if the default value is set
		{{.SubstituteDefaultValue}}
		// Check other validators
		{{.CheckValidators}}
		
		parsed.{{.FieldName}}, convErr = strconv.Atoi({{.FieldName}}Raw)
	)
	`))

	handlerTpl = template.Must(template.New("handler").Parse(`
	func (api *{{.ApiTypeName}}) handler{{.ApiMethodName}}(w http.ResponseWriter, r *http.Request) {
		var err *ApiError
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
		resultStr := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resultStr)
	}
	`))

	checkAuthenticationStr = `
		// checking authentication
		err = checkAuth(w, r)
		if err != nil {
			handleError(err, w)
			return
		}
		deserializedParams := deserialize()
	`
)

//{"url": "/user/profile", "auth": false}

type apigenConfig struct {
	URL string `json:"url"`
	Auth bool `json:"auth"`
}

/*
type OtherCreateParams struct {
	Username string `apivalidator:"required,min=3"`
	Name     string `apivalidator:"paramname=account_name"`
	Class    string `apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`
	Level    int    `apivalidator:"min=1,max=50"`
}
*/

func generateDeserializer(out io.Writer, paramType *ast.StructType) {
	// for _, field := range paramType.Fields.List {
	// 	fieldName := field.Names[0].Name

	// 	if field.Tag != nil {
	// 		tag := reflect.StructTag(field.Tag.Value)
	// 		if tag.Get("apivalidator") == "-" {
				
	// 		}
	// 	}
	// }
}

func getApiMethodInfo(fDecl *ast.FuncDecl) (apiTypeName string, methodName string, paramsTypeName string) {
	apiTypeName = fDecl.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
	methodName = fDecl.Name.Name
	paramsTypeName = fDecl.Type.Params.List[1].Type.(*ast.Ident).Name
	return
}

func generateHandler(out io.Writer, fDecl *ast.FuncDecl, config apigenConfig) {
	apiTypeName, apiMethodName, paramsTypeName := getApiMethodInfo(fDecl)

	fmt.Printf("Generating handler for %s.%s(%s)\n", apiTypeName, apiMethodName, paramsTypeName)

	authCheck := ""
	if config.Auth {
		authCheck = checkAuthenticationStr
	}

	handlerTpl.Execute(out, handlerTplArg{
		ApiMethodName: apiMethodName,
		ApiTypeName: apiTypeName,
		CheckAuthBlock: authCheck,
		ParamsTypeName: paramsTypeName,
	})
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

		generateHandler(out, fDecl, config)
	}
}