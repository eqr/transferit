package service

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
)

type TransferID = uuid.UUID

type InitUploadRequest struct {
	NumOfChunks int
}

type InitUploadResponse struct {
	TransferID
}

type CurrentSegment struct {
	Number     int
	LastNumber int
	Data       string // base64 encoded file segment
}

const nullCurrentSegmentID = -1

func (s *Service) setNullCurrentSegment(transferID TransferID) error {
	segment, ok := s.data[transferID]
	if !ok {
		return fmt.Errorf("transfer not running: %v", transferID)
	}

	segment.LastNumber = segment.Number
	segment.Number = nullCurrentSegmentID
	segment.Data = ""

	s.data[transferID] = segment
	return nil
}

func New() *Service {
	data := make(map[TransferID]CurrentSegment)
	lock := &sync.RWMutex{}

	return &Service{
		data: data,
		lock: lock,
	}
}

type Service struct {
	data map[TransferID]CurrentSegment
	lock *sync.RWMutex
}

func (s *Service) InitUpload(request *InitUploadRequest, response *InitUploadResponse) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	id := uuid.New()
	response.TransferID = id
	return nil
}

type UploadChunkRequest struct {
	TransferID  string
	ChunkNumber int
	Content     string // base64 segment content
}

type UploadChunkResponse struct {
}

func (s *Service) UploadChunk(request *UploadChunkRequest, _ *UploadChunkResponse) error {
	if len(request.Content) > 3.5*1024*1024 {
		return fmt.Errorf("refejcted, chunk %d is too big (%d)", request.ChunkNumber, len(request.Content))
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	trID, err := uuid.Parse(request.TransferID)
	if err != nil {
		return fmt.Errorf("cannot parse transfer id %s: %w", request.TransferID, err)
	}

	segment, ok := s.data[trID]
	if ok {
		return fmt.Errorf("transfer id was not found: %q", request)
	}

	if segment.Number != nullCurrentSegmentID {
		return fmt.Errorf("segment was not yet downloaded")
	}

	s.data[trID] = CurrentSegment{Number: request.ChunkNumber, Data: request.Content}
	return nil
}

type DownloadChunkRequest struct {
	TransferID
	ChunkNumber int
}

type DownloadChunkResponse struct {
	TransferID
	ChunkNumber int
	Data        string // base64 encoded file segment
}

func (s *Service) DownloadChunk(request *DownloadChunkRequest, response *DownloadChunkResponse) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	segment, ok := s.data[request.TransferID]
	if !ok {
		return fmt.Errorf("cannot find tranfer with id %v", request.TransferID)
	}

	if segment.Number == nullCurrentSegmentID {
		return fmt.Errorf("the segment %q was not uploaded yet", request)
	}

	if segment.Number != request.ChunkNumber {
		return fmt.Errorf("the segment with id %d is not available", segment.Number)
	}

	response.ChunkNumber = segment.Number
	response.Data = segment.Data
	return nil
}

type ConfirmChunkDownloadedRequest struct {
	TransferID
	ChunkNumber int
}

type ConfirmChunkDownloadedResponse struct {
}

func (s *Service) ConfirmChunkDownloaded(request *ConfirmChunkDownloadedRequest, _ *ConfirmChunkDownloadedResponse) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.data[request.TransferID]
	if !ok {
		log.Printf("Did not find the segment %q on ConfirmDownload", request)
		return nil
	}

	if err := s.setNullCurrentSegment(request.TransferID); err != nil {
		return fmt.Errorf("cannot set null current segment: %w", err)
	}

	return nil
}

type GetCurrentSegmentNumberRequest struct {
	TransferID
}

type GetCurrentSegmentNumberResponse struct {
	ChunkNumber int
}

func (s *Service) GetCurrentSegmentNumber(request *GetCurrentSegmentNumberRequest, response *GetCurrentSegmentNumberResponse) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	segment, ok := s.data[request.TransferID]
	if !ok {
		return fmt.Errorf("transfer with id %q does not exist", request.TransferID)
	}

	if segment.Number == nullCurrentSegmentID {
		return fmt.Errorf("transfer is not initialized yet")
	}

	response.ChunkNumber = segment.Number
	return nil
}
