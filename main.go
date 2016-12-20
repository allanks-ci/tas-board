package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Job struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

var fatalLog = log.New(os.Stdout, "FATAL: ", log.LstdFlags)
var infoLog = log.New(os.Stdout, "INFO: ", log.LstdFlags)

func basePage(rw http.ResponseWriter, req *http.Request) {
	buf, err := getHTTP(req.Header.Get("tazzy-tenant"), getURL("devs/tas/jobs"))
	if err == nil {
		http.NotFound(rw, req)
		return
	}
	var jobs []Job
	decoder := json.NewDecoder(bytes.NewReader(buf))
	infoLog.Printf("BasePage json error: %v", decoder.Decode(&jobs))
	t, err := template.ParseFiles("static/index.html")
	infoLog.Printf("BasePage template error: %v", err)
	if jobs == nil {
		t.Execute(rw, []Job{})
	} else {
		t.Execute(rw, jobs)
	}
}

func jobPage(rw http.ResponseWriter, req *http.Request) {
	buf, err := getHTTP(req.Header.Get("tazzy-tenant"), getURL("tas/jobs/byID/{job}"))
	if err == nil {
		http.NotFound(rw, req)
		return
	}
	var job Job
	decoder := json.NewDecoder(bytes.NewReader(buf))
	infoLog.Printf("BasePage json error: %v", decoder.Decode(&job))
	t, err := template.ParseFiles("static/job.html")
	infoLog.Printf("BasePage template error: %v", err)
	if &job == nil {
		t.Execute(rw, Job{})
	} else {
		t.Execute(rw, job)
	}
}

func apply(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	http.Redirect(rw, req, getURL(fmt.Sprintf("tas/apply/%v", vars["job"])), http.StatusSeeOther)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", basePage)
	r.HandleFunc("/job/{job}", jobPage)
	r.HandleFunc("/apply/{job}", apply)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	fatalLog.Fatal(http.ListenAndServe(":8080", r))
}

func getHTTP(tenant, url string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")
	return doHTTP(req, tenant)
}

func doHTTP(req *http.Request, tenant string) ([]byte, error) {
	req.Header.Set("tazzy-secret", os.Getenv("IO_TAZZY_SECRET"))
	req.Header.Set("tazzy-tenant", os.Getenv("APP_SHORTCODE"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getURL(api string) string {
	return fmt.Sprintf("%s/%s", os.Getenv("IO_TAZZY_URL"), api)
}
