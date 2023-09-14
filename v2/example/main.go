// Command example runs a sample webserver that uses go-i18n/v2/i18n.
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/abdullahskartal/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var page = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<body>

<h1>{{.Title}}</h1>

{{range .Paragraphs}}<p>{{.}}</p>{{end}}

</body>
</html>
`))

const (
	filePattern = "active.%s.toml"
)

var (
	defaultLanguage = language.Turkish
	bundle          *i18n.Bundle
)

func main() {
	countryLanguagePair := map[string][]string{
		"TR": {"tr", "en"},
		"GB": {"tr", "en"},
	}
	if err := CreateDefaultBundle("./lang", countryLanguagePair); err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		acceptLanguage := r.Header.Get("Accept-Language")
		countryCode := r.Header.Get("countryCode")
		localizer := i18n.NewLocalizer(bundle, acceptLanguage)

		name := r.FormValue("name")
		if name == "" {
			name = "Bob"
		}

		localizer.SetCountryCode(countryCode)
		helloPerson := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "HelloPerson",
				Other: "Hello {{.Name}}",
			},
			TemplateData: map[string]string{
				"Name": name,
			},
		})

		err := page.Execute(w, map[string]interface{}{
			"Title": helloPerson,
		})
		if err != nil {
			panic(err)
		}
	})

	fmt.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CreateDefaultBundle(folderPath string, countryLanguagesPairs map[string][]string) error {
	b := i18n.NewBundle("tr", defaultLanguage)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	for countryCode, languages := range countryLanguagesPairs {
		for _, lang := range languages {
			fullPath := generateFullPath(folderPath, strings.ToLower(countryCode), fmt.Sprintf(filePattern, lang))
			_, err := b.LoadMessageFile(fullPath, strings.ToLower(countryCode))
			if err != nil {
				return err
			}
		}
	}

	bundle = b
	return nil
}

func generateFullPath(fileNames ...string) string {
	return filepath.Join(fileNames...)
}
