// Package httpx provides HTTP utilities including JSend response formatting
package httpx

import (
	"encoding/json"
	"net/http"
)

// JSendSuccess represents a successful JSend response
type JSendSuccess struct {
	Data   any    `json:"data" swaggertype:"object"`
	Status string `json:"status" example:"success"`
}

// JSendFail represents a client error JSend response (validation errors, missing fields, etc.)
type JSendFail struct {
	Data   map[string]any `json:"data"`
	Status string         `json:"status" example:"fail"`
}

// JSendFailPhoneInvalid represents a phone validation error example
type JSendFailPhoneInvalid struct {
	Status string             `json:"status" example:"fail"`
	Data   JSendFailPhoneData `json:"data"`
}

// JSendFailPhoneData represents phone validation error data
type JSendFailPhoneData struct {
	Phone string `json:"phone" example:"Invalid phone format (use E.164 format: +525512345678)"`
}

// JSendFailEmailExists represents an email already registered error example
type JSendFailEmailExists struct {
	Status string             `json:"status" example:"fail"`
	Data   JSendFailEmailData `json:"data"`
}

// JSendFailEmailData represents email already exists error data
type JSendFailEmailData struct {
	Email string `json:"email" example:"Email already registered"`
}

// JSendFailPhoneExists represents a phone already registered error example
type JSendFailPhoneExists struct {
	Status string                   `json:"status" example:"fail"`
	Data   JSendFailPhoneExistsData `json:"data"`
}

// JSendFailPhoneExistsData represents phone already exists error data
type JSendFailPhoneExistsData struct {
	Phone string `json:"phone" example:"Phone number already registered"`
}

// JSendFailOTPInvalid represents an invalid OTP error example
type JSendFailOTPInvalid struct {
	Status string           `json:"status" example:"fail"`
	Data   JSendFailOTPData `json:"data"`
}

// JSendFailOTPData represents invalid OTP error data
type JSendFailOTPData struct {
	OTP string `json:"otp" example:"Invalid OTP code"`
}

// JSendFailOTPExpired represents an expired OTP error example
type JSendFailOTPExpired struct {
	Status string                  `json:"status" example:"fail"`
	Data   JSendFailOTPExpiredData `json:"data"`
}

// JSendFailOTPExpiredData represents expired OTP error data
type JSendFailOTPExpiredData struct {
	OTP string `json:"otp" example:"OTP has expired - please request a new one"`
}

// JSendFailPhoneNotVerified represents a phone not verified error example
type JSendFailPhoneNotVerified struct {
	Status string                        `json:"status" example:"fail"`
	Data   JSendFailPhoneNotVerifiedData `json:"data"`
}

// JSendFailPhoneNotVerifiedData represents phone not verified error data
type JSendFailPhoneNotVerifiedData struct {
	Phone string `json:"phone" example:"Phone not verified - please verify OTP first"`
}

// JSendFailAgeRestriction represents an age restriction error example
type JSendFailAgeRestriction struct {
	Status string                      `json:"status" example:"fail"`
	Data   JSendFailAgeRestrictionData `json:"data"`
}

// JSendFailAgeRestrictionData represents age restriction error data
type JSendFailAgeRestrictionData struct {
	BirthDate string `json:"birth_date" example:"User must be at least 18 years old"`
}

// JSendFailRFCInvalid represents an invalid RFC format error example
type JSendFailRFCInvalid struct {
	Status string                  `json:"status" example:"fail"`
	Data   JSendFailRFCInvalidData `json:"data"`
}

// JSendFailRFCInvalidData represents invalid RFC format error data
type JSendFailRFCInvalidData struct {
	Error string `json:"error" example:"formato de RFC inválido (debe tener 13 caracteres alfanuméricos)"`
}

// JSendFailZipCodeInvalid represents an invalid ZIP code error example
type JSendFailZipCodeInvalid struct {
	Status string                      `json:"status" example:"fail"`
	Data   JSendFailZipCodeInvalidData `json:"data"`
}

// JSendFailZipCodeInvalidData represents invalid ZIP code error data
type JSendFailZipCodeInvalidData struct {
	Error string `json:"error" example:"formato de código postal inválido (debe tener 5 dígitos)"`
}

// JSendFailFiscalRegimeInvalid represents an invalid fiscal regime error example
type JSendFailFiscalRegimeInvalid struct {
	Status string                           `json:"status" example:"fail"`
	Data   JSendFailFiscalRegimeInvalidData `json:"data"`
}

// JSendFailFiscalRegimeInvalidData represents invalid fiscal regime error data
type JSendFailFiscalRegimeInvalidData struct {
	Error string `json:"error" example:"régimen fiscal inválido - valores permitidos: general, simplificado_confianza, actividad_empresarial, arrendamiento, salarios, incorporacion_fiscal"`
}

// JSendFailInvalidJSON represents an invalid JSON body error example
type JSendFailInvalidJSON struct {
	Status string                   `json:"status" example:"fail"`
	Data   JSendFailInvalidJSONData `json:"data"`
}

// JSendFailInvalidJSONData represents invalid JSON body error data
type JSendFailInvalidJSONData struct {
	Error string `json:"error" example:"Cuerpo de solicitud inválido"`
}

// JSendFailDocumentType represents an invalid document type error example
type JSendFailDocumentType struct {
	Status string                    `json:"status" example:"fail"`
	Data   JSendFailDocumentTypeData `json:"data"`
}

// JSendFailDocumentTypeData represents invalid document type error data
type JSendFailDocumentTypeData struct {
	Type string `json:"type" example:"Tipo de documento inválido - valores permitidos: circulation_card, ine_front, ine_back, driver_license_front, driver_license_back, profile_photo, fiscal_certificate"`
}

// JSendFailFileSize represents a file too large error example
type JSendFailFileSize struct {
	Status string                `json:"status" example:"fail"`
	Data   JSendFailFileSizeData `json:"data"`
}

// JSendFailFileSizeData represents file too large error data
type JSendFailFileSizeData struct {
	File string `json:"file" example:"Archivo demasiado grande - máximo 10 MB permitido"`
}

// JSendFailFileType represents an unsupported file type error example
type JSendFailFileType struct {
	Status string                `json:"status" example:"fail"`
	Data   JSendFailFileTypeData `json:"data"`
}

// JSendFailFileTypeData represents unsupported file type error data
type JSendFailFileTypeData struct {
	File string `json:"file" example:"Tipo de archivo no permitido - solo se permiten imágenes JPEG/PNG y archivos PDF"`
}

// JSendFailUserIDInvalid represents an invalid user ID format error example
type JSendFailUserIDInvalid struct {
	Status string                     `json:"status" example:"fail"`
	Data   JSendFailUserIDInvalidData `json:"data"`
}

// JSendFailUserIDInvalidData represents invalid user ID error data
type JSendFailUserIDInvalidData struct {
	UserID string `json:"user_id" example:"Formato de ID de usuario inválido"`
}

// JSendFailDocumentIDInvalid represents an invalid document ID format error example
type JSendFailDocumentIDInvalid struct {
	Status string                         `json:"status" example:"fail"`
	Data   JSendFailDocumentIDInvalidData `json:"data"`
}

// JSendFailDocumentIDInvalidData represents invalid document ID error data
type JSendFailDocumentIDInvalidData struct {
	DocumentID string `json:"document_id" example:"Formato de ID de documento inválido"`
}

// JSendError represents a server error JSend response (database errors, external service failures, etc.)
type JSendError struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Failed to connect to database"`
	Code    int    `json:"code,omitempty" example:"500"`
}

// RespondSuccess sends a JSend success response
func RespondSuccess(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(JSendSuccess{
		Status: "success",
		Data:   data,
	}); err != nil {
		// Log encoding error but don't try to write another response as headers are already sent
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RespondFail sends a JSend fail response (client errors)
func RespondFail(w http.ResponseWriter, statusCode int, data map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(JSendFail{
		Status: "fail",
		Data:   data,
	}); err != nil {
		// Log encoding error but don't try to write another response as headers are already sent
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// RespondError sends a JSend error response (server errors)
func RespondError(w http.ResponseWriter, statusCode int, message string, code ...int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := JSendError{
		Status:  "error",
		Message: message,
	}

	if len(code) > 0 {
		errResp.Code = code[0]
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		// Log encoding error but don't try to write another response as headers are already sent
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DecodeJSON decodes JSON from request body into the provided struct
func DecodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
