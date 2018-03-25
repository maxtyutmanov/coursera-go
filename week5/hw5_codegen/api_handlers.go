package main

import "fmt"
import "net/http"
import "strconv"
import "encoding/json"



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


func (api *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) {
	var err *ApiError
	
	
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
	resultStr := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resultStr))
}

func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	var err *ApiError
	
	// checking http method
	if r.Method != "POST" {
		handleError(&ApiError{http.StatusNotAcceptable,fmt.Errorf("bad method")})
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
	resultStr := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resultStr))
}

func (api *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
	var err *ApiError
	
	// checking http method
	if r.Method != "POST" {
		handleError(&ApiError{http.StatusNotAcceptable,fmt.Errorf("bad method")})
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
	resultStr := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resultStr))
}
//Deserializer for struct ProfileParams

func deserializeProfileParams(r *http.Request) (ProfileParams, *ApiError) {
	model := ProfileParams{}
	
	var valStr string

	valStr = r.FormValue("login")

	if valStr == "" {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Login must me not empty")}
	}

	model.Login = valStr


	return model, nil
}
//Deserializer for struct CreateParams

func deserializeCreateParams(r *http.Request) (CreateParams, *ApiError) {
	model := CreateParams{}
	
	var valStr string

	valStr = r.FormValue("login")

	if valStr == "" {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Login must me not empty")}
	}

	model.Login = valStr

	if len(model.Login) < 10 {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Login len must be >= 10")}
	}


	valStr = r.FormValue("full_name")

	model.Name = valStr


	valStr = r.FormValue("status")

	
	if valStr == "" {
		valStr = "user"
	}

	switch (valStr) {
	case "user", "moderator", "admin":
		//ok!
	default:
		err := fmt.Errorf("Status must be one of [user, moderator, admin]")
		return nil, &ApiError{http.StatusBadRequest, err}
	}

	model.Status = valStr


	valStr = r.FormValue("age")

	model.Age, _ = strconv.Atoi(valStr)

	if model.Age < 0 {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Age must be >= 0")}
	}

	if model.Age > 128 {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Age must be <= 128")}
	}


	return model, nil
}
//Deserializer for struct OtherCreateParams

func deserializeOtherCreateParams(r *http.Request) (OtherCreateParams, *ApiError) {
	model := OtherCreateParams{}
	
	var valStr string

	valStr = r.FormValue("username")

	if valStr == "" {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Username must me not empty")}
	}

	model.Username = valStr

	if len(model.Username) < 3 {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Username len must be >= 3")}
	}


	valStr = r.FormValue("account_name")

	model.Name = valStr


	valStr = r.FormValue("class")

	
	if valStr == "" {
		valStr = "warrior"
	}

	switch (valStr) {
	case "warrior", "sorcerer", "rouge":
		//ok!
	default:
		err := fmt.Errorf("Class must be one of [warrior, sorcerer, rouge]")
		return nil, &ApiError{http.StatusBadRequest, err}
	}

	model.Class = valStr


	valStr = r.FormValue("level")

	model.Level, _ = strconv.Atoi(valStr)

	if model.Level < 1 {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Level must be >= 1")}
	}

	if model.Level > 50 {
		return nil, &ApiError{http.StatusBadRequest, fmt.Errorf("Level must be <= 50")}
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
		handleError(&ApiError{http.StatusBadRequest, fmt.Errorf("unknown method")})
	}
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.handlerCreate(w, r)
	default:
		handleError(&ApiError{http.StatusBadRequest, fmt.Errorf("unknown method")})
	}
}
