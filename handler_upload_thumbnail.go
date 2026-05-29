package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't parse Data", err)
		return
	}

	// 1. Rename the second return value to 'header'
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get thumbnail", err)
		return
	}
	defer file.Close()

	// 2. Get the raw string from the header
	contentType := header.Header.Get("Content-Type")

	// 3. Parse that string into a dedicated 'mediaType' string variable
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", err)
		return
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", nil)
		return
	}

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get data from thumbnail", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Didnt find video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User is not video owner", err)
		return
	}

	fileExtension := strings.Split(header.Header.Get("Content-Type"), "/")[1]

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		respondWithError(w, 401, "Couldn't created random key", err)
		return
	}

	fileName := base64.RawURLEncoding.EncodeToString(key)

	uniqueFilePath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s.%s", fileName, fileExtension))

	newFile, err := os.Create(uniqueFilePath)
	if err != nil {
		respondWithError(w, 401, "Couldn't create File", err)
		return
	}

	_, err = io.Copy(newFile, file)
	if err != nil {
		respondWithError(w, 401, "Couldn't copy File data", err)
		return
	}

	newThumbnail := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, fileName, fileExtension)

	video.ThumbnailURL = &newThumbnail

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusNotModified, "Couödn't update Video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
