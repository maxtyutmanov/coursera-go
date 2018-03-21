package main

// это программа для которой ваш кодогенератор будет писать код
// запускать через go test -v, как обычно

// этот код закомментирован чтобы он не светился в тестовом покрытии

import (
	"fmt"
	"net/http"
	"strings"
	"strconv"
)

func validateRequired(fieldName string, fieldValue string) (*ApiError) {
	if (fieldValue == "") {
		return &ApiError{http.StatusBadRequest, fmt.Errorf("%s must me not empty", fieldName)}
	}
	return nil
}

func validateMin_string(fieldName string, fieldValue string, min int) (*ApiError) {
	if (len(fieldValue) < min) {
		return &ApiError{http.StatusBadRequest, fmt.Errorf("%s len must be >= %v", fieldName, min)}
	}
	return nil
}

func validateMin_int(fieldName string, fieldValue string, min int) (*ApiError) {
	intval, _ := strconv.Atoi(fieldValue)
	if (intval < min) {
		return &ApiError{http.StatusBadRequest, fmt.Errorf("%s must be >= %v", fieldName, min)}
	}
	return nil
}

func validateMax_string(fieldName string, fieldValue string, max int) (*ApiError) {
	if (len(fieldValue) < max) {
		return &ApiError{http.StatusBadRequest, fmt.Errorf("%s len must be <= %v", fieldName, max)}
	}
	return nil
}

func validateMax_int(fieldName string, fieldValue string, max int) (*ApiError) {
	intval, _ := strconv.Atoi(fieldValue)
	if (intval > max) {
		return &ApiError{http.StatusBadRequest, fmt.Errorf("%s must be <= %v", fieldName, max)}
	}
	return nil
}

func validateEnum_string(fieldName string, 
	fieldValue string, allowedValues []string, defaultValue string) (*ApiError) {

	for _, val := range allowedValues {
		if val == fieldValue {
			return nil
		}
	}

	err := fmt.Errorf("%s must be one of [%s]", fieldName, strings.Join(allowedValues, ", "))
	return &ApiError{http.StatusBadRequest, err}
}

func (params *CreateParams) ParseAndValidate(r *http.Request) (*ApiError) {
	
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

func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	//check auth
	

	//parse params

	params := CreateParams{}
	
	
}

func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	
	if (r.URL == "/user/create" && r.Method == http.MethodPost) {
		
	}
}

func main() {
	// будет вызван метод ServeHTTP у структуры MyApi
	http.Handle("/user/", NewMyApi())

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
