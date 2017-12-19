package constantcontact

import (
	"context"
	"net/http"
	"time"
)

// ListService handles communication with the list related
// methods of the API.
//
// API docs: http://developer.constantcontact.com/docs/contact-list-api/contactlist-collection.html
type ListService service

// List represents a contact list
type List struct {
	ID           string     `json:"id,omitempty"`
	Name         string     `json:"name,omitempty"`
	Status       string     `json:"status,omitempty"`
	CreatedDate  *time.Time `json:"created_date,omitempty"`
	ModifiedDate *time.Time `json:"modified_date,omitempty"`
	ContactCount int        `json:"contact_count,omitempty"`
}

//GetAll returns all contact lists
func (s *ListService) GetAll(ctx context.Context) ([]*List, *http.Response, error) {
	u := "lists"

	req, err := s.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	var lists []*List
	resp, err := s.client.Do(ctx, req, &lists)
	if err != nil {
		return nil, resp, err
	}

	return lists, resp, nil
}
