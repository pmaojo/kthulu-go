// @kthulu:module:verifactu
// @kthulu:generated:true
package verifactu

import (
	"go.uber.org/fx"

	"kitchen-sink-test/internal/modules/verifactu/api"
	"kitchen-sink-test/internal/modules/verifactu/store"
	"kitchen-sink-test/internal/modules/verifactu/core"
)

// Providers returns the Fx providers for the verifactu module
func Providers() fx.Option {
        return fx.Options(
                fx.Provide(
                        store.NewVerifactuRepository,
                        core.NewVerifactuService,
                        api.NewVerifactuHandler,
                ),
        )
}
