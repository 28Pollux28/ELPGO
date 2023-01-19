package intent

import (
	image2 "Projet/image"
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"reflect"
	"strings"
)

type Jobs int

func (j Jobs) Filter(img *image.RGBA, red, green, blue, alpha float64) {
	image2.Filter(img, red, green, blue, alpha)
}

type Intent struct {
	jobs           []func()
	img            *image.RGBA
	initColorModel color.Model
}

// The reader will likely be either a file or a network stream with function to apply to an image that will be encoded in bytes in the stream
func ReadIntent(reader io.Reader) (intent Intent, err error) {
	newReader := bufio.NewReader(reader)
	for {
		line, err := newReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return intent, err
			}
		}
		if line == "" {
			continue
		} else if line == "\n" {
			continue
		} else if line == "image\n" {
			fmt.Println("image")
			var imgStr string
			//while line != "endImg"
			for {
				line, err := newReader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						panic("EOF")
					} else {
						return intent, err
					}
				}
				if line == "endimage\n" {
					break
				}
				imgStr += line
			}
			intent.img, intent.initColorModel = readImage(imgStr)
			fmt.Println(intent.img.Stride)
			image2.SaveImage("te2st.png", intent.img)
			intent.jobs = append(intent.jobs, func() {
				//parse line and add job to intent
				j := reflect.ValueOf(Jobs(0))
				j.MethodByName(line).Call([]reflect.Value{reflect.ValueOf(intent.img), reflect.ValueOf(1.0), reflect.ValueOf(1.0), reflect.ValueOf(1.0), reflect.ValueOf(1.0)})
			})
		}
	}
	return intent, nil
}

// ReadImage reads an image from a *bufio.Reader and returns it in RGBA color space.
// It also returns the initial color.Model of the image.
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
