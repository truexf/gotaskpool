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

type mypriortask struct {
	priorty int
}

func (m *mypriortask) Run() {
	fmt.Printf("%s, mytask[%d] is running.\n", time.Now().String(), m.priorty)	
}

func main() {
{
		pool := gotaskpool.NewPriorTaskPool(5,1024,10)

		pool.Start()
		for {
			pool.PushTask(0,&mypriortask{priorty: 0}, -1)
			pool.PushTask(1,&mypriortask{priorty: 1}, -1)
			pool.PushTask(2,&mypriortask{priorty: 2}, -1)
			pool.PushTask(3,&mypriortask{priorty: 3}, -1)
			pool.PushTask(4,&mypriortask{priorty: 4}, -1)

			pool.PushTask(0,&mypriortask{priorty: 0}, -1)
			pool.PushTask(1,&mypriortask{priorty: 1}, -1)
			pool.PushTask(2,&mypriortask{priorty: 2}, -1)
			pool.PushTask(3,&mypriortask{priorty: 3}, -1)
			pool.PushTask(4,&mypriortask{priorty: 4}, -1)

			pool.PushTask(0,&mypriortask{priorty: 0}, -1)
			pool.PushTask(1,&mypriortask{priorty: 1}, -1)
			pool.PushTask(2,&mypriortask{priorty: 2}, -1)
			pool.PushTask(3,&mypriortask{priorty: 3}, -1)
			pool.PushTask(4,&mypriortask{priorty: 4}, -1)
			fmt.Println("===================")
			<-time.After(time.Second)
		}
		return 
	}

	{
		pool := gotaskpool.NewTaskPool(1024,10)
		pool.Start()
		for{
			pool.PushTask(&mytask{}, -1)
			//<-time.After(time.Second)	
		}

		return
	}


	

}