package main

import (
	"log"

	"github.com/emrecanterzi/wisp/internal/api"
	"github.com/emrecanterzi/wisp/internal/memory"
	"github.com/emrecanterzi/wisp/internal/sstable"
	"github.com/emrecanterzi/wisp/internal/wal"
)

func main() {
	srv := api.NewAPI()

	w, err := wal.NewWAL("data/wal")
	if err != nil {
		log.Panic(err)
	}
	sm, err := sstable.NewSSTable("data/sstable")
	if err != nil {
		log.Panic(err)
	}
	mem := memory.NewMemory(w, sm)
	err = mem.Startup()
	if err != nil {
		log.Panic(err)
	}

	memoryHandler := memory.NewHandler(srv, mem)
	memoryHandler.RegisterHandlers()

	err = srv.Start()
	if err != nil {
		log.Panic(err)
	}
}
