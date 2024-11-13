package file

import (
	"github.com/Superdanda/hade/framework"
	"os"
)

type HadeFileService struct {
	container framework.Container
}

func NewHadeFileService(container framework.Container) *HadeFileService {
	return &HadeFileService{container: container}
}

func (h HadeFileService) UploadFile(fileName string, file os.File) error {
	//TODO implement me
	panic("implement me")
}

func (h HadeFileService) DownloadFile(fileName string) (os.File, error) {
	//TODO implement me
	panic("implement me")
}

func (h HadeFileService) DeleteFile(fileName string) error {
	//TODO implement me
	panic("implement me")
}
