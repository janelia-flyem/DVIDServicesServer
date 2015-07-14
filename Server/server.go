package Server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// Contain URI location for interface
	interfacePath  = "/interface/"
	servicesPath   = "/services/"
	servicePath    = "/service/"
	statusPath     = "/jobstatus/"
	sparkScript    = "spark_launch_wrapper"
	numNodes       = "16"
	workflowscript = "launchworkflow.py"
)

// contains settings for launching spark script
var executableParams ExeParams

// location of service server
var webAddress string

// location of web console source
var webConsole string

// location of service workflows directory
var sparkWorkflowsLocation string

// all jobs on the server (in memory DB)
var JobManager *jobManager

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

// badRequest is a helper for printing an http error message
func badRequest(w http.ResponseWriter, msg string) {
	fmt.Println(msg)
	http.Error(w, msg, http.StatusBadRequest)
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
		http.ServeFile(w, r, webConsole+"/"+r.URL.Path[:])
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
	pathlist, requestType, err := parseURI(r, statusPath)
	if err != nil || len(pathlist) != 1 {
		badRequest(w, "Error: incorrectly formatted request")
		return
	}

	// get copy of spark job
	jobinfo, found := JobManager.GetJobStatus(pathlist[0])
	if !found {
		badRequest(w, "Error: job id not found")
		return
	}

	if requestType == "get" {
		// send job status
		outputData := make(map[string]interface{})
		outputData["job_status"] = jobinfo.status
		outputData["job_message"] = jobinfo.message
		outputData["sparkAddr"] = jobinfo.spark_address
		if jobinfo.status == "Finished" || jobinfo.status == "Error" {
			outputData["runtime"] = jobinfo.runtime
		} else {
			outputData["runtime"] = time.Now().Unix() - jobinfo.runtime
		}
		outputData["config"] = jobinfo.configuration

		jsonbytes, _ := json.Marshal(outputData)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(jsonbytes))

		return
	} else if requestType == "post" {
		// grab job status (need status and spark callback)
		decoder := json.NewDecoder(r.Body)
		var json_data map[string]interface{}
		err = decoder.Decode(&json_data)
		if err != nil {
			badRequest(w, "poorly formed JSON")
			return
		}

		if status, found := json_data["job_status"]; !found {
			badRequest(w, "no status provided")
			return
		} else {
			jobinfo.status, _ = status.(string)
			if jobinfo.status == "Finished" || jobinfo.status == "Error" {
				jobinfo.runtime = time.Now().Unix() - jobinfo.runtime
			}
		}

		if spark, found := json_data["sparkAddr"]; !found {
			badRequest(w, "no spark callback provided")
			return
		} else {
			jobinfo.spark_address, _ = spark.(string)
		}

		if message, found := json_data["job_message"]; found {
			jobinfo.message, _ = message.(string)
		}

		// load new status into job manager
		JobManager.SetJobStatus(jobinfo.GetID(), jobinfo)

		fmt.Fprintf(w, "")
	} else {
		badRequest(w, "must be get or post")
	}
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
	output, err := exec.Command("python", sparkWorkflowsLocation+"/"+workflowscript, "-w").Output()

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
		output, err := exec.Command("python", sparkWorkflowsLocation+"/"+workflowscript, "-d", pathlist[0]).Output()

		if err != nil {
			badRequest(w, "failure to find schema for given service")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(output))
	} else if requestType == "post" {
		// launch job

		decoder := json.NewDecoder(r.Body)
		var json_data map[string]interface{}
		err = decoder.Decode(&json_data)
		if err != nil {
			badRequest(w, "poorly formed JSON")
			return
		}

		// create job
		spark_job := NewSparkJob(pathlist[0], json_data)

		JobManager.SetJobStatus(spark_job.GetID(), *spark_job)

		// launch job
		err := spark_job.StartJob(executableParams, webAddress)
		if err != nil {
			badRequest(w, err.Error())
			return
		}

		// write back call back
		w.Header().Set("Content-Type", "application/json")
		jsonbytes, _ := json.Marshal(map[string]interface{}{"callBack": statusPath + spark_job.GetID()})
		fmt.Fprintf(w, string(jsonbytes))

	} else {
		badRequest(w, "only supports gets and posts")
	}
}

// Serve is the main server function call that creates http server and handlers
func Serve(port int, config_file string) {
	// initialize job list
	JobManager = NewJobManager()

	// read and parse configuration file
	config_handle, _ := os.Open(config_file)
	decoder := json.NewDecoder(config_handle)
	config_data := make(map[string]interface{})
	decoder.Decode(&config_data)
	config_handle.Close()

	remoteMachine := ""
	if mach, found := config_data["remote-machine"]; found {
		remoteMachine = mach.(string)
	}
	remoteUser := ""
	if ruser, found := config_data["remote-user"]; found {
		remoteUser = ruser.(string)
	}

	// might not be necessary if scripts are installed in
	// system bin directories
	remoteEnv := make([]string, 0)
	if renv, found := config_data["remote-environment"]; found {
		env_list := renv.([]interface{})
		for _, envsing := range env_list {
			remoteEnv = append(remoteEnv, envsing.(string))
		}
	}

	// get log path (error if doesn't exist)
	logDirectory := ""
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

	// initialize ExeParams
	executableParams = ExeParams{logDirectory, remoteMachine, remoteUser, sparkScript, numNodes, remoteEnv}

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
