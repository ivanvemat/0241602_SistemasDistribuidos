package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var enc = binary.BigEndian

const lenWidth = 8
const storeSize = 64

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(file *os.File) (*store, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := uint64(fileInfo.Size())

	return &store{
		File: file,
		size: size,
		buf:  bufio.NewWriter(file),
	}, nil
}

func (s *store) Append(data []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size
	// Escribir tamaño de "data"
	if err := binary.Write(s.buf, enc, uint64(len(data))); err != nil {
		return 0, 0, err
	}

	bytesWritten, err := s.buf.Write(data)
	if err != nil {
		return 0, 0, err
	}

	bytesWritten += lenWidth
	s.size += uint64(bytesWritten)
	return uint64(bytesWritten), pos, nil
}

func (s *store) Read(pos uint64) (read []byte, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	if _, err = s.File.Seek(int64(pos), 0); err != nil {
		return nil, err
	}

	var length uint64
	// Recuperando tamaño del siguiente en "length"
	if err := binary.Read(s.File, enc, &length); err != nil {
		return nil, err
	}

	p := make([]byte, length)
	if _, err := s.File.Read(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return err
	}
	if err := s.File.Close(); err != nil {
		return err
	}

	return nil
}
