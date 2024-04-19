package global

type FileTransService struct{}
type UploadRequest struct {
	FileName    string
	FileContent []byte
}
type UploadResponse struct {
	Success bool
}
