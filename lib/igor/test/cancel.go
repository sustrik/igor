package main

import __context "context"
import __igor "github.com/sustrik/igor/lib/igor"

import "fmt"
import "time"

import "github.com/sustrik/igor/lib/igor"

func __main(__ctx __context.Context) {
	n := __igor.NewNursery()
	defer n.Close(__ctx)
	{
		__nursery := n
		__nursery.Start__()
		go func() {
			__err := foo(__nursery.Context__())
			__nursery.Stop__(__err)
		}()
	}

	igor.Sleep(__ctx, 500*time.Millisecond)
}

func foo(__ctx __context.Context) error {
	err := igor.Sleep(__ctx, time.Second)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	fmt.Println("Done!")
	return nil
}
func main() {
	__main(__context.Background())
}
