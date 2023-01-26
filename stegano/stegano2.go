package stegano

import (
	"Projet/bits"
	"fmt"
	"image"
	"math"
	"sync"
	"time"
)

func calcOneBits(args ...any) []any {
	keyBits := args[0].(*[]byte)
	keyPart := args[1].(int)
	nOnes := args[2].(*[]int)
	for i := 8 * keyPart; i < 8*(keyPart+1); i++ {
		(*nOnes)[keyPart] += bits.GetBit(keyBits, i)
	}
	return nil
}

func applyZigZag(args ...any) []any {
	chunk := args[0].(*[]byte)
	zigzag := args[1].(*[]byte)
	newChunk := args[2].(*[]byte)
	for i := 0; i < 64; i++ {
		bits.SetBit(newChunk, i, bits.GetBit(chunk, int((*zigzag)[i])-1))
	}
	return nil
}

func decodeZigZag(args ...any) []any {
	chunk := args[0].(*[]byte)
	zigzag := args[1].(*[]byte)
	newChunk := args[2].(*[]byte)
	for i := 0; i < 64; i++ {
		bits.SetBit(newChunk, int((*zigzag)[i])-1, bits.GetBit(chunk, i))
	}
	return nil
}

func applyXOR(args ...any) []any {
	chunk := args[0].(*[]byte)
	key := args[1].(*[]byte)
	newChunk := args[2].(*[]byte)
	i := args[3].(int)
	for j := 0; j < 64; j++ {
		bits.SetBit(newChunk, j, bits.GetBit(chunk, j)^bits.GetBit(key, (i%4)*16+j%16))
		//(*chunk)[j] ^ (*key)[i%4+j%2]
	}

	return nil
}

func getBlock(args ...any) []any {
	img := args[0].(*image.RGBA)
	bNum := args[1].(int)
	blocks := args[2].(*[][][]byte)
	(*blocks)[bNum] = make([][]byte, 8)
	bX := img.Stride * 3 / 4 / 8
	for i := 0; i < 8; i++ {
		(*blocks)[bNum][i] = make([]byte, 8)
		//fmt.Println(img.Stride, ((bNum/bX)*8+i)*img.Stride+32*(i%bX)+bNum%3, ((bNum/bX)*8+i)*img.Stride+32*(i%bX)+bNum%3+32)
		for j := 0; j < 8; j++ {
			bNumMod3 := bNum % 3
			//fmt.Println(bNum, bX, len(img.Pix), i, j, ((bNum/bX)*8+i)*img.Stride+8*((bNum-bNumMod3)%bX+(bNum%bX)/3)+bNumMod3+j*4, ((bNum/bX)*8+i)*img.Stride, 8*((bNum-bNumMod3)%bX+(bNum%bX)/3), bNumMod3+j*4)
			(*blocks)[bNum][i][j] = img.Pix[((bNum/bX)*8+i)*img.Stride+8*((bNum-bNumMod3)%bX+(bNum%bX)/3)+bNumMod3+j*4]
		}
		//copy((*blocks)[bNum][i], img.Pix[((bNum/bX)*8+i)*img.Stride+32*(bNum%bX)+bNum%3:((bNum/bX)*8+i)*img.Stride+32*(bNum%bX)+bNum%3+32:4])
	}
	return nil
}

func cosines(args ...any) []any {
	cosinesArr := args[0].(*[][]float64)
	i := args[1].(int)
	(*cosinesArr)[i] = make([]float64, 8)
	for j := 0; j < 8; j++ {
		(*cosinesArr)[i][j] = math.Cos((2*float64(i) + 1) * float64(j) * math.Pi / 16)
	}
	return nil
}

func dct(args ...any) []any {
	block := args[0].(*[][]byte)
	newBlock := args[1].(*[][]float64)
	cosinesArr := args[2].(*[][]float64)
	for i := 0; i < 8; i++ {
		(*newBlock)[i] = make([]float64, 8)
		for j := 0; j < 8; j++ {
			sum := 0.0
			for x := 0; x < 8; x++ {
				for y := 0; y < 8; y++ {
					sum += float64((*block)[x][y]) * (*cosinesArr)[x][i] * (*cosinesArr)[y][j]
				}
			}
			(*newBlock)[i][j] = sig(i, 8) * sig(j, 8) * sum
		}
	}
	return nil
}

func sig(x, n int) float64 {
	if x == 0 {
		return 1 / math.Sqrt(float64(n))
	} else {
		return math.Sqrt(2 / float64(n))
	}
}

func getEnergy(args ...any) []any {
	block := args[0].(*[][]float64)
	energy := args[1].(*[]float64)
	index := args[2].(int)
	sum := 0.0
	for i := 0; i < 64; i++ {
		sum += math.Abs((*block)[i/8][i%8])
	}
	(*energy)[index] = sum
	return nil
}

func quantize(args ...any) []any {
	block := args[0].(*[][]float64)
	qdctBlock := args[1].(*[][]int)
	QMatrix := args[2].(*[8][8]int)
	factor := args[3].(float64)
	for i := 0; i < 8; i++ {
		(*qdctBlock)[i] = make([]int, 8)
		for j := 0; j < 8; j++ {
			(*qdctBlock)[i][j] = int((*block)[i][j]) / int(float64((*QMatrix)[i][j])*factor)
		}
	}
	return nil
}

func getDCCoeffs(args ...any) []any {
	block := args[0].(*[][]int)
	dcCoeffs := args[1].(*[]int)
	zigzag := args[2].(*[]byte)
	for i := 0; i < 64; i++ {
		(*dcCoeffs)[(*zigzag)[i]-1] = (*block)[i/8][i%8]
	}
	return nil
}

func embedMessage(args ...any) []any {
	dcCoeffs := args[0].(*[]int)
	encodedMessage := args[1].(*[]byte)
	messageIndex := args[2].(int)
	privateKeyBytes := args[3].(*[]byte)
	for i := 1; i <= 8; i++ {
		//fmt.Println("embedMessage", i, messageIndex, (*dcCoeffs)[i], bits.GetBit(privateKeyBytes, (messageIndex)%64))
		if messageIndex >= len(*encodedMessage)*8 {
			break
		}
		if bits.GetBit(privateKeyBytes, messageIndex%64) == 1 {
			if (*dcCoeffs)[i] < 0 {
				val := int(math.Abs(float64((*dcCoeffs)[i])))
				//set the LSB of val to encodedMessage[messageIndex]
				if bits.GetBit(encodedMessage, messageIndex) == 1 {
					val = val | 1
				} else {
					val = val &^ 1
				}
				(*dcCoeffs)[i] = -val
			} else {
				//set the LSB of val to encodedMessage[messageIndex]
				if bits.GetBit(encodedMessage, messageIndex) == 1 {
					(*dcCoeffs)[i] = int((*dcCoeffs)[i]) | 1
				} else {
					(*dcCoeffs)[i] = int((*dcCoeffs)[i]) &^ 1
				}
			}
			messageIndex++
		}
		//fmt.Println((*dcCoeffs)[i], bits.GetBit(privateKeyBytes, (messageIndex)%64))
	}
	return nil
}

func invGetDCCoeffs(args ...any) []any {
	block := args[0].(*[][]int)
	dcCoeffs := args[1].(*[]int)
	zigzag := args[2].(*[]byte)
	for i := 0; i < 8; i++ {
		(*block)[i] = make([]int, 8)
		for j := 0; j < 8; j++ {
			(*block)[i][j] = (*dcCoeffs)[(*zigzag)[i*8+j]-1]
		}
	}
	return nil
}

func invQuantize(args ...any) []any {
	QDCTBlocks := args[0].(*[][]int)
	invDCTBlocks := args[1].(*[][]float64)
	QMatrix := args[2].(*[8][8]int)
	factor := args[3].(float64)
	for i := 0; i < 8; i++ {
		(*invDCTBlocks)[i] = make([]float64, 8)
		for j := 0; j < 8; j++ {
			(*invDCTBlocks)[i][j] = float64((*QDCTBlocks)[i][j]) * float64(QMatrix[i][j]) * factor
		}
	}
	return []any{}

}

func invDCT(args ...any) []any {
	block := args[0].(*[][]float64)
	invBlock := args[1].(*[][]float64)
	cosineArr := args[2].(*[][]float64)
	for i := 0; i < 8; i++ {
		(*invBlock)[i] = make([]float64, 8)
		for j := 0; j < 8; j++ {
			sum := 0.0
			for k := 0; k < 8; k++ {
				for l := 0; l < 8; l++ {
					sum += (*cosineArr)[i][k] * (*cosineArr)[j][l] * (*block)[k][l] * sig(k, 8) * sig(l, 8)
				}
			}
			(*invBlock)[i][j] = sum
		}
	}
	return nil
}

func convertBlockToImage(args ...any) []any {
	invBlock := args[0].(*[][]float64)
	imgResult := args[1].(*image.RGBA)
	bNum := args[2].(int)
	bX := imgResult.Stride * 3 / 4 / 8
	bNumMod3 := bNum % 3
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			imgResult.Pix[((bNum/bX)*8+i)*imgResult.Stride+8*((bNum-bNumMod3)%bX+(bNum%bX)/3)+bNumMod3+j*4] = byte((*invBlock)[i][j])
			if bNum%3 == 2 {
				imgResult.Pix[((bNum/bX)*8+i)*imgResult.Stride+8*((bNum-bNumMod3)%bX+(bNum%bX)/3)+bNumMod3+j*4+1] = 255
			}
		}
	}
	return nil
}

func Encde(message string, key [8]byte, img *image.RGBA, qualityFactor int) *image.RGBA {
	nWorkers := 12
	jobQueue := make(chan Job, nWorkers+1)
	resultQueue := make(chan []any, nWorkers+1)
	var wg sync.WaitGroup
	workers := make([]Worker, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = Worker{Id: i, JobQueue: jobQueue, ResultQueue: resultQueue, WG: &wg}
		go workers[i].run()
	}

	//message = "Hello World ! It's a beautiful day to try and do steganography!"

	//img := image2.LoadImage("./test/512.png")
	//privateKeyBytes := []byte{48, 130, 2, 94, 2, 1, 0, 2}
	privateKeyBytes := key[:]

	tinitial := time.Now()
	//convert message to bytes
	messageBytes := append([]byte(message), 26)

	fmt.Println("messageBytes", messageBytes)
	//cut the message into 16 bytes chunks and add padding if needed

	nOnes := make([]int, 8)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			wg.Add(1)
			jobQueue <- Job{f: calcOneBits, args: []any{&privateKeyBytes, i, &nOnes}, OutputResults: false}
		}
	}()
	wg.Wait()

	nBlocks := 3 * (img.Bounds().Max.X / 8) * (img.Bounds().Max.Y / 8)
	var bitsPer4Blocks int
	for i := 0; i < len(nOnes); i++ {
		bitsPer4Blocks += nOnes[i]
	}
	//if (nBlocks/4)*bitsPer4Blocks < len(messageBytes)*8 {
	//	fmt.Println("Message is too long for this image")
	//	return nil
	//}

	chunks := make([][]byte, 0)
	for i := 0; i < len(messageBytes); i += 8 {
		chunk := messageBytes[i:int(math.Min(float64(i+8), float64(len(messageBytes))))]
		if len(chunk) < 8 {
			chunk = append(chunk, make([]byte, 8-len(chunk))...)
		}
		chunks = append(chunks, chunk)
	}

	var zigzag0 []byte = []byte{
		1, 2, 6, 7, 15, 16, 28, 29,
		3, 5, 8, 14, 17, 27, 30, 43,
		4, 9, 13, 18, 26, 31, 42, 44,
		10, 12, 19, 25, 32, 41, 45, 54,
		11, 20, 24, 33, 40, 46, 53, 55,
		21, 23, 34, 39, 47, 52, 56, 61,
		22, 35, 38, 48, 51, 57, 60, 62,
		36, 37, 49, 50, 58, 59, 63, 64,
	}

	var zigzag1 []byte = []byte{
		64, 63, 59, 58, 50, 49, 37, 36,
		61, 60, 57, 51, 48, 38, 35, 22,
		62, 55, 52, 47, 39, 34, 23, 21,
		55, 53, 46, 40, 33, 24, 20, 11,
		54, 45, 41, 32, 25, 19, 12, 10,
		44, 42, 31, 26, 18, 13, 9, 4,
		43, 30, 27, 17, 14, 8, 5, 3,
		29, 28, 16, 15, 7, 6, 2, 1,
	}

	var zigzag2 []byte = []byte{
		29, 28, 16, 15, 7, 6, 2, 1,
		43, 30, 27, 17, 14, 8, 5, 3,
		44, 42, 31, 26, 18, 13, 9, 4,
		54, 45, 41, 32, 25, 19, 12, 10,
		55, 53, 46, 40, 33, 24, 20, 11,
		61, 56, 52, 47, 39, 34, 23, 21,
		62, 60, 57, 51, 48, 38, 35, 22,
		64, 63, 59, 58, 50, 49, 37, 36,
	}

	var zigzag3 []byte = []byte{
		36, 37, 49, 50, 58, 59, 63, 64,
		22, 35, 38, 48, 51, 57, 60, 62,
		21, 23, 34, 39, 47, 52, 56, 61,
		11, 20, 24, 33, 40, 46, 53, 55,
		10, 12, 19, 25, 32, 41, 45, 54,
		4, 9, 13, 18, 26, 31, 42, 44,
		3, 5, 8, 14, 17, 27, 30, 43,
		1, 2, 6, 7, 15, 16, 28, 29,
	}

	var zigzag [][]byte = [][]byte{zigzag0, zigzag1, zigzag2, zigzag3}
	newChunks := make([][]byte, len(chunks))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(chunks); i++ {
			newChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyZigZag, args: []any{&chunks[i], &zigzag[i%4], &newChunks[i]}, OutputResults: false}
		}
	}()
	wg.Wait()

	xorChunks := make([][]byte, len(chunks))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(chunks); i++ {
			xorChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyXOR, args: []any{&newChunks[i], &privateKeyBytes, &xorChunks[i], i}, OutputResults: false}
		}
	}()
	wg.Wait()

	//convert chunks to bytes
	encodedMessage := make([]byte, 0)
	for i := 0; i < len(xorChunks); i++ {
		encodedMessage = append(encodedMessage, xorChunks[i]...)
	}
	fmt.Println("Encoded message: ", encodedMessage)
	//segment the image into 8x8 blocks

	blocks := make([][][]byte, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			blocks[i] = make([][]byte, 64)
			wg.Add(1)
			jobQueue <- Job{f: getBlock, args: []any{img, i, &blocks}, OutputResults: false}
		}
	}()
	wg.Wait()
	//fmt.Println("Blocks: ", blocks[0])

	cosineArr := make([][]float64, 8)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			wg.Add(1)
			jobQueue <- Job{f: cosines, args: []any{&cosineArr, i}, OutputResults: false}
		}
	}()
	wg.Wait()

	//apply DCT to each block
	dctBlocks := make([][][]float64, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			dctBlocks[i] = make([][]float64, 8)
			wg.Add(1)
			jobQueue <- Job{f: dct, args: []any{&blocks[i], &dctBlocks[i], &cosineArr}, OutputResults: false}
		}
	}()
	wg.Wait()
	//fmt.Println("DCT Blocks: ", dctBlocks[0])
	//fmt.Println("\n")

	QMatrix := [8][8]int{
		{16, 11, 10, 16, 24, 40, 51, 61},
		{12, 12, 14, 19, 26, 58, 60, 55},
		{14, 13, 16, 24, 40, 57, 69, 56},
		{14, 17, 22, 29, 51, 87, 80, 62},
		{18, 22, 37, 56, 68, 109, 103, 77},
		{24, 35, 55, 64, 81, 104, 113, 92},
		{49, 64, 78, 87, 103, 121, 120, 101},
		{72, 92, 95, 98, 112, 100, 103, 99},
	}

	var factor float64
	if qualityFactor > 50 {
		factor = float64(100-qualityFactor) / 50.0
	} else {
		factor = 50.0 / float64(qualityFactor)
	}

	QDCTBlocks := make([][][]int, nBlocks)
	//QDCTBlocks = dctBlocks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			QDCTBlocks[i] = make([][]int, 8)
			wg.Add(1)
			jobQueue <- Job{f: quantize, args: []any{&dctBlocks[i], &QDCTBlocks[i], &QMatrix, factor}, OutputResults: false}
		}
	}()
	wg.Wait()
	//fmt.Println("QDCTBlocks", QDCTBlocks[0])

	dcCoeffs := make([][]int, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			dcCoeffs[i] = make([]int, 64)
			wg.Add(1)
			jobQueue <- Job{f: getDCCoeffs, args: []any{&QDCTBlocks[i], &dcCoeffs[i], &zigzag[0]}, OutputResults: false}
		}
	}()
	wg.Wait()
	//fmt.Println("dcCoeffs", dcCoeffs[0])
	//embed the message into the DC coefficients
	wg.Add(1)
	go func() {
		defer wg.Done()
		var messageIndex int
		for i := 0; i < nBlocks; i++ {
			wg.Add(1)
			jobQueue <- Job{f: embedMessage, args: []any{&dcCoeffs[i], &encodedMessage, messageIndex, &privateKeyBytes}, OutputResults: false}
			messageIndex += nOnes[i%8]
			if messageIndex >= len(encodedMessage)*8 {
				break
			}
		}
	}()
	wg.Wait()
	//fmt.Println("dcCoeffs", dcCoeffs[0])
	invQDCTBlocks := make([][][]int, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			invQDCTBlocks[i] = make([][]int, 8)
			wg.Add(1)
			jobQueue <- Job{f: invGetDCCoeffs, args: []any{&invQDCTBlocks[i], &dcCoeffs[i], &zigzag[0]}, OutputResults: false}
		}
	}()
	wg.Wait()
	//fmt.Println("invQDCTBlocks", invQDCTBlocks[0])

	invDCTBlocks := make([][][]float64, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			invDCTBlocks[i] = make([][]float64, 8)
			wg.Add(1)
			jobQueue <- Job{f: invQuantize, args: []any{&invQDCTBlocks[i], &invDCTBlocks[i], &QMatrix, factor}, OutputResults: false}
		}
	}()
	wg.Wait()

	//fmt.Println("invDCTBlocks Encode", invDCTBlocks[0])

	invBlocks := make([][][]float64, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			invBlocks[i] = make([][]float64, 8)
			wg.Add(1)
			jobQueue <- Job{f: invDCT, args: []any{&invDCTBlocks[i], &invBlocks[i], &cosineArr}, OutputResults: false}
		}
	}()
	wg.Wait()

	///////////////////////////////////////////////////////////
	invBlocks2 := make([][][]byte, len(invBlocks))
	for i := 0; i < nBlocks; i++ {
		invBlocks2[i] = make([][]byte, 8)
		for j := 0; j < 8; j++ {
			invBlocks2[i][j] = make([]byte, 8)
			for k := 0; k < 8; k++ {
				invBlocks2[i][j][k] = byte(math.Round(invBlocks[i][j][k]))
			}
		}
	}
	fmt.Print("invBlocks2", invBlocks2[0][0])
	fmt.Println("invBlocks", invBlocks[0][0])
	dctBlocks2 := make([][][]float64, nBlocks)
	//dctBlocks2 = invBlocks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			dctBlocks2[i] = make([][]float64, 8)
			wg.Add(1)
			jobQueue <- Job{f: dct, args: []any{&invBlocks2[i], &dctBlocks2[i], &cosineArr}, OutputResults: false}
		}
	}()
	wg.Wait()

	fmt.Print("dctBlocks2 ", dctBlocks2[0][0])
	fmt.Print("invDCTBlocks ", invDCTBlocks[0][0])

	QDCTBlocks2 := make([][][]int, nBlocks)
	//QDCTBlocks2 = dctBlocks2
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			QDCTBlocks2[i] = make([][]int, 8)
			wg.Add(1)
			jobQueue <- Job{f: quantize, args: []any{&dctBlocks2[i], &QDCTBlocks2[i], &QMatrix, factor}, OutputResults: false}
		}
	}()
	wg.Wait()

	dcCoeffs2 := make([][]int, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			dcCoeffs2[i] = make([]int, 64)
			wg.Add(1)
			jobQueue <- Job{f: getDCCoeffs, args: []any{&QDCTBlocks2[i], &dcCoeffs2[i], &zigzag[0]}, OutputResults: false}
		}
	}()
	wg.Wait()

	//fmt.Println("dcCoeffs", dcCoeffs2[0])

	// extract LSB of each of the 8 first DC coefficients if key is 1 at bit i mod64
	lsb := make([]byte, 8)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var messageIndex int
		for i := 0; i < nBlocks; i++ {
			wg.Add(1)
			jobQueue <- Job{f: deEmbedMessage, args: []any{&dcCoeffs2[i], &lsb, messageIndex, &privateKeyBytes}, OutputResults: false}
			messageIndex += nOnes[i%8]
			if messageIndex >= 64 {
				break
			}
		}
	}()
	wg.Wait()
	//fmt.Println("lsb", lsb)
	chunks2 := make([][]byte, 0)
	for i := 0; i < len(lsb); i += 8 {
		chunk := lsb[i:int(math.Min(float64(i+8), float64(len(lsb))))]
		if len(chunk) < 8 {
			chunk = append(chunk, make([]byte, 8-len(chunk))...)
		}
		chunks2 = append(chunks2, chunk)
	}
	//fmt.Println("Chunks2", chunks2[0])
	//fmt.Println("xorChunks", xorChunks[0])

	xorChunks2 := make([][]byte, len(chunks))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(chunks2); i++ {
			fmt.Println(chunks2[i])
			xorChunks2[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyXOR, args: []any{&chunks2[i], &privateKeyBytes, &xorChunks2[i], i}, OutputResults: false}
		}
	}()
	wg.Wait()

	fmt.Println("xorChunks2", xorChunks2[0])
	fmt.Println("newChunks", newChunks[0])

	//wg.Add(1)
	//go func() {
	//	defer wg.Done()
	//	for i := 0; i < len(chunks); i++ {
	//		xorChunks[i] = make([]byte, 8)
	//		wg.Add(1)
	//		jobQueue <- Job{f: applyXOR, args: []any{&newChunks[i], &privateKeyBytes, &xorChunks[i], i}, OutputResults: false}
	//	}
	//}()
	//wg.Wait()
	newChunks2 := make([][]byte, len(chunks))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(chunks2); i++ {

			newChunks2[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: decodeZigZag, args: []any{&xorChunks2[i], &zigzag[i%4], &newChunks2[i]}, OutputResults: false}
		}
	}()
	wg.Wait()

	fmt.Println("newChunks2", newChunks2[0])
	fmt.Println("chunks", chunks[0])
	messageBytes2 := make([]byte, 0)
	for i := 0; i < len(newChunks2); i++ {
		messageBytes2 = append(messageBytes2, newChunks2[i]...)
	}
	// remove padding after byte value 26
	for i := 0; i < len(messageBytes2); i++ {
		if messageBytes2[i] == 26 {
			messageBytes2 = messageBytes2[:i]
			break
		}
	}

	fmt.Println("messageBytes2", messageBytes2)
	fmt.Println("message", messageBytes)
	fmt.Println("message is ", string(messageBytes2))

	/////////////////////////////////////////////

	//fmt.Println("invBlocks", invBlocks[0])

	// convert the blocks back to image
	imgResult := image.NewRGBA(img.Rect)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			wg.Add(1)
			jobQueue <- Job{f: convertBlockToImage, args: []any{&invBlocks[i], imgResult, i}, OutputResults: false}
		}
	}()
	wg.Wait()
	tfinal := time.Now()
	//fmt.Println(encodedMessage)
	fmt.Println("Time taken: ", tfinal.Sub(tinitial))
	//// save the image
	//f, err := os.Create("result.png")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//defer f.Close()
	//png.Encode(f, imgResult)
	close(jobQueue)
	close(resultQueue)
	return imgResult
}

func Decde(key [8]byte, img *image.RGBA, qualityFactor int) string {
	nWorkers := 12
	jobQueue := make(chan Job, nWorkers+1)
	resultQueue := make(chan []any, nWorkers+1)
	var wg sync.WaitGroup
	workers := make([]Worker, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = Worker{Id: i, JobQueue: jobQueue, ResultQueue: resultQueue, WG: &wg}
		go workers[i].run()
	}

	var zigzag0 []byte = []byte{
		1, 2, 6, 7, 15, 16, 28, 29,
		3, 5, 8, 14, 17, 27, 30, 43,
		4, 9, 13, 18, 26, 31, 42, 44,
		10, 12, 19, 25, 32, 41, 45, 54,
		11, 20, 24, 33, 40, 46, 53, 55,
		21, 23, 34, 39, 47, 52, 56, 61,
		22, 35, 38, 48, 51, 57, 60, 62,
		36, 37, 49, 50, 58, 59, 63, 64,
	}

	var zigzag1 []byte = []byte{
		64, 63, 59, 58, 50, 49, 37, 36,
		61, 60, 57, 51, 48, 38, 35, 22,
		62, 55, 52, 47, 39, 34, 23, 21,
		55, 53, 46, 40, 33, 24, 20, 11,
		54, 45, 41, 32, 25, 19, 12, 10,
		44, 42, 31, 26, 18, 13, 9, 4,
		43, 30, 27, 17, 14, 8, 5, 3,
		29, 28, 16, 15, 7, 6, 2, 1,
	}

	var zigzag2 []byte = []byte{
		29, 28, 16, 15, 7, 6, 2, 1,
		43, 30, 27, 17, 14, 8, 5, 3,
		44, 42, 31, 26, 18, 13, 9, 4,
		54, 45, 41, 32, 25, 19, 12, 10,
		55, 53, 46, 40, 33, 24, 20, 11,
		61, 56, 52, 47, 39, 34, 23, 21,
		62, 60, 57, 51, 48, 38, 35, 22,
		64, 63, 59, 58, 50, 49, 37, 36,
	}

	var zigzag3 []byte = []byte{
		36, 37, 49, 50, 58, 59, 63, 64,
		22, 35, 38, 48, 51, 57, 60, 62,
		21, 23, 34, 39, 47, 52, 56, 61,
		11, 20, 24, 33, 40, 46, 53, 55,
		10, 12, 19, 25, 32, 41, 45, 54,
		4, 9, 13, 18, 26, 31, 42, 44,
		3, 5, 8, 14, 17, 27, 30, 43,
		1, 2, 6, 7, 15, 16, 28, 29,
	}

	var zigzag [][]byte = [][]byte{zigzag0, zigzag1, zigzag2, zigzag3}
	keyBytes := key[:]
	nOnes := make([]int, 8)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			wg.Add(1)
			jobQueue <- Job{f: calcOneBits, args: []any{&keyBytes, i, &nOnes}, OutputResults: false}
		}
	}()
	wg.Wait()

	//fmt.Println("nOnes", nOnes)

	nBlocks := 3 * (img.Bounds().Max.X / 8) * (img.Bounds().Max.Y / 8)
	//fmt.Println("nBlocks", nBlocks)
	// get the blocks
	blocks := make([][][]byte, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			blocks[i] = make([][]byte, 64)
			wg.Add(1)
			jobQueue <- Job{f: getBlock, args: []any{img, i, &blocks}, OutputResults: false}
		}
	}()
	wg.Wait()

	cosineArr := make([][]float64, 8)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			wg.Add(1)
			jobQueue <- Job{f: cosines, args: []any{&cosineArr, i}, OutputResults: false}
		}
	}()
	wg.Wait()

	//apply DCT to each block
	dctBlocks := make([][][]float64, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			dctBlocks[i] = make([][]float64, 8)
			wg.Add(1)
			jobQueue <- Job{f: dct, args: []any{&blocks[i], &dctBlocks[i], &cosineArr}, OutputResults: false}
		}
	}()
	wg.Wait()

	//fmt.Println("dctBlocks: ", dctBlocks[0])

	QMatrix := [8][8]int{
		{16, 11, 10, 16, 24, 40, 51, 61},
		{12, 12, 14, 19, 26, 58, 60, 55},
		{14, 13, 16, 24, 40, 57, 69, 56},
		{14, 17, 22, 29, 51, 87, 80, 62},
		{18, 22, 37, 56, 68, 109, 103, 77},
		{24, 35, 55, 64, 81, 104, 113, 92},
		{49, 64, 78, 87, 103, 121, 120, 101},
		{72, 92, 95, 98, 112, 100, 103, 99},
	}

	var factor float64
	if qualityFactor > 50 {
		factor = float64(100-qualityFactor) / 50.0
	} else {
		factor = 50.0 / float64(qualityFactor)
	}

	QDCTBlocks := make([][][]int, nBlocks)
	//QDCTBlocks = dctBlocks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			QDCTBlocks[i] = make([][]int, 8)
			wg.Add(1)
			jobQueue <- Job{f: quantize, args: []any{&dctBlocks[i], &QDCTBlocks[i], &QMatrix, factor}, OutputResults: false}
		}
	}()
	wg.Wait()

	dcCoeffs := make([][]int, nBlocks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nBlocks; i++ {
			dcCoeffs[i] = make([]int, 64)
			wg.Add(1)
			jobQueue <- Job{f: getDCCoeffs, args: []any{&QDCTBlocks[i], &dcCoeffs[i], &zigzag[0]}, OutputResults: false}
		}
	}()
	wg.Wait()

	//fmt.Println("dcCoeffs", dcCoeffs[0])

	// extract LSB of each of the 8 first DC coefficients if key is 1 at bit i mod64
	lsb := make([]byte, 8)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var messageIndex int
		for i := 0; i < nBlocks; i++ {
			wg.Add(1)
			jobQueue <- Job{f: deEmbedMessage, args: []any{&dcCoeffs[i], &lsb, messageIndex, &keyBytes}, OutputResults: false}
			messageIndex += nOnes[i%8]
			if messageIndex >= 64 {
				break
			}
		}
	}()
	wg.Wait()
	//fmt.Println("lsb", lsb)
	chunks := make([][]byte, 0)
	for i := 0; i < len(lsb); i += 8 {
		chunk := lsb[i:int(math.Min(float64(i+8), float64(len(lsb))))]
		if len(chunk) < 8 {
			chunk = append(chunk, make([]byte, 8-len(chunk))...)
		}
		chunks = append(chunks, chunk)
	}

	xorChunks := make([][]byte, len(chunks))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(chunks); i++ {
			xorChunks[i] = make([]byte, 64)
			wg.Add(1)
			jobQueue <- Job{f: applyXOR, args: []any{&chunks[i], &keyBytes, &xorChunks[i], i}, OutputResults: false}
		}
	}()
	wg.Wait()

	newChunks := make([][]byte, len(chunks))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(chunks); i++ {
			newChunks[i] = make([]byte, 64)
			wg.Add(1)
			jobQueue <- Job{f: decodeZigZag, args: []any{&xorChunks[i], &zigzag[i%4], &newChunks[i]}, OutputResults: false}
		}
	}()
	wg.Wait()

	messageBytes := make([]byte, 0)
	for i := 0; i < len(newChunks); i++ {
		messageBytes = append(messageBytes, newChunks[i]...)
	}
	// remove padding after byte value 26
	for i := 0; i < len(messageBytes); i++ {
		if messageBytes[i] == 26 {
			messageBytes = messageBytes[:i]
			break
		}
	}

	return string(messageBytes)
}

func deEmbedMessage(args ...any) []any {
	dcCoeffs := args[0].(*[]int)
	lsb := args[1].(*[]byte)
	messageIndex := args[2].(int)
	key := args[3].(*[]byte)

	for j := 1; j < 9; j++ {
		//fmt.Println("deEmbedMessage", j, messageIndex, (*dcCoeffs)[j], bits.GetBit(key, (messageIndex)%64))
		//fmt.Print(" ", messageIndex)
		if bits.GetBit(key, (messageIndex)%64) == 1 {
			if (*dcCoeffs)[j] < 0 {
				val := int(math.Abs(float64((*dcCoeffs)[j])))
				if val&1 == 1 {
					bits.SetBit(lsb, messageIndex, 1)
				} else {
					bits.SetBit(lsb, messageIndex, 0)
				}
			} else {
				if int((*dcCoeffs)[j])&1 == 1 {
					bits.SetBit(lsb, messageIndex, 1)
				} else {
					bits.SetBit(lsb, messageIndex, 0)
				}
			}
			messageIndex++
		}
	}
	//fmt.Println()
	return nil
}
