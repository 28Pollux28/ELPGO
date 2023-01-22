package main

import (
	image2 "Projet/image"
	"Projet/stegano"
	"fmt"
	"image"
	"image/png"
	"os"
	"sync"
	//"time"
)

func worker(wg *sync.WaitGroup, img *image.RGBA, jobs chan []uint /*, results chan<- []uint8*/) {
	for i := <-jobs; i != nil; i = <-jobs {
		image2.Filter(img, i[0], i[1], 1, 0, 0, -1)
		wg.Done()
	}

}

func main() {
	//t1 := time.Now().UnixMilli()
	//imgFile, _ := os.ReadFile("./test/4.png")
	//f, _ := os.Create("./test/intent.txt")
	//f.Write([]byte("image\n"))
	//f.Write(imgFile)
	//f.Write([]byte("\nendimage\n"))
	//f.Close()
	//
	//f, _ = os.Open("./test/intent.txt")
	//defer f.Close()
	//intent.ReadIntent(f)

	//img := loadImage("./test/largeImage.png")
	////fmt.Println(img.Stride)
	//t1 := time.Now().UnixMilli()
	//var wg sync.WaitGroup
	//chO := make(chan []uint, 50)
	////chI := make(chan []uint8)
	//
	//for i := 0; i < 48; i++ {
	//	go func() {
	//		worker(&wg, img, chO)
	//	}()
	//}
	//go func() {
	//	for i := 0; i < img.Rect.Dy(); i++ {
	//		wg.Add(1)
	//
	//		chO <- []uint{uint(i*img.Stride + 0), uint((i+1)*img.Stride - 1)}
	//	}
	//	fmt.Println("i")
	//}()
	//
	//wg.Wait()
	////image2.Filter(img, 1.5, -1.0, 0, -1)
	//t2 := time.Now().UnixMilli()
	//fmt.Println("Time taken:", t2-t1, "ms")
	//saveImage("./test/test.png", img)
	stegano.Main("test")

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
