package search

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/gin-gonic/gin"
)

const gitBlobInfoIndexName = "gitblob"

type gitBlobInfo struct {
	RepoName string `json:"repoName"`
	Revision string `json:"revision"`
	FilePath string `json:"filePath"`
	BlobID   string `json:"blobID"`
	Language string `json:"language"`

	Contents []byte `json:"contents"`
}

func (g *gitBlobInfo) ID() string {
	return fmt.Sprintf("%s:%s", g.RepoName, g.BlobID)
}

func Index(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var gitBlobInfo gitBlobInfo
		if err := c.BindJSON(&gitBlobInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		contents := make([]byte, base64.StdEncoding.DecodedLen(len(gitBlobInfo.Contents)))
		_, err := base64.RawStdEncoding.Decode(contents, gitBlobInfo.Contents)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		gitBlobInfo.Contents = contents

		es, err := elasticsearch.NewTypedClient(elasticsearch.Config{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		response, err := es.Index(gitBlobInfoIndexName).
			Request(&gitBlobInfo).
			Id(gitBlobInfo.ID()).
			Refresh(refresh.Waitfor).
			Do(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}
}

func Query(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		repoName := c.Query("repoName")
		revision := c.Query("revision")
		filePath := c.Query("filePath")
		blobID := c.Query("blobID")
		language := c.Query("language")
		queryString := c.Query("query")

		es, err := elasticsearch.NewTypedClient(elasticsearch.Config{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		boolQuery := types.BoolQuery{}

		queryAppend(boolQuery.Filter, "repoName", repoName)
		queryAppend(boolQuery.Filter, "filePath", filePath)
		queryAppend(boolQuery.Filter, "revision", revision)
		queryAppend(boolQuery.Must, "blobID", blobID)
		queryAppend(boolQuery.Must, "language", language)
		queryAppend(boolQuery.Must, "contents", queryString)

		res, err := es.Search().Index(gitBlobInfoIndexName).TrackTotalHits("true").
			Request(&search.Request{
				Query: &types.Query{
					Bool: &boolQuery,
				},
			}).Do(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		c.JSON(http.StatusOK, res)
	}
}

func queryAppend(query []types.Query, prop, value string) []types.Query {
	if value != "" {
		query = append(query, types.Query{
			Term: map[string]types.TermQuery{prop: {Value: value}},
		})
	}
	return query
}
