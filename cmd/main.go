package main

import (
	"github.com/emrecanterzi/wisp/internal/api"
	"github.com/emrecanterzi/wisp/internal/memory"
	"github.com/emrecanterzi/wisp/internal/skiplist"
	"github.com/emrecanterzi/wisp/internal/wal"
)

func main() {
	API := api.NewAPI()

	skipList := skiplist.NewSkipList(4)
	wal, err := wal.NewWAL()
	if err != nil {
		panic(err)
	}
	mem := memory.NewMemory(skipList, wal)

	memoryHandler := memory.NewHandler(API, mem)
	memoryHandler.RegisterHandlers()

	err = API.Start()
	if err != nil {
		panic(err)
	}
}
