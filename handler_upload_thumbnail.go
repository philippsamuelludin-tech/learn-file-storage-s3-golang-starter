package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

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

	file, mt, err := r.FormFile("thumbnail")

	defer file.Close()

	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get data from thumbnail", err)
		return
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get image Byte Data", err)
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

	imageDataAsString := base64.StdEncoding.EncodeToString(imageData)

	dataURL := fmt.Sprintf("data:%s;base64,%s", mt.Header.Get("Content-Type"), imageDataAsString)

	video.ThumbnailURL = &dataURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusNotModified, "Couödn't update Video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
