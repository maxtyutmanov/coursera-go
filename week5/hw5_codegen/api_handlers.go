package main

import "fmt"
import "net/http"
import "strconv"
import "encoding/json"



type finalResponse struct {
	Error string `json:"error"`
	Response interface{} `json:"response,omitempty"`
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


func (api *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	var err error
	
	
	deserializedParams, err := deserializeProfileParams(r)
	if err != nil {
		handleError(err, w)
		return
	}
	result, err := api.Profile(r.Context(), deserializedParams)
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

func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	var err error
	
	// checking http method
	if r.Method != "POST" {
		handleError(ApiError{http.StatusNotAcceptable,fmt.Errorf("bad method")}, w)
		return
	}

	
	// checking authentication
	err = checkAuth(w, r)
	if err != nil {
		handleError(err, w)
		return
	}

	deserializedParams, err := deserializeCreateParams(r)
	if err != nil {
		handleError(err, w)
		return
	}
	result, err := api.Create(r.Context(), deserializedParams)
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

func (api *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	var err error
	
	// checking http method
	if r.Method != "POST" {
		handleError(ApiError{http.StatusNotAcceptable,fmt.Errorf("bad method")}, w)
		return
	}

	
	// checking authentication
	err = checkAuth(w, r)
	if err != nil {
		handleError(err, w)
		return
	}

	deserializedParams, err := deserializeOtherCreateParams(r)
	if err != nil {
		handleError(err, w)
		return
	}
	result, err := api.Create(r.Context(), deserializedParams)
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
//Deserializer for struct ProfileParams

func deserializeProfileParams(r *http.Request) (ProfileParams, error) {
	model := ProfileParams{}
	
	var valStr string

	if r.Method == "GET" {
		valStr = r.URL.Query().Get("login")
	} else {
		valStr = r.FormValue("login")
	}

	if valStr == "" {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("login must me not empty")}
	}

	model.Login = valStr


	return model, nil
}
//Deserializer for struct CreateParams

func deserializeCreateParams(r *http.Request) (CreateParams, error) {
	model := CreateParams{}
	
	var valStr string

	if r.Method == "GET" {
		valStr = r.URL.Query().Get("login")
	} else {
		valStr = r.FormValue("login")
	}

	if valStr == "" {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("login must me not empty")}
	}

	model.Login = valStr

	if len(model.Login) < 10 {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("login len must be >= 10")}
	}


	if r.Method == "GET" {
		valStr = r.URL.Query().Get("full_name")
	} else {
		valStr = r.FormValue("full_name")
	}

	model.Name = valStr


	if r.Method == "GET" {
		valStr = r.URL.Query().Get("status")
	} else {
		valStr = r.FormValue("status")
	}

	
	if valStr == "" {
		valStr = "user"
	}

	switch (valStr) {
	case "user", "moderator", "admin":
		//ok!
	default:
		err := fmt.Errorf("status must be one of [user, moderator, admin]")
		return model, ApiError{http.StatusBadRequest, err}
	}

	model.Status = valStr


	if r.Method == "GET" {
		valStr = r.URL.Query().Get("age")
	} else {
		valStr = r.FormValue("age")
	}

	var convErr error
	model.Age, convErr = strconv.Atoi(valStr)
	if convErr != nil {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("age must be int")}
	}

	if model.Age < 0 {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("age must be >= 0")}
	}

	if model.Age > 128 {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("age must be <= 128")}
	}


	return model, nil
}
//Deserializer for struct OtherCreateParams

func deserializeOtherCreateParams(r *http.Request) (OtherCreateParams, error) {
	model := OtherCreateParams{}
	
	var valStr string

	if r.Method == "GET" {
		valStr = r.URL.Query().Get("username")
	} else {
		valStr = r.FormValue("username")
	}

	if valStr == "" {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("username must me not empty")}
	}

	model.Username = valStr

	if len(model.Username) < 3 {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("username len must be >= 3")}
	}


	if r.Method == "GET" {
		valStr = r.URL.Query().Get("account_name")
	} else {
		valStr = r.FormValue("account_name")
	}

	model.Name = valStr


	if r.Method == "GET" {
		valStr = r.URL.Query().Get("class")
	} else {
		valStr = r.FormValue("class")
	}

	
	if valStr == "" {
		valStr = "warrior"
	}

	switch (valStr) {
	case "warrior", "sorcerer", "rouge":
		//ok!
	default:
		err := fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")
		return model, ApiError{http.StatusBadRequest, err}
	}

	model.Class = valStr


	if r.Method == "GET" {
		valStr = r.URL.Query().Get("level")
	} else {
		valStr = r.FormValue("level")
	}

	var convErr error
	model.Level, convErr = strconv.Atoi(valStr)
	if convErr != nil {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("level must be int")}
	}

	if model.Level < 1 {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("level must be >= 1")}
	}

	if model.Level > 50 {
		return model, ApiError{http.StatusBadRequest, fmt.Errorf("level must be <= 50")}
	}


	return model, nil
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.handlerProfile(w, r)
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		handleError(ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}, w)
	}
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		handleError(ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}, w)
	}
}
