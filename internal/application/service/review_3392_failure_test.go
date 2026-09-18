package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

func TestReview3392PresidioLateFailureKeepsIndex(t *testing.T) {
	for _, failAt := range []int{0, 2, 3} {
		t.Run(map[int]string{0: "success", 2: "body_failure", 3: "question_failure"}[failAt], func(t *testing.T) {
			calls := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls == failAt {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				_ = json.NewEncoder(w).Encode([]any{})
			}))
			defer srv.Close()
			kb := desensitizedMobileKB()
			kb.DesensitizationConfig.Engine = types.DesensitizationEnginePresidio
			engine := &capturingRetrieveEngine{}
			chunk := &types.Chunk{ID: "chunk", TenantID: 1, KnowledgeID: "knowledge", KnowledgeBaseID: "kb", Content: "body 13800138000", IsEnabled: true, ChunkType: types.ChunkTypeText}
			require.NoError(t, chunk.SetDocumentMetadata(&types.DocumentChunkMetadata{GeneratedQuestions: []types.GeneratedQuestion{{ID: "q1", Question: "question 13800138000"}}}))
			cfg := &config.Config{Desensitization: &config.DesensitizationRuntimeConfig{PresidioAnalyzerURL: srv.URL}}
			svc := &chunkService{knowledgeRepo: editableChunkKnowledgeRepo{}, kbRepository: desensitizeChunkKBRepo{kb: kb}, modelService: stubEmbeddingModelService{}, retrieveEngine: capturingRetrieveRegistry{engine: engine}, config: cfg}
			err := svc.syncChunkIndex(editorChunkContext(), chunk)
			if failAt != 0 {
				require.Error(t, err)
				require.Equal(t, failAt, calls)
				require.Empty(t, engine.deleted)
				require.Empty(t, engine.indexed)
			} else {
				require.NoError(t, err)
				require.Equal(t, 3, calls)
				require.Equal(t, []string{"chunk"}, engine.deleted)
				require.Len(t, engine.indexed, 2)
				for _, item := range engine.indexed {
					require.NotContains(t, item.Content, "13800138000")
				}
			}
		})
	}
}

func TestReview3392MaskFailureDoesNotPersistEdits(t *testing.T) {
	for _, op := range []string{"body", "question"} {
		t.Run(op, func(t *testing.T) {
			kb := desensitizedMobileKB()
			kb.DesensitizationConfig.Engine = types.DesensitizationEnginePresidio
			repo := &editableChunkRepo{chunk: &types.Chunk{ID: "chunk", TenantID: 1, KnowledgeID: "knowledge", KnowledgeBaseID: "kb", Content: "old body", SourceContent: "old body", ChunkType: types.ChunkTypeText, IsEnabled: true, IndexStatus: "ready"}}
			svc := &chunkService{chunkRepository: repo, knowledgeRepo: editableChunkKnowledgeRepo{}, kbRepository: desensitizeChunkKBRepo{kb: kb}}
			body := "13800138000"
			if op == "body" {
				_, err := svc.UpdateDocumentChunk(editorChunkContext(), "chunk", &body, nil, nil)
				require.Error(t, err)
			} else {
				_, err := svc.UpsertGeneratedQuestion(editorChunkContext(), "chunk", "", body)
				require.Error(t, err)
			}
			require.Equal(t, "old body", repo.chunk.Content)
			require.Zero(t, repo.chunk.ContentRevision)
			require.Empty(t, repo.chunk.Metadata)
		})
	}
}
