package world

import "testing"

func TestWorld(t *testing.T) {
	w, err := Open("/Users/roberthencke/Downloads/world/")
	if err != nil {
		t.Error(err)
	}
	err = w.Close()
	if err != nil {
		t.Error(err)
	}
}
