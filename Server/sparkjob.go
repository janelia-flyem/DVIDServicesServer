package Server

import (
	//"fmt"
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
	log_dir        string   // location for log files
	remote_machine string   // location of remote machine
	remote_user    string   // name of remote user
	cluster_script string   // name of cluster launch script
	num_nodes      string   // number of spark nodes to launch 
	remote_env     []string // slice of environment variables to set
}

type sparkJob struct {
	service_type  string // what service is being run
	job_id        string // auto-generated job ID
	log_loc       string // where are results stored
	status        string // current status ("Waiting")
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

	job.spark_address = ""

	return job
}

func (job *sparkJob) GetID() string {
	return job.job_id
}

func (job *sparkJob) StartJob(exe_params ExeParams, web_address string) error {
	job.log_loc = exe_params.log_dir + "/" + job.job_id + "/"

	var err error
	err = nil

	if exe_params.remote_machine == "" {
		_, err2 := exec.Command(exe_params.cluster_script, exe_params.num_nodes, job.service_type, job.log_loc, web_address+"/jobstatus/"+job.job_id).Output()
		err = err2
	} else {
		var argument_str string
		for _, envvar := range exe_params.remote_env {
			// assume shell allows for export of variables
			argument_str += "export " + envvar + "; "
		}
		argument_str += (exe_params.cluster_script)
		_, err2 := exec.Command("ssh", exe_params.remote_user+"@"+exe_params.remote_machine, argument_str, exe_params.num_nodes, job.service_type, job.log_loc, web_address+"/jobstatus/"+job.job_id).Output()
		err = err2
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
