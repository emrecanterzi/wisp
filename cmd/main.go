package main

import (
	"github.com/emrecanterzi/wisp/internal/api"
	"github.com/emrecanterzi/wisp/internal/memory"
	"github.com/emrecanterzi/wisp/internal/skiplist"
)

func main() {
	API := api.NewAPI()
	skipList := skiplist.NewSkipList(4)
	mem := memory.NewMemory(skipList)
	memoryHandler := memory.NewHandler(API, mem)
	memoryHandler.RegisterHandlers()

	API.Start()
}
