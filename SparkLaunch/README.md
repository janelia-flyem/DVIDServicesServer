# Plugins for Launching and Running a Spark Job

This directory contains plugins for launching a spark cluster and runnning a job.  Each
plugin should have an executable that is called by the server (as specified in the config.json).

The plugin should handle the following input parameters:

* number of spark worker nodes
* name of the workflow plugin to use in DVIDSparkServices
* the job id given by the server (this can be used to create a unique log name)
* a callback address.  /config contains the config file for running DVIDSparkServices, POSTs can be done to update the status of the job and indicate the location of the master node.  The POST should be a JSON with a the following format:
    {
        "sparkAddr": "address of master spark node",
        "job_status": "state of current job",
        "job_message": "message from current job"
    }
* the location of the spark services workflow launch script
* the location of the spark services python executable

Each plugin directory will contain a README with information about any custom parameters that need to be set.
