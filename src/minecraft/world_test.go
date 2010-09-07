package world

import "testing"
// import "os"
// import "fmt"
func TestWorld(t *testing.T) {
	_, err := Open("/Users/roberthencke/Downloads/world/")
	if err!=nil{
		t.Error(err)
	}
}
