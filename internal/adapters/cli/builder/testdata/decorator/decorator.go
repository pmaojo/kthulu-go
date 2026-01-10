package decorator

import "github.com/pmaojo/kthulu-go/internal/adapters/cli/builder/testdata/interfaces"

func DecorateService(s interfaces.Service) interfaces.Service { return s }
