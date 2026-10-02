package main

import (
	"context"
	"log"
	"time"

	client "github.com/nihitdev/Aether-Log/sdk/go"
)

func main() {
	c, err := client.New(client.Config{Address: "127.0.0.1:8080"})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.Send(ctx, []byte("hello from Go")); err != nil {
		log.Fatal(err)
	}
}
