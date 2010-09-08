package world

import "minecraft/nbt"
import "minecraft/error"

import "fmt"
import "io/ioutil"
import "os"
import "path"

const (
	leveldat    = "level.dat"
	sessionlock = "session.lock"
)

type World struct {
	dir      string
	lockmsec int64
	leveldat map[string]interface{}
	lockfd   *os.File
}

func Open(worlddir string) (w *World, err os.Error) {
	w = &World{dir: worlddir}
	if err = w.verifyFormat(); err != nil {
		err = error.NewError("could not verify world format", err)
		return
	}
	if err = w.lock(); err != nil {
		err = error.NewError("unable to obtain lock on world", err)
		return
	}
	if _, w.leveldat, err = nbt.Load(path.Join(w.dir, leveldat)); err != nil {
		err = error.NewError("could not read level", err)
		return
	}
	return
}

func (world *World) Close() os.Error {
	return world.unlock()
}


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
		err = error.NewError("could not read world directory contents", nil)
		return
	}

	for _, f := range files {
		if f.IsRegular() {
			switch f.Name {
			case leveldat:
				hasLevelDat = true
			case sessionlock:
				hasSessionLock = true
			}
		}
	}

	if !hasLevelDat {
		err = error.NewError(fmt.Sprint("world is missing ", leveldat), nil)
		return
	}
	if !hasSessionLock {
		err = error.NewError(fmt.Sprint("world is missing ", sessionlock), nil)
		return
	}
	return
}

func (world *World) lock() (err os.Error) {
	if world.lockfd != nil {
		panic("lock fd already exists... should never happen")
	}
	sessionLockPath := path.Join(world.dir, sessionlock)
	world.lockfd, err = os.Open(sessionLockPath, os.O_RDWR|os.O_ASYNC, 0000)
	if err != nil {
		error.NewError(fmt.Sprint("could not open ", sessionlock), nil)
	}
	// minecraft's locking mechanism is peculiar.
	// It writes the current system time in milliseconds since 1970 to the file.
	// It then watches the file for changes.  If a change is monitored, it aborts.

	// This has strange implications, such as the LAST process to open the world owns it,
	// not the first.

	// but hey, when in rome...
	sec, nsec, err := os.Time()
	if err != nil {
		err = error.NewError("couldn't get the current time..?!", err)
		return
	}

	world.lockmsec = (sec * 1000) + (nsec / 1000000)
	err = nbt.WriteInt64(world.lockfd, world.lockmsec)
	if err != nil {
		err = error.NewError("could not write timestamp to session lock", err)
		return
	}
	return
}

func (world *World) verifyLock() (err os.Error) {
	_, err = world.lockfd.Seek(0, 0)
	if err != nil {
		err = error.NewError("could not seek to beginning of session lock", err)
		return
	}
	msec, err := nbt.ReadInt64(world.lockfd)
	if err != nil {
		err = error.NewError("could not read timestamp from session lock", err)
		return
	}
	if msec != world.lockmsec {
		err = error.NewError("someone else has opened this world :(", nil)
		return
	}
	return
}

func (world *World) unlock() os.Error {
	return world.lockfd.Close()
}
