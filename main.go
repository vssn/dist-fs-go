package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/vssn/dist-fs-go/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", "")
	s3 := makeServer(":5000", ":3000", ":4000")

	go func() {
		log.Fatal(s1.Start())
	}()

	go func() {
		log.Fatal(s2.Start())
	}()

	time.Sleep(4 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)

	/* 	for i := 0; i < 20; i++ {

		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		s3.Store(key, data)

		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	} */

	readFilenames(s3)

}

func readFilenames(s3 *FileServer) {
	fmt.Println("One filename per line:")
	scanner := bufio.NewScanner(os.Stdin)

	var lines []string
	for {
		scanner.Scan()
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		lines = append(lines, line)
	}

	err := scanner.Err()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("output:")
	for _, l := range lines {
		fmt.Printf("Attempt to read %s:\n", l)
		b, err := readFile(l)

		if err != nil {
			log.Fatal(err)
		}

		data := bytes.NewReader(b)

		s3.Store(l, data)

		fmt.Println("File stored.")

		readFilenames(s3)

	}
}

func readFile(filename string) ([]byte, error) {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))

	return b, nil

}
