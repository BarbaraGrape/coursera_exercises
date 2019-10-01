package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	for _, j := range jobs {
		out := make(chan interface{})
		go j(in, out)
		//time.Sleep(20 * time.Millisecond)
		in = out
	}
	for _ = range in {
	}
}

type HashOrigin struct {
	Origin string
	Hash   string
}

func Md5Hash(in <-chan string) chan HashOrigin {
	out := make(chan HashOrigin)

	go func() {
		for srcData := range in {
			resultStr := DataSignerMd5(srcData)
			out <- HashOrigin{srcData, resultStr}
			fmt.Println("Data: ", srcData, "SingleHash result: ", resultStr)
		}
		close(out)
	}()
	return out

}

func SingleHash(in <-chan HashOrigin) chan string {
	out := make(chan string)
	go func() {
		for hashOrigin := range in {
			var wg sync.WaitGroup
			hashes := make([]string, 2)
			fmt.Println("DATA: ", hashOrigin.Origin)
			wg.Add(2)
			go func() {
				hashes[1] = DataSignerCrc32(hashOrigin.Hash)
				fmt.Println("CRC + MD5: ", hashes[1])
				wg.Done()
			}()

			go func() {
				hashes[0] = DataSignerCrc32(hashOrigin.Origin)
				fmt.Println("CRC: ", hashes[0])
				wg.Done()
			}()
			wg.Wait()

			resultStr := strings.Join(hashes, "~")
			out <- resultStr
			fmt.Println("Data: ", hashOrigin.Origin, "SingleHash result: ", resultStr)
		}
		close(out)
	}()
	return out

}

func MultiHash(in chan string) chan string {

	type Result struct {
		Index int
		Hash  string
	}
	out := make(chan string)
	go func() {
		for srcData := range in {
			//var hashIterator int64
			//srcData = srcData.(string)
			ch := make(chan Result, 1)
			for i := 0; i < 6; i++ {
				go func(index int) {
					s := strconv.Itoa(index) + srcData
					s = DataSignerCrc32(s)
					fmt.Println("Data: ", srcData, " \tStep: ", index, "\tResult", s)
					ch <- Result{index, s}
				}(i)
			}
			var res []Result
			for i := 0; i < 6; i++ {
				res = append(res, <-ch)
			}
			sort.Slice(res, func(i, j int) bool {
				return res[i].Index < res[j].Index
			})
			resultSlice := make([]string, len(res))
			for i, r := range res {
				resultSlice[i] = r.Hash
			}
			resMultiHash := strings.Join(resultSlice, "")
			//fmt.Println(res)
			//fmt.Println(resMultiHash)
			out <- resMultiHash
			fmt.Println("Data: ", srcData, "\tMultihash result: ", resMultiHash)
		}
		close(out)
	}()
	return out
}

func CombineResults(in chan string) string {
	combineResult := make([]string, 0)
	for srcData := range in {
		combineResult = append(combineResult, srcData)
	}
	sort.Slice(combineResult, func(i, j int) bool {
		return combineResult[i] < combineResult[j]
	})
	return strings.Join(combineResult, "_")

}

func gen(inputData []int) chan string {
	out := make(chan string)
	go func() {
		for _, num := range inputData {
			out <- strconv.Itoa(num)
		}
		close(out)
	}()
	return out
}

func merge(chans ...chan string) chan string {
	out := make(chan string)
	var wg sync.WaitGroup
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(c chan string) {
			for value := range c {
				out <- value
			}
			wg.Done()
		}(ch)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	//fmt.Println(DataSignerMd5(strconv.Itoa(0)))
	inputData := []int{0, 1, 1, 2, 3}
	start := time.Now()

	in := gen(inputData)
	md5 := Md5Hash(in)
	outChannels := make([]chan string, len(inputData))
	for i := 0; i < len(inputData); i++ {
		singleHash := SingleHash(md5)
		multiHash := MultiHash(singleHash)
		outChannels[i] = multiHash
	}
	mergeHash := merge(outChannels...)
	combineResult := CombineResults(mergeHash)
	fmt.Println(combineResult)
	end := time.Since(start)
	fmt.Println(end)
}
