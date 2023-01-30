package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	_ "image/png"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	IP   = "localhost" // IP local
	PORT = "8000"      // Port utilisé
)

const BUFFERSIZE = 4096

func main() {
	args := os.Args
	action := args[1]
	key := args[2]
	keyBytes := []byte(key)
	fmt.Println(keyBytes)
	var imgFile *os.File
	var data []byte
	if action == "encode" {
		if len(args) < 5 {
			fmt.Println("Usage: go run devopssec_client.go <encode|decode> <key> <image_file> <data_file>")
			os.Exit(1)
		}
		var err2 error
		imgFile, err2 = os.Open(args[3])
		if err2 != nil {
			fmt.Println("Error while reading image file")
			os.Exit(1)
		}
		var err error
		data, err = os.ReadFile(args[4])
		if err != nil {
			fmt.Println("Error while reading data file")
			os.Exit(1)
		}
	} else if action == "decode" {
		if len(args) < 4 {
			fmt.Println("Usage: go run devopssec_client.go <encode|decode> <key> <image_file>")
			os.Exit(1)
		}
		var err2 error
		imgFile, err2 = os.Open(args[3])
		if err2 != nil {
			fmt.Println("Error while reading image file")
			os.Exit(1)
		}
	}
	var wg sync.WaitGroup

	// Connexion au serveur
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", IP, PORT))
	if err != nil {
		fmt.Println(err)
	}

	wg.Add(2)

	go func(data *[]byte, img *os.File, key *[]byte) { // goroutine dédiée à l'entrée utilisateur
		defer wg.Done()

		//Send data
		if action == "encode" {
			_, err := conn.Write([]byte("startData\n"))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = conn.Write(*data)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = conn.Write([]byte("\nendData\n"))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		_, err2 := conn.Write([]byte("startImage\n"))
		if err2 != nil {
			fmt.Println(err2)
			os.Exit(1)
		}
		sendFileToClient(conn, img)
		//fmt.Println(n)
		_, err2 = conn.Write([]byte("\nendImage\n"))
		fmt.Println("Image sent")
		if err2 != nil {
			fmt.Println(err2)
			os.Exit(1)
		}
		_, err4 := conn.Write([]byte("startAction\n"))
		if err4 != nil {
			fmt.Println(err4)
			os.Exit(1)
		}
		_, err4 = conn.Write([]byte(action + "\n"))
		if err4 != nil {
			fmt.Println(err4)
			os.Exit(1)
		}
		_, err4 = conn.Write([]byte("\nendAction\n"))
		if err4 != nil {
			fmt.Println(err4)
			os.Exit(1)
		}

		_, err3 := conn.Write([]byte("startKey\n"))
		if err3 != nil {
			fmt.Println(err3)
			os.Exit(1)
		}
		_, err3 = conn.Write(*key)
		if err3 != nil {
			fmt.Println(err3)
			os.Exit(1)
		}
		_, err3 = conn.Write([]byte("\nendKey\n"))
		if err3 != nil {
			fmt.Println(err3)
			os.Exit(1)
		}
		fmt.Println("Data sent")
	}(&data, imgFile, &keyBytes)

	go func() { // goroutine dédiée à la reception des messages du serveur
		defer wg.Done()
		reader := bufio.NewReader(conn)
		var imgStr string
		var dataStr string
		var imgOuData string
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if message == "startImage\n" {
				fmt.Println("Reception d'une image")
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							panic("EOF")
						} else {
							panic(err)
						}
					}
					if line == "endImage\n" {
						fmt.Println("Image reçue")
						imgOuData = "image"
						break
					}
					imgStr += line
				}
			} else if message == "startData\n" {
				fmt.Println("Reception de data")
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							panic("EOF")
						} else {
							panic(err)
						}
					}
					if line == "endData\n" {
						fmt.Println("Data reçue")
						imgOuData = "data"
						break
					}
					dataStr += line
				}
			} else if message == "end\n" {
				break
			}
		}
		if imgOuData == "image" {
			img, _ := readImage(imgStr)
			f, err := os.Create("output.png")
			defer f.Close()
			if err != nil {
				fmt.Println(err)
			}
			png.Encode(f, img)
		} else if imgOuData == "data" {
			f, err := os.Create("outputDecodePFR.png")
			defer f.Close()
			if err != nil {
				fmt.Println(err)
			}
			_, err = f.Write([]byte(dataStr))
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	wg.Wait()

}

func sendFileToClient(connection net.Conn, file *os.File) {
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file!")
	for {
		_, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("File has been sent!")
	file.Close()
	return
}

func readImage(str string) (*image.RGBA, color.Model) {
	//convert string to reader
	reader := strings.NewReader(str)
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println(err)
	}
	if img.ColorModel() != color.RGBAModel {
		imga := image.NewRGBA(img.Bounds())
		draw.Draw(imga, img.Bounds(), img, img.Bounds().Min, draw.Src)
		return imga, img.ColorModel()
	}
	return img.(*image.RGBA), img.ColorModel()
}
