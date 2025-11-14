package overrides

import "backend/internal/cli/builder/testdata/interfaces"

type MockService struct{}

func (MockService) Do() {}

func NewMockService() interfaces.Service { return MockService{} }
