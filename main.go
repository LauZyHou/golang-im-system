package main

func main() {
	// use "nc 127.0.0.1 8888" to test
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}