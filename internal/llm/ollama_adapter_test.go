package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaAdapter_Call(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		prompt         Prompt
		expectedResult string
		expectedError  string
	}{
		{
			name:           "Valid prompt returns response containing expected substring",
			responseBody:   `{"message": {"content": "The answer to 2+2 is 4."}}`,
			responseStatus: http.StatusOK,
			prompt: Prompt{
				System: "System message",
				User:   "What is 2+2?",
			},
			expectedResult: "4", // Check for substring in the result
			expectedError:  "",
		},
		{
			name:           "Empty prompt is invalid",
			responseBody:   "",
			responseStatus: http.StatusBadRequest,
			prompt: Prompt{
				System: "",
				User:   "",
			},
			expectedResult: "",
			expectedError:  "both system and user prompts cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json") // Ensure correct content type
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			adapter := &OllamaAdapter{
				baseURL:     server.URL,
				model:       "llama3.2:latest",
				maxTokens:   100,
				temperature: 0.7,
			}

			result, err := adapter.Call(context.Background(), tt.prompt)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedResult != "" {
				assert.Contains(t, result.Content, tt.expectedResult)
			} else {
				assert.Equal(t, tt.expectedResult, result.Content)
			}
		})
	}
}
func TestOllamaAdapter_Embeddings(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		texts          []string
		expectedLength int
		expectedError  string
	}{
		{
			name:           "Valid text returns embedding",
			responseBody:   `{"embedding": [0.1, 0.2, 0.3, 0.4, 0.5]}`,
			responseStatus: http.StatusOK,
			texts:          []string{"Hello world"},
			expectedLength: 5, // Length of the embedding vector
			expectedError:  "",
		},
		{
			name:           "Empty texts returns empty result",
			responseBody:   "",
			responseStatus: http.StatusOK,
			texts:          []string{},
			expectedLength: 0,
			expectedError:  "",
		},
		{
			name:           "API error returns error",
			responseBody:   `{"error": "Model not found"}`,
			responseStatus: http.StatusNotFound,
			texts:          []string{"Hello world"},
			expectedLength: 0,
			expectedError:  "Ollama embeddings API error for text 0: {\"error\": \"Model not found\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			adapter := &OllamaAdapter{
				baseURL:        server.URL,
				model:          "llama3.2:latest",
				embeddingModel: "nomic-embed-text:latest",
				maxTokens:      100,
				temperature:    0.7,
			}

			result, err := adapter.Embeddings(context.Background(), tt.texts)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedLength == 0 && len(tt.texts) == 0 {
				assert.Equal(t, [][]float64{}, result)
			} else if tt.expectedLength > 0 {
				assert.Len(t, result, len(tt.texts))
				if len(result) > 0 {
					assert.Len(t, result[0], tt.expectedLength)
				}
			}
		})
	}
}

func TestOllamaAdapter_SetEmbeddingModel(t *testing.T) {
	adapter := &OllamaAdapter{
		embeddingModel: "nomic-embed-text:latest",
	}

	// Test setting a new model
	adapter.SetEmbeddingModel("all-minilm:latest")
	assert.Equal(t, "all-minilm:latest", adapter.embeddingModel)

	// Test that empty string doesn't change the model
	adapter.SetEmbeddingModel("")
	assert.Equal(t, "all-minilm:latest", adapter.embeddingModel)
}