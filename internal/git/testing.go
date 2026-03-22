package git

import "context"

// MockClient is a test double for Client. Set BlameData and ChurnData to
// control what each method returns for a given file path. Set BlameErr or
// ChurnErr to simulate failures.
type MockClient struct {
	BlameData map[string]map[int]BlameInfo
	BlameErr  map[string]error
	ChurnData map[string]int
	ChurnErr  map[string]error
}

// NewMockClient returns a MockClient with all maps initialised.
func NewMockClient() *MockClient {
	return &MockClient{
		BlameData: make(map[string]map[int]BlameInfo),
		BlameErr:  make(map[string]error),
		ChurnData: make(map[string]int),
		ChurnErr:  make(map[string]error),
	}
}

// Blame returns the pre-loaded data for filePath, or an error if one is set.
func (m *MockClient) Blame(_ context.Context, _, filePath string) (map[int]BlameInfo, error) {
	if err, ok := m.BlameErr[filePath]; ok {
		return nil, err
	}
	return m.BlameData[filePath], nil
}

// Churn returns the pre-loaded count for filePath, or an error if one is set.
func (m *MockClient) Churn(_ context.Context, _, filePath string) (int, error) {
	if err, ok := m.ChurnErr[filePath]; ok {
		return 0, err
	}
	return m.ChurnData[filePath], nil
}

// ValidateRepo always succeeds in the mock.
func (m *MockClient) ValidateRepo(_ context.Context, _ string) error {
	return nil
}

// Ensure MockClient satisfies the Client interface at compile time.
var _ Client = (*MockClient)(nil)
