package constantcontact

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// ContactService handles communication with the contact related
// methods of the API.
//
// API docs: http://developer.constantcontact.com/docs/contacts-api/contacts-collection.html
type ContactService service

// Contact is a single contact in CC
type Contact struct {
	ID             string         `json:"id,omitempty"`
	Addresses      []Address      `json:"addresses,omitempty"`
	CellPhone      string         `json:"cell_phone,omitempty"`
	Confirmed      bool           `json:"confirmed,omitempty"`
	CreatedDate    *time.Time     `json:"created_date,omitempty"`
	EmailAddresses []EmailAddress `json:"email_addresses,omitempty"`
	ModifiedDate   *time.Time     `json:"modified_date,omitempty"`
	CompanyName    string         `json:"company_name,omitempty"`
	CustomFields   []CustomField  `json:"custom_fields,omitempty"`
	Fax            string         `json:"fax,omitempty"`
	FirstName      string         `json:"first_name,omitempty"`
	HomePhone      string         `json:"home_phone,omitempty"`
	JobTitle       string         `json:"job_title,omitempty"`
	LastName       string         `json:"last_name,omitempty"`
	Lists          []ContactList  `json:"lists,omitempty"`
	PrefixName     string         `json:"prefix_name,omitempty"`
	Source         string         `json:"source,omitempty"`
	SourceDetails  string         `json:"source_details,omitempty"`
	Status         string         `json:"status,omitempty"`
	WorkPhone      string         `json:"work_phone,omitempty"`
}

// Address is a contact address
type Address struct {
	ID            string `json:"id,omitempty"`
	AddressType   string `json:"address_type,omitempty"`
	City          string `json:"city,omitempty"`
	CountryCode   string `json:"country_code,omitempty"`
	Line1         string `json:"line1,omitempty"`
	Line2         string `json:"line2,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	State         string `json:"state,omitempty"`
	StateCode     string `json:"state_code,omitempty"`
	SubPostalCode string `json:"sub_postal_code,omitempty"`
}

//CustomField is a custom field in the contact.  There can be up to 15
type CustomField struct {
	Label string `json:"label,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

//EmailAddress is the email address of a contact.  It's only
//possible to create one through the API
type EmailAddress struct {
	ID            string     `json:"id,omitempty"`
	ConfirmStatus string     `json:"confirm_status,omitempty"`
	EmailAddress  string     `json:"email_address,omitempty"`
	OptInDate     *time.Time `json:"opt_in_date,omitempty"`
	OptInSource   string     `json:"opt_in_source,omitempty"`
	OptOutDate    *time.Time `json:"opt_out_date,omitempty"`
	OptOutSource  string     `json:"opt_out_source,omitempty"`
	Status        string     `json:"status,omitempty"`
}

// BulkImport is a structure for bulk importing contact data
// http://developer.constantcontact.com/docs/bulk_activities_api/bulk-activities-import-contacts.html
type BulkImport struct {
	ImportData  []interface{} `json:"import_data,omitempty"`
	ColumnNames []string      `json:"column_names,omitempty"`
	Lists       []string      `json:"lists,omitempty"`
}

// ContactList is a list the contact belongs to
type ContactList struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
}

// ImportResponse shows the result of a bulk import request
type ImportResponse struct {
	ID           string `json:"id,omitempty"`
	Type         string `json:"type,omitempty"`
	ErrorCount   int    `json:"error_count,omitempty"`
	ContactCount int    `json:"contact_count,omitempty"`
}

//resultsResponse is a temp interface to parse the
//"results" json key and return the actual results
type resultsResponse struct {
	Results []*Contact `json:"results,omitempty"`
}

// Create a contact
//
// http://developer.constantcontact.com/docs/contacts-api/contacts-collection.html?method=POST
func (s *ContactService) Create(ctx context.Context, contact *Contact) (*Contact, *Response, error) {
	u := "contacts"
	req, err := s.client.NewRequest("POST", u, contact)
	if err != nil {
		return nil, nil, err
	}
	c := new(Contact)
	resp, err := s.client.Do(ctx, req, c)
	if err != nil {
		return nil, resp, err
	}

	return c, resp, nil
}

//GetAll returns all contacts
func (s *ContactService) GetAll(ctx context.Context) ([]*Contact, *Response, error) {
	u := "contacts"

	return s.Get(ctx, u)
}

//Get calls the URL passed to it and attempts to convert to a list of contacts
func (s *ContactService) Get(ctx context.Context, nextURL string) ([]*Contact, *Response, error) {
	req, err := s.client.NewRequest("GET", nextURL, nil)
	if err != nil {
		return nil, nil, err
	}

	var record resultsResponse
	resp, err := s.client.Do(ctx, req, &record)
	if err != nil {
		return nil, resp, err
	}

	return record.Results, resp, nil
}

//Update modifies the contact according to the object it was sent
func (s *ContactService) Update(ctx context.Context, contact *Contact) (*Contact, *Response, error) {
	u := fmt.Sprintf("contacts/%s", contact.ID)
	req, err := s.client.NewRequest("PUT", u, contact)
	if err != nil {
		return nil, nil, err
	}

	c := new(Contact)
	resp, err := s.client.Do(ctx, req, c)
	if err != nil {
		return nil, resp, err
	}

	return c, resp, nil
}

//Import adds or updates multiple contacts
//
//http://developer.constantcontact.com/docs/bulk_activities_api/bulk-activities-import-contacts.html
func (s *ContactService) Import(ctx context.Context, contacts *BulkImport) (*ImportResponse, *Response, error) {
	u := "activities/addcontacts"

	req, err := s.client.NewRequest("POST", u, contacts)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to create request")
	}

	importResp := new(ImportResponse)
	resp, err := s.client.Do(ctx, req, importResp)
	if err != nil {
		return nil, resp, errors.Wrapf(err, "Bulk import failed")
	}

	return importResp, resp, nil
}
