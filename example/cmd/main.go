package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/russellrollins/regrev"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	// Too many repeats will be ugly in the web UI
	reverse, err := regrev.NewRegexReverser(
		regrev.MaxRepeats(5),
	)
	if err != nil {
		return err
	}

	port, envSet := os.LookupEnv("PORT")
	if !envSet {
		// use Russell's favorite port.
		port = ":8910"
	}

	getTemplate, err := template.New("get").Parse(strings.TrimSpace(`
<head>
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
</head>
<body>
  <div class=container>
    <div>
      <p>
        Enter a regular expression below, and regrev will attempt to produce a string that matches it!
      </p>

      <form action="/" method="post">
        <input id="regexp" type="text" placeholder="h?e+ll{4,5}o!" name="regexp" size="100">
        <button type="submit">Reverse Regular Expression</button>
      </form>
    </div>
  </div>
</body>
`))
	if err != nil {
		return err
	}

	responseTemplate, err := template.New("post").Parse(strings.TrimSpace(`
<head>
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
</head>
<body>
  <div class="container">
    <div>
      You submitted {{.Input}}
    </div>
    {{ if .RegexSucceeded }}
      <div>
        <p>
          regrev produced the string: {{.Response}} in response.
        </p>
        {{ if .Matches }}
          <p class="alert-success">A matching string!<p>
        {{ else }}
          <p class="alert-danger">A string that doesn't match, unfortunately. Looks like there's more cases to account for!</p>
        {{ end }}
        <a href="/"><button type="button" class="btn btn-primary">one 'mo 'gain?</button></a>
      </div>
    {{ else }}
      <div>
        That's not a valid regexp. Pretty unfair of you, in my opinion.
      </div>
    {{ end }}
  </div>
</body>
`))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_ = getTemplate.Execute(w, nil)
		}
		if r.Method == "POST" {
			inputReg := r.FormValue("regexp")
			var (
				regexSucceeded bool
				response       string
				matches        bool
			)
			reg, err := regexp.Compile(inputReg)
			if err == nil {
				regexSucceeded = true
				resp, err := reverse.Reverse(reg)
				if err != nil {
					matches = false
					response = ""
				} else {
					response = resp
					matches = reg.MatchString(resp)
				}
			}

			_ = responseTemplate.Execute(w, struct {
				Input          string
				RegexSucceeded bool
				Response       string
				Matches        bool
			}{
				inputReg,
				regexSucceeded,
				response,
				matches,
			})
		}
	})

	return http.ListenAndServe(port, nil)
}
