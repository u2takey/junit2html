package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"strings"

	"github.com/Masterminds/sprig"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	ID          string            `xml:"-"`
	Outcome     string            `xml:"-"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitFailure contains data related to a failed test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type Junit struct {
	Report *JUnitTestSuites
	Title  string
}

func NewJunit(xmlData []byte, title string) (*Junit, error) {
	report := &JUnitTestSuites{}
	singleSuite := &JUnitTestSuite{}
	err := xml.Unmarshal(xmlData, report)
	if err != nil {
		err = xml.Unmarshal(xmlData, singleSuite)
		if err != nil {
			return nil, err
		}
		report.Suites = append(report.Suites, *singleSuite)
	}
	for i := range report.Suites {
		for j := range report.Suites[i].TestCases {
			report.Suites[i].TestCases[j].ID = fmt.Sprintf("id_%d_%d", i, j)
			if report.Suites[i].TestCases[j].Failure != nil {
				report.Suites[i].TestCases[j].Outcome = "FAILED"
			} else if report.Suites[i].TestCases[j].SkipMessage != nil {
				report.Suites[i].TestCases[j].Outcome = "SKIPPED"
			} else {
				report.Suites[i].TestCases[j].Outcome = "SUCCESS"
			}
		}
	}

	// fmt.Println("loaded junit xml: \n", prettyJson(report))
	return &Junit{Report: report, Title: title}, nil
}

func (j *Junit) Html() ([]byte, error) {
	styleCss, err := templates.ReadFile("templates/style.css")
	if err != nil {
		return nil, err
	}
	html, err := templates.ReadFile("templates/report.html")
	if err != nil {
		return nil, err
	}
	htmlStr := strings.ReplaceAll(string(html), "{{style.css}}", string(styleCss))

	tmpl, err := template.New("html").Funcs(sprig.FuncMap()).Parse(htmlStr)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	err = tmpl.Execute(buf, j)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// debug
func prettyJson(a interface{}) string {
	data, _ := json.MarshalIndent(a, "", " ")
	return string(data)
}
