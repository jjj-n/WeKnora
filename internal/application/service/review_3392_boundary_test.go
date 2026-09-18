package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

type review3392KBRepo struct {
	interfaces.KnowledgeBaseRepository
	kb *types.KnowledgeBase
}

func (r review3392KBRepo) GetKnowledgeBaseByID(context.Context, string) (*types.KnowledgeBase, error) {
	return r.kb, nil
}

type review3392Engine struct{ parentChildRetrieveEngine }

func (e *review3392Engine) DeleteByChunkIDList(context.Context, []string, int, string) error {
	return nil
}

func TestReview3392ReindexMasksBodyAndQuestions(t *testing.T) {
	kb := enabledBuiltinKB()
	kb.IndexingStrategy = types.IndexingStrategy{VectorEnabled: true}
	kb.EmbeddingModelID = "embedding-1"
	engine := &review3392Engine{}
	svc := &chunkService{
		kbRepository:   review3392KBRepo{kb: kb},
		knowledgeRepo:  &parentChildKnowledgeRepo{knowledge: &types.Knowledge{ID: "doc-1", KnowledgeBaseID: kb.ID, Title: "合同 13800138000"}},
		modelService:   parentChildModelService{embedder: parentChildEmbedder{}},
		retrieveEngine: parentChildRetrieveRegistry{engine: engine},
	}
	tenant := &types.Tenant{ID: 1, RetrieverEngines: types.RetrieverEngines{Engines: []types.RetrieverEngineParams{{RetrieverType: types.VectorRetrieverType, RetrieverEngineType: types.PostgresRetrieverEngineType}}}}
	ctx := context.WithValue(context.Background(), types.TenantInfoContextKey, tenant)
	chunk := &types.Chunk{ID: "chunk-1", TenantID: 1, KnowledgeID: "doc-1", KnowledgeBaseID: kb.ID, ChunkType: types.ChunkTypeText, IsEnabled: true, Content: "编辑后的正文 联系13800138000"}
	if err := chunk.SetDocumentMetadata(&types.DocumentChunkMetadata{GeneratedQuestions: []types.GeneratedQuestion{{ID: "q-1", Question: "13800138000 是谁的电话？"}}}); err != nil {
		t.Fatal(err)
	}
	if err := svc.syncChunkIndex(ctx, chunk); err != nil {
		t.Fatal(err)
	}
	if len(engine.indexed) != 2 {
		t.Fatalf("expected body and question indexes, got %d", len(engine.indexed))
	}
	for _, item := range engine.indexed {
		if strings.Contains(item.Content, "13800138000") {
			t.Errorf("PII reached index input %s: %s", item.SourceID, item.Content)
		}
	}
}
