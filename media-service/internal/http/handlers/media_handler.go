package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"team6-managing.mni.thm.de/Commz/media-service/internal/auth"
	"team6-managing.mni.thm.de/Commz/media-service/internal/media"
	"team6-managing.mni.thm.de/Commz/media-service/internal/utils"
)

var CodesToLog = map[int]bool{
	http.StatusBadRequest:          true,
	http.StatusUnauthorized:        true,
	http.StatusNotFound:            true,
	http.StatusInternalServerError: true,
}

type MediaHandler struct {
	mda  *media.MediaService
	auth *auth.AuthService
}

func (c *MediaHandler) RegisterRoutes(router *mux.Router) {
	HandleFunc(router, "/", c.UploadPicture, "POST")

	HandleFunc(router, "/version", c.getVersion, "GET")
	HandleFunc(router, "/{imageName}", c.GetPicture, "GET")

}

// @Summary Get the service Version
// @Success 200 {object} string "version"
// @Router /version [get]
func (c *MediaHandler) getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(utils.VERSION))
}

// @Summary Upload a image
// @Description Upload a image
// @Tags images
// @Accept image/jpeg
// @Produce json
// @Param commz-token header string true "Authenticated user JWT token"
// @Success 200 {object} string "Message Updated"
// @Failure 400 {object} utils.ServiceError "Invalid request body"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router / [post]
// @Security ApiKeyAuth
func (c *MediaHandler) UploadPicture(w http.ResponseWriter, r *http.Request) {
	img, err := io.ReadAll(r.Body)

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		c.handleErrors(fmt.Errorf("content type cannot be empty"), w)
		return
	}

	// get user id from context
	user := r.Context().Value("user-id")

	// upload image
	name, err := c.mda.UploadPicture(context.Background(), user.(string), contentType, img)
	if c.handleErrors(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)

	response := utils.UploadResponse{
		Name:    name,
		Success: true,
	}

	utils.SendJsonResponse(w, response)
}

// @Summary Get a image by imageName
// @Description Get a image by its Name
// @Tags images
// @Accept json
// @Produce image/jpeg
// @Param imageName path string true "Name of the image"
// @Success 200 {object} string "Image"
// @Failure 400 {object} utils.ServiceError "Invalid request body"
// @Failure 401 {object} utils.ServiceError "Unauthorized"
// @Failure 404 {object} utils.ServiceError "Image not found"
// @Failure 500 {object} utils.ServiceError "Internal server error"
// @Router /{imageName} [get]
func (c *MediaHandler) GetPicture(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pictureID := vars["imageName"]
	picture, err := c.mda.GetPicture(context.Background(), pictureID)
	if c.handleErrors(err, w) {
		return
	}
	utils.ServeImage(w, picture)
}

func (c *MediaHandler) handleErrors(err error, w http.ResponseWriter) bool {
	// check if error is a custom error
	customError, ok := err.(*utils.ServiceError)
	if ok {
		// marshal error into json and send status from error as statuscode
		w.WriteHeader(customError.StatusCode)
		w.Write(customError.Bytes())
		return true
	}

	if err != nil {
		customError = &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Err:        "Internal server error",
			Message:    err.Error(),
		}
		utils.SendJsonError(w, customError)
		return true
	}
	return false
}

func (c *MediaHandler) error(w http.ResponseWriter, error string, code int) {
	err := utils.NewError(error, code)
	if CodesToLog[code] {
		logger.Error().Err(err)
	}
	c.handleErrors(err, w)
}
