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

	for {
		switch inputChoice() {
		case "1":
			readFilenames(s3)
		case "2":
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

			for _, l := range lines {
				if err := s3.store.Delete(s3.ID, l); err != nil {
					log.Fatal(err)
				}
			}

		case "3":
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

			for _, l := range lines {
				r, err := s3.Get(l)

				if err != nil {
					fmt.Printf("File %s not found in storage", l)
					break
				}

				b, err := io.ReadAll(r)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(string(b))
			}

		default:
			fmt.Println("No input accepted. \n\n")
		}

		time.Sleep(500 * time.Millisecond)
	}

}

func inputChoice() string {
	fmt.Println(`Please select what you would like to do:
	[1] Add files
	[2] Delete files
	[3] Get files
	`)
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	line := scanner.Text()

	return line
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
