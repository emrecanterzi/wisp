package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a value by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		return runGet(key)
	},
}

func runGet(key string) error {
	params := url.Values{}
	params.Set("key", key)
	fullURL := fmt.Sprintf("http://localhost:8080/?%s", params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println(string(body))
	return nil
}
