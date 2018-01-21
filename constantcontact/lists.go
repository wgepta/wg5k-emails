package constantcontact

import (
	"context"
	"fmt"
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
func (s *ListService) GetAll(ctx context.Context) ([]*List, *Response, error) {
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

// Create a contact list.
//
// http://developer.constantcontact.com/docs/contact-list-api/contactlist-collection.html?method=POST
func (s *ListService) Create(ctx context.Context, list *List) (*List, *Response, error) {
	u := "lists"
	req, err := s.client.NewRequest("POST", u, list)
	if err != nil {
		return nil, nil, err
	}
	l := new(List)
	resp, err := s.client.Do(ctx, req, l)
	if err != nil {
		return nil, resp, err
	}

	return l, resp, nil
}

// Delete a list.
//
// http://developer.constantcontact.com/docs/contact-list-api/contactlist-resource.html?method=DELETE
func (s *ListService) Delete(ctx context.Context, id string) (*Response, error) {
	u := fmt.Sprintf("lists/%v", id)
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
