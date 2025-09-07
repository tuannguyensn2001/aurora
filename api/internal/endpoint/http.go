package endpoint

import (
	"context"
	"encoding/json"
	"net/http"

	"api/internal/dto"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// MakeHTTPHandler creates HTTP handlers for all endpoints
func MakeHTTPHandler(endpoints Endpoints) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	// Attribute routes - matching NestJS controller paths with api prefix
	r.Methods("POST").Path("/api/v1/attributes").Handler(httptransport.NewServer(
		endpoints.CreateAttribute,
		decodeCreateAttributeRequest,
		encodeCreatedResponse,
		options...,
	))

	r.Methods("GET").Path("/api/v1/attributes").Handler(httptransport.NewServer(
		endpoints.GetAllAttributes,
		decodeGetAllAttributesRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/api/v1/attributes/{id}").Handler(httptransport.NewServer(
		endpoints.GetAttributeByID,
		decodeGetAttributeByIDRequest,
		encodeResponse,
		options...,
	))

	r.Methods("PATCH").Path("/api/v1/attributes/{id}").Handler(httptransport.NewServer(
		endpoints.UpdateAttribute,
		decodeUpdateAttributeRequest,
		encodeResponse,
		options...,
	))

	r.Methods("DELETE").Path("/api/v1/attributes/{id}").Handler(httptransport.NewServer(
		endpoints.DeleteAttribute,
		decodeDeleteAttributeRequest,
		encodeNoContentResponse,
		options...,
	))

	r.Methods("PATCH").Path("/api/v1/attributes/{id}/increment-usage").Handler(httptransport.NewServer(
		endpoints.IncrementAttributeUsageCount,
		decodeIncrementAttributeUsageCountRequest,
		encodeNoContentResponse,
		options...,
	))

	r.Methods("PATCH").Path("/api/v1/attributes/{id}/decrement-usage").Handler(httptransport.NewServer(
		endpoints.DecrementAttributeUsageCount,
		decodeDecrementAttributeUsageCountRequest,
		encodeNoContentResponse,
		options...,
	))

	// Segment routes - matching NestJS controller paths with api prefix
	r.Methods("POST").Path("/api/v1/segments").Handler(httptransport.NewServer(
		endpoints.CreateSegment,
		decodeCreateSegmentRequest,
		encodeCreatedResponse,
		options...,
	))

	r.Methods("GET").Path("/api/v1/segments").Handler(httptransport.NewServer(
		endpoints.GetAllSegments,
		decodeGetAllSegmentsRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/api/v1/segments/{id}").Handler(httptransport.NewServer(
		endpoints.GetSegmentByID,
		decodeGetSegmentByIDRequest,
		encodeResponse,
		options...,
	))

	r.Methods("PATCH").Path("/api/v1/segments/{id}").Handler(httptransport.NewServer(
		endpoints.UpdateSegment,
		decodeUpdateSegmentRequest,
		encodeResponse,
		options...,
	))

	r.Methods("DELETE").Path("/api/v1/segments/{id}").Handler(httptransport.NewServer(
		endpoints.DeleteSegment,
		decodeDeleteSegmentRequest,
		encodeNoContentResponse,
		options...,
	))

	return r
}

// Decode functions
func decodeCreateAttributeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req dto.CreateAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return CreateAttributeRequest{Request: req}, nil
}

func decodeGetAllAttributesRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return GetAllAttributesRequest{}, nil
}

func decodeGetAttributeByIDRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}
	return GetAttributeByIDRequest{ID: id}, nil
}

func decodeUpdateAttributeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}

	var req dto.UpdateAttributeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return UpdateAttributeRequest{ID: id, Request: req}, nil
}

func decodeDeleteAttributeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}
	return DeleteAttributeRequest{ID: id}, nil
}

func decodeIncrementAttributeUsageCountRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}
	return IncrementAttributeUsageCountRequest{ID: id}, nil
}

func decodeDecrementAttributeUsageCountRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}
	return DecrementAttributeUsageCountRequest{ID: id}, nil
}

// Encode functions
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeCreatedResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(response)
}

func encodeNoContentResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// encodeError handles errors from endpoints
func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Determine status code based on error message
	statusCode := http.StatusInternalServerError
	errorType := "Internal Server Error"

	errMsg := err.Error()
	if contains(errMsg, "not found") {
		statusCode = http.StatusNotFound
		errorType = "Not Found"
	} else if contains(errMsg, "already exists") || contains(errMsg, "required") || contains(errMsg, "cannot delete") {
		statusCode = http.StatusConflict
		errorType = "Conflict"
	} else if contains(errMsg, "invalid") || contains(errMsg, "bad request") {
		statusCode = http.StatusBadRequest
		errorType = "Bad Request"
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error:   errorType,
		Message: errMsg,
	})
}

// Helper function to check if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Segment decode functions
func decodeCreateSegmentRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req dto.CreateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return CreateSegmentRequest{Request: req}, nil
}

func decodeGetAllSegmentsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return GetAllSegmentsRequest{}, nil
}

func decodeGetSegmentByIDRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}
	return GetSegmentByIDRequest{ID: id}, nil
}

func decodeUpdateSegmentRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}

	var req dto.UpdateSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return UpdateSegmentRequest{ID: id, Request: req}, nil
}

func decodeDeleteSegmentRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, err := ParseID(vars["id"])
	if err != nil {
		return nil, err
	}
	return DeleteSegmentRequest{ID: id}, nil
}
