package util

type GSpatialIndex map[Vec2]map[string]struct{}

func (s GSpatialIndex) Add(id string, pos Vec2) {
	if s[pos] == nil {
		s[pos] = make(map[string]struct{})
	}

	s[pos][id] = struct{}{}
}

func (s GSpatialIndex) Remove(id string, pos Vec2) {
	// Cannot delete if it doesnt exist
	if s[pos] == nil {
		return
	}

	// Run delete
	delete(s[pos], id)

	if len(s[pos]) == 0 {
		delete(s, pos)
	}
}

func (s GSpatialIndex) Update(id string, oldPos Vec2, newPos Vec2) {
	if oldPos.Equal(newPos) {
		return
	}

	s.Remove(id, oldPos)
	s.Add(id, newPos)
}

func (s GSpatialIndex) QueryPos(x, y int) map[string]struct{} {
	return s[Vec2{X: x, Y: y}]
}

func (s GSpatialIndex) QueryPosRange(minX, minY, maxX, maxY int) map[string]struct{} {
	ids := make(map[string]struct{})
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for id := range s.QueryPos(x, y) {
				ids[id] = struct{}{}
			}
		}
	}

	return ids
}
