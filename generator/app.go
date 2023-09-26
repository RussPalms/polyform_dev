package generator

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/EliCDavis/polyform/generator/room"

	_ "embed"
)

type App struct {
	Name        string
	Version     string
	Description string
	WebScene    *room.WebScene
	Authors     []Author
	Generator   Generator
}

type pageData struct {
	Title       string
	Version     string
	Description string
}

//go:embed server.html
var indexData []byte

func (a App) Serve(host, port string) error {

	serverStarted := time.Now()

	pageToServe := pageData{
		Title:       a.Name,
		Version:     a.Version,
		Description: a.Description,
	}

	webscene := a.WebScene
	if webscene == nil {
		webscene = room.DefaultWebScene()
	}

	movelVersion := 0

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// d, err := os.ReadFile("generator/server.html")
		// if err != nil {
		// 	panic(err)
		// }

		t := template.New("")
		// _, err = t.Parse(string(d))
		_, err := t.Parse(string(indexData))
		if err != nil {
			panic(err)
		}
		t.Execute(w, pageToServe)
		// w.Write(indexData)
	})

	http.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(a.Generator.Schema())
		if err != nil {
			panic(err)
		}
		w.Write(data)
	})

	http.HandleFunc("/scene", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(webscene)
		if err != nil {
			panic(err)
		}
		w.Write(data)
	})

	http.HandleFunc("/started", func(w http.ResponseWriter, r *http.Request) {
		time := serverStarted.Format("2006-01-02 15:04:05")
		w.Write([]byte(fmt.Sprintf("{ \"time\": \"%s\" }", time)))
	})

	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		profile := Profile{}
		if err := json.Unmarshal(body, &profile); err != nil {
			panic(err)
		}
		err := a.Generator.ApplyProfile(profile)
		if err != nil {
			panic(err)
		}
		movelVersion++
		w.Write([]byte("{}"))
	})

	http.HandleFunc("/producer/", func(w http.ResponseWriter, r *http.Request) {
		// params, _ := url.ParseQuery(r.URL.RawQuery)

		generatorToUse := a.Generator
		components := strings.Split(r.URL.Path, "/")
		for i := 2; i < len(components)-1; i++ {
			newGen, ok := generatorToUse.SubGenerators[components[i]]
			if !ok {
				panic(fmt.Errorf("no sub generator exists: %s", components[i]))
			}
			generatorToUse = newGen
		}

		producerToLoad := path.Base(r.URL.Path)

		producer, ok := generatorToUse.Producers[producerToLoad]
		if !ok {
			panic(fmt.Errorf("no producer registered for: %s", producerToLoad))
		}
		artifact, err := producer(&Context{
			Parameters: generatorToUse.Parameters,
		})
		if err != nil {
			panic(err)
		}
		artifact.Write(w)
	})

	hub := room.NewHub(webscene, &movelVersion)
	go hub.Run()

	http.Handle("/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	}))

	connection := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Serving over: http://%s\n", connection)
	return http.ListenAndServe(connection, nil)
}

func (a App) Run() error {

	argsWithoutProg := os.Args[1:]

	switch strings.ToLower(argsWithoutProg[0]) {
	case "generate":
		generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
		a.Generator.initialize(generateCmd)
		folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
		if err := generateCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return a.Generator.run(*folderFlag)

	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		a.Generator.initialize(serveCmd)
		portFlag := serveCmd.String("port", "8080", "port to serve over")
		hostFlag := serveCmd.String("host", "localhost", "interface to bind to")

		if err := serveCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return a.Serve(*hostFlag, *portFlag)

	default:
		fmt.Fprintf(os.Stdout, "unrecognized command %s", argsWithoutProg[0])
	}

	return nil
}
