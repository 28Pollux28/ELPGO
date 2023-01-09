package main

import (
	image2 "Projet/image"
	"fmt"
	"image"
	"image/png"
	"os"
)

func main() {
	img := loadImage("./test/imageTest.png")
	image2.Filter(img, -1, -1, 128, -1)
	saveImage("./test/test.png", img)
}

func loadImage(filePath string) *image.NRGBA {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Error happened when opening file "+filePath+" :", err)
		return nil
	}
	img, _, err2 := image.Decode(imgFile)
	if err2 != nil {
		fmt.Println("Error happened when parsing image :", err)
		return nil
	}
	return img.(*image.NRGBA)
}

func saveImage(filePath string, img *image.NRGBA) {
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
