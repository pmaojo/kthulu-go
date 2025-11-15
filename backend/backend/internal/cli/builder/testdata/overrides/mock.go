package overrides

import "github.com/kthulu/kthulu-go/backend/internal/cli/builder/testdata/interfaces"

type MockService struct{}

func (MockService) Do() {}

func NewMockService() interfaces.Service { return MockService{} }
