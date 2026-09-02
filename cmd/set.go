package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a value by key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]
		return runSet(key, value)
	},
}

func runSet(key, value string) error {
	params := url.Values{}
	params.Set("key", key)
	params.Set("value", value)

	fullURL := fmt.Sprintf("http://localhost:8080/?%s", params.Encode())

	req, err := http.NewRequest("POST", fullURL, nil)
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

	fmt.Println("Value set successfully")
	return nil
}
