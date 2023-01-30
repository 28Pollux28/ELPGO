package main

import (
	image2 "Projet/image"
	"Projet/stegano"
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

func gestionErreur(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

const (
	IP2   = "localhost" // IP local
	PORT2 = "8000"      // Port utilisé
)

func main() {

	fmt.Println("Lancement du serveur ...")
	var wg sync.WaitGroup
	go func() { // goroutine dédiée à l'entrée utilisateur
		defer wg.Done()
		for {
			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if text == "exit\r\n" {
				fmt.Println("Fermeture du serveur")
				os.Exit(0)
			}
		}
	}()

	// on écoute sur le port 8000
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", IP2, PORT2))
	if err != nil {
		fmt.Println(err)
	}

	// On accepte les connexions entrantes sur le port 8000
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}
		// Information sur les clients qui se connectent
		fmt.Println("Un client est connecté depuis", conn.RemoteAddr())
		go func() { // goroutine dédiée à la reception des messages du serveur
			var data string
			var imgStr string
			var key string
			var action string
			reader := bufio.NewReader(conn)
			read := true
			for read {
				message, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Client déconnecté")
					break
				}
				if message == "startData\n" {
					fmt.Println("Reception de données")
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
							fmt.Println("Données reçues")
							break
						}
						data += line
					}
				} else if message == "startImage\n" {
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
							break
						}
						imgStr += line
					}
				} else if message == "startAction\n" {
					fmt.Println("Reception d'une action")
					for {
						line, err := reader.ReadString('\n')
						if err != nil {
							if err == io.EOF {
								panic("EOF")
							} else {
								panic(err)
							}
						}
						if line == "endAction\n" {
							fmt.Println("Action reçue")
							break
						}
						action += line[:len(line)-1]
					}
				} else if message == "startKey\n" {
					fmt.Println("Reception d'une clé")
					for {
						line, err := reader.ReadString('\n')
						if err != nil {
							if err == io.EOF {
								panic("EOF")
							} else {
								panic(err)
							}
						}
						if line == "endKey\n" {
							fmt.Println("Clé reçue")
							read = false
							break
						}
						key += line[:len(line)-1]
					}
				}
			}
			img, _ := readImage(imgStr)
			if action == "encode" {
				encode, err := stegano.Encode([]byte(data), []byte(key), img)
				if err != nil {
					return
				}
				conn.Write([]byte("startImage\n"))
				err2 := png.Encode(conn, encode)

				conn.Write([]byte("\nendImage\n"))
				conn.Write([]byte("end\n"))
				if err2 != nil {
					return
				}
			} else if action == "decode" {
				img = image2.LoadImage("output.png")
				decode, err := stegano.Decode(img, []byte(key))
				if err != nil {
					return
				}
				conn.Write([]byte("startData\n"))
				conn.Write(decode)
				conn.Write([]byte("\nendData\n"))
				conn.Write([]byte("end\n"))
			}
		}()
	}
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
