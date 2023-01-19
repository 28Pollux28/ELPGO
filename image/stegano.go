package image

import (
	mathUtil "Projet/utils"
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

func Main() {
	img := LoadImage("./test/webb.png")
	qualityFactor := 50

	nWorkers := 100
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
			var matrix [8][8]float64
			for i := 0; i < 8; i++ {
				for j := 0; j < 8; j++ {
					matrix[i][j] = float64(i*8 + j)
				}
			}
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
			//fmt.Println(coeffVector)
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
