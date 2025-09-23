package vknorm

import (
	"context"
	"encoding/json"
	"io"

	"vk-normalizer/internal/contract"
)

type Producer interface {
	Send(_ context.Context, topic string, data []byte) error
}

type Parser interface {
	ParseVKMessage(data []byte) (contract.NormalizedMessage, error)
}

type Downloader interface {
	DownloadFile(url string) (io.ReadSeeker, int64, error)
}

type Uploader interface {
	Upload(ctx context.Context, input contract.UploadFileRequest) (string, error)
}

type Option struct {
	KafkaTopic string
	Uploader   Uploader
	Parser     Parser
	Downloader Downloader
	Producer   Producer
}

type VkNormalizer struct {
	parser     Parser
	producer   Producer
	downloader Downloader
	uploader   Uploader
	kafkaTopic string
}

func NewVkNormalizer(opt Option) *VkNormalizer {
	return &VkNormalizer{
		parser:     opt.Parser,
		producer:   opt.Producer,
		downloader: opt.Downloader,
		uploader:   opt.Uploader,
		kafkaTopic: opt.KafkaTopic,
	}
}

func (t *VkNormalizer) Handle(ctx context.Context, raw []byte) error {
	msg, err := t.parser.ParseVKMessage(raw)
	if err != nil {
		return err
	}

	for i, media := range msg.Media {

		reader, size, err := t.downloader.DownloadFile(media.VkUrl)
		if err != nil {
			return err
		}

		s3URL, err := t.uploader.Upload(ctx, contract.UploadFileRequest{
			Reader:      reader,
			Size:        size,
			ContentType: media.MimeType,
		})
		if err != nil {
			return err
		}

		msg.Media[i].S3URL = s3URL
	}

	out, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return t.producer.Send(ctx, t.kafkaTopic, out)
}
