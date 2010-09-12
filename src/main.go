package main

import "minecraft/world"

import "net"
import "fmt"
import "os"

const MAX_CLIENTS = 64

func main() {
	//	runtime.GOMAXPROCS(64)
	servAddr := net.TCPAddr{Port: 25565}
	world, err := world.Open("/Users/roberthencke/Downloads/world/")
	if err != nil {
		fmt.Printf("can't open world; err=%s\n", err.String())
		os.Exit(1)

	}
	_, err = newServer(&servAddr, world)
	if err != nil {
		fmt.Printf("can't create server; err=%s\n", err.String())
		os.Exit(1)
	}
	fmt.Printf("Press enter to stop.\n")
	fmt.Scanln()
}
