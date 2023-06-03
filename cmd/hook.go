package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	hook_config "github.com/adlternative/tinygithub/pkg/config/hook"
	"github.com/adlternative/tinygithub/pkg/service/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

//var hookConfigFile string

// hookCmd represents the hook command
var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "transport git hook to server",
	Long:  `http client, transport git hook to server hook api`,
	Args:  cobra.ExactArgs(3),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return fmt.Errorf("viper bind hookCmd flags failed with %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := hook_config.Init()
		if err != nil {
			return fmt.Errorf("server hook config init failed: %v", err)
		}

		if !viper.GetBool(hook_config.PostReceiveMode) {
			return fmt.Errorf("only support post-receive hook now")
		}

		// 准备需要传递的参数
		request := &protocol.PostReceiveRequest{
			OldOid:  args[0],
			NewOid:  args[1],
			RefName: args[2],
			Repositry: &protocol.Repository{
				UserName: viper.GetString(hook_config.UserName),
				RepoName: viper.GetString(hook_config.RepoName),
			},
		}

		// 将参数转换为JSON格式
		jsonData, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON data: %w", err)
		}

		// 创建HTTP请求
		url := fmt.Sprintf("http://%s:%s/internal/post-receive", viper.GetString(hook_config.ServerIp), viper.GetString(hook_config.ServerPort))
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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
			log.Info(postReceiveResponse.Message)
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
	rootCmd.AddCommand(hookCmd)

	hookCmd.PersistentFlags().Bool(hook_config.PostReceiveMode, false, "run post-receive hook")
	hookCmd.PersistentFlags().String(hook_config.ServerIp, "localhost", "server ip")
	hookCmd.PersistentFlags().String(hook_config.ServerPort, "8083", "server port")
	hookCmd.PersistentFlags().String(hook_config.LogLevel, "info", "log level")
	hookCmd.PersistentFlags().String(hook_config.LogFile, "~/.tinygithub/hook/hook.log", "log file")
	hookCmd.PersistentFlags().String(hook_config.UserName, "", "user name")
	hookCmd.PersistentFlags().String(hook_config.RepoName, "", "repo name")
}
