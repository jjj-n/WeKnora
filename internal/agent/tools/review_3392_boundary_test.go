package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/modelcontext"
	"github.com/Tencent/WeKnora/internal/types"
)

func TestReview3392ReadDocumentModelBoundary(t *testing.T) {
	for _, mode := range []string{"manual_metadata", "url_source"} {
		t.Run(mode, func(t *testing.T) {
			tool, repo := newReadDocumentFixture(1)
			kb := &types.KnowledgeBase{ID: "kb-1", DesensitizationConfig: &types.DesensitizationConfig{Enabled: true, Engine: types.DesensitizationEngineBuiltin, EntityTypes: []string{types.DesensitizationEntityCNMobile}}}
			tool.knowledgeBaseService = stubKBLookup{kb: kb}
			doc := tool.knowledgeService.(*readDocKnowledgeService).docs["doc-1"]
			doc.Title = "合同 13800138000"
			repo.ordered[0].Content = "联系 <手机号>"
			if mode == "manual_metadata" {
				doc.Type = types.KnowledgeTypeManual
				if err := doc.SetManualMetadata(types.NewManualKnowledgeMetadata("原始正文 联系13800138000", "publish", 1)); err != nil {
					t.Fatal(err)
				}
			} else {
				doc.Type = "url"
				doc.Source = "https://example.com/contact/13800138000"
			}
			result, err := tool.Execute(context.Background(), json.RawMessage(`{"id":"doc-1","limit":1}`))
			if err != nil || !result.Success {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			output := modelcontext.NewRegistry(true).ModelToolResultForTool("read_document", result)
			if strings.Contains(output, "13800138000") {
				t.Fatalf("PII reached final model input despite masked chunk: %s", output)
			}
		})
	}
}
