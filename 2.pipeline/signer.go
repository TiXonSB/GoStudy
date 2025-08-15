package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {

	var wg sync.WaitGroup
	var md5Mutex sync.Mutex

	for data := range in {
		wg.Add(1)

		go func(data interface{}) {
			defer wg.Done()

			strData := fmt.Sprintf("%v", data)

			crc32chan := make(chan string)
			go func(data string) {
				crc32chan <- DataSignerCrc32(data)
			}(strData)

			md5chan := make(chan string)
			go func(data string) {
				md5Mutex.Lock() // Защита от перегрева
				md5hash := DataSignerMd5(data)
				md5Mutex.Unlock()
				md5chan <- md5hash
			}(strData)
			md5hash := <-md5chan

			crc32md5Chan := make(chan string)
			go func(data string) {
				crc32md5Chan <- DataSignerCrc32(data)
			}(md5hash)

			crc32Data := <-crc32chan
			crc32md5Data := <-crc32md5Chan

			result := crc32Data + "~" + crc32md5Data
			out <- result

		}(data)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := sync.WaitGroup{}

	for input := range in {
		wg.Add(1)

		go func(data interface{}) {
			defer wg.Done()

			strData := fmt.Sprintf("%v", data)
			results := make([]string, 6)

			innerWg := sync.WaitGroup{}

			for th := 0; th < 6; th++ {
				innerWg.Add(1)

				go func(th int) {
					defer innerWg.Done()
					results[th] = DataSignerCrc32(fmt.Sprintf("%d%s", th, strData))
				}(th)
			}
			innerWg.Wait()

			out <- strings.Join(results, "")

		}(input)
	}
	wg.Wait()
}

// Комбинирует результаты
func CombineResults(in, out chan interface{}) {
	var results []string

	for input := range in {
		results = append(results, fmt.Sprintf("%v", input))
	}

	sort.Strings(results)
	// fmt.Println(strings.Join(results, "_")) // Отладочная печать
	out <- strings.Join(results, "_")
}

// конвеер
func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})

	for _, currentJob := range jobs {
		out := make(chan interface{})

		go func(job job, in, out chan interface{}) {
			defer close(out)
			job(in, out)
		}(currentJob, in, out)
		in = out
	}

	<-in
}
