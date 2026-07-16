package main

import (
	"github.com/emrecanterzi/wisp/internal/api"
	"github.com/emrecanterzi/wisp/internal/memory"
	"github.com/emrecanterzi/wisp/internal/skiplist"
	"github.com/emrecanterzi/wisp/internal/wal"
)

func main() {
	srv := api.NewAPI()

	skipList := skiplist.NewSkipList(4)
	w, err := wal.NewWAL()
	if err != nil {
		panic(err)
	}
	mem := memory.NewMemory(skipList, w)
	err = mem.Startup()
	if err != nil {
		panic(err)
	}

	memoryHandler := memory.NewHandler(srv, mem)
	memoryHandler.RegisterHandlers()

	err = srv.Start()
	if err != nil {
		panic(err)
	}
}
