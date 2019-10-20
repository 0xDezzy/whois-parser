/*
 * Copyright 2014-2019 Li Kexian
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Go module for domain whois info parse
 * https://www.likexian.com/
 */

package whoisparser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/likexian/gokit/assert"
)

var (
	dotJPReplacer = regexp.MustCompile(`\n\[(.+?)\][\ ]*(.+?)?`)
	searchDomain  = regexp.MustCompile(`(?i)\[?Domain(\s+name)?\]?\s*\:?\s*([a-z0-9\-]+)\.([a-z]{2,})`)
)

// Prepare do prepare the whois info for parsing
func Prepare(text string) string {
	text = strings.Replace(text, "\r", "", -1)
	text = strings.Replace(text, "\t", " ", -1)

	m := searchDomain.FindStringSubmatch(text)
	if len(m) > 0 {
		switch strings.ToLower(m[3]) {
		case "edu":
			return prepareEDU(text)
		case "int":
			return prepareINT(text)
		case "mo":
			return prepareMO(text)
		case "hk":
			return prepareHK(text)
		case "tw":
			return prepareTW(text)
		case "ch":
			return prepareCH(text)
		case "it":
			return prepareIT(text)
		case "fr", "re", "tf", "yt", "pm", "wf":
			return prepareFR(text)
		case "ru", "su":
			return prepareRU(text)
		case "jp":
			return prepareJP(text)
		case "uk":
			return prepareUK(text)
		case "kr":
			return prepareKR(text)
		}
	}

	return text
}

// prepareEDU do prepare the .edu domain
func prepareEDU(text string) string {
	tokens := map[string][]string{
		"Registrant:": {
			"Organization",
			"Address",
			"Address",
			"Address",
		},
		"Administrative Contact:": {
			"Name",
			"Organization",
			"Address",
			"Address",
			"Address",
			"Phone",
			"Email",
		},
		"Technical Contact:": {
			"Name",
			"Organization",
			"Address",
			"Address",
			"Address",
			"Phone",
			"Email",
		},
	}

	token := ""
	index := 0

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if strings.HasSuffix(v, ":") {
			token = ""
			index = 0
		}
		if _, ok := tokens[v]; ok {
			token = v
		} else {
			if token == "" {
				result += "\n" + v
			} else {
				result += fmt.Sprintf("\n%s %s: %s", token[:len(token)-1], tokens[token][index], v)
				index += 1
			}
		}
	}

	return result
}

// prepareINT do prepare the .int domain
func prepareINT(text string) string {
	token := ""
	result := ""

	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			token = ""
			continue
		}
		if strings.Contains(v, ":") {
			vs := strings.Split(v, ":")
			if strings.TrimSpace(vs[0]) == "organisation" {
				if token == "" {
					token = "registrant"
				}
			}
			if strings.TrimSpace(vs[0]) == "contact" {
				token = strings.TrimSpace(vs[1])
			} else {
				if token != "" {
					v = fmt.Sprintf("%s %s", token, v)
				}
			}
		}
		result += "\n" + v
	}

	return result
}

// prepareMO do prepare the .mo domain
func prepareMO(text string) string {
	tokens := map[string]string{
		"Registrant:":           "Registrant",
		"Admin Contact(s):":     "Admin",
		"Billing Contact(s):":   "Billing",
		"Technical Contact(s):": "Technical",
	}

	token := ""
	result := ""

	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			token = ""
			continue
		}
		if v[0] == '-' {
			continue
		}
		for _, s := range []string{"Record created on", "Record expires on"} {
			if strings.HasPrefix(v, s) {
				v = strings.Replace(v, s, s+":", 1)
			}
		}
		if _, ok := tokens[v]; ok {
			token = tokens[v]
		} else {
			if token != "" {
				v = fmt.Sprintf("%s %s", token, v)
			}
		}
		result += "\n" + v
	}

	return result
}

// prepareHK do prepare the .hk domain
func prepareHK(text string) string {
	tokens := map[string]string{
		"Registrant Contact Information:":     "Registrant",
		"Administrative Contact Information:": "Admin",
		"Technical Contact Information:":      "Technical",
		"Name Servers Information:":           "Name Servers:",
	}

	dateTokens := []string{
		"Domain Name Commencement Date",
		"Expiry Date",
	}

	token := ""
	addressToken := false
	text = strings.Replace(text, "\n\n", "\n", -1)

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			token = ""
			continue
		}
		field := ""
		if strings.Contains(v, ":") {
			vs := strings.SplitN(v, ":", 2)
			field = strings.TrimSpace(vs[0])
			if strings.Contains(field, "(") {
				field = strings.Split(field, "(")[0]
				v = fmt.Sprintf("%s: %s", field, vs[1])
			}
			addressToken = field == "Address"
			if field == "Registrar Contact Information" {
				re := regexp.MustCompile(`Email\:\s+([^\s]+)(\s+Hotline\:(.*))?`)
				m := re.FindStringSubmatch(vs[1])
				if len(m) == 4 {
					v = ""
					if m[1] != "" {
						v += "Registrar Contact Email: " + m[1] + "\n"
					}
					if m[3] != "" {
						v += "Registrar Contact Phone: " + m[3] + "\n"
					}
					v = strings.TrimSpace(v)
				}
			}
			if field == "Family name" {
				vv := strings.TrimSpace(vs[1])
				if vv != "" && vv != "." {
					result += " " + vv
				}
				continue
			}
		} else {
			if addressToken {
				result += ", " + v
				continue
			}
		}
		if _, ok := tokens[v]; ok {
			token = tokens[v]
		} else {
			if token != "" && !assert.IsContains(dateTokens, field) {
				v = fmt.Sprintf("%s %s", token, v)
			}
		}
		result += "\n" + v
	}

	return result
}

// prepareTW do prepare the .tw domain
func prepareTW(text string) string {
	tokens := map[string][]string{
		"Registrant:": {
			"Organization",
			"Name,Email",
			"Phone",
			"Fax",
			"Address",
			"Address",
			"Address",
		},
		"Administrative Contact:": {
			"Name,Email",
			"Phone",
			"Fax",
		},
		"Technical Contact:": {
			"Name,Email",
			"Phone",
			"Fax",
		},
		"Contact:": {
			"Name",
			"Email",
		},
	}

	token := ""
	index := -1

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		for _, s := range []string{"Record created on", "Record expires on"} {
			if strings.HasPrefix(v, s) {
				v = strings.Replace(v, s, s+":", 1)
			}
		}
		if strings.Contains(v, ":") {
			token = ""
			index = -1
		}
		if _, ok := tokens[v]; ok {
			token = v
		} else {
			if token == "" {
				result += "\n" + v
			} else {
				index += 1
				tokenName := token[:len(token)-1]
				indexName := tokens[token][index]
				if tokenName == "Contact" {
					tokenName = "Registrant Contact"
				}
				if strings.Contains(indexName, ",") {
					ins := strings.Split(indexName, ",")
					re := regexp.MustCompile(`(.*)\s+([^\s]+@[^\s]+)`)
					m := re.FindStringSubmatch(v)
					if len(m) == 3 {
						result += fmt.Sprintf("\n%s %s: %s", tokenName, ins[0], strings.TrimSpace(m[1]))
						result += fmt.Sprintf("\n%s %s: %s", tokenName, ins[1], strings.TrimSpace(m[2]))
					} else {
						result += fmt.Sprintf("\n%s %s: %s", tokenName, ins[0], strings.TrimSpace(v))
					}
					continue
				}
				result += fmt.Sprintf("\n%s %s: %s", tokenName, indexName, v)
			}
		}
	}

	return result
}

// prepareCH do prepare the .ch domain
func prepareCH(text string) string {
	tokens := []string{
		"Domain name",
		"Holder",
		"Technical contact",
		"Registrar",
		"DNSSEC",
		"Name servers",
		"First registration date",
	}

	splits := map[string]string{
		"Holder":            "Registrant organization, Registrant name, Registrant street",
		"Technical contact": "Technical organization, Technical name, Technical street",
	}

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		found := false
		for _, t := range tokens {
			if strings.HasPrefix(strings.ToLower(v)+" ", strings.ToLower(t+" ")) {
				found = true
				result += fmt.Sprintf("\n%s: %s", strings.TrimSpace(t), strings.TrimSpace(v[len(t):]))
				break
			}
		}
		if !found {
			result += ", " + v
		}
	}

	results := []string{}
	for _, v := range strings.Split(result, "\n") {
		if !strings.Contains(v, ":") {
			continue
		}
		vs := strings.Split(v, ":")
		if sp, ok := splits[vs[0]]; ok {
			vv := strings.Split(vs[1], ", ")
			ss := strings.Split(sp, ", ")
			if len(vv) > len(ss) {
				vv[len(ss)-1] = strings.Join(vv[len(ss)-1:], ", ")
				vv = vv[:len(ss)]
			}
			for k := range vv {
				results = append(results, fmt.Sprintf("%s: %s", ss[k], vv[k]))
			}
		} else {
			results = append(results, v)
		}
	}

	text = strings.Join(results, "\n")
	text = strings.Replace(text, ": ,", ":", -1)

	return text
}

// prepareIT do prepare the .it domain
func prepareIT(text string) string {
	topTokens := []string{
		"Registrant",
		"Admin Contact",
		"Technical Contacts",
		"Registrar",
		"Nameservers",
	}

	topToken := ""
	subToken := ""

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if assert.IsContains(topTokens, v) {
			topToken = v + " "
			subToken = ""
		} else {
			if v[0] != '*' && strings.Contains(v, ":") {
				vs := strings.Split(v, ":")
				subToken = vs[0]
			} else {
				if subToken != "" {
					result += ", " + v
					continue
				}
			}
			if topToken != "" && !strings.Contains(v, ":") {
				result += fmt.Sprintf("\n%s: %s", topToken, v)
			} else {
				result += fmt.Sprintf("\n%s%s", topToken, v)
			}
		}
	}

	return result
}

// prepareFR do prepare the .fr domain
func prepareFR(text string) string {
	dsToken := "dsl-id"
	hdlToken := "nic-hdl"
	regToken := "registrar"

	tokens := map[string]string{
		"holder-c": "holder",
		"admin-c":  "admin",
		"tech-c":   "tech",
	}

	token := ""
	newBlock := false
	hdls := map[string]string{}

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			newBlock = true
			continue
		}

		vs := strings.Split(v, ":")
		if newBlock && strings.TrimSpace(vs[0]) == regToken {
			token = regToken + " "
			v = fmt.Sprintf("name: %s", strings.TrimSpace(vs[1]))
		}

		newBlock = false
		if t, ok := tokens[strings.TrimSpace(vs[0])]; ok {
			hdls[t] = strings.TrimSpace(vs[1])
		}

		if strings.TrimSpace(vs[0]) == dsToken && strings.TrimSpace(vs[1]) != "" {
			v += "\nDNSSEC: signed"
		}

		if strings.TrimSpace(vs[0]) == hdlToken {
			for _, kk := range Keys(hdls) {
				if strings.TrimSpace(vs[1]) == hdls[kk] {
					token = kk + " "
					delete(hdls, kk)
					break
				}
			}
		}

		result += fmt.Sprintf("\n%s%s", token, v)
	}

	return result
}

// prepareRU do prepare the .ru domain
func prepareRU(text string) string {
	tokens := map[string]string{
		"person": "Registrant Name",
		"e-mail": "Registrant Email",
		"org":    "Registrant Organization",
	}

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !strings.Contains(v, ":") {
			continue
		}
		vs := strings.Split(v, ":")
		if vv, ok := tokens[strings.TrimSpace(vs[0])]; ok {
			v = fmt.Sprintf("%s: %s", vv, vs[1])
		}
		result += v + "\n"
	}

	return result
}

// prepareJP do prepare the .jp domain
func prepareJP(text string) string {
	text = dotJPReplacer.ReplaceAllString(text, "\n$1: $2")

	adminToken := "Contact Information"
	addressToken := "Postal Address"

	token := ""
	prefixToken := ""

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if strings.Contains(v, ":") {
			vs := strings.Split(v, ":")
			token = strings.TrimSpace(vs[0])
			if token == adminToken {
				prefixToken = "admin "
			}
		} else {
			if token == addressToken {
				result += ", " + v
				continue
			}
		}
		result += "\n" + prefixToken + v
	}

	return result
}

// prepareUK do prepare the .uk domain
func prepareUK(text string) string {
	tokens := map[string]string{
		"URL": "Registrar URL",
	}

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if strings.Contains(v, ":") {
			vs := strings.SplitN(v, ":", 2)
			if vv, ok := tokens[strings.TrimSpace(vs[0])]; ok {
				v = fmt.Sprintf("%s: %s", vv, vs[1])
			}
		}
		result += "\n" + v
	}

	return result
}

// prepareKR do prepare the .kr domain
func prepareKR(text string) string {
	english := "# ENGLISH"
	tokens := map[string]string{
		"Administrative Contact(AC)": "Administrative Contact Name",
		"AC E-Mail":                  "Administrative Contact E-Mail",
		"AC Phone Number":            "Administrative Contact Phone Number",
	}

	pos := strings.Index(text, english)
	if pos != -1 {
		text = text[pos+len(english):]
	}

	result := ""
	for _, v := range strings.Split(text, "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if v[0] == '\'' || v[0] == '-' {
			continue
		}
		if strings.Contains(v, ":") {
			vs := strings.SplitN(v, ":", 2)
			if vv, ok := tokens[strings.TrimSpace(vs[0])]; ok {
				v = fmt.Sprintf("%s: %s", vv, vs[1])
			}
		}
		result += "\n" + v
	}

	return result
}
