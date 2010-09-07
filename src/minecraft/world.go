package world

// import "minecraft/nbt"
import "minecraft/error"
import "os"
import "io/ioutil"
// import "syscall"
// import "path"
// import "fmt"

type World struct {
	dir string
}

func Open(worlddir string) (w *World, err os.Error) {
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


func (world *World) verifyFormat() (err os.Error) {
	// We don't want to go crazy vetting every byte, but we can at least do a sanity check
	// for how the folder structure should look.  It is important we don't touch any files,
	// so if this world is in use by another process, things don't go terribly wrong.
	fi, err := os.Stat(world.dir)
	if err != nil {
		err = error.NewError("could not stat world directory", err)
		return
	}

	if !fi.IsDirectory() {
		return error.NewError("expected a directory, didn't get one", nil)
	}
	var hasLevelDat, hasSessionLock bool

	files, err := ioutil.ReadDir(world.dir)
	if err != nil {
		err = error.NewError("could not read world directory contents",nil)
		return
	}

	for _, f := range files {
		if f.IsRegular() {
			switch f.Name {
			case "level.dat":
				hasLevelDat = true
			case "session.lock":
				hasSessionLock = true
			}
		}
	}

	if !hasLevelDat {
		err = error.NewError("world is missing level.dat",nil)
		return
	}
	if !hasSessionLock {
		err = error.NewError("world is missing session.lock",nil)
		return
	}
	return
}

func (world *World) lock() (err os.Error) {
	return
}

func (world *World) unlock() {

}
