package main

import (
	"fmt"
	"os"
)

func main() {
	a := img_to_byte("128.png")
	fmt.Println(a)
	var ch string
	for i := 0; i < len(a); i++ {
		ch += a[i]
	}
	ecriture(ch)
}

//imgFile, _ := os.ReadFile("./test/4.png")
//f, _ := os.Create("./test/intent.txt")
//f.Write([]byte("image\n"))
//f.Write(imgFile)
//f.Write([]byte("\nendimage\n"))
//f.Close()

func img_to_byte(img string) []byte {
	imgFile, _ := os.ReadFile(img)
	f, _ := os.Create("intent.txt")
	f.Write([]byte("image\n"))
	f.Write(imgFile)
	f.Write([]byte("\nendimage\n"))
	f.Close()
	return imgFile
}

func ecriture(txt string) {
	file, err := os.OpenFile("img.txt", os.O_CREATE|os.O_WRONLY, 0600) //os.O_APPEND pour ne pas ecraser le fichier
	defer file.Close()                                                 // on ferme automatiquement à la fin de notre programme

	file.Sync() //effacer les ecritures

	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(txt)
	if err != nil {
		panic(err)
	}
}
