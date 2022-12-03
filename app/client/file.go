package client

import (
	"fmt"
	"log"
	"net/rpc"
	"os"

	"github.com/eqr/transferit/app/service"
)

const batchSize = 5 * 1024 * 1024

func upload(filePath string, c *rpc.Client) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}

	buf := make([]byte, batchSize)
	batchNumber := 0

	initReq := &service.InitUploadRequest{}
	initResp := &service.InitUploadResponse{}
	err = c.Call("InitUpload", initReq, initResp)
	if err != nil {
		return fmt.Errorf("cannot init upload file %s: %w", filePath, err)
	}

	for {
		bytesRead, err := f.Read(buf)
		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}

		if bytesRead == 0 {
			log.Printf("reached end of file %s", filePath)
			break
		}

		log.Printf("sending batch %d of file %s", batchNumber, filePath)

		batchNumber++
	}

	return nil
}

func download(id service.TransferID) (string, error) {
	return "", nil
}
