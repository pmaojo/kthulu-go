package overrides

import "github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/builder/testdata/interfaces"

type MockService struct{}

func (MockService) Do() {}

func NewMockService() interfaces.Service { return MockService{} }
