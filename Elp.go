package main

import (
	image2 "Projet/image"
	"Projet/stegano"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"
)

func main() {
	data, _ := os.ReadFile("./test/512.png") //"Bonjour Gaspard, prochaine étape, envoie moi ta carte bleue :)" /*orld ! It's a beautiful day to try and do steganography!*/
	privateKeyBytes := []byte{48, 130, 2, 94, 2, 1, 0, 2}
	img := image2.LoadImage("./test/webb.png")
	for i := 0; i < 10; i++ {
		//fmt.Println("sending message : ", data)
		stegoImg, err := stegano.Encode(data, privateKeyBytes, img)
		if err != nil {
			fmt.Println(err)
		} else {
			if stegoImg != nil {
				image2.SaveImage("./output.png", stegoImg)
			}
		}
		time.Sleep(1 * time.Second)

		message2, _ := stegano.Decode(stegoImg, privateKeyBytes)
		create, err := os.Create("./output4121512.png")
		if err != nil {
			fmt.Println(err)
		}
		create.Write(message2)
		create.Close()
	}

	//fmt.Println("got message: ", string(message2))

}

func loadImage(filePath string) *image.RGBA {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Error happened when opening file "+filePath+" :", err)
		return nil
	}
	img, _, err2 := image.Decode(imgFile)
	//png.Encode(os.Stdout, img)
	//convert := color.NRGBAModel.Convert(img.At(0, 0))
	//get the nrgba of the img at the 0,0
	//fmt.Println(convert, convert[1], convert[2], convert[3])

	//fmt.Println(r>>8, g>>8)
	if err2 != nil {
		fmt.Println("Error happened when parsing image :", err)
		return nil
	}
	return img.(*image.RGBA)
}

func saveImage(filePath string, img *image.RGBA) {
	imgFile, err := os.Create(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Error happened when creating file "+filePath+" :", err)
		return
	}
	err2 := png.Encode(imgFile, img.SubImage(img.Rect))
	if err2 != nil {
		fmt.Println("Error happened when encoding image :", err)
		return
	}
}
