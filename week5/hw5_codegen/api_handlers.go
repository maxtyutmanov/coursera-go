package main

import "fmt"
import "net/http"
import "strings"
import "strconv"


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
		w.Write([]byte(resultStr)
	}
	
	func (api *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
		var err *ApiError
		
		// checking authentication
		err = checkAuth(w, r)
		if err != nil {
			handleError(err, w)
			return
		}
		deserializedParams := deserialize()
	
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
		w.Write([]byte(resultStr)
	}
	
	func (api *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) {
		var err *ApiError
		
		// checking authentication
		err = checkAuth(w, r)
		if err != nil {
			handleError(err, w)
			return
		}
		deserializedParams := deserialize()
	
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
		w.Write([]byte(resultStr)
	}
	