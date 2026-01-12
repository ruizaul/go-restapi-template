package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"tacoshare-delivery-api/internal/documents/models"
	"tacoshare-delivery-api/internal/documents/services"
	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"
	"tacoshare-delivery-api/pkg/validator"

	"github.com/google/uuid"
)

const (
	errDocumentsNotFound = "documentos no encontrados"
)

// DocumentHandler handles document HTTP requests
type DocumentHandler struct {
	documentService *services.DocumentService
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

// CreateDocument creates a new document record
// Deprecated: Use UpdateDocument (PATCH /documents/me) instead - it works as upsert
func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario no encontrado en el contexto")
		return
	}

	var req models.CreateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de solicitud inválido",
		})
		return
	}

	doc, err := h.documentService.CreateDocument(userID, &req)
	if err != nil {

		// Client-side validation errors (400 Bad Request)
		if err.Error() == "el usuario ya tiene documentos registrados" ||
			err.Error() == "usuario no encontrado" ||
			err.Error() == "formato de RFC inválido (debe tener 13 caracteres alfanuméricos)" ||
			err.Error() == "formato de código postal inválido (debe tener 5 dígitos)" ||
			err.Error() == "régimen fiscal inválido - valores permitidos: general, simplificado_confianza, actividad_empresarial, arrendamiento, salarios, incorporacion_fiscal" {
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
			return
		}

		// Server-side errors (500 Internal Server Error)
		httpx.RespondError(w, http.StatusInternalServerError, "Error al crear documentos")
		return
	}

	httpx.RespondSuccess(w, http.StatusCreated, doc)
}

// GetMyDocuments godoc
//
//	@Summary		Get my documents
//	@Description	Retrieve the authenticated user's complete document record including: **Document URLs** (circulation_card_url, ine_front_url, ine_back_url, driver_license_front_url, driver_license_back_url, profile_photo_url), **Vehicle Info** (vehicle_brand, vehicle_model, license_plate), **Fiscal Data** (fiscal_name, fiscal_rfc, fiscal_zip_code, fiscal_regime, fiscal_street, fiscal_ext_number, fiscal_int_number, fiscal_neighborhood, fiscal_city, fiscal_state, fiscal_certificate_url), and **Review Status** (reviewed: boolean indicating admin approval). Returns null if user has not created a document record yet. The reviewed field is set by admins via PATCH /documents/{user_id}/review after manual verification - true means documents are approved and driver can receive orders, false means pending review or rejected.
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.DocumentResponse	"Document record retrieved successfully with all fields, or null if no documents exist yet"	example({"status": "success", "data": {"id": "550e8400-e29b-41d4-a716-446655440000", "user_id": "123e4567-e89b-12d3-a456-426614174000", "vehicle_brand": "Honda", "vehicle_model": "CBR 250", "license_plate": "ABC-123-XYZ", "ine_front_url": "https://storage.example.com/docs/ine_front.jpg", "reviewed": false, "created_at": "2025-01-15T10:30:00Z", "updated_at": "2025-01-15T11:00:00Z"}})
//	@Failure		401	{object}	httpx.JSendError		"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		500	{object}	httpx.JSendError		"Internal server error - database query failed or connection error"
//	@Security		BearerAuth
//	@Router			/documents/me [get]
func (h *DocumentHandler) GetMyDocuments(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario no encontrado en el contexto")
		return
	}

	doc, err := h.documentService.GetDocumentByUserID(userID)
	if err != nil {
		if err.Error() == errDocumentsNotFound {
			// Return null instead of 404 - no documents yet is a valid state
			httpx.RespondSuccess(w, http.StatusOK, nil)
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener documentos")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, doc)
}

// GetMyDocumentStatus retrieves the document status for the authenticated user
// Deprecated: Use GetMyDocuments (GET /documents/me) instead - the reviewed field is included there
func (h *DocumentHandler) GetMyDocumentStatus(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario no encontrado en el contexto")
		return
	}

	status, err := h.documentService.GetDocumentStatus(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener estado de documentos")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, status)
}

// UpdateDocument godoc
//
//	@Summary		Create or update my documents (upsert)
//	@Description	Create or partially update the authenticated user's document record using upsert pattern (creates if doesn't exist, updates if exists). **All fields are optional** - send only the fields you want to create/update. **Typical workflow**: 1) Upload file via POST /documents/upload to get URL, 2) PATCH this endpoint with {"ine_front_url": "returned_url"} to save. Supports progressive submission - add vehicle info first, then fiscal data later, etc. **Validation rules**: RFC format (13 alphanumeric characters following SAT pattern: ^[A-ZÑ&]{3,4}\d{6}[A-Z0-9]{3}$), ZIP code (exactly 5 digits), fiscal_regime enum (general, simplificado_confianza, actividad_empresarial, arrendamiento, salarios, incorporacion_fiscal). **Note**: The reviewed field cannot be modified by users - only admins can change it via PATCH /documents/{user_id}/review after manual verification. Returns the complete document record after operation.
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.UpdateDocumentRequest		true	"Document fields to create/update - all optional, send only what you need"	example({"vehicle_brand": "Honda", "vehicle_model": "CBR 250", "license_plate": "ABC-123-XYZ", "ine_front_url": "https://storage.example.com/docs/ine_front.jpg", "fiscal_rfc": "PEGJ850101ABC"})
//	@Success		200		{object}	models.DocumentResponse				"Documents created or updated successfully - returns complete document record with all fields"
//	@Failure		400		{object}	httpx.JSendFailInvalidJSON			"Invalid JSON body - malformed JSON or syntax error"
//	@Failure		400		{object}	httpx.JSendFailRFCInvalid			"Invalid RFC format - debe tener 13 caracteres alfanuméricos siguiendo patrón SAT"
//	@Failure		400		{object}	httpx.JSendFailZipCodeInvalid		"Invalid ZIP code format - debe tener exactamente 5 dígitos"
//	@Failure		400		{object}	httpx.JSendFailFiscalRegimeInvalid	"Invalid fiscal regime - valores permitidos: general, simplificado_confianza, actividad_empresarial, arrendamiento, salarios, incorporacion_fiscal"
//	@Failure		401		{object}	httpx.JSendError					"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		500		{object}	httpx.JSendError					"Internal server error - database upsert operation failed or connection error"
//	@Security		BearerAuth
//	@Router			/documents/me [patch]
func (h *DocumentHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario no encontrado en el contexto")
		return
	}

	var req models.UpdateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de solicitud inválido",
		})
		return
	}

	doc, err := h.documentService.UpdateDocument(userID, &req)
	if err != nil {
		// Validation errors or user not found
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, doc)
}

// DeleteDocument godoc
//
//	@Summary		Delete my documents
//	@Description	Permanently delete the authenticated user's entire document record from the database. **This action is irreversible** - all data will be lost including vehicle information, all document URLs, and fiscal data. The files themselves remain in R2 storage but are no longer linked to the user account. After deletion, user can create a fresh document record via PATCH /documents/me and re-upload files. **Use cases**: User wants to restart document submission process from scratch, user wants to remove all stored document information, or driver is switching to a different vehicle and needs to resubmit all documents. Returns 404 if no document record exists for the user.
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httpx.JSendSuccess	"Documents deleted successfully - user can now create new document record"	example({"status": "success", "data": {"message": "Documentos eliminados exitosamente"}})
//	@Failure		401	{object}	httpx.JSendError	"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		404	{object}	httpx.JSendFail		"Documents not found - user has no document record to delete"	example({"status": "fail", "data": {"error": "Documentos no encontrados"}})
//	@Failure		500	{object}	httpx.JSendError	"Internal server error - database delete operation failed or connection error"
//	@Security		BearerAuth
//	@Router			/documents/me [delete]
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario no encontrado en el contexto")
		return
	}

	err := h.documentService.DeleteDocument(userID)
	if err != nil {
		if err.Error() == errDocumentsNotFound {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"error": "Documentos no encontrados",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al eliminar documentos")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]string{
		"message": "Documentos eliminados exitosamente",
	})
}

// GetDocumentByUserID godoc
//
//	@Summary		Get user documents (Admin)
//	@Description	**Admin-only endpoint** to retrieve document record for any user by their UUID. Used in driver onboarding workflow: admin reviews submitted documents, verifies authenticity (INE, driver license, circulation card match real identity), checks fiscal compliance (RFC format, tax certificate validity), and inspects vehicle documentation. Returns complete document record including: **Vehicle Info** (brand, model, license plate), **Document URLs** (INE front/back, driver license front/back, circulation card, profile photo - stored in R2), **Fiscal Data** (fiscal name, RFC, tax regime, complete address, zip code, tax certificate URL), and **Review Status** (reviewed boolean). **Access control**: Requires admin role JWT - non-admin users receive 403 Forbidden. The user_id path parameter must be valid UUID format. **Workflow**: 1) GET this endpoint to inspect documents, 2) Manually verify documents for authenticity, 3) PATCH /documents/{user_id}/review to approve/reject. Returns 404 if user has no document record.
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string							true	"User ID in UUID format"	example(550e8400-e29b-41d4-a716-446655440000)
//	@Success		200		{object}	models.DocumentResponse			"Document record retrieved successfully with all fields and review status"
//	@Failure		400		{object}	httpx.JSendFailUserIDInvalid	"Invalid user ID format - must be valid UUID"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		403		{object}	httpx.JSendError				"Forbidden - user does not have admin role"						example({"status": "error", "message": "Acceso denegado - se requiere rol de administrador", "code": 403})
//	@Failure		404		{object}	httpx.JSendFail					"Documents not found - specified user has no document record"	example({"status": "fail", "data": {"error": "Documentos no encontrados"}})
//	@Failure		500		{object}	httpx.JSendError				"Internal server error - database query failed or connection error"
//	@Security		BearerAuth
//	@Router			/documents/{user_id} [get]
func (h *DocumentHandler) GetDocumentByUserID(w http.ResponseWriter, r *http.Request) {
	// Get user ID from path parameter
	userIDParam := r.PathValue("user_id")
	if !validator.IsValidUUID(userIDParam) {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"user_id": "Formato de ID de usuario inválido",
		})
		return
	}

	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"user_id": "Error al parsear ID de usuario",
		})
		return
	}

	doc, err := h.documentService.GetDocumentByUserID(userID)
	if err != nil {
		if err.Error() == errDocumentsNotFound {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"error": "Documentos no encontrados",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener documentos")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, doc)
}

// MarkAsReviewed godoc
//
//	@Summary		Review user documents (Admin)
//	@Description	**Admin-only endpoint** to mark a user's documents as reviewed (approved) or unreviewed (rejected/pending) after manual verification. **Driver onboarding workflow**: 1) Admin calls GET /documents/{user_id} to retrieve all submitted documents, 2) Admin manually verifies: vehicle registration authenticity, INE/driver license validity and photo match, fiscal certificate legitimacy and RFC format correctness, vehicle insurance and circulation card validity, 3) Admin calls this endpoint with {"reviewed": true} to approve or {"reviewed": false} to reject. **Business logic impact**: Setting reviewed=true typically activates the driver account and makes them eligible for automatic order assignment. Setting reviewed=false may trigger notification to driver to resubmit documents. **Access control**: Requires admin role JWT - non-admin users receive 403 Forbidden. The user_id path parameter must be valid UUID. Request body must contain boolean 'reviewed' field. Returns success message with updated review status.
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string							true																			"User ID in UUID format"							example(550e8400-e29b-41d4-a716-446655440000)
//	@Param			request	body		object{reviewed=bool}			true																			"Review status - true to approve, false to reject"	example({"reviewed": true})
//	@Success		200		{object}	httpx.JSendSuccess				"Review status updated successfully - driver account status may be affected"	example({"status": "success", "data": {"message": "Estado de revisión actualizado exitosamente", "reviewed": true}})
//	@Failure		400		{object}	httpx.JSendFailUserIDInvalid	"Invalid user ID format - must be valid UUID"
//	@Failure		400		{object}	httpx.JSendFailInvalidJSON		"Invalid request body - malformed JSON or missing reviewed field"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		403		{object}	httpx.JSendError				"Forbidden - user does not have admin role"								example({"status": "error", "message": "Acceso denegado - se requiere rol de administrador", "code": 403})
//	@Failure		404		{object}	httpx.JSendFail					"Documents not found - specified user has no document record to review"	example({"status": "fail", "data": {"error": "Documentos no encontrados"}})
//	@Failure		500		{object}	httpx.JSendError				"Internal server error - database update operation failed or connection error"
//	@Security		BearerAuth
//	@Router			/documents/{user_id}/review [patch]
func (h *DocumentHandler) MarkAsReviewed(w http.ResponseWriter, r *http.Request) {
	// Get user ID from path parameter
	userIDParam := r.PathValue("user_id")
	if !validator.IsValidUUID(userIDParam) {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"user_id": "Formato de ID de usuario inválido",
		})
		return
	}

	userID, parseErr := uuid.Parse(userIDParam)
	if parseErr != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"user_id": "Error al parsear ID de usuario",
		})
		return
	}

	// Parse request body
	var req struct {
		Reviewed bool `json:"reviewed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de solicitud inválido",
		})
		return
	}

	err := h.documentService.MarkAsReviewed(userID, req.Reviewed)
	if err != nil {
		if err.Error() == errDocumentsNotFound {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"error": "Documentos no encontrados",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al actualizar estado de revisión")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message":  "Estado de revisión actualizado exitosamente",
		"reviewed": req.Reviewed,
	})
}

// GetAllDocuments godoc
//
//	@Summary		Get all documents (Admin)
//	@Description	**Admin-only endpoint** to retrieve all user documents with pagination. Returns paginated list of document records with complete information including vehicle data, document URLs, fiscal information, and review status. Used by admins to view pending document submissions that need review. **Pagination**: Use query parameters `page` (default: 1) and `limit` (default: 20, max: 100). **Access control**: Requires admin role JWT - non-admin users receive 403 Forbidden. Documents are ordered by creation date (newest first). Returns empty array if no documents exist in the system.
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int							false	"Page number (default: 1)"					minimum(1)	default(1)
//	@Param			limit	query		int							false	"Items per page (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Success		200		{object}	models.DocumentListResponse	"Successfully retrieved paginated documents with metadata"
//	@Failure		400		{object}	httpx.JSendFail				"Invalid pagination parameters"
//	@Failure		401		{object}	httpx.JSendError			"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		403		{object}	httpx.JSendError			"Forbidden - user does not have admin role"	example({"status": "error", "message": "Acceso denegado - se requiere rol de administrador", "code": 403})
//	@Failure		500		{object}	httpx.JSendError			"Internal server error - database query failed or connection error"
//	@Security		BearerAuth
//	@Router			/documents [get]
func (h *DocumentHandler) GetAllDocuments(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}

	documents, totalCount, err := h.documentService.GetAllDocuments(page, limit)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener documentos")
		return
	}

	// Return empty array if no documents found
	if documents == nil {
		documents = []*models.UserDocument{}
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrevious := page > 1

	// Build next/previous URLs
	baseURL := "/api/v1/documents"
	var nextURL, previousURL string
	if hasNext {
		nextURL = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page+1, limit)
	}
	if hasPrevious {
		previousURL = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page-1, limit)
	}

	// Build response
	response := map[string]any{
		"items": documents,
		"pagination": models.PaginationMetadata{
			CurrentPage: page,
			PerPage:     limit,
			TotalItems:  totalCount,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
			NextURL:     nextURL,
			PreviousURL: previousURL,
		},
	}

	httpx.RespondSuccess(w, http.StatusOK, response)
}

// UpdateDocumentByID godoc
//
//	@Summary		Update document status by ID (Admin)
//	@Description	**Admin-only endpoint** to update a document's review status by its document ID (not user ID). Similar to PATCH /documents/{user_id}/review but uses document ID for direct document manipulation. **Use case**: Admin dashboard where documents are listed with their IDs - admin can approve/reject directly using the document ID without needing to look up the user ID. Request body must contain boolean 'reviewed' field. **Access control**: Requires admin role JWT - non-admin users receive 403 Forbidden. The document_id path parameter must be valid UUID. Returns success message with updated review status. Returns 404 if document ID doesn't exist.
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Param			document_id	path		string								true									"Document ID in UUID format"						example(550e8400-e29b-41d4-a716-446655440000)
//	@Param			request		body		object{reviewed=bool}				true									"Review status - true to approve, false to reject"	example({"reviewed": true})
//	@Success		200			{object}	httpx.JSendSuccess					"Review status updated successfully"	example({"status": "success", "data": {"message": "Estado de revisión actualizado exitosamente", "reviewed": true}})
//	@Failure		400			{object}	httpx.JSendFailDocumentIDInvalid	"Invalid document ID format - must be valid UUID"
//	@Failure		400			{object}	httpx.JSendFailInvalidJSON			"Invalid request body - malformed JSON or missing reviewed field"
//	@Failure		401			{object}	httpx.JSendError					"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		403			{object}	httpx.JSendError					"Forbidden - user does not have admin role"			example({"status": "error", "message": "Acceso denegado - se requiere rol de administrador", "code": 403})
//	@Failure		404			{object}	httpx.JSendFail						"Document not found - document ID doesn't exist"	example({"status": "fail", "data": {"error": "Documento no encontrado"}})
//	@Failure		500			{object}	httpx.JSendError					"Internal server error - database update operation failed or connection error"
//	@Security		BearerAuth
//	@Router			/documents/{document_id} [patch]
func (h *DocumentHandler) UpdateDocumentByID(w http.ResponseWriter, r *http.Request) {
	// Get document ID from path parameter
	docIDParam := r.PathValue("document_id")
	if !validator.IsValidUUID(docIDParam) {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"document_id": "Formato de ID de documento inválido",
		})
		return
	}

	docID, parseErr := uuid.Parse(docIDParam)
	if parseErr != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"document_id": "Error al parsear ID de documento",
		})
		return
	}

	// Parse request body
	var req struct {
		Reviewed bool `json:"reviewed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de solicitud inválido",
		})
		return
	}

	err := h.documentService.UpdateDocumentByID(docID, req.Reviewed)
	if err != nil {
		if err.Error() == "documento no encontrado" {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"error": "Documento no encontrado",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al actualizar estado de revisión")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]any{
		"message":  "Estado de revisión actualizado exitosamente",
		"reviewed": req.Reviewed,
	})
}
