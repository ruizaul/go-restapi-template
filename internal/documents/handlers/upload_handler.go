package handlers

import (
	"net/http"
	"strings"

	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"
	"tacoshare-delivery-api/pkg/storage"

	"github.com/google/uuid"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
)

// DocumentType represents the type of document being uploaded
type DocumentType string

const (
	// DocumentTypeCirculationCard represents a vehicle circulation card document
	DocumentTypeCirculationCard DocumentType = "circulation_card"
	// DocumentTypeINEFront represents the front of an INE (Mexican ID) document
	DocumentTypeINEFront DocumentType = "ine_front"
	// DocumentTypeINEBack represents the back of an INE (Mexican ID) document
	DocumentTypeINEBack DocumentType = "ine_back"
	// DocumentTypeDriverLicenseFront represents the front of a driver's license
	DocumentTypeDriverLicenseFront DocumentType = "driver_license_front"
	// DocumentTypeDriverLicenseBack represents the back of a driver's license
	DocumentTypeDriverLicenseBack DocumentType = "driver_license_back"
	// DocumentTypeProfilePhoto represents a profile photo document
	DocumentTypeProfilePhoto DocumentType = "profile_photo"
	// DocumentTypeFiscalCertificate represents a fiscal certificate document
	DocumentTypeFiscalCertificate DocumentType = "fiscal_certificate"
)

// UploadHandler handles document upload HTTP requests
type UploadHandler struct {
	r2Client *storage.R2Client
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(r2Client *storage.R2Client) *UploadHandler {
	return &UploadHandler{r2Client: r2Client}
}

// UploadDocument godoc
//
//	@Summary		Upload document file
//	@Description	Upload a single document file to Cloudflare R2 storage using multipart/form-data. **Maximum file size: 10 MB**. **Supported formats: JPEG, PNG, PDF**. The file is stored in R2 with path structure: documents/{user_id}/{doc_type}/{uuid}.{ext}. Returns the public URL which should be saved via PATCH /documents/me. **Workflow**: 1) Upload file here to get URL, 2) Call PATCH /documents/me with {"ine_front_url": "returned_url"} to save in database. Each document type corresponds to a specific field in the document record (circulation_card → circulation_card_url, ine_front → ine_front_url, etc.).
//	@Tags			documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file														true																		"Document file - max 10 MB, formats: JPEG/PNG/PDF"
//	@Param			type	formData	string														true																		"Document type - determines storage path and field name"	Enums(circulation_card, ine_front, ine_back, driver_license_front, driver_license_back, profile_photo, fiscal_certificate)
//	@Success		200		{object}	object{status=string,data=object{url=string,type=string}}	"File uploaded successfully to R2 - returns public URL and document type"	example({"status": "success", "data": {"url": "https://pub-abc123.r2.dev/documents/550e8400-e29b-41d4-a716-446655440000/ine_front/a1b2c3d4.jpg", "type": "ine_front"}})
//	@Failure		400		{object}	httpx.JSendFailDocumentType									"Invalid document type - valores permitidos: circulation_card, ine_front, ine_back, driver_license_front, driver_license_back, profile_photo, fiscal_certificate"
//	@Failure		400		{object}	httpx.JSendFailFileSize										"File too large - máximo 10 MB permitido"
//	@Failure		400		{object}	httpx.JSendFailFileType										"Unsupported file type - solo se permiten imágenes JPEG/PNG y archivos PDF"
//	@Failure		401		{object}	httpx.JSendError											"Unauthorized - missing or invalid JWT token in Authorization header"
//	@Failure		500		{object}	httpx.JSendError											"Internal server error - failed to upload file to R2 storage or S3 API error"
//	@Security		BearerAuth
//	@Router			/documents/upload [post]
func (h *UploadHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Check if R2 client is available
	if h.r2Client == nil {
		httpx.RespondError(w, http.StatusServiceUnavailable, "Servicio de almacenamiento no disponible. Por favor contacte al administrador.", 0)
		return
	}

	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario no encontrado en el contexto")
		return
	}

	// Parse multipart form (max 10 MB)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"file": "Error al procesar archivo - tamaño máximo 10 MB permitido",
		})
		return
	}

	// Get document type from form data
	docTypeParam := r.FormValue("type")
	if docTypeParam == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"type": "Tipo de documento no proporcionado",
		})
		return
	}
	docType := DocumentType(docTypeParam)

	// Validate document type
	validTypes := map[DocumentType]bool{
		DocumentTypeCirculationCard:    true,
		DocumentTypeINEFront:           true,
		DocumentTypeINEBack:            true,
		DocumentTypeDriverLicenseFront: true,
		DocumentTypeDriverLicenseBack:  true,
		DocumentTypeProfilePhoto:       true,
		DocumentTypeFiscalCertificate:  true,
	}

	if !validTypes[docType] {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"type": "Tipo de documento inválido - valores permitidos: circulation_card, ine_front, ine_back, driver_license_front, driver_license_back, profile_photo, fiscal_certificate",
		})
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"file": "Archivo no proporcionado o inválido",
		})
		return
	}
	defer func() {
		_ = file.Close()
	}()

	// Validate file size
	if header.Size > maxUploadSize {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"file": "Archivo demasiado grande - máximo 10 MB permitido",
		})
		return
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowedTypes[contentType] {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"file": "Tipo de archivo no permitido - solo se permiten imágenes JPEG/PNG y archivos PDF",
		})
		return
	}

	// Construct folder path: documents/{user_id}/{doc_type}/
	folder := strings.Join([]string{"documents", userID.String(), string(docType)}, "/")

	// Upload to R2
	fileURL, err := h.r2Client.UploadFile(r.Context(), file, header, folder)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al subir archivo")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]string{
		"url":  fileURL,
		"type": string(docType),
	})
}
