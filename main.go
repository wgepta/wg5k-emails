package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/pkg/errors"
	"github.com/rbriski/wg5k/constantcontact"
	"github.com/rbriski/wg5k/racemine"
)

// Unregistered is the ID of the unreg list
const Unregistered = "1268645980"

// Registered is the ID of the reg list
const Registered = "1756200534"

// ContactList is a hash of contacts with the email address as the key
type ContactList map[string]*constantcontact.Contact

// func main() {
// 	apiKey := os.Getenv("CC_API_KEY")
// 	accessToken := os.Getenv("CC_ACCESS_TOKEN")
// 	context := context.Background()

// 	cc := constantcontact.NewClient(nil, apiKey, accessToken)
// 	lists, _, err := cc.Lists.GetAll(context)
// 	if err != nil {
// 		log.Fatalf("%v\n", err)
// 	}

// 	for _, list := range lists {
// 		fmt.Printf("%s [%s]: %s\n", list.Name, list.ID, list.Status)
// 	}
// }

// func main() {
// 	apiKey := os.Getenv("CC_API_KEY")
// 	accessToken := os.Getenv("CC_ACCESS_TOKEN")
// 	context := context.Background()

// 	cc := constantcontact.NewClient(nil, apiKey, accessToken)
// 	resp, err := cc.Lists.Delete(context, "1914568736")
// 	if err != nil {
// 		log.Fatalf("%v\n", err)
// 	}
// 	fmt.Printf("%s\n", resp.Status)
// }

func main() {
	cmd := os.Args[0:]

	if len(cmd) == 1 {
		log.Fatal("Please enter a command")
		os.Exit(1)
	}

	var err error
	if cmd[1] == "lists" {
		err = listCommand()
	}
	if cmd[1] == "load" {
		err = loadUnregistered()
	}
	if cmd[1] == "racers" {
		err = downloadRacers()
	}
	if cmd[1] == "update" {
		err = updateContacts()
	}
	if cmd[1] == "contacts" {
		err = syncToLocal()
	}
	if cmd[1] == "output" {
		err = contactCommand()
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(0)

	// contacts, _, err := cc.Contacts.GetAll(context)

	// addresses := make([]interface{}, len(contacts))
	// for ic, c := range contacts {
	// 	a := make([]string, len(c.EmailAddresses))
	// 	for i, e := range c.EmailAddresses {
	// 		a[i] = e.EmailAddress
	// 	}
	// 	addresses[ic] = map[string][]string{"email_addresses": a}
	// }
	// fmt.Printf("%v\n", addresses)

	// imp := &constantcontact.BulkImport{
	// 	ImportData:  addresses,
	// 	ColumnNames: []string{"Email Address"},
	// 	Lists:       []string{"2110141979"},
	// }
	// ir, resp, err := cc.Contacts.Import(context, imp)
	// if err != nil {
	// 	log.Fatalf("%v\n", err)
	// }

	// fmt.Printf("\n\n\n")
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Printf("%d => %v\n", resp.StatusCode, body)
	// fmt.Printf(" * * * * * * * * * * * \n")
	// fmt.Printf("%v\n", ir)
	// fmt.Printf(" - - -- - - - - \n")
	// fmt.Printf("%s\n", resp.Header.Get("Location"))

}

func newCC() (*constantcontact.Client, context.Context) {
	apiKey := os.Getenv("CC_API_KEY")
	accessToken := os.Getenv("CC_ACCESS_TOKEN")
	context := context.Background()

	cc := constantcontact.NewClient(nil, apiKey, accessToken)

	return cc, context
}

func loadUnregistered() error {
	cc, context := newCC()

	contacts, err := readContacts()
	if err != nil {
		return errors.Wrap(err, "could not read contacts")
	}

	var addresses []interface{}
	for address := range *contacts {
		addresses = append(addresses, map[string][]string{"email_addresses": []string{address}})
	}

	imp := &constantcontact.BulkImport{
		ImportData:  addresses,
		ColumnNames: []string{"Email Address"},
		Lists:       []string{Unregistered},
	}
	_, _, err = cc.Contacts.Import(context, imp)
	if err != nil {
		return errors.Wrap(err, "could not import contacts")
	}
	return nil
}

func listCommand() error {
	cc, context := newCC()
	lists, _, err := cc.Lists.GetAll(context)
	if err != nil {
		return errors.Wrap(err, "could not get all lists")
	}

	for _, list := range lists {
		if list.ID != Registered && list.ID != Unregistered {
			// cc.Lists.Delete(context, list.ID)
			fmt.Printf("%s [%s]: %s\n", list.Name, list.ID, list.Status)
		}
	}
	return nil
}

func syncToLocal() error {
	cc, context := newCC()

	contactList := make(map[string]*constantcontact.Contact)
	contacts, resp, err := cc.Contacts.GetAll(context)
	if err != nil {
		return errors.Wrap(err, "could not get contacts")
	}

	for {
		for idx, contact := range contacts {
			if idx%100 == 0 {
				fmt.Printf("*")
			}
			contactList[contact.EmailAddresses[0].EmailAddress] = contact
		}
		if resp.Next == "" {
			break
		}
		contacts, resp, err = cc.Contacts.Get(context, resp.Next)
		if err != nil {
			return errors.Wrap(err, "could not get next page of contacts")
		}
	}

	fmt.Println()
	fmt.Printf("Writing %d contacts to contacts.gob", len(contactList))
	err = writeGob("./contacts.gob", contactList)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func contactCommand() error {
	contacts, err := readContacts()
	if err != nil {
		return errors.Wrap(err, "could not read contacts")
	}
	fmt.Printf("EMAIL\tFNAME\tLNAME\n")
	for _, contact := range *contacts {
		fmt.Printf("%s\t%s\t%s\n", contact.EmailAddresses[0].EmailAddress, contact.FirstName, contact.LastName)
	}

	return nil
}

func loadRacers(fp string) (*ContactList, error) {
	xlsx, err := excelize.OpenFile(fp)
	if err != nil {
		return nil, errors.Wrap(err, "could not open racer Excel file")
	}

	// Get all the rows in the Sheet1.
	rows := xlsx.GetRows("Sheet1")
	header := make(map[int]string)

	racerContacts := make(ContactList)
	for i, row := range rows {
		if i == 0 {
			for idx, val := range row {
				header[idx] = val
			}
			continue
		}

		r := make(map[string]string)
		for idx, val := range row {
			r[header[idx]] = val
		}
		if _, ok := racerContacts[r["Email"]]; ok {
			fmt.Printf("%s has already been used, skipping.\n", r["Email"])
			continue
		}
		addItem := constantcontact.EmailAddress{
			EmailAddress: r["Email"],
		}
		address := []constantcontact.EmailAddress{addItem}

		racerContacts[r["Email"]] = &constantcontact.Contact{
			EmailAddresses: address,
			FirstName:      r["FirstName"],
			LastName:       r["LastName"],
		}
	}

	return &racerContacts, nil
}

func updateContacts() error {
	rf, err := getLatestRacers()
	if err != nil {
		return errors.Wrap(err, "could not get latest racer file")
	}

	racers, err := loadRacers(rf)
	if err != nil {
		return errors.Wrap(err, "could not load racer contacts")
	}

	// Get all existing contacts
	contacts, err := readContacts()
	if err != nil {
		return err
	}

	err = nil
	for rEmail, rContact := range *racers {
		if contact, ok := (*contacts)[rEmail]; ok {
			log.Printf("Updating contact: %s", rEmail)
			err = addExistingContact(contact)
		} else {
			log.Printf("Adding new contact: %s", rEmail)
			err = addNewContact(rContact)
		}
	}

	return err
}
func addNewContact(racer *constantcontact.Contact) error {
	cc, context := newCC()
	l := constantcontact.ContactList{ID: Registered, Status: "ACTIVE"}
	racer.Lists = []constantcontact.ContactList{l}
	fmt.Printf("%v\n", racer.EmailAddresses)
	_, _, err := cc.Contacts.Create(context, racer)
	if err != nil {
		return errors.Wrap(err, "could not create contact")
	}

	return nil
}

func addExistingContact(racer *constantcontact.Contact) error {
	cc, context := newCC()
	l := constantcontact.ContactList{ID: Registered, Status: "ACTIVE"}
	racer.Lists = []constantcontact.ContactList{l}
	_, _, err := cc.Contacts.Update(context, racer)
	if err != nil {
		return errors.Wrap(err, "could not update contact")
	}
	return nil
}

func downloadRacers() error {
	rc := racemine.NewClient(os.Getenv("RM_USERNAME"), os.Getenv("RM_PASSWORD"))
	links, err := rc.GetAllExports()
	if err != nil {
		return errors.Wrap(err, "could not get all RM export links")
	}

	err = rc.NewExport()
	if err != nil {
		return errors.Wrap(err, "could not create a new RM export")
	}
	for {
		fmt.Print("Checking if export has completed ... ")
		newLinks, err := rc.GetAllExports()
		if err != nil {
			return errors.Wrap(err, "could not get all RM export links")
		}
		if len(newLinks) > len(links) {
			fmt.Print("yes\n")
			links = newLinks
			break
		}
		fmt.Print("not yet\n")
		time.Sleep(30)
	}

	b, err := url.Parse("https://directors.racemine.com")
	if err != nil {
		return errors.Wrap(err, "Could not parse the RM URL")
	}
	for _, u := range links {
		l, err := url.Parse(u)
		if err != nil {
			return errors.Wrapf(err, "Could not parse RM URL (%s)", string(u))
		}

		f, err := download(b.ResolveReference(l))
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		fmt.Printf("Downloaded file to %s\n", f)
	}

	return nil
}

func getLatestRacers() (string, error) {
	files, err := ioutil.ReadDir("./downloads/")
	if err != nil {
		return "", errors.Wrap(err, "could not read downloads")
	}

	f := files[len(files)-1]
	return fmt.Sprintf("./downloads/%s", f.Name()), nil
}

func download(u *url.URL) (string, error) {
	p := "./downloads/"

	name := path.Base(u.Path)
	err := os.MkdirAll(p, 0755)
	if err != nil {
		return "", err
	}

	downloadPath := path.Join(p, name)
	out, err := os.Create(downloadPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return downloadPath, nil
}

func readContacts() (*ContactList, error) {
	contacts := new(ContactList)
	err := readGob("./contacts.gob", contacts)
	if err != nil {
		return nil, errors.Wrap(err, "could not read gob")
	}

	return contacts, nil
}

func writeGob(filePath string, object interface{}) error {
	file, err := os.Create(filePath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(object)
	}
	file.Close()
	return err
}

func readGob(filePath string, object interface{}) error {
	file, err := os.Open(filePath)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	file.Close()
	return err
}
