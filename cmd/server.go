package cmd

import (
	"log"

	"github.com/emrecanterzi/wisp/internal/api"
	"github.com/emrecanterzi/wisp/internal/memory"
	"github.com/emrecanterzi/wisp/internal/sstable"
	"github.com/emrecanterzi/wisp/internal/wal"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the wisp server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func runServer() error {
	srv := api.NewAPI()

	w, err := wal.NewWAL("data/wal")
	if err != nil {
		return err
	}
	sm, err := sstable.NewSSTable("data/sstable")
	if err != nil {
		return err
	}
	mem := memory.NewMemory(w, sm)
	if err := mem.Startup(); err != nil {
		return err
	}

	memoryHandler := memory.NewHandler(srv, mem)
	memoryHandler.RegisterHandlers()

	log.Println("wisp server starting")
	return srv.Start()
}
