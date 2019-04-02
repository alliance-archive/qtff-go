package qtff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSampleToChunkData_SampleChunk(t *testing.T) {
	data := &SampleToChunkData{
		NumberOfEntries: 3,
		Entries: []SampleToChunkDataEntry{
			{1, 3},
			{3, 1},
			{5, 1},
		},
	}

	assert.Equal(t, 1, data.SampleChunk(1))
	assert.Equal(t, 1, data.SampleChunk(2))
	assert.Equal(t, 1, data.SampleChunk(3))
	assert.Equal(t, 2, data.SampleChunk(4))
	assert.Equal(t, 2, data.SampleChunk(5))
	assert.Equal(t, 2, data.SampleChunk(6))
	assert.Equal(t, 3, data.SampleChunk(7))
	assert.Equal(t, 4, data.SampleChunk(8))
	assert.Equal(t, 5, data.SampleChunk(9))
	assert.Equal(t, 6, data.SampleChunk(10))
}

func TestSampleToChunkData_ChunkFirstSample(t *testing.T) {
	data := &SampleToChunkData{
		NumberOfEntries: 3,
		Entries: []SampleToChunkDataEntry{
			{1, 3},
			{3, 1},
			{5, 1},
		},
	}

	assert.Equal(t, 1, data.ChunkFirstSample(1))
	assert.Equal(t, 4, data.ChunkFirstSample(2))
	assert.Equal(t, 7, data.ChunkFirstSample(3))
	assert.Equal(t, 8, data.ChunkFirstSample(4))
	assert.Equal(t, 9, data.ChunkFirstSample(5))
	assert.Equal(t, 10, data.ChunkFirstSample(6))
}
