package Server

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// randomHex computes a random hash for storing service results
func randomHex() (randomStr string) {
	randomStr = ""
	for i := 0; i < 8; i++ {
		val := rand.Intn(16)
		randomStr += strconv.FormatInt(int64(val), 16)
	}
	return
}

// ExeParams contains execution params for spark jobs
type ExeParams struct {
	remote_machine         string   // location of remote machine
	remote_user            string   // name of remote user
	cluster_script         string   // name of cluster launch script
	num_nodes              string   // number of spark nodes to launch
	remote_env             []string // slice of environment variables to set
	cluster_workflowscript string   // location of workflow launch script
	cluster_python         string   // location of cluster python
}

type sparkJob struct {
	service_type  string // what service is being run
	job_id        string // auto-generated job ID
	status        string // current status ("Waiting")
	runtime       int64  // initial time stamp before completion, then time in seconds
	message       string
	configuration map[string]interface{}
	spark_address string
}

func NewSparkJob(service_name string, config map[string]interface{}) *sparkJob {
	job := new(sparkJob)
	job.service_type = service_name

	// make random job id
	job.job_id = randomHex()

	// grab a timestamp (could overflow but is just used for a unique stamp)
	tstamp := int(time.Now().Unix())
	job.job_id = job.job_id + "-" + strconv.Itoa(tstamp)

	job.status = "Waiting"
	job.message = ""
	job.configuration = config
	job.runtime = time.Now().Unix()

	job.spark_address = ""

	return job
}

func (job *sparkJob) GetID() string {
	return job.job_id
}

// StartJob executes the spark launch script with the following
// parameters: <num nodes> <workflow name> <job id>
// <callback address> <spark services workflow script>
// <python location>
func (job *sparkJob) StartJob(exe_params ExeParams, web_address string) error {
	var err error
	err = nil

	if exe_params.remote_machine == "" {
		out, err2 := exec.Command(exe_params.cluster_script, exe_params.num_nodes, job.service_type, job.job_id, "http://"+web_address+"/jobstatus/"+job.job_id, exe_params.cluster_workflowscript, exe_params.cluster_python).Output()
		err = err2
		if err2 != nil {
			err = fmt.Errorf(string(out))
		}
	} else {
		var argument_str string
		for _, envvar := range exe_params.remote_env {
			// assume shell allows for export of variables
			argument_str += "export " + envvar + "; "
		}

		argument_str += (exe_params.cluster_script)
		argument_str += " " + exe_params.num_nodes + " " + job.service_type + " " + job.job_id + " " + "http://" + web_address + "/jobstatus/" + job.job_id + " " + exe_params.cluster_workflowscript + " " + exe_params.cluster_python

		out, err2 := exec.Command("ssh", exe_params.remote_user+"@"+exe_params.remote_machine, argument_str).Output()
		err = err2
		if err2 != nil {
			err = fmt.Errorf(string(out))
		}
	}

	return err
}

type jobManager struct {
	job_list map[string]sparkJob
	mutex    *sync.Mutex
}

func NewJobManager() *jobManager {
	return &jobManager{make(map[string]sparkJob), &sync.Mutex{}}
}

func (manager *jobManager) SetJobStatus(jobid string, jobinfo sparkJob) {
	manager.mutex.Lock()
	manager.job_list[jobid] = jobinfo
	manager.mutex.Unlock()
}

func (manager *jobManager) GetJobStatus(jobid string) (sparkJob, bool) {
	manager.mutex.Lock()
	// retrieve a copy of sparkJob
	var jobinfo sparkJob
	jobinfo, found := manager.job_list[jobid]
	manager.mutex.Unlock()

	return jobinfo, found
}
