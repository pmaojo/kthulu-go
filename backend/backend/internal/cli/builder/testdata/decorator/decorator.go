package decorator

import "github.com/kthulu/kthulu-go/backend/internal/cli/builder/testdata/interfaces"

func DecorateService(s interfaces.Service) interfaces.Service { return s }
