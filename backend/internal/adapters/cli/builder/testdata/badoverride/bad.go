package badoverride

// BadService does not implement interfaces.Service.
type BadService struct{}

func (BadService) SomethingElse() {}

func NewBadService() interface{} { return BadService{} }
