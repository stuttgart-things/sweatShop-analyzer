package main

import (
	"github.com/stuttgart-things/sweatShop-analyzer/internal"
	"github.com/stuttgart-things/sweatShop-analyzer/stream"
)

func main() {

	// PRINT BANNER + VERSION INFO
	internal.PrintBanner()

	// POLL STREAM
	stream.PollRedisStreams()

}
