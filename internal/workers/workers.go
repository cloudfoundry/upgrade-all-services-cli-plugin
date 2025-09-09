package workers

import "sync"

type worker func()

func Run(count int, w worker) {
	var wg sync.WaitGroup
	wg.Add(count)

	for range count {
		go func() {
			w()
			wg.Done()
		}()
	}

	wg.Wait()
}
