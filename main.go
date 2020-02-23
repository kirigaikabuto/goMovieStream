package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
)

var (
	fileuploaddir = "./assets/media/"
)

func main() {
	router := mux.NewRouter()
	router.Handle("/todo/", http.StripPrefix("./public/src/css", http.FileServer(http.Dir("./public/src/css"))))
	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/media/{mId}/stream/", streamHandler).Methods("GET")
	router.HandleFunc("/media/{mId}/stream/{segName}", streamHandler).Methods("GET")
	router.HandleFunc("/upload_form/", uploadForm).Methods("GET")
	router.HandleFunc("/upload", uploadFile)
	fmt.Println("Server is started")
	http.ListenAndServe("0.0.0.0:8000", router)
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./public/html/index.html")
	t.Execute(w, nil)
}
func uploadForm(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./public/html/form_upload.html")
	t.Execute(w, nil)
}
func uploadFile(w http.ResponseWriter, r *http.Request) {
	file, handle, err := r.FormFile("file")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	//mimeType := handle.Header.Get("Content-Type")

	// switch mimeType {
	// case "image/jpeg":
	//     saveFile(w, file, handle)
	// case "image/png":
	//     saveFile(w, file, handle)
	// default:
	//     jsonResponse(w, http.StatusBadRequest, "The format file is not valid.")
	// }
	saveFile(w, file, handle)
}
func saveFile(w http.ResponseWriter, file multipart.File, handle *multipart.FileHeader) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
		return
	}

	videoname := strings.Split(handle.Filename, ".")[0]
	err = os.Mkdir(fileuploaddir+videoname, 0666)
	if err != nil {
		log.Fatal(err)
		return
	}
	folder_hls := fileuploaddir + videoname + "/hls/"
	err = os.Mkdir(fileuploaddir+videoname+"/hls/", 0666)
	if err != nil {
		log.Fatal(err)
		return
	}
	file_url := fileuploaddir + videoname + "/" + handle.Filename

	err = ioutil.WriteFile(file_url, data, 0666)
	if err != nil {
		log.Fatal(err)
		return
	}
	//ffmpeg -i input.mp4 -profile:v baseline -level 3.0 -s 640x360 -start_number 0 -hls_time 10 -hls_list_size 0 -f hls index.m3u8
	fmt.Println(file_url)
	go func(file_url, folder_hls, videoname string) {
		cmd := exec.Command("ffmpeg", "-i", file_url, "-profile:v", "baseline", "-level", "3.0", "-s", "640x360", "-start_number", "0", "-hls_time", "10", "-hls_list_size", "0", "-f", "hls", folder_hls+"playlist"+videoname+".m3u8")
		cmd.Run()
	}(file_url, folder_hls, videoname)
	jsonResponse(w, http.StatusCreated, "File uploaded successfully!.")
}

func jsonResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprint(w, message)
}
func streamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mId := vars["mId"]

	fmt.Println("asdsad")
	segName, ok := vars["segName"]
	if !ok {
		mediaBase := getMediaBase(mId)
		fmt.Println(mediaBase)
		m3u8Name := fmt.Sprintf("playlist%s.m3u8", mId)
		serverHlsM3u8(w, r, mediaBase, m3u8Name)
	} else {
		mediaBase := getMediaBase(mId)
		fmt.Println(mediaBase)
		serverHlsTs(w, r, mediaBase, segName)
	}
}
func getMediaBase(mId string) string {
	mediaRoot := "assets\\media"
	return fmt.Sprintf("%s\\%s", mediaRoot, mId)
}
func serverHlsM3u8(w http.ResponseWriter, r *http.Request, mediaBase, m3u8Name string) {
	enableCors(&w)
	mediaFile := fmt.Sprintf("%s\\hls\\%s", mediaBase, m3u8Name)
	fmt.Println(mediaFile)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "application/x-mpegURL")

}
func serverHlsTs(w http.ResponseWriter, r *http.Request, mediaBase, segName string) {
	enableCors(&w)
	mediaFile := fmt.Sprintf("%s\\hls\\%s", mediaBase, segName)
	fmt.Println(mediaFile)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "video/MP2T")

}
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
