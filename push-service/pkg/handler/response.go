package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"
)

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type ApiResponse struct {
	Message          string   `json:"message"`
	Error            bool     `json:"error"`
	Data             any      `json:"data,omitempty"`
	ValidationErrors []string `json:"validationErrors,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("failed to encode JSON response", logger.WithError(err))
	}
}

// writes a success response with default status code 200 OK
func RespondWithSuccess(w http.ResponseWriter, message string, data any) {
	RespondWithSuccessAndStatus(w, http.StatusOK, message, data)
}

// writes a success response with a custom status code
func RespondWithSuccessAndStatus(w http.ResponseWriter, status int, message string, data any) {
	response := ApiResponse{
		Message: message,
		Error:   false,
		Data:    data,
	}

	WriteJSON(w, status, response)
}

// writes an error response
func RespondWithError(w http.ResponseWriter, status int, message string, err error) {
	response := ApiResponse{
		Message: message,
		Error:   true,
	}

	if err != nil {

		logger.Error(message, logger.WithError(err))
	}

	WriteJSON(w, status, response)
}

// writes a validation error response
func RespondWithValidationError(w http.ResponseWriter, validationErrors any) {

	var fieldErrors []string

	// check if it's the expected type
	if errs, ok := validationErrors.(validator.ValidationErrors); ok {
		for _, err := range errs {

			fieldErrors = append(fieldErrors, GetValidationErrorMessage(err))
		}
	}

	response := ApiResponse{
		Message:          "Validation Error",
		Error:            true,
		ValidationErrors: fieldErrors,
	}

	logger.Info("validation failed",
		logger.Fields{"response": response})

	WriteJSON(w, http.StatusBadRequest, response)
}

// returns a human readable validation error message
func GetValidationErrorMessage(err validator.FieldError) string {
	fieldName := err.Field()
	switch err.Tag() {
	case "required":
		return "Field '" + fieldName + "' is required"
	case "email":
		return "Field '" + fieldName + "' must be a valid email address"
	case "min":
		return "Field '" + fieldName + "' must be at least " + err.Param() + " characters long"
	case "max":
		return "Field '" + fieldName + "' must be at most " + err.Param() + " characters long"
	case "oneof":
		return "Field '" + fieldName + "' must be one of: " + err.Param()
	default:
		return "Field '" + fieldName + "' has an invalid value"
	}
}
