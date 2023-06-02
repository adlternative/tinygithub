package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/service/protocol"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var (
	postReceive bool
)

// hookCmd represents the hook command
var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "transport git hook to server",
	Long:  `http client, transport git hook to server hook api`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !postReceive {
			return fmt.Errorf("only support post-receive hook now")
		}

		// 准备需要传递的参数
		request := &protocol.PostReceiveRequest{
			OldOid:  args[0],
			NewOid:  args[1],
			RefName: args[2],
		}

		// 将参数转换为JSON格式
		jsonData, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON data: %w", err)
		}

		// 创建HTTP请求
		req, err := http.NewRequest("POST", "http://localhost:8083/internal/post-receive", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// 设置HTTP请求头
		req.Header.Set("Content-Type", "application/json")

		// 发送HTTP请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send HTTP request: %w", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		switch resp.StatusCode {
		case http.StatusOK:
			postReceiveResponse := &protocol.PostReceiveResponse{}
			err = json.NewDecoder(resp.Body).Decode(&postReceiveResponse)
			if err != nil {
				return fmt.Errorf("failed to decode JSON data: %v", err)
			}
			fmt.Printf(postReceiveResponse.Message)
		default:
			postReceiveError := &protocol.PostReceiveError{}
			err = json.NewDecoder(resp.Body).Decode(&postReceiveError)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to decode JSON data: %v\n", err)
			}
			return fmt.Errorf("hook failed with: %s", postReceiveError.Error)
		}
		return nil
	},
}

func init() {
	hookCmd.PersistentFlags().BoolVar(&postReceive, "post-receive", false, "run post-receive hook")
	rootCmd.AddCommand(hookCmd)
}
