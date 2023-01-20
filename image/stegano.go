package image

import (
	"Projet/bits"
	mathUtil "Projet/utils"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"image"
	"math"
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

func idct(args ...any) []any {
	block := args[0].(*[][]int)
	dctBlock := args[1].(*[][]float64)
	for i := 0; i < 8; i++ {
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
	qualityFactor := args[3].(int)
	k := args[4].(int)
	i := args[5].(int)
	j := args[6].(int)

	if qualityFactor > 50 {
		(*QDCTBlocks)[k][i][j] = math.Round((*dctBlocks)[k][i][j] / float64((100-qualityFactor)/50*(*QMatrix)[i][j]))
	} else {
		(*QDCTBlocks)[k][i][j] = math.Round((*dctBlocks)[k][i][j] * float64(50/qualityFactor*(*QMatrix)[i][j]))
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

func Main() {
	img := LoadImage("./test/128.png")
	qualityFactor := 50
	message := "Hello, world!azertyuioihgfdsdfghjhgfdefghjiuytrfedfgvhjutfvgbhytrfvgbhytrfvbghytrfdcvgbhytrfcvgtrfdcvbgtrfdcfvgtrdcvfgtrfdcvgtrfdcvfgtrfdcvgytfdcvfgtrfdcvgtrfvgtyrfcvgytfvghuyhjkolhbjklhghjigfdcfvgbhdsdfqsdfresdfresxdcfgtfdcfvghgfvbnjkgbnjkgvbnjhgvfcderfghfdcvbhgfdrtghjygtfdsedrfghjiuygtfdsdefghjhgfdertghjk"
	message += message
	message += message
	messageBytes := []byte(message)
	messLen := len(messageBytes)

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

	CConj := make([]byte, 0)

	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}

	// marshal private key to bytes
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	//tst := x509.MarshalPKCS8PrivateKey(&privateKey)

	publicKeyBytes := x509.MarshalPKCS1PublicKey((*privateKey).Public().(*rsa.PublicKey))
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

	nWorkers := 12
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
	for k := 0; k < nBlock; k++ {
		if entropy[k] > meanEntropy {
			QDCTBlocks[k] = make([][]float64, 8)
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 8; i++ {
					QDCTBlocks[k][i] = make([]float64, 8)
					for j := 0; j < 8; j++ {
						wg.Add(1)
						jobQueue <- Job{f: quantize, args: []any{&QDCTBlocks, &dctBlocks, &QMatrix, qualityFactor, k, i, j}, OutputResults: false}
					}
				}
			}()
			wg.Wait()
			coeffVector := make([]float64, 64)
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
		}
	}

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
