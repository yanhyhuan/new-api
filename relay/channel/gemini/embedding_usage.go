package gemini

import (
	"fmt"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"strings"
)

// Never substitute an inaccurate local estimate for the multimodal embedding
// tariff. Unknown or missing modality details cannot be settled safely.
func validateEmbedding2Usage(model string, usage *dto.GeminiEmbeddingUsageMetadata) error {
	if model != "gemini-embedding-2" && model != "gemini-embedding-2-preview" {
		return nil
	}
	if usage == nil || usage.PromptTokenCount <= 0 {
		return fmt.Errorf("embedding usage metadata missing")
	}
	total := 0
	for _, d := range usage.PromptTokenDetails {
		switch strings.ToUpper(d.Modality) {
		case "TEXT", "IMAGE", "DOCUMENT", "AUDIO":
		default:
			return fmt.Errorf("unsupported embedding usage modality: %s", d.Modality)
		}
		if d.TokenCount < 0 {
			return fmt.Errorf("invalid embedding modality token count")
		}
		total += d.TokenCount
	}
	if total != usage.PromptTokenCount {
		return fmt.Errorf("incomplete embedding modality usage")
	}
	return nil
}
