package main

import "sync"
import "strconv"
import "sort"
import "fmt"

var md5QuotaChan chan struct{}

type MultiHashResultEntry struct {
	OrderNum int
	Crc32 string
}

func ExecutePipeline(jobs... job) {
	md5QuotaChan = make(chan struct{}, 1)
	outChans := make([]chan interface{}, len(jobs))
	
	for jobIx, _ := range jobs {
		outChans[jobIx] = make(chan interface{}, 10)
	}

	var currentIn chan interface{}
	for jobIx, currentJob := range jobs {
		go func(in, out chan interface{}, jobToRun job) {
			defer close(out)
			jobToRun(in, out)
		}(currentIn, outChans[jobIx], currentJob)
		currentIn = outChans[jobIx]
	}

	for _ = range outChans[len(outChans) - 1] {
	}
}

func SingleHash(in, out chan interface{}) {
	//DataSignerCrc32
	//crc32(data)+"~"+crc32(md5(data))
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go SingleHashWorker(data, out, wg)
	}
	wg.Wait()
}

func SingleHashWorker(data interface{}, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	dataInt, ok := data.(int)
	if !ok {
		panic("Conversion in SingleHash failed")
	}
	dataStr := strconv.Itoa(dataInt)
	fmt.Println("Calculating SingleHash for", dataStr)
	md5QuotaChan <- struct{}{}
	md5 := DataSignerMd5(dataStr)
	<-md5QuotaChan
	
	crc32plainIn := make(chan string)
	crc32md5In := make(chan string)

	crc32Func := func(sourceStr string, crc32out chan string) {
		crc32out <- DataSignerCrc32(sourceStr)
	}

	go crc32Func(dataStr, crc32plainIn)
	go crc32Func(md5, crc32md5In)

	dataCrc32 := <-crc32plainIn
	md5Crc32 := <-crc32md5In


	out <- dataCrc32 + "~" + md5Crc32
	fmt.Println("Done calculating SingleHash for", dataStr)
}

func MultiHash(in, out chan interface{}) {
	/*
	MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки), 
	где th=0..5 ( т.е. 6 хешей на каждое входящее значение ), потом берёт конкатенацию результатов в 
	порядке расчета (0..5), где data - то что пришло на вход (и ушло на выход из SingleHash)
	*/
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go MultiHashWorker(data, out, wg)
	}
	wg.Wait()
}

func MultiHashWorker(data interface{}, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	dataStr, ok := data.(string)
	if !ok {
		panic("Conversion in MultiHash failed")
	}

	fmt.Println("Calculating MultiHash for", dataStr)

	crc32in := make(chan MultiHashResultEntry)
	for i := 0; i <= 5; i++ {
		go Crc32Worker(i, dataStr, crc32in)
	}

	crcs := make([]MultiHashResultEntry, 6, 6)
	alreadyRead := 0
	for crcEntry := range crc32in {
		crcs = append(crcs, crcEntry)
		alreadyRead++
		if alreadyRead == 6 {
			break
		}
	}
	sort.Slice(crcs, func (i, j int) bool {
		return crcs[i].OrderNum <= crcs[j].OrderNum
	})
	var result string
	for _, crc := range crcs {
		result += crc.Crc32
	}
	
	fmt.Println("Done calculating MultiHash for", dataStr)
	out <- result
}

func Crc32Worker(orderNum int, dataStr string, out chan MultiHashResultEntry) {
	res := DataSignerCrc32(strconv.Itoa(orderNum) + dataStr)
	out <- MultiHashResultEntry {
		Crc32: res,
		OrderNum: orderNum,
	}
	fmt.Println("CRC32 Worker done for orderNum", orderNum)
}

func CombineResults(in, out chan interface{}) {
	fmt.Println("Starting combining results")
	/*
	CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/), 
	объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
	*/
	var resultSlice = make([]string, 0, 0)
	for dataItem := range in {
		dataStr, ok := dataItem.(string)
		if !ok {
			panic("Conversion in CombineResults failed")
		}
		fmt.Println("CombineResults got data item", dataStr)
		resultSlice = append(resultSlice, dataStr)
	}
	sort.Strings(resultSlice)
	var result string
	for ix, itemStr := range resultSlice {
		result += itemStr
		if ix != len(resultSlice) - 1 {
			result += "_"
		}
	}
	fmt.Println("Done combining results")
	out <- result
}