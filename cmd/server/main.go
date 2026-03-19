package main

import (
	"context"
)

func main() {
	if err := runServer(context.Background()); err != nil {
		panic(err)
	}
}
