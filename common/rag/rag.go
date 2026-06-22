package rag

import (
	"SamaraAI/common/redis"
	redisPkg "SamaraAI/common/redis"
	"SamaraAI/internal/config"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	embeddingArk "github.com/cloudwego/eino-ext/components/embedding/ark"
	redisIndexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	redisRetriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	redisCli "github.com/redis/go-redis/v9"
)

type RAGIndexer struct {
	embedding embedding.Embedder
	indexer   *redisIndexer.Indexer
	filename  string
}

type RAGQuery struct {
	embedding embedding.Embedder
	username  string
	filenames []string
}

func UserUploadDir(username string) string {
	cfg := config.Get()
	base := cfg.Rag.UploadDir
	if base == "" {
		base = "./uploads"
	}
	return filepath.Join(base, username)
}

func ragChunkSettings() (chunkSize, chunkOverlap int) {
	cfg := config.Get().Rag
	chunkSize = cfg.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 800
	}
	chunkOverlap = cfg.ChunkOverlap
	if chunkOverlap < 0 {
		chunkOverlap = 100
	}
	return chunkSize, chunkOverlap
}

// NewRAGIndexer 构建知识库索引
func NewRAGIndexer(filename, embeddingModel string) (*RAGIndexer, error) {
	ctx := context.Background()
	conf := config.Get()
	apiKey := conf.OpenAI.ApiKey
	dimension := conf.Rag.Dimension

	embedConfig := &embeddingArk.EmbeddingConfig{
		BaseURL: conf.Rag.BaseUrl,
		APIKey:  apiKey,
		Model:   embeddingModel,
	}
	embedder, err := embeddingArk.NewEmbedder(ctx, embedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	if err := redisPkg.InitRedisIndex(ctx, filename, dimension); err != nil {
		return nil, fmt.Errorf("failed to init redis index: %w", err)
	}

	rdb := redisPkg.Rdb
	indexerConfig := &redisIndexer.IndexerConfig{
		Client:    rdb,
		KeyPrefix: redis.GenerateIndexNamePrefix(filename),
		BatchSize: 10,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redisIndexer.Hashes, error) {
			source := ""
			if s, ok := doc.MetaData["source"].(string); ok {
				source = s
			}
			return &redisIndexer.Hashes{
				Key: fmt.Sprintf("%s:%s", filename, doc.ID),
				Field2Value: map[string]redisIndexer.FieldValue{
					"content":  {Value: doc.Content, EmbedKey: "vector"},
					"metadata": {Value: source},
				},
			}, nil
		},
	}
	indexerConfig.Embedding = embedder

	idx, err := redisIndexer.NewIndexer(ctx, indexerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	return &RAGIndexer{
		embedding: embedder,
		indexer:   idx,
		filename:  filename,
	}, nil
}

// IndexFile 读取文件内容、切块并创建向量索引
func (r *RAGIndexer) IndexFile(ctx context.Context, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	chunkSize, chunkOverlap := ragChunkSettings()
	chunks := chunkText(string(content), chunkSize, chunkOverlap)
	if len(chunks) == 0 {
		return fmt.Errorf("file is empty: %s", filePath)
	}

	docs := make([]*schema.Document, 0, len(chunks))
	for i, chunk := range chunks {
		docs = append(docs, &schema.Document{
			ID:      fmt.Sprintf("chunk_%d", i+1),
			Content: chunk,
			MetaData: map[string]any{
				"source": filePath,
				"chunk":  i + 1,
				"total":  len(chunks),
			},
		})
	}

	if _, err = r.indexer.Store(ctx, docs); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}
	return nil
}

// DeleteIndex 删除指定文件的知识库索引
func DeleteIndex(ctx context.Context, filename string) error {
	if err := redisPkg.DeleteRedisIndex(ctx, filename); err != nil {
		return fmt.Errorf("failed to delete redis index: %w", err)
	}
	return nil
}

// ListUserFilenames 列出用户已上传的知识库文件（内部存储名）
func ListUserFilenames(username string) ([]string, error) {
	files, err := ListUserFiles(username)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no valid file found for user %s", username)
	}
	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.FileID)
	}
	return names, nil
}

// NewRAGQuery 创建 RAG 查询器（支持检索用户全部已上传文档）
func NewRAGQuery(ctx context.Context, username string) (*RAGQuery, error) {
	cfg := config.Get()
	apiKey := cfg.OpenAI.ApiKey

	embedConfig := &embeddingArk.EmbeddingConfig{
		BaseURL: cfg.Rag.BaseUrl,
		APIKey:  apiKey,
		Model:   cfg.Rag.EmbeddingModel,
	}
	embedder, err := embeddingArk.NewEmbedder(ctx, embedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	filenames, err := ListUserFilenames(username)
	if err != nil {
		return nil, err
	}

	return &RAGQuery{
		embedding: embedder,
		username:  username,
		filenames: filenames,
	}, nil
}

// EnsureFileIndexed 若 Redis 索引缺失则从本地文件重建（例如更换 Redis 实例后）
func EnsureFileIndexed(ctx context.Context, username, fileID string) error {
	if redisPkg.IndexExists(ctx, fileID) {
		return nil
	}
	return reindexFile(ctx, username, fileID)
}

func reindexFile(ctx context.Context, username, fileID string) error {
	filePath := filepath.Join(UserUploadDir(username), fileID)
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("local file missing for %s: %w", fileID, err)
	}
	log.Printf("RAG reindex: rebuilding index for %s (user=%s)", fileID, username)
	_ = DeleteIndex(ctx, fileID)
	indexer, err := NewRAGIndexer(fileID, config.Get().Rag.EmbeddingModel)
	if err != nil {
		return err
	}
	return indexer.IndexFile(ctx, filePath)
}

func newRetriever(ctx context.Context, filename string, embedder embedding.Embedder) (retriever.Retriever, error) {
	rdb := redisPkg.Rdb
	indexName := redis.GenerateIndexName(filename)

	retrieverConfig := &redisRetriever.RetrieverConfig{
		Client:       rdb,
		Index:        indexName,
		Dialect:      2,
		ReturnFields: []string{"content", "metadata", "distance"},
		TopK:         5,
		VectorField:  "vector",
		DocumentConverter: func(ctx context.Context, doc redisCli.Document) (*schema.Document, error) {
			resp := &schema.Document{
				ID:       doc.ID,
				Content:  "",
				MetaData: map[string]any{},
			}
			for field, val := range doc.Fields {
				if field == "content" {
					resp.Content = val
				} else {
					resp.MetaData[field] = val
				}
			}
			return resp, nil
		},
	}
	retrieverConfig.Embedding = embedder
	return redisRetriever.NewRetriever(ctx, retrieverConfig)
}

func docDistance(doc *schema.Document) float64 {
	if doc == nil || doc.MetaData == nil {
		return 999
	}
	switch v := doc.MetaData["distance"].(type) {
	case string:
		d, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return d
		}
	case float64:
		return v
	case float32:
		return float64(v)
	}
	return 999
}

// RetrieveDocuments 从用户全部文档索引中检索最相关片段
func (r *RAGQuery) RetrieveDocuments(ctx context.Context, query string) ([]*schema.Document, error) {
	var allDocs []*schema.Document

	for _, filename := range r.filenames {
		if err := EnsureFileIndexed(ctx, r.username, filename); err != nil {
			log.Printf("EnsureFileIndexed %s failed: %v", filename, err)
			continue
		}

		rtr, err := newRetriever(ctx, filename, r.embedding)
		if err != nil {
			return nil, fmt.Errorf("failed to create retriever for %s: %w", filename, err)
		}
		docs, err := rtr.Retrieve(ctx, query)
		if err != nil {
			log.Printf("retrieve %s failed, retry after reindex: %v", filename, err)
			if reindexErr := reindexFile(ctx, r.username, filename); reindexErr == nil {
				if rtr, err = newRetriever(ctx, filename, r.embedding); err == nil {
					docs, err = rtr.Retrieve(ctx, query)
				}
			} else {
				log.Printf("reindex %s failed: %v", filename, reindexErr)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve documents from %s: %w", filename, err)
		}
		for _, doc := range docs {
			if doc.MetaData == nil {
				doc.MetaData = map[string]any{}
			}
			doc.MetaData["filename"] = filename
		}
		allDocs = append(allDocs, docs...)
	}

	sort.Slice(allDocs, func(i, j int) bool {
		return docDistance(allDocs[i]) < docDistance(allDocs[j])
	})

	topK := 5
	if len(allDocs) > topK {
		allDocs = allDocs[:topK]
	}
	return allDocs, nil
}

// BuildRAGPrompt 构建包含检索文档的提示词
func BuildRAGPrompt(query string, docs []*schema.Document) string {
	if len(docs) == 0 {
		return fmt.Sprintf(`用户问题：%s

说明：系统未检索到已上传文档中的相关内容。请明确告知用户：当前知识库中没有找到与问题匹配的文档片段，建议确认是否已上传文档并选择了 RAG 模式。不要假装已经阅读了文件。`, query)
	}

	contextText := ""
	for i, doc := range docs {
		source := ""
		if doc.MetaData != nil {
			if s, ok := doc.MetaData["source"].(string); ok {
				source = s
			}
		}
		if source != "" {
			contextText += fmt.Sprintf("[文档 %d | 来源 %s]: %s\n\n", i+1, source, doc.Content)
		} else {
			contextText += fmt.Sprintf("[文档 %d]: %s\n\n", i+1, doc.Content)
		}
	}

	return fmt.Sprintf(`基于以下参考文档回答用户的问题。如果文档中没有相关信息，请说明无法找到相关信息。

参考文档：
%s

用户问题：%s

请提供准确、完整的回答：`, contextText, query)
}
