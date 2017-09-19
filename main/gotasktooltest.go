package main

import (
	"github.com/truexf/gotaskpool"
	"fmt"
	"time"
)

type mytask struct {

}

func (m *mytask) Run() {
	fmt.Printf("%s, mytask is running.\n", time.Now().String())
}

func main() {
	pool := gotaskpool.NewTaskPool(1024,10)
	pool.Start()
	for{
		pool.PushTask(&mytask{}, -1)
		//<-time.After(time.Second)	
	}

	return
	

}