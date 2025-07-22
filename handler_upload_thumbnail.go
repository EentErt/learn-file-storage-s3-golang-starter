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

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form", err)
		return
	}

	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get file", err)
		return
	}
	defer file.Close()

	mediaType := fileHeader.Header.Get("Content-Type")
	mimeType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unrecognized file type", err)
		return
	}

	if mimeType != "image/jpeg" && mimeType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Cannot use file type as a thumbnail", err)
		return
	}

	/*
		imageData, err := io.ReadAll(file)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Unable to read file", err)
			return
		}
	*/

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	fileType := strings.Split(mediaType, "/")[1]
	fileNameBytes := make([]byte, 32)
	rand.Read(fileNameBytes)
	fileName := fmt.Sprint(base64.RawURLEncoding.EncodeToString(fileNameBytes), ".", fileType)

	path := filepath.Join(cfg.assetsRoot, "/", fileName)

	newFile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create file", err)
		return
	}

	io.Copy(newFile, file)
	/*
		thumb := base64.StdEncoding.EncodeToString(imageData)

		dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, thumb)
	*/

	/*
		thumb := thumbnail{
			data:      imageData,
			mediaType: mediaType,
		}

		videoThumbnails[videoID] = thumb

		thumbURL := fmt.Sprintf("http://localhost:8091/api/thumbnails/%s", videoID.String())
		video.ThumbnailURL = &thumbURL
	*/

	thumbURL := fmt.Sprintf("http://localhost:8091/assets/%s", fileName)
	video.ThumbnailURL = &thumbURL

	cfg.db.UpdateVideo(video)

	respondWithJSON(w, http.StatusOK, video)
}
