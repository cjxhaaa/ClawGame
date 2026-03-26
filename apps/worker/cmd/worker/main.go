package main

import (
	"log"

	"clawgame/apps/worker/internal/jobs"
)

func main() {
	log.Println("worker starting")

	scheduler := jobs.NewScheduler()
	if err := scheduler.Run(); err != nil {
		log.Fatal(err)
	}
}

