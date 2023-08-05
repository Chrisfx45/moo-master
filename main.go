package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/speecan/moo/game"
)

var (
	difficulty = 4    // moo digit number <= 10
	benchNum   = 1    // how many benchmark
	workers    = 16   // run workers at the same time
	queueSize  = 1000 // queue size for worker depend on PC memory size

	totalEstimates = 0
	totalDuration  = time.Duration(0)

	wg    = &sync.WaitGroup{}
	mutex = &sync.Mutex{}
)

func main() {
	// init
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	game.DebugMode = false
	queue := make(chan game.Estimate, queueSize)
	wg.Add(workers)
	totalHits := 0
	totalBlows := 0

	// run benchmark for moo estimater
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for fn := range queue {
				startTime := time.Now()
				count := runOnce(fn) // 1回数当てゲームを行う
				duration := time.Since(startTime)

				// かかった時間の集計
				mutex.Lock()
				totalEstimates += count
				totalDuration += duration
				mutex.Unlock()
			}
		}()
	}

	for n := 0; n < benchNum; n++ {
		// queue <- sample.EstimateWithRandom(difficulty)

		//「func(difficulty int) game.Estimate 」queueに入力した
		// Estimate with human code
		queue <- func(difficulty int) game.Estimate {

			return func(fn game.Question) (res []int) {
				var input string
				var hit int
				var blow int

				fmt.Print("?: ")
				fmt.Fscanln(os.Stdin, &input)
				guess := game.Str2Int(strings.Split(input, ""))
				hit, blow = fn(guess)
				//hit と　blow 回数がカウントされる
				totalHits += hit
				totalBlows += blow
				//一回答える時 hit と　blow 数を教える
				fmt.Println(guess, "Hit:", hit, "Blow:", blow)
				return guess
			}
		}(difficulty)

		//estimate with idiot algo

		/* queue <- func(difficulty int) game.Estimate {

			return func(fn game.Question) (res []int) {

				r := game.GetMooNum(difficulty)
				hit, blow := fn(r)
				totalHits += hit
				totalBlows += blow

				return r

			}

		}(difficulty)
		*/

		// estimate idot algo2
		/*
			queue <- func(difficulty int) game.Estimate {
				query := make([][]int, 0)
				isDuplicated := func(i []int) bool {
					for _, v := range query {
						if game.Equals(v, i) {
							return true
						}
					}
					return false
				}
				return func(fn game.Question) (res []int) {
					var r []int
					for {
						r = game.GetMooNum(difficulty)
						if !isDuplicated(r) {
							break
						}
					}
					hit, blow := fn(r)
					totalHits += hit
					totalBlows += blow
					query = append(query, r)
					return r
				}
			}(difficulty)
		*/
	}

	close(queue)
	wg.Wait()

	// result
	fmt.Println()
	fmt.Println("avg. spent:", totalDuration/time.Duration(benchNum), "avg. estimates count:", float64(totalEstimates)/float64(benchNum))
	fmt.Println("total Hits:", totalBlows, "Blows: ", totalHits)
}

func runOnce(estimateFn game.Estimate) int {
	// set game difficulty to 4
	g := game.NewGame(difficulty)

	count := 0
	q := g.GetQuestion(&count)
	fmt.Println("answer:", g.GetAnswer())
	for {
		// loop until hit the answer
		res := estimateFn(q)

		if g.Equals(res) {
			break
		}
	}

	fmt.Println("total questions:", count)

	return count
}
