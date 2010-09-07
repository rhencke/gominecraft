package world

// import "minecraft/nbt"
import "minecraft/error"
import "os"
import "io/ioutil"
import "syscall"

type World struct {
	dir *os.FileInfo
}

func Open(worlddir *os.FileInfo) (w *World, err os.Error) {
	w = &World{dir: worlddir}
	if err = w.verifyFormat(); err != nil {
		return
	}
	if err = w.lock(); err != nil {
		err = error.NewError("unable to obtain exclusive lock on world", err)
		return
	}
	return
}

// func (world *World) Close() {
// 	return
// }


func (world *World) verifyFormat() os.Error {
	// We don't want to go crazy vetting every byte, but we can at least do a sanity check
	// for how the folder structure should look.  It is important we don't touch any files,
	// so if this world is in use by another process, things don't go terribly wrong.
	if (!world.dir.IsDirectory()){
		return error.NewError("expected a directory, didn't get one")
	}
	#err you were here
}

func (world *World) lock() os.Error {

}

func (world *World) unlock() {

}
