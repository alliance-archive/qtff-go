package qtff

import (
	"encoding/binary"
	"io"
	"io/ioutil"
)

const (
	AtomTypeMOOV FourCC = 0x6d6f6f76
	AtomTypeMDAT FourCC = 0x6d646174
)

type Atom struct {
	Type FourCC
	Size int64

	Data *io.SectionReader
}

func (a *Atom) ParseData() (interface{}, error) {
	if factory, ok := dataTypes[a.Type]; !ok {
		return nil, nil
	} else if b, err := ioutil.ReadAll(a.Data); err != nil {
		return nil, err
	} else {
		ret := factory()
		if err := ret.UnmarshalBinary(b); err != nil {
			return nil, err
		}
		return ret, nil
	}
}

type AtomReader struct {
	r      io.ReaderAt
	offset int64
	err    error
}

func (r *AtomReader) Next() *Atom {
	if r.err != nil {
		return nil
	}

	var header [8]byte
	if n, err := r.r.ReadAt(header[:], r.offset); n == 0 && err == io.EOF {
		return nil
	} else if n < len(header) {
		r.err = err
		return nil
	}
	atom := &Atom{
		Type: FourCC(binary.BigEndian.Uint32(header[4:])),
		Size: int64(binary.BigEndian.Uint32(header[:])),
	}
	if atom.Size == 1 {
		var extendedSize [8]byte
		if n, err := r.r.ReadAt(extendedSize[:], r.offset+int64(len(header))); n < len(extendedSize) {
			r.err = err
			return nil
		}
		atom.Size = int64(binary.BigEndian.Uint64(extendedSize[:]))
		atom.Data = io.NewSectionReader(r.r, r.offset+16, atom.Size-16)
	} else {
		atom.Data = io.NewSectionReader(r.r, r.offset+8, atom.Size-8)
	}
	r.offset += int64(atom.Size)
	return atom
}

func (r *AtomReader) Error() error {
	return r.err
}

func NewAtomReader(r io.ReaderAt) *AtomReader {
	return &AtomReader{
		r: r,
	}
}
