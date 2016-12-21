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

type application struct {
	Job   string `json:"job"`
	Email string `json:"email"`
}

type attributes struct {
	Email string `json:"tas.personal.email"`
}

var fatalLog = log.New(os.Stdout, "FATAL: ", log.LstdFlags)
var infoLog = log.New(os.Stdout, "INFO: ", log.LstdFlags)

func basePage(rw http.ResponseWriter, req *http.Request) {
	buf, err := getHTTP(req.Header.Get("tazzy-tenant"), getURL("devs/tas/jobs"))
	if err != nil {
		errorHandler(rw, req, 404, err)
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
	vars := mux.Vars(req)
	buf, err := getHTTP(req.Header.Get("tazzy-tenant"), getURL(fmt.Sprintf("devs/tas/jobs/byID/%v", vars["job"])))
	if err != nil {
		errorHandler(rw, req, 404, err)
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
	email := getEmail(req.Header.Get("tazzy-tenant"), req.Header.Get("tazzy-saml"))
	app := application{
		Job:   vars["job"],
		Email: email,
	}
	data, err := json.Marshal(&app)
	if err != nil {
		infoLog.Printf("Apply json error: %v", err)
		http.Error(rw, "Could not serialize input", http.StatusInternalServerError)
		return
	}
	_, err = postHTTP(req.Header.Get("tazzy-tenant"), getURL("devs/allan/submit"), data)
	infoLog.Printf("Apply post error: %v", err)
	t, err := template.ParseFiles("static/thanks.html")
	infoLog.Printf("Apply template error: %v", err)
	t.Execute(rw, nil)
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int, err error) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprintf(w, "404 Not found\nError: %v", err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", basePage)
	r.HandleFunc("/job/{job}", jobPage)
	r.HandleFunc("/apply/{job}", apply)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	fatalLog.Fatal(http.ListenAndServe(":8080", r))
}

func getEmail(tenant, saml string) string {
	url := getURL(fmt.Sprintf("core/tenants/%s/saml/assertions/byKey/%s/json", tenant, saml))
	jsonAttr, err := getHTTP(tenant, url)
	infoLog.Printf("GetEmail json error", err)
	if err != nil {
		return ""
	}

	var attr attributes
	infoLog.Printf("GetEmail attr error", json.Unmarshal(jsonAttr, &attr))
	return attr.Email
}

func getHTTP(tenant, url string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")
	return doHTTP(req, tenant)
}

func postHTTP(tenant, url string, data []byte) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	return doHTTP(req, tenant)
}

func doHTTP(req *http.Request, tenant string) ([]byte, error) {
	req.Header.Set("tazzy-secret", os.Getenv("IO_TAZZY_SECRET"))
	req.Header.Set("tazzy-tenant", tenant)
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
