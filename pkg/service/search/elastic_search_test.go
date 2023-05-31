package search

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestDeleteIndex(t *testing.T) {
	r := require.New(t)

	// 创建 HTTP 请求
	req, err := http.NewRequest("DELETE", "http://localhost:8083/api/v2/_search_test/index", nil)
	r.NoError(err)

	// 发送 HTTP 请求并处理响应
	client := &http.Client{}
	resp, err := client.Do(req)
	r.NoError(err)

	_, err = io.Copy(os.Stdout, resp.Body)
	r.NoError(err)

	err = resp.Body.Close()
	r.NoError(err)
}

func TestCreateIndex(t *testing.T) {
	r := require.New(t)
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", "http://localhost:8083/api/v2/_search_test/index", nil)
	r.NoError(err)

	// 发送 HTTP 请求并处理响应
	client := &http.Client{}
	resp, err := client.Do(req)
	r.NoError(err)

	_, err = io.Copy(os.Stdout, resp.Body)
	r.NoError(err)

	err = resp.Body.Close()
	r.NoError(err)
}

func TestCreateDocs(t *testing.T) {
	r := require.New(t)

	// 创建一个包含多个 GitBlobInfo 实例的切片
	blobs := []GitBlobInfo{
		{
			RepoName: "my-repo",
			Revision: "main",
			FilePath: "path/to/file1",
			BlobID:   "abc123",
			Language: "go",
			Contents: "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
		},
		{
			RepoName: "my-repo",
			Revision: "dev",
			FilePath: "path/to/file2",
			BlobID:   "def456",
			Language: "python",
			Contents: "print('Hello, World!')",
		},
		{
			RepoName: "my-repo",
			Revision: "main",
			FilePath: "path/to/file3",
			BlobID:   "ghi789",
			Language: "java",
			Contents: "public class Main {\n    public static void main(String[] args) {\n        System.out.println(\"Hello, World!\");\n    }\n}",
		},
	}

	// 遍历 blobs 切片并发送 HTTP 请求
	for _, blob := range blobs {
		jsonBlob, err := json.Marshal(blob)
		r.NoError(err)

		// 创建 HTTP 请求
		req, err := http.NewRequest("POST", "http://localhost:8083/api/v2/_search_test/docs", bytes.NewBuffer(jsonBlob))
		r.NoError(err)

		req.Header.Set("Content-Type", "application/json")

		// 发送 HTTP 请求并处理响应
		client := &http.Client{}
		resp, err := client.Do(req)
		r.NoError(err)

		_, err = io.Copy(os.Stdout, resp.Body)
		r.NoError(err)

		err = resp.Body.Close()
		r.NoError(err)
	}
}

func TestQueryDocs(t *testing.T) {
	r := require.New(t)

	blob := &GitBlobInfo{
		RepoName: "my-repo",
		Contents: "main",
	}
	jsonBlob, err := json.Marshal(blob)
	r.NoError(err)

	req, err := http.NewRequest("POST", "http://localhost:8083/api/v2/_search_test/search", bytes.NewBuffer(jsonBlob))
	r.NoError(err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	r.NoError(err)

	_, err = io.Copy(os.Stdout, resp.Body)
	r.NoError(err)

	err = resp.Body.Close()
	r.NoError(err)
}
