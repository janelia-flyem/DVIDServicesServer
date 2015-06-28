package Server

import (
	"encoding/json"
	//"time"
        "fmt"
	//"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	// Contain URI location for interface
	interfacePath  = "/interface/"
	servicesPath = "/services/"
        servicePath = "/service/"
        statusPath = "/jobstatus/"
        sparkScript = "spark_launch_wrapper"
        workflowscript = "launchworkflow.py"
)

// Directory containing temporary results from segmentation (root + /.calclabels/)
var logDirectory string

// machine where clusterscript is installed
var remoteMachine string

// location of service server 
var webAddress string

// user for remote program
var remoteUser string

// location of web console source 
var webConsole string

// location of service workflows directory
var sparkWorkflowsLocation string

// environment for remote command
var remoteEnv []string

// parseURI is a utility function for retrieving parts of the URI
func parseURI(r *http.Request, prefix string) ([]string, string, error) {
	requestType := strings.ToLower(r.Method)
	prefix = strings.Trim(prefix, "/")
	path := strings.Trim(r.URL.Path, "/")
	prefix_list := strings.Split(prefix, "/")
	url_list := strings.Split(path, "/")
	var path_list []string

	if len(prefix_list) > len(url_list) {
		return path_list, requestType, fmt.Errorf("Incorrectly formatted URI")
	}

	for i, val := range prefix_list {
		if val != url_list[i] {
			return path_list, requestType, fmt.Errorf("Incorrectly formatted URI")
		}
	}

	if len(prefix_list) < len(url_list) {
		path_list = url_list[len(prefix_list):]
	}

	return path_list, requestType, nil
}

// badRequest is a halper for printing an http error message
func badRequest(w http.ResponseWriter, msg string) {
	fmt.Println(msg)
	http.Error(w, msg, http.StatusBadRequest)
}

// randomHex computes a random hash for storing service results
func randomHex() (randomStr string) {
	randomStr = ""
	for i := 0; i < 8; i++ {
		val := rand.Intn(16)
		randomStr += strconv.FormatInt(int64(val), 16)
	}
	return
}

// InterfaceHandler returns the RAML interface for any request at
// the /interface URI.
func interfaceHandler(w http.ResponseWriter, r *http.Request) {
	// allow resources to be accessed via ajax
	w.Header().Set("Content-Type", "application/raml+yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, ramlInterface)
}

// frontHandler handles GET requests to "/"
func frontHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) > 1 {
            http.ServeFile(w, r, webConsole + "/" + r.URL.Path[:])  
            return
        }

        _, requestType, err := parseURI(r, "/")
        if err != nil {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}
	if requestType != "get" {
		badRequest(w, "only supports gets")
		return
	}

	w.Header().Set("Content-Type", "text/html")

	//formHTMLsub := strings.Replace(formHTML, "DEFAULT", webAddress, 1)
	fmt.Fprintf(w, formHTML)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	_, _, err := parseURI(r, statusPath)
	//pathlist, _, err := parseURI(r, statusPath)
	if err != nil {
		badRequest(w, "Error: incorrectly formatted request")
                return
        }

        // ?! not complete
        fmt.Fprintf(w, "{\"job_status\": \"Started\", \"job_message\": \"I am running\"}")
}        
	
func servicesHandler(w http.ResponseWriter, r *http.Request) {
	pathlist, requestType, err := parseURI(r, servicesPath)
	if err != nil || len(pathlist) != 0 {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}
	if requestType != "get" {
		badRequest(w, "only supports gets")
		return
	}

        // grab services from python
        output, err := exec.Command("python", sparkWorkflowsLocation + "/" + workflowscript, "-w").Output()
        
        if err != nil {
            badRequest(w, "internal failure to retrieve services")
            return
        }

	w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, string(output))
}

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	pathlist, requestType, err := parseURI(r, servicePath)
	if err != nil || len(pathlist) != 1 {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}
        
	if requestType == "get" {
            // grab service json schema from python
            output, err := exec.Command("python", sparkWorkflowsLocation + "/" + workflowscript, "-d", pathlist[0]).Output()
            
            if err != nil {
                badRequest(w, "failure to find schema for given service")
                return
            }

            w.Header().Set("Content-Type", "application/json")
            fmt.Fprintf(w, string(output))
        } else {
            // ?! create job id, call spark launch scripts, provide callback
        }
}


// Serve is the main server function call that creates http server and handlers
func Serve(port int, config_file string) {
        // read and parse configuration file
        config_handle, _ := os.Open(config_file)
        decoder := json.NewDecoder(config_handle)
	config_data := make(map[string]interface{})
        decoder.Decode(&config_data)
        config_handle.Close()
       
        remoteMachine = "" 
        if mach, found := config_data["remote-machine"]; found {
            remoteMachine = mach.(string)
        }
        remoteUser = "" 
        if ruser, found := config_data["remote-user"]; found {
            remoteUser = ruser.(string)
        }

        // might not be necessary if scripts are installed in
        // system bin directories
        if renv, found := config_data["remote-environment"]; found {
                env_list := renv.([]interface{})
                for _, envsing := range env_list {
                        remoteEnv = append(remoteEnv, envsing.(string))
                }
        }

        // get log path (error if doesn't exist)
        logDirectory = "" 
        if ldir, found := config_data["log-dir"]; found {
            logDirectory = ldir.(string)
        } else {
            fmt.Println("No log file specified.  Exiting...")
            os.Exit(-1)
        }

        // get spark workflow script locations (error if doesn't exist)
        sparkWorkflowsLocation = "" 
        if wdir, found := config_data["workflow-dir"]; found {
            sparkWorkflowsLocation = wdir.(string)
        } else {
            fmt.Println("No workflows location specfied.  Exiting...")
            os.Exit(-1)
        }

	hname, _ := os.Hostname()
	webAddress = hname + ":" + strconv.Itoa(port)

	fmt.Printf("Web server address: %s\n", webAddress)
	fmt.Printf("Running...\n")

	httpserver := &http.Server{Addr: webAddress}

	// serve out raml
	http.HandleFunc(interfacePath, interfaceHandler)

        // get available services
        http.HandleFunc(servicesPath, servicesHandler)

        // handle specific services
        http.HandleFunc(servicePath, serviceHandler)

        // show updates to job status
        http.HandleFunc(statusPath, statusHandler)

        // get location of front-end (don't enable front-end if it doesn't exist
	// perform calclabel service
        webConsole = "" 
        if cname, found := config_data["web-console"]; found {
            webConsole = cname.(string)
            
            // front page containing simple form
            http.HandleFunc("/", frontHandler)
	}

	// exit server if user presses Ctrl-C
	go func() {
		sigch := make(chan os.Signal)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch
		fmt.Println("Exiting...")
		os.Exit(0)
	}()

	httpserver.ListenAndServe()
}
