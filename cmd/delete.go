package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [key]",
	Short: "Delete a value by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		return runDelete(key)
	},
}

func runDelete(key string) error {
	params := url.Values{}
	params.Set("key", key)

	fullURL := fmt.Sprintf("http://localhost:8080/?%s", params.Encode())

	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Key Deleted")
	return nil
}
