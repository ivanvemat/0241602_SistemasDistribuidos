package index

import (
	"encoding/binary"
	"io"
	"os"

	. "server"

	"github.com/tysonmote/gommap"
)

var (
	enc             = binary.BigEndian
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth uint64 = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (idx *index, err error) {
	fileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	f.Truncate(int64(c.Segment.MaxIndexBytes))
	mmap, err := gommap.Map(f.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &index{
		file: f,
		size: uint64(fileInfo.Size()),
		mmap: mmap,
	}, nil
}

func (idx *index) Name() string {
	return idx.file.Name()
}

func (idx *index) Read(in int64) (off uint32, pos uint64, err error) {
	if idx.size == 0 {
		return 0, 0, io.EOF
	}

	if in == -1 {
		in = int64(idx.size/entWidth) - 1
	}

	posInFile := uint64(in) * entWidth

	if posInFile >= idx.size {
		return 0, 0, io.EOF
	}

	off = enc.Uint32(idx.mmap[posInFile : posInFile+offWidth])
	pos = enc.Uint64(idx.mmap[posInFile+offWidth : posInFile+entWidth])

	return off, pos, nil
}

func (idx *index) Write(off uint32, pos uint64) error {
	if uint64(len(idx.mmap))-idx.size < entWidth {
		return io.EOF
	}

	enc.PutUint32(idx.mmap[idx.size:idx.size+offWidth], off)
	enc.PutUint64(idx.mmap[idx.size+offWidth:idx.size+entWidth], pos)
	idx.size += entWidth

	return nil
}

func (idx *index) Close() error {
	if err := idx.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}

	if err := idx.mmap.UnsafeUnmap(); err != nil {
		return err
	}

	if err := idx.file.Truncate(int64(idx.size)); err != nil {
		return err
	}

	if err := idx.file.Close(); err != nil {
		return err
	}

	return nil
}
