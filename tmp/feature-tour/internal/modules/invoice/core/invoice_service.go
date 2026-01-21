// @kthulu:service:invoice
package core

type invoiceService struct {
	repo InvoiceRepository
}

func NewInvoiceService(repo InvoiceRepository) InvoiceService {
	return &invoiceService{repo: repo}
}

func (s *invoiceService) CreateInvoice(entity *Invoice) error {
	// Add business logic here
	return s.repo.Create(entity)
}

func (s *invoiceService) GetInvoiceByID(id uint) (*Invoice, error) {
	return s.repo.GetByID(id)
}

func (s *invoiceService) UpdateInvoice(entity *Invoice) error {
	// Add business logic here
	return s.repo.Update(entity)
}

func (s *invoiceService) DeleteInvoice(id uint) error {
	// Add business logic here
	return s.repo.Delete(id)
}

func (s *invoiceService) ListInvoices(filter SearchFilter) ([]*Invoice, error) {
	return s.repo.List(filter)
}
