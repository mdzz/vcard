package main

import (
	"bitbucket.org/llg/vcard"
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func ask(msg string, format ...interface{}) string {
	fmt.Printf(msg+"\n", format...)
	stdin := bufio.NewReader(os.Stdin)
	line, _, _ := stdin.ReadLine()
	if line != nil {
		return string(line)
	}
	return ""
}

func removeAdditionalName(ab *vcard.AddressBook) {
	for i := 0; i < len(ab.Contacts); i++ {
		contact := &(ab.Contacts[i])
		if len(contact.AdditionalNames) > 0 {
			fmt.Println("---------------------------")
			fmt.Printf("%v\n", contact)
			contact.AdditionalNames = []string{}

		}
	}
}

func capitalizeString(s string) string {
	return strings.Title(strings.ToLower(s))
}

func capitalizeStrings(ss []string) []string {
	for i, s := range ss {
		ss[i] = capitalizeString(s)
	}
	return ss
}

func capitalize(ab *vcard.AddressBook) {
	for i := 0; i < len(ab.Contacts); i++ {
		contact := &(ab.Contacts[i])
		contact.FormattedName = capitalizeString(contact.FormattedName)
		contact.GivenNames = capitalizeStrings(contact.GivenNames)
		contact.FamilyNames = capitalizeStrings(contact.FamilyNames)
	}
}

func integrateAdditionalName(ab *vcard.AddressBook) {
	for i := 0; i < len(ab.Contacts); i++ {
		contact := &(ab.Contacts[i])
		if len(contact.AdditionalNames) > 0 {
			fmt.Println("---------------------------")
			fmt.Printf("%v\n", contact)
			/*msg := "Integrate Additional Name in \n"
			if len(contact.GivenNames) > 0 {
				msg += "given name (g)?\n"
			}
			if len(contact.FamilyNames) > 0 {
				msg += "family name (f) ?\n"
			}*/
			f := "f"
			switch f /*ask(msg)*/ {
			case "g":
				contact.GivenNames[0] += " " + contact.AdditionalNames[0]
				fmt.Printf("result: %s\n", contact.GivenNames[0])
				contact.FormattedName = displayStrings(contact.FamilyNames, contact.GivenNames)
				contact.AdditionalNames = []string{}
			case "f":
				contact.FamilyNames[0] += " " + contact.AdditionalNames[0]
				fmt.Printf("result: %s\n", contact.FamilyNames[0])
				contact.FormattedName = displayStrings(contact.FamilyNames, contact.GivenNames)
				contact.AdditionalNames = []string{}
			}

		}
	}
}

func switchFamilyNamesGivenName(ab *vcard.AddressBook) {
	for i := 0; i < len(ab.Contacts); i++ {
		contact := &(ab.Contacts[i])
		fmt.Println("---------------------------")
		fmt.Printf("%v\n", contact)
		msg := "Switch Family Names and Given Names (yes or no) ?\n"
		switch ask(msg) {
		case "y":
			tmp := contact.GivenNames
			contact.GivenNames = contact.FamilyNames
			contact.FamilyNames = tmp
			contact.FormattedName = displayStrings(contact.FamilyNames, contact.GivenNames)
			fmt.Printf("Given names: %s\n", contact.GivenNames)
			fmt.Printf("FamilyNames names: %s\n", contact.FamilyNames)

		}
	}
}

func displayStrings(sss ...[]string) (display string) {
	for i, ss := range sss {
		for j, s := range ss {
			display += s
			if j+1 < len(ss) {
				display += " "
			}
		}
		if i+1 < len(sss) && len(sss[i+1]) > 0 {
			display += " "
		}
	}
	return display
}

func indexOf(ars []string, value string) int {
	for i, s := range ars {
		if s == value {
			return i
		}
	}
	return -1
}

func mobilePhone(ab *vcard.AddressBook) {
	for i := 0; i < len(ab.Contacts); i++ {
		contact := &(ab.Contacts[i])
		for j := 0; j < len(contact.Telephones); j++ {
			phone := &(contact.Telephones[j])
			if len(phone.Number) > 2 && phone.Number[0:2] == "06" && indexOf(phone.Type, "CELL") == -1 && indexOf(phone.Type, "cell") == -1 {
				fmt.Println("---------------------------")
				fmt.Printf("%v\n", contact)
				//msg := "Is it a portable %s (yes or no) ?\n"
				yes := "y"
				switch yes /*ask(msg, phone)*/ {
				case "y":
					ipref := indexOf(phone.Type, "pref")
					if ipref != -1 {
						phone.Type = []string{"pref", "CELL"}
					} else {
						phone.Type = []string{"CELL"}
					}

				}
			}
		}
	}
}

func main() {
	var output io.Writer
	outputFilename := flag.String("o", "", "Output vcard file")
	flag.Parse()
	if *outputFilename == "" {
		output = os.Stdout
	} else {
		file, err := os.Create(*outputFilename)
		bufoutput := bufio.NewWriter(file)
		output = bufoutput
		defer file.Close()
		defer bufoutput.Flush()
		if err != nil {
			log.Printf("Can't create %s\n", *outputFilename)
			return
		}
	}
	var args []string
	if len(flag.Args()) > 0 {
		args = flag.Args()
	} else {
		args = []string{"contacts.vcf"}
	}
	var addressBook vcard.AddressBook
	for _, abpath := range args {
		f, err := os.Open(abpath)
		defer f.Close()
		if err != nil {
			log.Printf("Can't read file %s\n", abpath)
			return
		}
		reader := vcard.NewDirectoryInfoReader(f)
		addressBook.ReadFrom(reader)
		log.Printf("Read %s\n", abpath)
	}
	/*switchFamilyNamesGivenName(&addressBook)
	integrateAdditionalName(&addressBook)
	mobilePhone(&addressBook)
	removeAdditionalName(&addressBook)
	*/
	capitalize(&addressBook)
	writer := vcard.NewDirectoryInfoWriter(output)
	addressBook.WriteTo(writer)
	log.Printf("Write %s\n", *outputFilename)
}
