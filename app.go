package main

import (
	"log"
	"os"
	"time"

	"github.com/acim/test-consul-leader-election/pkg/cloud/consul"
	"github.com/robfig/cron"
)

func main() {
	me := os.Getenv("APP_MY_ID")
	sn := os.Getenv("SERVICE_NAME")
	var client *consul.Client
	var err error
	for {
		client, err = consul.NewClient("consul:8500", sn, me, 0)
		if err != nil {
			log.Printf("%s can't connect to consul: %v\n", me, err)
			time.Sleep(3 * time.Second)
		}
	}

	err = client.Register()
	if err != nil {
		log.Printf("%s can't register to consul: %v\n", me, err)
	}

	c := cron.New()
	c.AddFunc("0 * * * * *", func() {
		log.Printf("%s cron triggered", me)
		lost, err := client.Lock()
		if err != nil {
			log.Printf("%s can't acquire lock: %v\n", me, err)
			return
		}
		log.Printf("%s acquired lock - doing something for 10 seconds", me)

		select {
		case <-lost:
			log.Printf("%s lost lock", me)
		default:
			time.Sleep(10 * time.Second)
			err = client.Unlock()
			if err != nil {
				log.Printf("%s can't release lock: %v\n", me, err)
			}
		}
	})
	c.Start()
	defer c.Stop()

	select {}
}
