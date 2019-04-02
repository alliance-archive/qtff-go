package qtff

import (
	"encoding"
	"encoding/binary"
	"fmt"
)

var dataTypes = map[FourCC]func() encoding.BinaryUnmarshaler{}

type MediaHeaderData struct {
	TimeScale uint32
	Duration  uint32
}

func init() {
	dataTypes[FourCCFromString("mdhd")] = func() encoding.BinaryUnmarshaler { return &MediaHeaderData{} }
}

func (d *MediaHeaderData) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return fmt.Errorf("data too short")
	}
	d.TimeScale = binary.BigEndian.Uint32(b[12:])
	d.Duration = binary.BigEndian.Uint32(b[16:])
	return nil
}

type HandlerReferenceData struct {
	ComponentType    FourCC
	ComponentSubtype FourCC
}

func init() {
	dataTypes[FourCCFromString("hdlr")] = func() encoding.BinaryUnmarshaler { return &HandlerReferenceData{} }
}

func (d *HandlerReferenceData) UnmarshalBinary(b []byte) error {
	if len(b) < 24 {
		return fmt.Errorf("data too short")
	}
	d.ComponentType = FourCC(binary.BigEndian.Uint32(b[4:]))
	d.ComponentSubtype = FourCC(binary.BigEndian.Uint32(b[8:]))
	return nil
}

type ChunkOffset64Data struct {
	NumberOfEntries int
	Offsets         []uint64
}

func init() {
	dataTypes[FourCCFromString("co64")] = func() encoding.BinaryUnmarshaler { return &ChunkOffset64Data{} }
}

func (d *ChunkOffset64Data) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return fmt.Errorf("data too short")
	}
	d.NumberOfEntries = int(binary.BigEndian.Uint32(b[4:]))
	if len(b) < 8+d.NumberOfEntries*8 {
		return fmt.Errorf("data too short for entries")
	}
	d.Offsets = make([]uint64, d.NumberOfEntries)
	for i := 0; i < d.NumberOfEntries; i++ {
		d.Offsets[i] = binary.BigEndian.Uint64(b[8+i*8:])
	}
	return nil
}

type ChunkOffsetData struct {
	NumberOfEntries int
	Offsets         []uint32
}

func init() {
	dataTypes[FourCCFromString("stco")] = func() encoding.BinaryUnmarshaler { return &ChunkOffsetData{} }
}

func (d *ChunkOffsetData) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return fmt.Errorf("data too short")
	}
	d.NumberOfEntries = int(binary.BigEndian.Uint32(b[4:]))
	if len(b) < 8+d.NumberOfEntries*4 {
		return fmt.Errorf("data too short for entries")
	}
	d.Offsets = make([]uint32, d.NumberOfEntries)
	for i := 0; i < d.NumberOfEntries; i++ {
		d.Offsets[i] = binary.BigEndian.Uint32(b[8+i*4:])
	}
	return nil
}

type SampleSizeData struct {
	ConstantSampleSize int
	NumberOfEntries    int
	SampleSizes        []int
}

func init() {
	dataTypes[FourCCFromString("stsz")] = func() encoding.BinaryUnmarshaler { return &SampleSizeData{} }
}

func (d *SampleSizeData) SampleSize(n int) int {
	if d.ConstantSampleSize != 0 {
		return d.ConstantSampleSize
	}
	return d.SampleSizes[n-1]
}

func (d *SampleSizeData) UnmarshalBinary(b []byte) error {
	if len(b) < 12 {
		return fmt.Errorf("data too short")
	}
	d.ConstantSampleSize = int(binary.BigEndian.Uint32(b[4:]))
	if d.ConstantSampleSize == 0 {
		d.NumberOfEntries = int(binary.BigEndian.Uint32(b[8:]))
		if len(b) < 12+d.NumberOfEntries*4 {
			return fmt.Errorf("data too short for entries")
		}
		d.SampleSizes = make([]int, d.NumberOfEntries)
		for i := 0; i < d.NumberOfEntries; i++ {
			d.SampleSizes[i] = int(binary.BigEndian.Uint32(b[12+i*4:]))
		}
	}
	return nil
}

type SampleToChunkDataEntry struct {
	FirstChunk      int
	SamplesPerChunk int
}

type SampleToChunkData struct {
	NumberOfEntries int
	Entries         []SampleToChunkDataEntry
}

func init() {
	dataTypes[FourCCFromString("stsc")] = func() encoding.BinaryUnmarshaler { return &SampleToChunkData{} }
}

func (d *SampleToChunkData) ChunkFirstSample(n int) int {
	if len(d.Entries) == 1 {
		return (n-1)*d.Entries[0].SamplesPerChunk + 1
	}
	sampleOffset := 0
	for i := 1; i < len(d.Entries); i++ {
		e := d.Entries[i]
		prev := d.Entries[i-1]
		if e.FirstChunk >= n {
			return sampleOffset + (n-prev.FirstChunk)*prev.SamplesPerChunk + 1
		}
		sampleOffset += (e.FirstChunk - prev.FirstChunk) * prev.SamplesPerChunk
	}
	return sampleOffset + (n-d.Entries[len(d.Entries)-1].FirstChunk)*d.Entries[len(d.Entries)-1].SamplesPerChunk + 1
}

func (d *SampleToChunkData) SampleChunk(n int) int {
	if len(d.Entries) == 1 {
		return 1 + (n-1)/d.Entries[0].SamplesPerChunk
	}
	sampleOffset := 0
	for i := 1; i < len(d.Entries); i++ {
		e := d.Entries[i]
		prev := d.Entries[i-1]
		newSampleOffset := sampleOffset + (e.FirstChunk-prev.FirstChunk)*prev.SamplesPerChunk
		if newSampleOffset >= n {
			return prev.FirstChunk + (n-sampleOffset-1)/prev.SamplesPerChunk
		}
		sampleOffset = newSampleOffset
	}
	return d.Entries[len(d.Entries)-1].FirstChunk + (n-sampleOffset-1)/d.Entries[len(d.Entries)-1].SamplesPerChunk
}

func (d *SampleToChunkData) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return fmt.Errorf("data too short")
	}
	d.NumberOfEntries = int(binary.BigEndian.Uint32(b[4:]))
	d.Entries = make([]SampleToChunkDataEntry, d.NumberOfEntries)
	for i := range d.Entries {
		d.Entries[i].FirstChunk = int(binary.BigEndian.Uint32(b[8+i*12:]))
		d.Entries[i].SamplesPerChunk = int(binary.BigEndian.Uint32(b[8+i*12+4:]))
	}
	return nil
}
