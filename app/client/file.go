package client

import (
	"encoding/base64"
	"fmt"
	"io"
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

	batchNumber := 0

	initReq := &service.InitUploadRequest{}
	initResp := &service.InitUploadResponse{}
	err = c.Call("Service.InitUpload", initReq, initResp)
	if err != nil {
		return fmt.Errorf("cannot init upload file %s: %w", filePath, err)
	}

	log.Println("Tranfser id: ", initResp.TransferID)

	for {
		buf := make([]byte, batchSize)
		_, err := f.Read(buf)

		if err == io.EOF {
			log.Printf("reached end of file %s", filePath)
			return nil
		}

		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}

		encoded := base64.StdEncoding.EncodeToString(buf)
		fmt.Println(encoded)

		log.Printf("sending batch %d of file %s", batchNumber, filePath)
		uploadReq := service.UploadChunkRequest{
			TransferID:  initResp.TransferID.String(),
			ChunkNumber: batchNumber,
			Content:     encoded,
		}

		uploadResp := &service.UploadChunkResponse{}

		err = c.Call("Service.UploadChunk", uploadReq, uploadResp)
		if err != nil {
			return fmt.Errorf("cannot upload chunk %d (%v): %w", batchNumber, initResp.TransferID, err)
		}

		batchNumber++
	}

	return nil
}

func download(id service.TransferID) (string, error) {
	return "", nil
}
