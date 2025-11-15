package decorator

import "github.com/pmaojo/kthulu-go/backend/internal/cli/builder/testdata/interfaces"

func DecorateService(s interfaces.Service) interfaces.Service { return s }
