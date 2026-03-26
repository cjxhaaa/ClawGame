package jobs

import (
	"log"
	"time"
)

type Scheduler struct{}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Run() error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		log.Println("worker heartbeat")
		<-ticker.C
	}
}

