package git

import "context"

type MockClient struct {
	BlameData map[string]map[int]BlameInfo
	BlameErr  map[string]error
	ChurnData map[string]int
	ChurnErr  map[string]error
}

func NewMockClient() *MockClient {
	return &MockClient{
		BlameData: make(map[string]map[int]BlameInfo),
		BlameErr:  make(map[string]error),
		ChurnData: make(map[string]int),
		ChurnErr:  make(map[string]error),
	}
}

func (m *MockClient) Blame(_ context.Context, _, filePath string) (map[int]BlameInfo, error) {
	if err, ok := m.BlameErr[filePath]; ok {
		return nil, err
	}
	return m.BlameData[filePath], nil
}

func (m *MockClient) Churn(_ context.Context, _, filePath string) (int, error) {
	if err, ok := m.ChurnErr[filePath]; ok {
		return 0, err
	}
	return m.ChurnData[filePath], nil
}

func (m *MockClient) ValidateRepo(_ context.Context, _ string) error {
	return nil
}

var _ Client = (*MockClient)(nil)
