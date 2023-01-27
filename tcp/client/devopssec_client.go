package main

import (
	image2 "Projet/image"
	"fmt"
	"image"
	"image/color"
	"os"
)

const (
	IP   = "localhost" // IP local
	PORT = "8000"      // Port utilisé
)

func main() {
	//args := os.Args
	//fmt.Println(args)
	//fmt.Println(os.Getwd())
	//if len(args) < 4 {
	//	fmt.Println("Usage: go run client_final.go <data_file> <image_file> <key>")
	//	os.Exit(1)
	//}
	//data, err := os.Open(args[1])
	//if err != nil {
	//	fmt.Println("Error while reading data file")
	//	os.Exit(1)
	//}
	//defer data.Close()
	//img, err2 := os.Open("./test/webb.png")
	//if err2 != nil {
	//	fmt.Println("Error while reading image file")
	//	os.Exit(1)
	//}
	//defer img.Close()
	imgFile, err := os.Open("./test/webb.png")
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Error happened when opening file "+" :", err)
	}
	img, _, err2 := image.Decode(imgFile)
	if err2 != nil {
		fmt.Println("Error happened when decoding file "+" :", err2)
	}
	fmt.Println(img)
	_ = image2.ConvertImageToModel(img, color.RGBAModel).(*image.RGBA)
	//fmt.Println(imi)
	//if format != "png" {
	//	fmt.Println("Image file is not a png")
	//	os.Exit(1)
	//}
	//key := args[3]
	//keyBytes := []byte(key)
	//fmt.Println(keyBytes)
	//
	//var wg sync.WaitGroup
	//
	//// Connexion au serveurf
	//conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", IP, PORT))
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//wg.Add(2)

	//go func(data *[]byte, img io.Reader, key *[]byte) { // goroutine dédiée à l'entrée utilisateur
	//	defer wg.Done()
	//	for {
	//		//reader := bufio.NewReader(os.Stdin)
	//		//text, err := reader.ReadString('\n')
	//
	//		//if err != nil {
	//		//	fmt.Println(err)
	//		//}
	//		//if text == "exit\r\n" {
	//		//	conn.Close()
	//		//	os.Exit(0)
	//		//}
	//
	//		//Send data
	//		_, err := conn.Write([]byte("startData"))
	//		if err != nil {
	//			fmt.Println(err)
	//			os.Exit(1)
	//		}
	//		_, err = conn.Write(*data)
	//		if err != nil {
	//			fmt.Println(err)
	//			os.Exit(1)
	//		}
	//		_, err = conn.Write([]byte("endData"))
	//		if err != nil {
	//			fmt.Println(err)
	//			os.Exit(1)
	//		}
	//
	//	}
	//}(&data, img, &keyBytes)

	//go func() { // goroutine dédiée à la reception des messages du serveur
	//	defer wg.Done()
	//	for {
	//		message, err := bufio.NewReader(conn).ReadString('\n')
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//
	//		fmt.Print("serveur : " + message)
	//	}
	//}()
	//
	//wg.Wait()

}
