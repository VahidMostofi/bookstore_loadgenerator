package main

func (lg *LoadGenerator) worker() {
	for r := range lg.RequestsQueue {
		r2, ok := lg.MakeRequest(r, false)
		if !ok {
			// lg.RequestsQueue <- r2
			// fmt.Println("POSTPONED", r2.Name)
		} else {
			lg.Results <- r2
		}
	}
	lg.DoneWorkers <- 0
}
