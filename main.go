package main

import (
	"github.com/stuttgart-things/sweatShop-analyze/internal"

	"github.com/stuttgart-things/sweatShop-analyze/stream"
)

func main() {
	// PRINT BANNER + VERSION INFO
	internal.PrintBanner()
	stream.PollRedisStreams()

}
