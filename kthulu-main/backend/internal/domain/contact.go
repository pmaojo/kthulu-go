// @kthulu:module:contacts
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// Domain errors for contact module
var (
	ErrContactNotFound      = errors.New("contact not found")
	ErrContactAlreadyExists = errors.New("contact already exists")
	ErrInvalidContactType   = errors.New("invalid contact type")
	ErrContactInvalidEmail  = errors.New("invalid contact email address")
	ErrContactInvalidPhone  = errors.New("invalid contact phone number")
	ErrAddressNotFound      = errors.New("address not found")
	ErrPhoneNotFound        = errors.New("phone not found")
)

// ContactType represents the type of contact
type ContactType string

const (
	ContactTypeCustomer ContactType = "customer"
	ContactTypeSupplier ContactType = "supplier"
	ContactTypeLead     ContactType = "lead"
	ContactTypePartner  ContactType = "partner"
)

// AddressType represents the type of address
type AddressType string

const (
	AddressTypeBilling  AddressType = "billing"
	AddressTypeShipping AddressType = "shipping"
	AddressTypeOffice   AddressType = "office"
	AddressTypeHome     AddressType = "home"
	AddressTypeOther    AddressType = "other"
)

// PhoneType represents the type of phone number
type PhoneType string

const (
	PhoneTypeWork   PhoneType = "work"
	PhoneTypeMobile PhoneType = "mobile"
	PhoneTypeHome   PhoneType = "home"
	PhoneTypeFax    PhoneType = "fax"
	PhoneTypeOther  PhoneType = "other"
)

// Contact represents a business contact (customer, supplier, lead, or partner)
type Contact struct {
	ID             uint        `json:"id"`
	OrganizationID uint        `json:"organizationId" validate:"required"`
	Type           ContactType `json:"type" validate:"required,oneof=customer supplier lead partner"`
	CompanyName    string      `json:"companyName,omitempty" validate:"max=200"`
	FirstName      string      `json:"firstName,omitempty" validate:"max=100"`
	LastName       string      `json:"lastName,omitempty" validate:"max=100"`
	Email          string      `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone          string      `json:"phone,omitempty" validate:"max=20"`
	Mobile         string      `json:"mobile,omitempty" validate:"max=20"`
	Website        string      `json:"website,omitempty" validate:"omitempty,url,max=500"`
	TaxNumber      string      `json:"taxNumber,omitempty" validate:"max=50"`
	Notes          string      `json:"notes,omitempty"`
	IsActive       bool        `json:"isActive"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`

	// Related entities (loaded separately)
	Addresses []ContactAddress `json:"addresses,omitempty"`
	Phones    []ContactPhone   `json:"phones,omitempty"`
}

// ContactAddress represents an address for a contact
type ContactAddress struct {
	ID           uint        `json:"id"`
	ContactID    uint        `json:"contactId" validate:"required"`
	Type         AddressType `json:"type" validate:"required,oneof=billing shipping office home other"`
	AddressLine1 string      `json:"addressLine1" validate:"required,max=200"`
	AddressLine2 string      `json:"addressLine2,omitempty" validate:"max=200"`
	City         string      `json:"city" validate:"required,max=100"`
	State        string      `json:"state,omitempty" validate:"max=100"`
	Country      string      `json:"country" validate:"required,max=100"`
	PostalCode   string      `json:"postalCode,omitempty" validate:"max=20"`
	IsPrimary    bool        `json:"isPrimary"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
}

// ContactPhone represents a phone number for a contact
type ContactPhone struct {
	ID        uint      `json:"id"`
	ContactID uint      `json:"contactId" validate:"required"`
	Type      PhoneType `json:"type" validate:"required,oneof=work mobile home fax other"`
	Number    string    `json:"number" validate:"required,max=20"`
	Extension string    `json:"extension,omitempty" validate:"max=10"`
	IsPrimary bool      `json:"isPrimary"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewContact creates a new contact with validation
func NewContact(organizationID uint, contactType ContactType, companyName, firstName, lastName, email string) (*Contact, error) {
	contact := &Contact{
		OrganizationID: organizationID,
		Type:           contactType,
		CompanyName:    strings.TrimSpace(companyName),
		FirstName:      strings.TrimSpace(firstName),
		LastName:       strings.TrimSpace(lastName),
		Email:          strings.TrimSpace(strings.ToLower(email)),
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := contact.Validate(); err != nil {
		return nil, err
	}

	return contact, nil
}

// Validate validates the contact data
func (c *Contact) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}

	// Business rule: must have either company name or first/last name
	if c.CompanyName == "" && (c.FirstName == "" || c.LastName == "") {
		return errors.New("contact must have either company name or first and last name")
	}

	return nil
}

// GetDisplayName returns the display name for the contact
func (c *Contact) GetDisplayName() string {
	if c.CompanyName != "" {
		return c.CompanyName
	}
	return strings.TrimSpace(c.FirstName + " " + c.LastName)
}

// UpdateBasicInfo updates the basic contact information
func (c *Contact) UpdateBasicInfo(companyName, firstName, lastName, email, phone, mobile, website, taxNumber, notes string) error {
	c.CompanyName = strings.TrimSpace(companyName)
	c.FirstName = strings.TrimSpace(firstName)
	c.LastName = strings.TrimSpace(lastName)
	c.Email = strings.TrimSpace(strings.ToLower(email))
	c.Phone = strings.TrimSpace(phone)
	c.Mobile = strings.TrimSpace(mobile)
	c.Website = strings.TrimSpace(website)
	c.TaxNumber = strings.TrimSpace(taxNumber)
	c.Notes = strings.TrimSpace(notes)
	c.UpdatedAt = time.Now()

	return c.Validate()
}

// SetActive sets the active status of the contact
func (c *Contact) SetActive(active bool) {
	c.IsActive = active
	c.UpdatedAt = time.Now()
}

// ConvertToCustomer converts a lead to a customer
func (c *Contact) ConvertToCustomer() error {
	if c.Type != ContactTypeLead {
		return errors.New("only leads can be converted to customers")
	}
	c.Type = ContactTypeCustomer
	c.UpdatedAt = time.Now()
	return nil
}

// NewContactAddress creates a new contact address with validation
func NewContactAddress(contactID uint, addressType AddressType, addressLine1, addressLine2, city, state, country, postalCode string, isPrimary bool) (*ContactAddress, error) {
	address := &ContactAddress{
		ContactID:    contactID,
		Type:         addressType,
		AddressLine1: strings.TrimSpace(addressLine1),
		AddressLine2: strings.TrimSpace(addressLine2),
		City:         strings.TrimSpace(city),
		State:        strings.TrimSpace(state),
		Country:      strings.TrimSpace(country),
		PostalCode:   strings.TrimSpace(postalCode),
		IsPrimary:    isPrimary,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	validate := validator.New()
	if err := validate.Struct(address); err != nil {
		return nil, err
	}

	return address, nil
}

// Update updates the address information
func (ca *ContactAddress) Update(addressLine1, addressLine2, city, state, country, postalCode string) error {
	ca.AddressLine1 = strings.TrimSpace(addressLine1)
	ca.AddressLine2 = strings.TrimSpace(addressLine2)
	ca.City = strings.TrimSpace(city)
	ca.State = strings.TrimSpace(state)
	ca.Country = strings.TrimSpace(country)
	ca.PostalCode = strings.TrimSpace(postalCode)
	ca.UpdatedAt = time.Now()

	validate := validator.New()
	return validate.Struct(ca)
}

// SetPrimary sets this address as primary
func (ca *ContactAddress) SetPrimary(isPrimary bool) {
	ca.IsPrimary = isPrimary
	ca.UpdatedAt = time.Now()
}

// GetFullAddress returns the formatted full address
func (ca *ContactAddress) GetFullAddress() string {
	parts := []string{ca.AddressLine1}
	if ca.AddressLine2 != "" {
		parts = append(parts, ca.AddressLine2)
	}
	parts = append(parts, ca.City)
	if ca.State != "" {
		parts = append(parts, ca.State)
	}
	if ca.PostalCode != "" {
		parts = append(parts, ca.PostalCode)
	}
	parts = append(parts, ca.Country)
	return strings.Join(parts, ", ")
}

// NewContactPhone creates a new contact phone with validation
func NewContactPhone(contactID uint, phoneType PhoneType, number, extension string, isPrimary bool) (*ContactPhone, error) {
	phone := &ContactPhone{
		ContactID: contactID,
		Type:      phoneType,
		Number:    strings.TrimSpace(number),
		Extension: strings.TrimSpace(extension),
		IsPrimary: isPrimary,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	validate := validator.New()
	if err := validate.Struct(phone); err != nil {
		return nil, err
	}

	return phone, nil
}

// Update updates the phone information
func (cp *ContactPhone) Update(number, extension string) error {
	cp.Number = strings.TrimSpace(number)
	cp.Extension = strings.TrimSpace(extension)
	cp.UpdatedAt = time.Now()

	validate := validator.New()
	return validate.Struct(cp)
}

// SetPrimary sets this phone as primary
func (cp *ContactPhone) SetPrimary(isPrimary bool) {
	cp.IsPrimary = isPrimary
	cp.UpdatedAt = time.Now()
}

// GetFormattedNumber returns the formatted phone number with extension
func (cp *ContactPhone) GetFormattedNumber() string {
	if cp.Extension != "" {
		return cp.Number + " ext. " + cp.Extension
	}
	return cp.Number
}
