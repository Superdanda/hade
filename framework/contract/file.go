package contract

import "os"

type FileService interface {
	UploadFile(fileName string, file os.File) error
	DownloadFile(fileName string) (os.File, error)
	DeleteFile(fileName string) error
}
