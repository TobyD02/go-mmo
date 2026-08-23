package util

import "github.com/tobyd02/go-mmo/pkg/config"

type GSpatialIndex struct {
	positionIndex map[Vec2]map[string]struct{}
	chunkIndex    map[Vec2]map[string]struct{}
	chunkSize     int
}

func NewGSpatialIndex() GSpatialIndex {
	return GSpatialIndex{
		positionIndex: make(map[Vec2]map[string]struct{}),
		chunkIndex:    make(map[Vec2]map[string]struct{}),
		chunkSize:     config.ChunkSize,
	}
}

func (s GSpatialIndex) calculateChunk(pos Vec2) Vec2 {
	return Vec2{
		X: floorDiv(pos.X, s.chunkSize),
		Y: floorDiv(pos.Y, s.chunkSize),
	}
}

func floorDiv(value, divisor int) int {
	result := value / divisor
	remainder := value % divisor

	if remainder != 0 && value < 0 {
		result--
	}

	return result
}

func (s GSpatialIndex) addPositionIndex(id string, pos Vec2) {
	if s.positionIndex[pos] == nil {
		s.positionIndex[pos] = make(map[string]struct{})
	}

	s.positionIndex[pos][id] = struct{}{}
}

func (s GSpatialIndex) removePositionIndex(id string, pos Vec2) {
	if s.positionIndex[pos] == nil {
		return
	}

	delete(s.positionIndex[pos], id)

	if len(s.positionIndex[pos]) == 0 {
		delete(s.positionIndex, pos)
	}
}

func (s GSpatialIndex) addChunkIndex(id string, pos Vec2) {
	chunk := s.calculateChunk(pos)

	if s.chunkIndex[chunk] == nil {
		s.chunkIndex[chunk] = make(map[string]struct{})
	}

	s.chunkIndex[chunk][id] = struct{}{}
}

func (s GSpatialIndex) removeChunkIndex(id string, pos Vec2) {
	chunk := s.calculateChunk(pos)

	if s.chunkIndex[chunk] == nil {
		return
	}

	delete(s.chunkIndex[chunk], id)

	if len(s.chunkIndex[chunk]) == 0 {
		delete(s.chunkIndex, chunk)
	}
}

func (s GSpatialIndex) Add(id string, pos Vec2) {
	s.addPositionIndex(id, pos)
	s.addChunkIndex(id, pos)
}

func (s GSpatialIndex) Remove(id string, pos Vec2) {
	s.removePositionIndex(id, pos)
	s.removeChunkIndex(id, pos)
}

func (s GSpatialIndex) Update(id string, oldPos Vec2, newPos Vec2) {
	if oldPos.Equal(newPos) {
		return
	}

	// Position always changes.
	s.removePositionIndex(id, oldPos)
	s.addPositionIndex(id, newPos)

	oldChunk := s.calculateChunk(oldPos)
	newChunk := s.calculateChunk(newPos)

	// Only update the chunk index when crossing
	// into a different chunk.
	if oldChunk.Equal(newChunk) {
		return
	}

	s.removeChunkIndex(id, oldPos)
	s.addChunkIndex(id, newPos)
}

func (s GSpatialIndex) QueryPos(x, y int) map[string]struct{} {
	return s.positionIndex[Vec2{X: x, Y: y}]
}

func (s GSpatialIndex) QueryChunk(x, y int) map[string]struct{} {
	chunk := s.calculateChunk(Vec2{X: x, Y: y})
	return s.chunkIndex[chunk]
}

// QueryRange - queries within a range originating from pos
func (s GSpatialIndex) QueryRange(
	pos Vec2,
	rangeX int,
	rangeY int,
) map[string]struct{} {
	minPos := Vec2{
		X: pos.X - rangeX,
		Y: pos.Y - rangeY,
	}

	maxPos := Vec2{
		X: pos.X + rangeX,
		Y: pos.Y + rangeY,
	}

	minChunk := s.calculateChunk(minPos)
	maxChunk := s.calculateChunk(maxPos)

	ids := make(map[string]struct{})

	for chunkX := minChunk.X; chunkX <= maxChunk.X; chunkX++ {
		for chunkY := minChunk.Y; chunkY <= maxChunk.Y; chunkY++ {
			chunk := Vec2{
				X: chunkX,
				Y: chunkY,
			}

			for id := range s.chunkIndex[chunk] {
				ids[id] = struct{}{}
			}
		}
	}

	return ids
}
