package gemini

import (
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateEmbedding2Usage(t *testing.T) {
	for _, modality := range []string{"TEXT", "IMAGE", "DOCUMENT", "AUDIO"} {
		m := &dto.GeminiEmbeddingUsageMetadata{PromptTokenCount: 10, PromptTokenDetails: []dto.GeminiPromptTokensDetails{{Modality: modality, TokenCount: 10}}}
		require.NoError(t, validateEmbedding2Usage("gemini-embedding-2", m))
		if modality == "DOCUMENT" {
			require.Equal(t, "IMAGE", m.GeminiUsageMetadata().PromptTokensDetails[0].Modality)
			require.Equal(t, "DOCUMENT", m.PromptTokenDetails[0].Modality)
		}
	}
	require.Error(t, validateEmbedding2Usage("gemini-embedding-2", nil))
	require.Error(t, validateEmbedding2Usage("gemini-embedding-2", &dto.GeminiEmbeddingUsageMetadata{PromptTokenCount: 10}))
	require.Error(t, validateEmbedding2Usage("gemini-embedding-2", &dto.GeminiEmbeddingUsageMetadata{PromptTokenCount: 10, PromptTokenDetails: []dto.GeminiPromptTokensDetails{{Modality: "VIDEO", TokenCount: 10}}}))
	require.NoError(t, validateEmbedding2Usage("gemini-embedding-001", nil))
}
