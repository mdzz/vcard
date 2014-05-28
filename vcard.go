package vcard

import (
	"io/ioutil"
	"log"
	"strings"
)

type VCard struct {
	Version           string
	FormattedName     string
	FamilyNames       []string
	GivenNames        []string
	AdditionalNames   []string
	HonorificNames    []string
	HonorificSuffixes []string
	NickNames         []string
	Photo             Photo
	Birthday          string
	Addresses         []Address
	Telephones        []Telephone
	Emails            []Email
	Title             string
	Role              string
	Org               []string
	Categories        []string
	Note              string
	URL               string
	XJabbers          []XJabber
	// mac specific
	XABuid    string
	XABShowAs string
}

func displayStrings(ss []string) (display string) {
	for _, s := range ss {
		display += s + ", "
	}
	return display
}

func (v VCard) String() (s string) {
	s = "VCard version: " + v.Version + "\n"
	s += "FormattedName:" + v.FormattedName + "\n"
	s += "FamilyNames:" + displayStrings(v.FamilyNames) + "\n"
	s += "GivenNames:" + displayStrings(v.GivenNames) + "\n"
	s += "AdditionalNames:" + displayStrings(v.AdditionalNames) + "\n"
	return s
}

type Photo struct {
	Encoding string
	Type     string
	Value    string
	Data     string
}

func defaultAddressTypes() (types []string) {
	return []string{"Intl", "Postal", "Parcel", "Work"}
}

type DataType interface {
	GetType() []string
	HasType(t string) bool
}

type Address struct {
	Type            []string // default is Intl,Postal,Parcel,Work
	Label           string
	PostOfficeBox   string
	ExtendedAddress string
	Street          string
	Locality        string // e.g: city
	Region          string // e.g: state or province
	PostalCode      string
	CountryName     string
}

type Telephone struct {
	Type   []string
	Number string
}

type Email struct {
	Type    []string
	Address string
}

type XJabber struct {
	Type    []string
	Address string
}

const ( // Constant define address information index in directory information StructuredValue
	familyNames       = 0
	givenNames        = 1
	additionalNames   = 2
	honorificPrefixes = 3
	honorificSuffixes = 4
	nameSize          = honorificSuffixes + 1
)

const ( // Constant define address information index in directory information StructuredValue
	postOfficeBox   = 0
	extendedAddress = 1
	street          = 2
	locality        = 3
	region          = 4
	postalCode      = 5
	countryName     = 6
	addressSize     = countryName + 1
)

func getValueFromContentLine(index int, contentLine *ContentLine) ([]string, string) {
	maxIndex := len(contentLine.Value) - 1
	if maxIndex >= index {
		text := contentLine.Value[index].GetText()

		if strings.ToLower(contentLine.Params["ENCODING"].GetText()) == "quoted-printable" {
			bytes, err := ioutil.ReadAll(newQuotedPrintableReader(strings.NewReader(text)))
			if err != nil {
				return contentLine.Value[index], text
			}
			return contentLine.Value[index], string(bytes)
		} else {
			return contentLine.Value[index], text
		}
	}
	return nil, ""
}

func (vcard *VCard) ReadFrom(di *DirectoryInfoReader) {
	contentLine := di.ReadContentLine()
	for contentLine != nil {
		switch contentLine.Name {
		case "VERSION":
			fallthrough
		case "version":
			vcard.Version = contentLine.Value.GetText()
		case "END":
			fallthrough
		case "end":
			if contentLine.Value.GetText() == "VCARD" {
				return
			}
		case "FN":
			fallthrough
		case "fn":
			if vcard != nil {
				vcard.FormattedName = contentLine.Value.GetText()
			}
		case "N":
			fallthrough
		case "n":
			// NOTE not all vcard names contain all fields, some have more fields
			contentLineLength := len(contentLine.Value)
			if contentLineLength > 0 {
				vcard.FamilyNames, _ = getValueFromContentLine(familyNames, contentLine)
				vcard.GivenNames, _ = getValueFromContentLine(givenNames, contentLine)
				vcard.AdditionalNames, _ = getValueFromContentLine(additionalNames, contentLine)
				vcard.HonorificNames, _ = getValueFromContentLine(honorificPrefixes, contentLine)
				vcard.HonorificSuffixes, _ = getValueFromContentLine(honorificSuffixes, contentLine)
				if contentLineLength > nameSize {
					log.Printf("N data has more fields: %d\n", contentLineLength)
				} else if contentLineLength < nameSize {
					log.Printf("N data has less fields: %d\n", contentLineLength)
				}
			} else {
				log.Printf("Error: N data has no field\n")
			}
		case "NICKNAME":
			fallthrough
		case "nickname":
			vcard.NickNames = contentLine.Value.GetTextList()
		case "PHOTO":
			fallthrough
		case "photo":
			vcard.Photo.Encoding = contentLine.Params["ENCODING"].GetText()
			vcard.Photo.Type = contentLine.Params["TYPE"].GetText()
			vcard.Photo.Value = contentLine.Params["VALUE"].GetText()
			vcard.Photo.Data = contentLine.Value.GetText()
		case "BDAY":
			fallthrough
		case "bday":
			vcard.Birthday = contentLine.Value.GetText()
		case "ADR":
			fallthrough
		case "adr":
			// NOTE not all vcard addresses contain all fields, some have more fields
			contentLineLength := len(contentLine.Value)
			if contentLineLength > 0 {
				var address Address
				if param, ok := contentLine.Params["TYPE"]; ok {
					address.Type = param
				} else {
					address.Type = defaultAddressTypes()
				}
				_, address.PostOfficeBox = getValueFromContentLine(postOfficeBox, contentLine)
				_, address.ExtendedAddress = getValueFromContentLine(extendedAddress, contentLine)
				_, address.Street = getValueFromContentLine(street, contentLine)
				_, address.Locality = getValueFromContentLine(locality, contentLine)
				_, address.Region = getValueFromContentLine(region, contentLine)
				_, address.PostalCode = getValueFromContentLine(postalCode, contentLine)
				_, address.CountryName = getValueFromContentLine(countryName, contentLine)
				vcard.Addresses = append(vcard.Addresses, address)
				if contentLineLength > addressSize {
					log.Printf("ADR data has more fields: %d\n", contentLineLength)
				} else if contentLineLength < addressSize {
					log.Printf("ADR data has less fields: %d\n", contentLineLength)
				}
			} else {
				log.Printf("Error: ADR data has no field\n")
			}
		case "X-ABUID":
			fallthrough
		case "x-abuid":
			vcard.XABuid = contentLine.Value.GetText()
		case "TEL":
			fallthrough
		case "tel":
			var tel Telephone
			if param, ok := contentLine.Params["type"]; ok {
				tel.Type = param
			} else {
				tel.Type = []string{"voice"}
			}
			tel.Number = contentLine.Value.GetText()
			vcard.Telephones = append(vcard.Telephones, tel)
		case "EMAIL":
			fallthrough
		case "email":
			var email Email
			if param, ok := contentLine.Params["type"]; ok {
				email.Type = param
			} else {
				email.Type = []string{"HOME"}
			}
			email.Address = contentLine.Value.GetText()
			vcard.Emails = append(vcard.Emails, email)
		case "TITLE":
			fallthrough
		case "title":
			vcard.Title = contentLine.Value.GetText()
		case "ROLE":
			fallthrough
		case "role":
			vcard.Role = contentLine.Value.GetText()
		case "ORG":
			fallthrough
		case "org":
			vcard.Org = contentLine.Value.GetTextList()
		case "CATEGORIES":
			fallthrough
		case "categories":
			vcard.Categories = contentLine.Value.GetTextList()
		case "NOTE":
			fallthrough
		case "note":
			vcard.Note = contentLine.Value.GetText()
		case "URL":
			fallthrough
		case "url":
			vcard.URL = contentLine.Value.GetText()
		case "X-JABBER":
			fallthrough
		case "x-jabber":
			fallthrough
		case "X-GTALK":
			fallthrough
		case "x-gtalk":
			var jabber XJabber
			if param, ok := contentLine.Params["type"]; ok {
				jabber.Type = param
			} else {
				jabber.Type = []string{"HOME"}
			}
			jabber.Address = contentLine.Value.GetText()
			vcard.XJabbers = append(vcard.XJabbers, jabber)
		case "X-ABShowAs":
			vcard.XABShowAs = contentLine.Value.GetText()
		/*case "X-ABLabel":
		case "X-ABADR":
			// ignore*/
		default:
			log.Printf("Not read %s, %s: %s\n", contentLine.Group, contentLine.Name, contentLine.Value)
		}
		contentLine = di.ReadContentLine()
	}
}

func (vcard *VCard) WriteTo(di *DirectoryInfoWriter) {
	di.WriteContentLine(&ContentLine{"", "BEGIN", nil, StructuredValue{Value{"VCARD"}}})
	di.WriteContentLine(&ContentLine{"", "VERSION", nil, StructuredValue{Value{"3.0"}}})
	di.WriteContentLine(&ContentLine{"", "FN", nil, StructuredValue{Value{vcard.FormattedName}}})
	di.WriteContentLine(&ContentLine{"", "N", nil, StructuredValue{vcard.FamilyNames, vcard.GivenNames, vcard.AdditionalNames, vcard.HonorificNames, vcard.HonorificSuffixes}})
	if len(vcard.NickNames) != 0 {
		di.WriteContentLine(&ContentLine{"", "NICKNAME", nil, StructuredValue{vcard.NickNames}})
	}
	vcard.Photo.WriteTo(di)
	if len(vcard.Birthday) != 0 {
		di.WriteContentLine(&ContentLine{"", "BDAY", nil, StructuredValue{Value{vcard.Birthday}}})
	}
	for _, addr := range vcard.Addresses {
		addr.WriteTo(di)
	}
	for _, tel := range vcard.Telephones {
		tel.WriteTo(di)
	}
	for _, email := range vcard.Emails {
		email.WriteTo(di)
	}
	if len(vcard.Title) != 0 {
		di.WriteContentLine(&ContentLine{"", "TITLE", nil, StructuredValue{Value{vcard.Title}}})
	}
	if len(vcard.Role) != 0 {
		di.WriteContentLine(&ContentLine{"", "ROLE", nil, StructuredValue{Value{vcard.Role}}})
	}
	if len(vcard.Org) != 0 {
		di.WriteContentLine(&ContentLine{"", "ORG", nil, StructuredValue{vcard.Org}})
	}
	if len(vcard.Categories) != 0 {
		di.WriteContentLine(&ContentLine{"", "CATEGORIES", nil, StructuredValue{vcard.Categories}})
	}
	if len(vcard.Note) != 0 {
		di.WriteContentLine(&ContentLine{"", "NOTE", nil, StructuredValue{Value{vcard.Note}}})
	}
	if len(vcard.URL) != 0 {
		di.WriteContentLine(&ContentLine{"", "URL", nil, StructuredValue{Value{vcard.URL}}})
	}
	for _, jab := range vcard.XJabbers {
		jab.WriteTo(di)
	}
	if len(vcard.XABShowAs) != 0 {
		di.WriteContentLine(&ContentLine{"", "X-ABShowAs", nil, StructuredValue{Value{vcard.XABShowAs}}})
	}
	if len(vcard.XABuid) != 0 {
		di.WriteContentLine(&ContentLine{"", "X-ABUID", nil, StructuredValue{Value{vcard.XABuid}}})
	}
	di.WriteContentLine(&ContentLine{"", "END", nil, StructuredValue{Value{"VCARD"}}})
}

func (photo *Photo) WriteTo(di *DirectoryInfoWriter) {
	if len(photo.Data) == 0 {
		return
	}
	params := make(map[string]Value)
	if photo.Encoding != "" {
		params["ENCODING"] = Value{photo.Encoding}
	}
	if photo.Type != "" {
		params["type"] = Value{photo.Type}
	}
	if photo.Value != "" {
		params["VALUE"] = Value{photo.Value}
	}
	if photo.Encoding == "" && photo.Type == "" && photo.Value == "" {
		params["BASE64"] = Value{}
	}
	di.WriteContentLine(&ContentLine{"", "PHOTO", params, StructuredValue{Value{photo.Data}}})
}

func (addr *Address) WriteTo(di *DirectoryInfoWriter) {
	params := make(map[string]Value)
	params["type"] = addr.Type
	di.WriteContentLine(&ContentLine{"", "ADR", params, StructuredValue{Value{addr.PostOfficeBox}, Value{addr.ExtendedAddress}, Value{addr.Street}, Value{addr.Locality}, Value{addr.Region}, Value{addr.PostalCode}, Value{addr.CountryName}}})
}

func (tel *Telephone) WriteTo(di *DirectoryInfoWriter) {
	params := make(map[string]Value)
	params["type"] = tel.Type
	di.WriteContentLine(&ContentLine{"", "TEL", params, StructuredValue{Value{tel.Number}}})
}

func (email *Email) WriteTo(di *DirectoryInfoWriter) {
	params := make(map[string]Value)
	params["type"] = email.Type
	di.WriteContentLine(&ContentLine{"", "EMAIL", params, StructuredValue{Value{email.Address}}})
}

func (jab *XJabber) WriteTo(di *DirectoryInfoWriter) {
	params := make(map[string]Value)
	params["type"] = jab.Type
	di.WriteContentLine(&ContentLine{"", "X-JABBER", params, StructuredValue{Value{jab.Address}}})
}
