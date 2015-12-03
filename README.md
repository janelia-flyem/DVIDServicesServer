# DVIDServicesServer [![Picture](https://raw.github.com/janelia-flyem/janelia-flyem.github.com/master/images/HHMI_Janelia_Color_Alternate_180x40.png)](http://www.janelia.org)
## Serves Web Interface to Launch Spark-based EM Services

This package implements a REST interface that exposes available DVID services.  It uses the [DVIDServicesConsole](https://github.com/janelia-flyem/DVIDServicesConsole) as a user-friendly front-end.  A user can launch a DVID service which will run on a Spark cluster.  The server temporarily stores the status of Spark jobs.

## Installation Instruction

The server is written in Go.  Install Go and set GOPATH as appropriate.
The package can then be installed by:

    % go get github.com/janelia-flyem/DVIDServicesServer

The SparkLaunch/ directory contains plugins for launching a spark cluster that calls
DVIDSparkServices with the appropriate parameters.  See notes on configuration below. 

[DVIDSparkServices](https://github.com/janelia-flyem/DVIDSparkServices) should be installed
where the server is running and also on the cluster running Spark.

The DVIDServicesServer uses [DVIDServicesConsole](https://github.com/janelia-flyem/DVIDServicesConsole) as a web-front end.  If this is not installed and pointed to, access to the server should be done directly through its REST api.

Set $GOPATH/bin to the executable path.

## Running the Server

To launch the server:

    % DVIDServicesServer config.json

config.json contains several configurable paramters including the location of the DVIDServicesConsole
and DVIDSparkServices.  By default, this will launch the server at port 15000 of the current
machine (specify custom port with -port).  If you are not using the web front-end, please retrieve
the interface by querying * < SERVER ADDRESS > : 15000/interface *

## Configuration and Launching the Spark Cluster

Several configurations need to be set properly to run on your target environment.  Users need to modify config.json and modify the desired plugin in SparkLaunch/ as appropriate.

1. The top-level config file expects the location of the static javascript in DVIDServicesConsole (download the release version and point to 'dist').  

2. DVIDSparkServices must be downloaded and the workflows path specified.

3. An executable that launches the spark cluster should be provided.  Make sure this exetuable is in the PATH.

4. If the executable must be launched via ssh, save the ssh password token to allow access through the program.

5. Consult the README under SparkLaunch/.  A plugin currently exists for the Janelia SGE compute cluster.

## Architecture Notes
This server queries the workflow manager in DVIDSparkServices to see which services are
available and accesses their JSON-schema interface.  The DVIDServicesConsole provides a front-end
interface to the JSON-schema for the different services and allows the user
to submit their job via web form.  The server can be accessed programmatically
through the [RAML](http://raml.org) REST API defined in /interface.

The server maintains the history of all submitted applications in memory.  The server
currently does not support offline storage.  The front-end interface can query
JOB status.  The launching of the Spark cluster and application is done
by spark_launch_wrapper and spark_launch.  These scripts are tuned to the Janelia
compute cluster.  The server calls spark_launch_wrapper which in turn
calls spark_launch, which is installed on the spark cluster.  spark_launch
starts the Spark cluster and calls the Spark workflow.  It also communicates
with the server by sending status information and querying the job configuration.

##TODO

* Option to automatically email user when application finishes.
* Better handle job failure situation.
* (Optional) Persist job status to disk.




