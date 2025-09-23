package vk

import (
	"bytes"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
)

type Downloader struct {
	Api *resty.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		Api: resty.New(),
	}
}

func (d *Downloader) DownloadFile(url string) (io.ReadSeeker, int64, error) {
	resp, err := d.Api.R().Get(url)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode() != 200 {
		return nil, 0, fmt.Errorf("bad status: %d", resp.StatusCode())
	}

	data := resp.Body()
	reader := bytes.NewReader(data)

	return reader, int64(len(data)), nil
}
