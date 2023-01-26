package stegano

import (
	"Projet/bits"
	"Projet/utils"
	"fmt"
	"image"
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
	for job, ok := <-w.JobQueue; ok && job.f != nil; job = <-w.JobQueue {
		result := job.f(job.args...)
		if job.OutputResults {
			w.ResultQueue <- result
		}
		w.WG.Done()
	}
}

func breakIn8ByteChunks(args ...any) []any {
	data := args[0].(*[]byte)
	chunks := args[1].(*[][]byte)
	chunkNum := args[2].(int)
	//(*chunks)[*chunkNum] = make([]byte, 8)
	(*chunks)[chunkNum] = (*data)[chunkNum*8 : utils.Min((chunkNum+1)*8, len(*data))]
	if len((*chunks)[chunkNum]) < 8 {
		(*chunks)[chunkNum] = append((*chunks)[chunkNum], make([]byte, 8-len((*chunks)[chunkNum]))...)
	}
	return nil
}

func applyPermutation(args ...any) []any {
	chunk := args[0].(*[]byte)
	newChunk := args[1].(*[]byte)
	perm := args[2].(*[]byte)
	for i := 0; i < 64; i++ {
		bits.SetBit(newChunk, i, bits.GetBit(chunk, int((*perm)[i])))
	}
	return nil
}

func applyRevPermutation(args ...any) []any {
	chunk := args[0].(*[]byte)
	newChunk := args[1].(*[]byte)
	perm := args[2].(*[]byte)
	for i := 0; i < 64; i++ {
		bits.SetBit(newChunk, int((*perm)[i]), bits.GetBit(chunk, i))
	}
	return nil
}

func applyXor(args ...any) []any {
	xorChunk := args[0].(*[]byte)
	newChunk := args[1].(*[]byte)
	keyChunk := args[2].(*[]byte)
	for i := 0; i < 8; i++ {
		(*xorChunk)[i] = (*newChunk)[i] ^ (*keyChunk)[i]
	}
	return nil
}

func writeChunkToImage(args ...any) []any {
	img := args[0].(*image.RGBA)
	chunk := args[1].(*[]byte)
	i := args[2].(int)
	sequence := args[3].(*[]int)
	for j := 0; j < 64; j += 1 {
		//fmt.Println("i", i, "j", j, "sequence", (*sequence)[j])
		bits.SetBit(&img.Pix, (*sequence)[j]+i*512, bits.GetBit(chunk, j))
	}
	return nil
}

func Encode(data []byte, key []byte, img *image.RGBA) (*image.RGBA, error) {
	if len(data)+4 > len(img.Pix)/8 {
		return nil, fmt.Errorf("data too long")
	}
	nWorkers := 12
	jobQueue := make(chan Job, nWorkers+1)
	resultQueue := make(chan []any, nWorkers+1)
	var wg sync.WaitGroup
	workers := make([]Worker, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = Worker{Id: i, JobQueue: jobQueue, ResultQueue: resultQueue, WG: &wg}
		go workers[i].run()
	}
	t1 := time.Now()
	perm := []byte{
		29, 42, 31, 40, 2, 17, 35, 55, 39, 53, 47, 26, 15, 6, 13, 49, 3, 7, 63, 52, 18, 16, 4, 14, 59, 9, 43, 51, 45,
		12, 46, 37, 56, 1, 25, 44, 48, 58, 21, 34, 27, 22, 33, 20, 19, 8, 23, 41, 32, 10, 62, 11, 36, 28, 0, 57,
		24, 60, 5, 50, 30, 38, 54, 61,
	}

	permKey := []byte{
		36, 54, 8, 46, 52, 25, 9, 14, 44, 18, 49, 47, 50, 32, 3, 7, 12, 13, 42, 15, 23, 41, 1, 20, 62, 5, 17, 0, 35,
		43, 22, 33, 30, 26, 19, 28, 59, 4, 63, 21, 58, 37, 27, 16, 53, 11, 31, 10, 61, 29, 55, 2, 57, 34, 39, 48,
		56, 60, 38, 6, 51, 24, 40, 45,
	}
	data = append(data, byte(26), byte(0), byte(26), byte(0))

	//number of chunks rounded up
	nChunks := (len(data) + 7) / 8
	// break data into 8 byte chunks
	chunks := make([][]byte, nChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i += 1 {
			chunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: breakIn8ByteChunks, args: []any{&data, &chunks, i}, OutputResults: false}
		}
	}()
	wg.Wait()
	nKeyChunks := (len(key) + 7) / 8
	keyChunks := make([][]byte, nKeyChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nKeyChunks; i += 1 {
			keyChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: breakIn8ByteChunks, args: []any{&key, &keyChunks, i}, OutputResults: false}
		}
	}()

	wg.Wait()

	newChunks := make([][]byte, nChunks)
	//Apply the permutation to the chunk bits
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i++ {
			newChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyPermutation, args: []any{&chunks[i], &newChunks[i], &perm}, OutputResults: false}
		}
	}()
	wg.Wait()

	newKeyChunks := make([][]byte, nKeyChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nKeyChunks; i++ {
			newKeyChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyPermutation, args: []any{&keyChunks[i], &newKeyChunks[i], &permKey}, OutputResults: false}
		}
	}()
	wg.Wait()

	//xor the data and key
	xorChunks := make([][]byte, nChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i++ {
			xorChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyXor, args: []any{&xorChunks[i], &newChunks[i], &newKeyChunks[i%nKeyChunks]}, OutputResults: false}
		}
	}()
	wg.Wait()

	sequence12 := []int{
		6, 7, 15, 23, 39, 46, 47, 55, 71, 79, 86, 87,
	}
	sequence64 := make([]int, 64)
	for i := 0; i < 64; i++ {
		sequence64[i] = sequence12[i%12] + (i/12)*96
	}

	//write the chunks to the image
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i++ {
			wg.Add(1)
			jobQueue <- Job{f: writeChunkToImage, args: []any{img, &xorChunks[i], i, &sequence64}, OutputResults: false}
		}
	}()
	wg.Wait()
	t2 := time.Now()
	fmt.Println("Time to encode:", t2.Sub(t1))
	return img, nil
}

func Decode(img *image.RGBA, key []byte) ([]byte, error) {
	nWorkers := 12
	jobQueue := make(chan Job, nWorkers+1)
	resultQueue := make(chan []any, nWorkers+1)
	var wg sync.WaitGroup
	workers := make([]Worker, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = Worker{Id: i, JobQueue: jobQueue, ResultQueue: resultQueue, WG: &wg}
		go workers[i].run()
	}
	t3 := time.Now()

	perm := []byte{
		29, 42, 31, 40, 2, 17, 35, 55, 39, 53, 47, 26, 15, 6, 13, 49, 3, 7, 63, 52, 18, 16, 4, 14, 59, 9, 43, 51, 45,
		12, 46, 37, 56, 1, 25, 44, 48, 58, 21, 34, 27, 22, 33, 20, 19, 8, 23, 41, 32, 10, 62, 11, 36, 28, 0, 57,
		24, 60, 5, 50, 30, 38, 54, 61,
	}

	permKey := []byte{
		36, 54, 8, 46, 52, 25, 9, 14, 44, 18, 49, 47, 50, 32, 3, 7, 12, 13, 42, 15, 23, 41, 1, 20, 62, 5, 17, 0, 35,
		43, 22, 33, 30, 26, 19, 28, 59, 4, 63, 21, 58, 37, 27, 16, 53, 11, 31, 10, 61, 29, 55, 2, 57, 34, 39, 48,
		56, 60, 38, 6, 51, 24, 40, 45,
	}

	sequence12 := []int{
		6, 7, 15, 23, 39, 46, 47, 55, 71, 79, 86, 87,
	}
	sequence64 := make([]int, 64)
	for i := 0; i < 64; i++ {
		sequence64[i] = sequence12[i%12] + (i/12)*96
	}

	//read the chunks from the image
	nChunks := len(img.Pix) / 64
	chunks := make([][]byte, nChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i++ {
			chunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: readChunkFromImage, args: []any{img, &chunks[i], i, &sequence64}, OutputResults: false}
		}
	}()
	wg.Wait()

	//break the key into 8 byte chunks
	nKeyChunks := (len(key) + 7) / 8
	keyChunks := make([][]byte, nKeyChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nKeyChunks; i += 1 {
			keyChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: breakIn8ByteChunks, args: []any{&key, &keyChunks, i}, OutputResults: false}
		}
	}()

	wg.Wait()

	newKeyChunks := make([][]byte, nKeyChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nKeyChunks; i++ {
			newKeyChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyPermutation, args: []any{&keyChunks[i], &newKeyChunks[i], &permKey}, OutputResults: false}
		}
	}()
	wg.Wait()

	//xor the data and key
	xorChunks := make([][]byte, nChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i++ {
			xorChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyXor, args: []any{&xorChunks[i], &chunks[i], &newKeyChunks[i%nKeyChunks]}, OutputResults: false}
		}
	}()
	wg.Wait()

	//apply the permutation to the data
	newChunks := make([][]byte, nChunks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i++ {
			newChunks[i] = make([]byte, 8)
			wg.Add(1)
			jobQueue <- Job{f: applyRevPermutation, args: []any{&xorChunks[i], &newChunks[i], &perm}, OutputResults: false}
		}
	}()
	wg.Wait()

	//concatenate the chunks
	data := make([]byte, nChunks*8)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < nChunks; i++ {
			wg.Add(1)
			jobQueue <- Job{f: concatenateChunks, args: []any{&data, &newChunks[i], i}, OutputResults: false}
		}
	}()
	wg.Wait()

	//trim the data if byte '26' is found
	for i := 0; i < len(data)-3; i++ {
		if data[i] == 26 && data[i+1] == 0 && data[i+2] == 26 && data[i+3] == 0 {
			t4 := time.Now()
			fmt.Println("time to decode ", t4.Sub(t3))
			return data[:i], nil
		}
	}
	t4 := time.Now()
	fmt.Println("time to decode ", t4.Sub(t3))
	return data, nil
}

func concatenateChunks(args ...any) []any {
	data := args[0].(*[]byte)
	chunk := args[1].(*[]byte)
	i := args[2].(int)
	for j := 0; j < 8; j++ {
		(*data)[i*8+j] = (*chunk)[j]
	}
	return nil
}

func readChunkFromImage(args ...any) []any {
	img := args[0].(*image.RGBA)
	chunk := args[1].(*[]byte)
	i := args[2].(int)
	sequence := args[3].(*[]int)
	for j := 0; j < 64; j += 1 {
		//fmt.Println("i", i, "j", j, "sequence", (*sequence)[j])
		bits.SetBit(chunk, j, bits.GetBit(&img.Pix, (*sequence)[j]+i*512))
	}
	return nil
}
