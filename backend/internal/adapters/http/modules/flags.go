package modules

import (
	flagcfg "github.com/pmaojo/kthulu-go/backend/internal/adapters/http/modules/flags"
	"go.uber.org/fx"
)

// FlagsModule provides configuration for request flags.
var FlagsModule = fx.Options(
	fx.Provide(flagcfg.LoadHeaderConfig),
)
