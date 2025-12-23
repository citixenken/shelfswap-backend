package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Service interface {
	Upload(file multipart.File, header *multipart.FileHeader) (string, error)
}

type SupabaseStorage struct {
	ProjectURL string
	SecretKey  string
	Bucket     string
}

func NewSupabaseStorage(projectURL, secretKey, bucket string) *SupabaseStorage {
	return &SupabaseStorage{
		ProjectURL: projectURL,
		SecretKey:  secretKey,
		Bucket:     bucket,
	}
}

func (s *SupabaseStorage) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// Read file content
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return "", err
	}

	// Create request to Supabase Storage API
	// POST /storage/v1/object/{bucket}/{path}
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.ProjectURL, s.Bucket, filename)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.SecretKey)
	req.Header.Set("Content-Type", header.Header.Get("Content-Type"))

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload image: %s", string(body))
	}

	// Return public URL
	// GET /storage/v1/object/public/{bucket}/{path}
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.ProjectURL, s.Bucket, filename)
	return publicURL, nil
}

// LocalStorage fallback for development if needed (optional, but good practice)
type LocalStorage struct {
	UploadDir string
}

func NewLocalStorage(uploadDir string) *LocalStorage {
	return &LocalStorage{UploadDir: uploadDir}
}

func (s *LocalStorage) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// Ensure upload directory exists
	if _, err := os.Stat(s.UploadDir); os.IsNotExist(err) {
		os.MkdirAll(s.UploadDir, 0755)
	}

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filepath := filepath.Join(s.UploadDir, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}
