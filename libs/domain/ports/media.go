package ports

import (
	"context"
	"io"
	"time"

	"whisper/libs/domain/entity"
	"whisper/libs/domain/valueobject"
)

type UploadFileCommand struct {
	MessageID valueobject.MessageID
	Filename  string
	MIME      string
	Size      int64
	Reader    io.Reader
}

type MediaUseCase interface {
	Upload(context.Context, UploadFileCommand) (entity.Attachment, error)
	GetPresignedURL(context.Context, string, time.Duration) (string, error)
}

type FileStorage interface {
	Upload(context.Context, string, io.Reader, int64, string) error
	PresignGetObject(context.Context, string, time.Duration) (string, error)
}

type AttachmentRepository interface {
	Create(context.Context, entity.Attachment) (entity.Attachment, error)
	GetByID(context.Context, valueobject.AttachmentID) (entity.Attachment, error)
}
