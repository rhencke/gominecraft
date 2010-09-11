package world

import "testing"

func TestWorld(t *testing.T) {
	w, err := Open("/Users/roberthencke/Downloads/world/")
	if err != nil {
		t.Error(err)
	}
	for i := int32(-128); i < 128; i++ {
		for j := int32(-128); j < 128; j++ {
			w.LoadChunk(i, j)
		}
	}
	err = w.Close()
	if err != nil {
		t.Error(err)
	}

}
