package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var token string
var busyMutex sync.Mutex

var earthDataUsername string = "YOUR USERNAME"
var earthDataPassword string = "YOUR PASSWORD"

func main() {
	token = getToken()
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/Loading", Loading)
	http.HandleFunc("/Done/", Done)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	http.ListenAndServe(":8080", nil)

}

var done map[string]bool = make(map[string]bool)

func Loading(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		coords := strings.Split(r.FormValue("coords"), ",")
		name := r.FormValue("name")
		fmt.Println(name)
		lat, _ := strconv.ParseFloat(coords[0], 64)
		lon, _ := strconv.ParseFloat(coords[1], 64)
		uniqueName := name
		uniqueName = strings.ReplaceAll(strings.ToLower(uniqueName), " ", "_")
		_, exists := done[uniqueName]
		id := 0
		for exists {
			id++
			uniqueName = name + strconv.Itoa(id)
			uniqueName = strings.ReplaceAll(strings.ToLower(uniqueName), " ", "_")
			_, exists = done[uniqueName]
		}

		done[uniqueName] = false
		go generateEverything(name, uniqueName, lat, lon)

		fmt.Println("load", done)
		http.Redirect(w, r, "/Done/"+uniqueName, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func Done(w http.ResponseWriter, r *http.Request) {

	var data template.HTML

	templates := template.Must(template.ParseFiles("done.html"))

	parts := strings.Split(r.URL.String(), "/")
	name := parts[len(parts)-1]
	ready, exists := done[name]
	if !exists {
		data = template.HTML("<p>Sorry, we can't seem to find a calendar under this name</p>")
	} else if !ready {
		data = template.HTML("<p>Your calendar is being generated, check back in a minute or two...</p>")
	} else {
		data = template.HTML(`<img src = "../resources/images/` + name + `.png" />`)
	}

	if err := templates.ExecuteTemplate(w, "done.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type TokenData struct {
	TokenType  string    `json:"token_type"`
	Token      string    `json:"token"`
	Expiration time.Time `json:"expiration"`
}

type StatusData struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type FileData struct {
	Sha256   string `json:"sha256"`
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	FileType string `json:"file_type"`
	S3URL    string `json:"s3_url"`
}

type FilesData struct {
	Files      []FileData `json:"files"`
	Created    string     `json:"created"`
	TaskID     string     `json:"task_id"`
	Updated    string     `json:"updated"`
	BundleType string     `json:"bundle_type"`
}

func getToken() string {
	url := "https://appeears.earthdatacloud.nasa.gov/api/login"
	r, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}
	creds := earthDataUsername + ":" + earthDataPassword
	r.Header.Add("Authorization", "Basic "+b64.StdEncoding.EncodeToString([]byte(creds)))
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	data := &TokenData{}
	derr := json.NewDecoder(res.Body).Decode(data)
	if derr != nil {
		panic(derr)
	}
	defer res.Body.Close()
	return data.Token
}

func startTask(lat float64, lon float64) string {
	postData := `{"params":{"coordinates":[{"category":"Test","id":"0","latitude":"` + fmt.Sprint(lat) + `","longitude":"` + fmt.Sprint(lon) + `"}],"dates":[{"endDate":"12-31-2022","startDate":"01-01-2022"}],"layers":
	[
		{"layer":"LST_Day_1km","product":"MOD11A2.061"}
	]},"task_name":"calendar","task_type":"point"}`
	fmt.Println(postData)
	url := "https://appeears.earthdatacloud.nasa.gov/api/task"
	r, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postData)))
	if err != nil {
		panic(err)
	}
	r.Header.Add("Authorization", "Bearer "+token)
	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	io.Copy(buf, res.Body)
	// check errors
	fmt.Println(buf.String())
	data := &StatusData{}
	derr := json.Unmarshal([]byte(buf.String()), data)
	if derr != nil {
		fmt.Println(buf.String())
	}
	defer res.Body.Close()
	return data.TaskID
}

func getStatus(taskId string) string {

	url := "https://appeears.earthdatacloud.nasa.gov/api/status/" + taskId
	r, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}
	r.Header.Add("Authorization", "Bearer "+token)
	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	io.Copy(buf, res.Body)
	// check errors

	data := &StatusData{}
	derr := json.Unmarshal([]byte(buf.String()), data)
	if derr != nil {
		fmt.Println(buf.String())
	}
	defer res.Body.Close()
	return data.Status
}

func downloadFiles(taskId string, folder string) {
	url := "https://appeears.earthdatacloud.nasa.gov/api/bundle/" + taskId
	r, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}
	r.Header.Add("Authorization", "Bearer "+token)
	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	io.Copy(buf, res.Body)
	// check errors

	data := &FilesData{}
	derr := json.Unmarshal([]byte(buf.String()), data)
	if derr != nil {
		fmt.Println(buf.String())
	}
	defer res.Body.Close()
	for _, f := range data.Files {
		if f.FileType == "csv" {
			fmt.Println("downloading " + f.FileName)
			downloadFile(taskId, f.FileID, f.FileName, folder)
		}
	}
}

func downloadFile(taskId string, fileId string, fileName string, folder string) {
	url := "https://appeears.earthdatacloud.nasa.gov/api/bundle/" + taskId + "/" + fileId
	r, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}
	r.Header.Add("Authorization", "Bearer "+token)
	r.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	buf := new(strings.Builder)
	io.Copy(buf, res.Body)
	// check errors

	err = os.WriteFile("data/"+folder+"/"+fileName, []byte(buf.String()), 0644)
	if err != nil {
		panic(err)
	}
}

func generateImage(name string) {
	CopyFile("data/"+name+"/calendar-MOD11A2-061-results.csv", "generator/data/calendar-MOD11A2-061-results.csv")
	_, err := exec.Command("./generator/kalender").Output()
	if err != nil {
		fmt.Println(err.Error())
	}

	time.Sleep(5 * time.Second)
	CopyFile("generator/output.png", "resources/images/"+name+".png")
}

func generateDataFile(name string, lat float64, lon float64) {
	err := os.WriteFile("generator/data/data.txt", []byte(name+"\n"+fmt.Sprint(lat)+"\n"+fmt.Sprint(lon)), 0644)
	if err != nil {
		panic(err)
	}
}

func generateEverything(name string, uniqueName string, lat float64, lon float64) {
	err := os.Mkdir("data/"+uniqueName, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	taskId := startTask(lat, lon)
	fmt.Println(taskId)
	for range time.Tick(time.Second * 20) {
		status := getStatus(taskId)
		fmt.Println(status)
		if status == "done" {
			break
		}
	}
	fmt.Println("Files ready!")
	downloadFiles(taskId, uniqueName)

	busyMutex.Lock()
	generateDataFile(name, lat, lon)
	generateImage(uniqueName)
	busyMutex.Unlock()

	done[uniqueName] = true
	fmt.Println("all done")
}

func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
