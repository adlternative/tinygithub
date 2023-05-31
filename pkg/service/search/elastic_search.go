package search

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/gin-gonic/gin"
)

const gitBlobInfoIndexName = "gitblobs"

type GitBlobInfo struct {
	RepoName string `json:"repoName"`
	Revision string `json:"revision"`
	FilePath string `json:"filePath"`
	BlobID   string `json:"blobID"`
	Language string `json:"language"`

	Contents string `json:"contents"`
}

func (g *GitBlobInfo) ID() string {
	return fmt.Sprintf("%s:%s", g.RepoName, g.BlobID)
}

func NewInSensitiveTextProperty() *types.TextProperty {
	inSensitive := types.NewTextProperty()

	lowercase := "standard_lowercase"
	inSensitive.Analyzer = &lowercase

	return inSensitive
}

func NewTextProperty() *types.TextProperty {
	// text
	// match -> insensitive part ok
	// term -> part ok(only lowercase ok)
	// term -> full x
	return types.NewTextProperty()
}

func NewTextCaseProperty() *types.TextProperty {
	property := types.NewTextProperty()

	keywordAnalyzer := "ngram_analyzer"
	property.Analyzer = &keywordAnalyzer
	return property
}

func NewKeyWordsProperty() *types.KeywordProperty {
	property := types.NewKeywordProperty()
	ignoreAbove := 256
	property.IgnoreAbove = &ignoreAbove
	return property
}

func NewLowercaseKeyWordsProperty() *types.KeywordProperty {
	property := types.NewKeywordProperty()

	normalizer := "lowercase_normalizer"
	property.Normalizer = &normalizer

	ignoreAbove := 256
	property.IgnoreAbove = &ignoreAbove

	return property
}

func NewLowercaseNormalizer() *types.CustomNormalizer {
	property := types.NewCustomNormalizer()
	filter := "lowercase"

	property.Filter = append(property.Filter, filter)

	return property
}

func NewTextWithKeyWordsProperty() *types.TextProperty {
	property := types.NewTextProperty()
	property.Fields = map[string]types.Property{
		"keyword": NewKeyWordsProperty(),
	}
	return property
}

func NewLowerCaseAnalyzer() *types.CustomAnalyzer {
	standardLowercaseAnalyzer := types.NewCustomAnalyzer()
	standardLowercaseAnalyzer.Tokenizer = "standard"
	standardLowercaseAnalyzer.Filter = []string{"lowercase"}
	return standardLowercaseAnalyzer
}

func NewKeywordAnalyzer() *types.CustomAnalyzer {
	standardLowercaseAnalyzer := types.NewCustomAnalyzer()
	standardLowercaseAnalyzer.Tokenizer = "keyword"
	standardLowercaseAnalyzer.Filter = []string{"lowercase"}
	return standardLowercaseAnalyzer
}

func NewIndexSettings() *types.IndexSettings {
	setting := types.NewIndexSettings()
	setting.Analysis = types.NewIndexSettingsAnalysis()

	ngramAnalyzer := types.NewCustomAnalyzer()
	ngramAnalyzer.Tokenizer = "ngram_tokenizer"

	ngramTokenizer := types.NewNGramTokenizer()
	ngramTokenizer.MinGram = 2
	ngramTokenizer.MaxGram = 3

	setting.Analysis.Analyzer = map[string]types.Analyzer{
		"ngram_analyzer": ngramAnalyzer,
	}
	setting.Analysis.Tokenizer = map[string]types.Tokenizer{
		"ngram_tokenizer": ngramTokenizer,
	}
	setting.Analysis.Normalizer = map[string]types.Normalizer{
		"lowercase_normalizer": types.NewLowercaseNormalizer(),
	}

	return setting
}

func CreateIndex(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		es, err := elasticsearch.NewTypedClient(elasticsearch.Config{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		res, err := es.Indices.Create(gitBlobInfoIndexName).
			Request(&create.Request{
				Settings: NewIndexSettings(),
				Mappings: &types.TypeMapping{
					Properties: map[string]types.Property{
						"blobID":   NewKeyWordsProperty(),
						"revision": NewKeyWordsProperty(),
						"language": NewLowercaseKeyWordsProperty(),
						"repoName": NewTextWithKeyWordsProperty(),
						"filePath": NewTextWithKeyWordsProperty(),

						"contents": NewTextCaseProperty(),
					},
				},
			}).
			Do(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res)
		return
	}
}

func DeleteIndex(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		es, err := elasticsearch.NewTypedClient(elasticsearch.Config{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		res, err := es.Indices.Delete(gitBlobInfoIndexName).Do(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, res)
		return
	}
}

func CreateDocs(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var gitBlobInfo GitBlobInfo
		if err := c.BindJSON(&gitBlobInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

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

func QueryDocs(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var gitBlobInfo GitBlobInfo
		if err := c.BindJSON(&gitBlobInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		log.Debugf("query info: %v", gitBlobInfo)

		es, err := elasticsearch.NewTypedClient(elasticsearch.Config{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		boolQuery := types.BoolQuery{}

		if gitBlobInfo.RepoName != "" {
			boolQuery.Filter = append(boolQuery.Filter, types.Query{
				Term: map[string]types.TermQuery{"repoName.keyword": {Value: gitBlobInfo.RepoName}},
			})
		}
		if gitBlobInfo.FilePath != "" {
			boolQuery.Filter = append(boolQuery.Filter, types.Query{
				Term: map[string]types.TermQuery{"filePath.keyword": {Value: gitBlobInfo.FilePath}},
			})
		}
		if gitBlobInfo.Revision != "" {
			boolQuery.Filter = append(boolQuery.Filter, types.Query{
				Term: map[string]types.TermQuery{"revision": {Value: gitBlobInfo.Revision}},
			})
		}
		if gitBlobInfo.BlobID != "" {
			boolQuery.Filter = append(boolQuery.Filter, types.Query{
				Term: map[string]types.TermQuery{"blobID": {Value: gitBlobInfo.BlobID}},
			})
		}
		if gitBlobInfo.Language != "" {
			boolQuery.Filter = append(boolQuery.Filter, types.Query{
				Term: map[string]types.TermQuery{"language": {Value: gitBlobInfo.Language}},
			})
		}
		if gitBlobInfo.Contents != "" {
			boolQuery.Must = append(boolQuery.Must,
				types.Query{
					Match: map[string]types.MatchQuery{
						"contents": {Query: gitBlobInfo.Contents},
					},
				})
		}

		res, err := es.Search().Index(gitBlobInfoIndexName).TrackTotalHits("true").
			Request(&search.Request{
				Query: &types.Query{
					Bool: &boolQuery,
				},
			}).Do(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		gitblobs := []*GitBlobInfo{}
		for _, hit := range res.Hits.Hits {
			var result *GitBlobInfo

			err := json.Unmarshal(hit.Source_, &result)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			gitblobs = append(gitblobs, result)
		}
		c.JSON(http.StatusOK, gin.H{
			"gitblobs": gitblobs,
		})
	}
}
