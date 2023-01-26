package image

import (
	"Projet/bits"
	mathUtil "Projet/utils"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"sync"
	"time"
)

type Job struct {
	f             func(...any) []any
	args          []any
	OutputResults bool
}

type Worker struct {
	Id          int
	JobQueue    chan Job
	ResultQueue chan []any
	WG          *sync.WaitGroup
}

func (w *Worker) run() {
	for job := <-w.JobQueue; job.f != nil; job = <-w.JobQueue {
		result := job.f(job.args...)
		if job.OutputResults {
			w.ResultQueue <- result
		}
		w.WG.Done()
	}
}

func substRange(args ...any) []any {
	img := args[0].(*image.RGBA)
	line := args[1].(int)
	arr := make([]int, 0)
	for i := line * img.Stride; i < (line+1)*img.Stride; i++ {
		if i%4 != 3 {
			arr = append(arr, int(img.Pix[i])-128)
		}
	}
	return []any{line, arr}
}

func break8x8(args ...any) []any {
	blocks := args[0].(*[][][]int)
	arr := args[1].(*[]int)
	n := args[2].(int)
	p := args[3].(int)
	k := args[4].(int)

	(*blocks)[k] = make([][]int, 8)
	for i := 0; i < 8; i++ {
		(*blocks)[k][i] = make([]int, 8)
		copy((*blocks)[k][i], (*arr)[n*((k/p)*8+i)+(k%p)*8:n*((k/p)*8+i)+k%p*8+8])
	}
	return nil
}

func merge8x8(args ...any) []any {
	blocks := args[0].(*[][][]int)
	arr := args[1].(*[]int)
	n := args[2].(int)
	p := args[3].(int)
	k := args[4].(int)

	for i := 0; i < 8; i++ {
		copy((*arr)[n*((k/p)*8+i)+(k%p)*8:n*((k/p)*8+i)+k%p*8+8], (*blocks)[k][i])
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
	block := args[0].(*[][]int)
	dctBlocks := args[1].(*[][][]float64)
	cosinesArr := args[2].(*[][]float64)
	k := args[3].(int)
	(*dctBlocks)[k] = make([][]float64, 8)
	for u := 0; u < 8; u++ {
		(*dctBlocks)[k][u] = make([]float64, 8)
		for v := 0; v < 8; v++ {
			sum := 0.0
			for i := 0; i < 8; i++ {
				for j := 0; j < 8; j++ {
					sum += float64((*block)[i][j]) * (*cosinesArr)[i][u] * (*cosinesArr)[j][v] //math.Cos((2*float64(i)+1)*float64(u)*math.Pi/16) * math.Cos((2*float64(j)+1)*float64(v)*math.Pi/16)
				}
			}
			(*dctBlocks)[k][u][v] = sum * sig(u) * sig(v) / 4
		}
	}
	return nil
}

func invdct(args ...any) []any {
	block := args[0].(*[][]int)
	dctBlock := args[1].(*[][]float64)
	for i := 0; i < 8; i++ {
		(*block)[i] = make([]int, 8)
		for j := 0; j < 8; j++ {
			sum := 0.0
			for u := 0; u < 8; u++ {
				for v := 0; v < 8; v++ {
					sum += sig(u) * sig(v) * (*dctBlock)[u][v] * math.Cos((2*float64(i)+1)*float64(u)*math.Pi/16) * math.Cos((2*float64(j)+1)*float64(v)*math.Pi/16)
				}
			}
			(*block)[i][j] = int(math.Round(sum / 4))
		}
	}
	return nil
}

func calcEntropy(args ...any) []any {
	block := args[0].(*[][]float64)
	entropy := 0.0
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			entropy += math.Pow(math.Abs((*block)[i][j]), 2)
		}
	}
	return []any{entropy}
}

func sig(x int) float64 {
	if x == 0 {
		return 1 / math.Sqrt(2)
	} else {
		return 1
	}
}

func quantize(args ...any) []any {
	QDCTBlocks := args[0].(*[][][]float64)
	dctBlocks := args[1].(*[][][]float64)
	QMatrix := args[2].(*[8][8]int)
	//qualityFactor := args[3].(int)
	k := args[3].(int)
	i := args[4].(int)
	j := args[5].(int)
	factor := args[6].(float64)
	(*QDCTBlocks)[k][i][j] = math.Round((*dctBlocks)[k][i][j] / (factor * float64((*QMatrix)[i][j])))
	return nil
}

func dequantize(args ...any) []any {
	invQDCTBlocks := args[0].(*[][][]float64)
	invDCTBlocks := args[1].(*[][][]float64)
	QMatrix := args[2].(*[8][8]int)
	k := args[3].(int)
	i := args[4].(int)
	j := args[5].(int)
	factor := args[6].(float64)
	for l := 0; l < 8; l++ {
		(*invDCTBlocks)[k][i][j] += (*invQDCTBlocks)[k][i][l] * (*invQDCTBlocks)[k][l][j] * float64((*QMatrix)[i][l]) * float64((*QMatrix)[l][j]) * factor * factor
	}
	return nil
}

func applyPermutation(args ...any) []any {
	permutation := args[0].(*[]int)
	key := args[1].([]int)
	permutationKey := args[2].(*[]int)
	for i := 0; i < len(*permutation); i++ {
		(*permutationKey)[i/8] = key[(*permutation)[i]/8] & (1 << uint((*permutation)[i]%8))
	}
	return nil
}

func R(x int, si int) int {
	if si == 0 {
		if x%2 == 0 {
			return x
		} else {
			return x - 1
		}
	} else {
		if x%2 == 0 {
			return x + 1
		} else {
			return x
		}
	}
}

func invR(y int) int {
	if y%2 == 0 {
		return 0
	} else {
		return 1
	}
}

func Main() {
	img := LoadImage("./test/128.png")
	qualityFactor := 50
	message := "Hello, world" //!azertyuioihgfdsdfghjhgfdefghjiuytrfedfgvhjutfvgbhytrfvgbhytrfvbghytrfdcvgbhytrfcvgtrfdcvbgtrfdcfvgtrdcvfgtrfdcvgtrfdcvfgtrfdcvgytfdcvfgtrfdcvgtrfvgtyrfcvgytfvghuyhjkolhbjklhghjigfdcfvgbhdsdfqsdfresdfresxdcfgtfdcfvghgfvbnjkgbnjkgvbnjhgvfcderfghfdcvbhgfdrtghjygtfdsedrfghjiuygtfdsdefghjhgfdertghjk"
	//message += message
	//message += message
	messageBytes := []byte(message)
	messLen := len(messageBytes)
	mode := 0 //0 - Embedding, 1 - Extraction
	messageDecodeBytes := make([]byte, 0)

	permutation128 := []int{
		41, 36, 52, 50, 32, 122, 127, 113,
		26, 106, 37, 28, 73, 16, 91, 63,
		103, 57, 44, 59, 102, 22, 62, 6,
		84, 7, 38, 116, 64, 12, 1, 100,
		34, 123, 30, 96, 43, 40, 109, 24,
		119, 2, 104, 60, 49, 69, 101, 94,
		112, 76, 114, 58, 74, 20, 47, 124,
		19, 46, 98, 99, 89, 31, 18, 8,
		3, 93, 67, 121, 108, 88, 27, 82,
		110, 81, 11, 53, 83, 72, 13, 77,
		61, 87, 56, 54, 78, 90, 9, 120,
		92, 107, 97, 29, 126, 86, 33, 23,
		48, 0, 10, 39, 70, 42, 79, 68,
		14, 118, 111, 21, 51, 4, 80, 66,
	}
	permutation96 := []int{
		43, 13, 76, 48, 44, 110, 67, 36,
		22, 46, 80, 30, 31, 66, 8, 11,
		101, 72, 68, 28, 87, 2, 26, 61,
		74, 14, 88, 33, 10, 9, 96, 57,
		104, 103, 83, 69, 4, 102, 99, 77,
		86, 20, 82, 62, 0, 70, 16, 29,
		64, 12, 49, 3, 111, 52, 40, 19,
		7, 108, 53, 60, 37, 93, 39, 79,
		73, 41, 54, 100, 109, 78, 63, 91,
		21, 84, 90, 106, 58, 23, 34, 97,
		1, 32, 51, 92, 98, 18, 27, 50,
		89, 47, 42, 81, 6, 94, 24, 56,
	}

	permutation1282 := []int{
		40, 59, 74, 60, 6, 67, 76, 70,
		0, 86, 22, 30, 86, 26, 56, 86,
		64, 61, 4, 47, 91, 23, 21, 35,
		92, 4, 79, 9, 12, 86, 6, 5,
		71, 84, 95, 75, 62, 6, 18, 8,
		25, 3, 65, 3, 54, 41, 52, 78,
		85, 1, 6, 63, 32, 3, 38, 34,
		57, 53, 31, 69, 76, 11, 26, 33,
		88, 4, 14, 51, 4, 45, 43, 26,
		55, 81, 80, 72, 56, 77, 42, 19,
		76, 58, 82, 7, 44, 66, 76, 76,
		26, 87, 76, 76, 28, 39, 10, 50,
		93, 2, 48, 13, 3, 20, 13, 36,
		89, 46, 13, 83, 37, 29, 49, 90,
		56, 4, 86, 16, 68, 4, 73, 6,
		76, 15, 56, 27, 17, 94, 24, 76,
	}
	const EOMLength = 16
	eom := []byte{234, 199, 233, 17, 128, 190, 27, 208, 137, 223, 186, 83, 41, 82, 107, 26}
	messageBytes = append(messageBytes, eom...)
	CConj := make([]byte, 0)

	// generate private key
	//privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}

	// marshal private key to bytes
	//privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	//fmt.Println("Private key: ", privateKeyBytes)
	//tst := x509.MarshalPKCS8PrivateKey(&privateKey)
	privateKeyBytes := []byte{48, 130, 2, 94, 2, 1, 0, 2, 129, 129, 0, 169, 220, 206, 138, 229, 219, 40, 160, 216, 236,
		57, 19, 232, 109, 185, 145, 220, 114, 38, 234, 241, 149, 113, 210, 149, 83, 41, 56, 98, 136, 70, 179, 16, 170,
		243, 93, 249, 205, 119, 51, 28, 221, 169, 131, 180, 248, 161, 167, 135, 15, 126, 194, 160, 228, 96, 36, 228,
		254, 80, 4, 136, 3, 95, 22, 104, 245, 195, 3, 116, 112, 73, 87, 172, 107, 255, 56, 69, 132, 211, 36, 24, 239,
		132, 153, 199, 26, 81, 18, 64, 188, 14, 91, 246, 208, 109, 64, 170, 66, 147, 254, 164, 39, 109, 162, 23, 137,
		181, 114, 68, 84, 126, 203, 225, 48, 171, 178, 25, 174, 199, 193, 225, 252, 63, 245, 59, 148, 108, 91, 2, 3, 1,
		0, 1, 2, 129, 129, 0, 133, 179, 112, 187, 177, 187, 52, 154, 142, 196, 57, 21, 43, 113, 26, 140, 238, 8, 200,
		11, 76, 6, 198, 165, 235, 181, 158, 143, 108, 55, 57, 246, 254, 90, 160, 222, 202, 158, 104, 129, 201, 233,
		203, 225, 8, 148, 95, 161, 158, 212, 154, 129, 21, 229, 76, 172, 29, 182, 243, 66, 237, 208, 65, 137, 248, 122,
		192, 162, 180, 77, 3, 111, 50, 115, 144, 227, 74, 243, 128, 247, 141, 26, 68, 198, 162, 59, 54, 204, 216, 160,
		166, 62, 177, 243, 54, 178, 161, 63, 173, 230, 77, 41, 23, 136, 58, 79, 255, 181, 230, 188, 153, 146, 18, 96,
		99, 190, 153, 7, 248, 15, 57, 99, 139, 144, 255, 170, 88, 105, 2, 65, 0, 198, 106, 196, 196, 190, 50, 130, 79,
		188, 12, 57, 28, 238, 200, 11, 214, 51, 41, 188, 252, 123, 239, 237, 97, 202, 32, 76, 70, 198, 80, 188, 28, 35,
		239, 60, 17, 7, 200, 54, 162, 250, 26, 108, 178, 104, 121, 221, 49, 106, 42, 89, 66, 186, 24, 12, 20, 147, 115,
		26, 89, 15, 116, 250, 157, 2, 65, 0, 219, 40, 154, 56, 70, 141, 117, 169, 108, 219, 246, 155, 152, 94, 178, 70,
		241, 201, 194, 170, 200, 59, 29, 14, 29, 240, 14, 14, 5, 185, 211, 154, 180, 208, 127, 35, 111, 143, 217, 196,
		24, 53, 55, 56, 196, 115, 214, 114, 153, 151, 171, 32, 208, 179, 78, 203, 117, 51, 125, 193, 40, 119, 245, 87,
		2, 65, 0, 132, 67, 210, 13, 48, 152, 108, 227, 136, 0, 65, 230, 54, 138, 101, 209, 144, 227, 134, 214, 108, 43,
		192, 251, 10, 9, 67, 175, 126, 45, 125, 103, 232, 208, 102, 35, 24, 35, 239, 191, 238, 166, 196, 196, 156, 254,
		119, 99, 164, 88, 188, 141, 205, 141, 144, 39, 251, 46, 164, 102, 175, 246, 19, 197, 2, 65, 0, 191, 136, 144,
		159, 182, 41, 83, 55, 171, 7, 226, 82, 193, 171, 161, 43, 23, 141, 57, 48, 128, 166, 9, 18, 153, 95, 127, 41,
		10, 32, 9, 171, 31, 115, 72, 105, 243, 202, 72, 139, 116, 140, 173, 162, 83, 46, 217, 176, 118, 67, 115, 47,
		206, 181, 166, 155, 113, 230, 122, 117, 33, 165, 21, 41, 2, 64, 53, 159, 93, 216, 165, 26, 36, 253, 116, 113, 5,
		119, 76, 214, 210, 194, 131, 246, 149, 233, 123, 74, 103, 181, 94, 17, 76, 17, 207, 55, 217, 104, 97, 111, 202,
		253, 123, 164, 18, 68, 55, 1, 134, 166, 166, 195, 47, 132, 163, 18, 189, 61, 189, 69, 121, 35, 16, 140, 140, 79,
		165, 85, 14, 46}
	//time.Sleep(10 * time.Second)
	//publicKeyBytes := x509.MarshalPKCS1PublicKey((*privateKey).Public().(*rsa.PublicKey))
	//fmt.Println("Public key: ", publicKeyBytes)

	publicKeyBytes := []byte{48, 129, 137, 2, 129, 129, 0, 169, 220, 206, 138, 229, 219, 40, 160, 216, 236, 57, 19, 232,
		109, 185, 145, 220, 114, 38, 234, 241, 149, 113, 210, 149, 83, 41, 56, 98, 136, 70, 179, 16, 170, 243, 93, 249,
		205, 119, 51, 28, 221, 169, 131, 180, 248, 161, 167, 135, 15, 126, 194, 160, 228, 96, 36, 228, 254, 80, 4,
		136, 3, 95, 22, 104, 245, 195, 3, 116, 112, 73, 87, 172, 107, 255, 56, 69, 132, 211, 36, 24, 239, 132, 153, 199,
		26, 81, 18, 64, 188, 14, 91, 246, 208, 109, 64, 170, 66, 147, 254, 164, 39, 109, 162, 23, 137, 181, 114, 68, 84,
		126, 203, 225, 48, 171, 178, 25, 174, 199, 193, 225, 252, 63, 245, 59, 148, 108, 91, 2, 3, 1, 0, 1}
	for len(CConj) < messLen {
		concatKey := make([]byte, 16)
		for i := 0; i < 8; i++ {
			concatKey[2*i] = privateKeyBytes[i]
			concatKey[2*i+1] = publicKeyBytes[i]
		}
		C0 := make([]byte, 2)
		for i := range concatKey {
			bits.SetBit(&C0, i, bits.GetBit(&concatKey, i*8+7))
		}

		permKey := make([]byte, 14)
		for i := 0; i < 112; i++ {
			bits.SetBit(&permKey, i, bits.GetBit(&concatKey, permutation128[i]))
		}

		cL := make([][]byte, 16)
		cR := make([][]byte, 16)
		for i := 0; i < 16; i++ {
			cL[i] = make([]byte, 7)
			cR[i] = make([]byte, 7)
		}
		copy(cL[0], permKey[:7])
		copy(cR[0], permKey[7:14])

		for i := 1; i < 16; i++ {
			copy(cL[i], cL[i-1])
			copy(cR[i], cR[i-1])
			if i == 1 || i == 2 || i == 9 {
				bits.LeftShift(&cL[i], 1)
				bits.LeftShift(&cR[i], 1)
			} else {
				bits.LeftShift(&cL[i], 2)
				bits.LeftShift(&cR[i], 2)
			}
		}
		cLR := make([][]byte, 15)
		for i := 0; i < 15; i++ {
			cLR[i] = make([]byte, 14)
			copy(cLR[i], cL[i+1])
			copy(cLR[i][7:], cR[i+1])
		}
		c := make([][]byte, 15)
		for i := 0; i < 15; i++ {
			c[i] = make([]byte, 12)
			for j := 0; j < 96; j++ {
				bits.SetBit(&c[i], j, bits.GetBit(&cLR[i], permutation96[j]))
			}
		}

		CXor := make([][]byte, 14)
		for j := 1; j < 15; j++ {
			CXor[j-1] = make([]byte, 14)
			for i := 0; i < 14; i++ {
				bits.SetBit(&CXor[j-1], i, bits.GetBit(&c[j-1], i)^bits.GetBit(&c[j], i))
			}
		}

		CChap := make([][]byte, 14)
		for j := 1; j < 15; j++ {
			CChap[j-1] = make([]byte, 16)
			for i := 0; i < 128; i++ {
				bits.SetBit(&CChap[j-1], i, bits.GetBit(&CXor[j-1], permutation1282[i]))
			}
			for i := 0; i < 16; i++ {
				CChap[j-1][i] ^= C0[i%2]
			}
		}

		for i := 0; i < 14; i++ {
			CConj = append(CConj, CChap[i]...)
		}
		if len(CConj) < messLen {
			privateKeyBytes = CChap[12][0:8]
			publicKeyBytes = CChap[13][8:16]
		}
		//fmt.Println(bits.DisplayBits(&concatKey))
		//fmt.Println(bits.DisplayBits(&permKey))
		//fmt.Println(bits.DisplayBits(&C0))
		//fmt.Println("cl & cr [0]")
		//fmt.Println(bits.DisplayBits(&cL[0]))
		//fmt.Println(bits.DisplayBits(&cR[0]))
		//fmt.Println("clr [0]")
		//fmt.Println(bits.DisplayBits(&cLR[0]))
		//fmt.Println("c [0]")
		//fmt.Println(bits.DisplayBits(&c[0]))
		//fmt.Println("cchap [0]")
		//fmt.Println(bits.DisplayBits(&CChap[0]))
	}
	fmt.Println(len(CConj), messLen)

	nWorkers := 1
	jobQueue := make(chan Job, nWorkers+1)
	resultQueue := make(chan []any, nWorkers+1)
	var wg sync.WaitGroup
	workers := make([]Worker, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = Worker{Id: i, JobQueue: jobQueue, ResultQueue: resultQueue, WG: &wg}
		go workers[i].run()
	}

	t1 := time.Now().UnixMicro()
	go func() {
		for i := 0; i < img.Rect.Dy(); i++ {
			wg.Add(1)
			jobQueue <- Job{f: substRange, args: []any{img, i}, OutputResults: true}
		}
	}()
	m := make(map[int][]int)
	for i := 0; i < img.Rect.Dy(); i++ {
		arr := <-resultQueue
		m[arr[0].(int)] = arr[1].([]int)
	}
	arr := make([]int, img.Rect.Dy()*img.Stride-img.Rect.Dy()*img.Rect.Dx())
	pos := 0
	for i := 0; i < img.Rect.Dy(); i++ {
		pos += copy(arr[pos:], m[i])
	}
	t2 := time.Now().UnixMicro()

	//split into 8x8 blocks
	n := img.Rect.Dy()
	p := n / 8
	nBlock := len(arr) / 64
	blocks := make([][][]int, nBlock)
	invBlocks := make([][][]int, nBlock)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 0; k < nBlock; k++ {
			wg.Add(1)

			jobQueue <- Job{f: break8x8, args: []any{&blocks, &arr, n, p, k}, OutputResults: false}
		}
	}()
	wg.Wait()
	t3 := time.Now().UnixMicro()
	//precalculate cosines values for dct
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
	t4 := time.Now().UnixMicro()
	//calculate the dct of each block
	dctBlocks := make([][][]float64, nBlock)
	invDCTBlocks := make([][][]float64, nBlock)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 0; k < nBlock; k++ {
			wg.Add(1)
			jobQueue <- Job{f: dct, args: []any{&blocks[k], &dctBlocks, &cosineArr, k}, OutputResults: false}
		}
	}()
	wg.Wait()
	t5 := time.Now().UnixMicro()
	//calculate entropy for each block
	entropy := make([]float64, nBlock)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 0; k < nBlock; k++ {
			wg.Add(1)
			jobQueue <- Job{f: calcEntropy, args: []any{&dctBlocks[k]}, OutputResults: true}
		}
	}()
	for i := 0; i < nBlock; i++ {
		entropy[i] = (<-resultQueue)[0].(float64)
	}

	meanEntropy := mathUtil.Mean(entropy)
	//initialize the Lohscheller Quantification matrix
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
	QDCTBlocks := make([][][]float64, nBlock)
	invQDCTBlocks := make([][][]float64, nBlock)
	iMess := 0
	l := 0
	for k := 0; k < nBlock; k++ {
		if entropy[k] > meanEntropy {
			QDCTBlocks[k] = make([][]float64, 8)
			invQDCTBlocks[k] = make([][]float64, 8)
			var factor float64
			if qualityFactor > 50 {
				factor = float64(100-qualityFactor) / 50.0
			} else {
				factor = 50.0 / float64(qualityFactor)
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 8; i++ {
					QDCTBlocks[k][i] = make([]float64, 8)
					invQDCTBlocks[k][i] = make([]float64, 8)
					for j := 0; j < 8; j++ {
						wg.Add(1)
						jobQueue <- Job{f: quantize, args: []any{&QDCTBlocks, &dctBlocks, &QMatrix, k, i, j, factor}, OutputResults: false}
					}
				}
			}()
			wg.Wait()
			coeffVector := make([]float64, 64)
			invCoeffVector := make([]float64, 64)
			copy(invCoeffVector, coeffVector)
			//zigzag scan
			x := 0
			y := 0
			dir := 1
			for i := 0; i < 64; i++ {
				coeffVector[i] = QDCTBlocks[k][x][y]
				if dir == 1 {
					if y == 0 && x != 7 {
						x++
						dir = 0
					} else if x == 7 {
						y++
						dir = 0
					} else {
						x++
						y--
					}
				} else {
					if x == 0 && y != 7 {
						y++
						dir = 1
					} else if y == 7 {
						x++
						dir = 1
					} else {
						x--
						y++
					}
				}
			}
			for j := 1; j < 8; j++ {
				if bits.GetBit(&CConj, l) == 1 {
					if mode == 0 {
						if coeffVector[j] < 0 {
							invCoeffVector[j] = float64(-R(-int(math.Round(coeffVector[j])), bits.GetBit(&messageBytes, iMess)))
						} else {
							invCoeffVector[j] = float64(R(int(math.Round(coeffVector[j])), bits.GetBit(&messageBytes, iMess)))
						}
					} else {
						if iMess/8 > len(messageDecodeBytes) {
							messageDecodeBytes = append(messageDecodeBytes, 0)
						}
						bits.SetBit(&messageDecodeBytes, iMess, invR(int(math.Abs(math.Round(coeffVector[j])))))
					}
					iMess++
				} else {
					invCoeffVector[j] = coeffVector[j]
				}
				l++
			}

			//inverse zigzag scan
			x = 0
			y = 0
			dir = 1
			for i := 0; i < 64; i++ {
				invQDCTBlocks[k][x][y] = invCoeffVector[i]
				if dir == 1 {
					if y == 0 && x != 7 {
						x++
						dir = 0
					} else if x == 7 {
						y++
						dir = 0
					} else {
						x++
						y--
					}
				} else {
					if x == 0 && y != 7 {
						y++
						dir = 1
					} else if y == 7 {
						x++
						dir = 1
					} else {
						x--
						y++
					}
				}
			}
			//multiply by the quantization matrix
			wg.Add(1)
			invDCTBlocks[k] = make([][]float64, 8)
			go func() {
				defer wg.Done()
				for i := 0; i < 8; i++ {
					invDCTBlocks[k][i] = make([]float64, 8)
					for j := 0; j < 8; j++ {
						wg.Add(1)
						jobQueue <- Job{f: dequantize, args: []any{&invQDCTBlocks, &invDCTBlocks, &QMatrix, k, i, j, factor}, OutputResults: false}
					}
				}
			}()
			wg.Wait()
			//inverse DCT
			wg.Add(1)
			go func() {
				defer wg.Done()
				wg.Add(1)
				invBlocks[k] = make([][]int, 8)
				jobQueue <- Job{f: invdct, args: []any{&invBlocks[k], &invDCTBlocks[k], &cosineArr, k}, OutputResults: false}
			}()
			wg.Wait()

		} else {
			invBlocks[k] = blocks[k]
		}
	}
	//merge blocks
	invArr := make([]int, len(arr))
	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 0; k < len(blocks); k++ {
			wg.Add(1)
			jobQueue <- Job{f: merge8x8, args: []any{&invBlocks, &invArr, n, p, k}, OutputResults: false}
		}
	}()
	wg.Wait()

	//draw image
	invImg := image.NewRGBA(img.Rect)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < len(invArr)/3; i++ {
			wg.Add(1)
			jobQueue <- Job{f: writeImg, args: []any{invImg, &invArr, i}, OutputResults: false}
		}
	}()
	wg.Wait()

	//save image
	file, _ := os.Create("output.png")
	png.Encode(file, invImg)

	t6 := time.Now().UnixMicro()
	fmt.Println("mean entropy:", meanEntropy)
	fmt.Println("Time taken:", t2-t1, "µs")
	fmt.Println("Time taken:", t3-t2, "µs")
	fmt.Println("Time taken:", t4-t3, "µs")
	fmt.Println("Time taken:", t5-t4, "µs")
	fmt.Println("Time taken:", t6-t5, "µs")
	fmt.Println("Total Time taken:", t6-t1, "µs")

	//SaveImage("./test/test.png", img)
}

func writeImg(args ...any) []any {
	invImg := args[0].(*image.RGBA)
	invArr := args[1].(*[]int)
	i := args[2].(int)
	invImg.Pix[i*4] = uint8((*invArr)[i] + 128)
	invImg.Pix[i*4+1] = uint8((*invArr)[i] + 128)
	invImg.Pix[i*4+2] = uint8((*invArr)[i] + 128)
	invImg.Pix[i*4+3] = 255
	return nil
}
