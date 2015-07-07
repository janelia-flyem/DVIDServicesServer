# DVIDServicesServer [![Picture](https://raw.github.com/janelia-flyem/janelia-flyem.github.com/master/images/HHMI_Janelia_Color_Alternate_180x40.png)](http://www.janelia.org)
## Serves Web Interface to Launch Spark-based EM Services

This package implements a REST interface that exposes available DVID services.  It uses the [DVIDServicesConsole](https://github.com/janelia-flyem/DVIDServicesConsole) as a user-friendly front-end.  A user can launch a DVID service which will run on a Spark cluster.  The server temporarily stores the status of Spark jobs.

## Installation Instruction

The server is written in Go.  Install Go and set GOPATH as appropriate.
The package can then be installed by:

    % go get github.com/janelia-flyem/DVIDServicesServer

The SparkLaunch/ directory contains two Python executables that should are called
from the server -- spark_launch_wrapper and spark_launch.  These should be installed
in the executable path.  These scripts launch the Spark cluster and Spark application.

[DVIDSparkServices](https://github.com/janelia-flyem/DVIDSparkServices) needs to be installed
on the target compute cluster and the workflows must be available to the server machine.

The DVIDServicesServer requires the webconsole to be installed [DVIDServicesConsole](https://github.com/janelia-flyem/DVIDServicesConsole).

Set $GOPATH/bin to the executable path.

## Running the Server

To launch the server:

    % DVIDServicesServer config.json

config.json contains several configurable paramters including the location of the DVIDServicesConsole
and DVIDSparkServices.  By default, this will launch the server at port 15000 of the current
machine (specify -port allows custom specification).

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
starts the Spark cluster and calls the Spark workflow.  It also provides status
information back to the server and queries the application configuration file.

** Configuration **

Users need to modify config.json, spark_launch_wrapper, and spark_launch as appropriate.
spark_launch_wrapper expects the following arguments:

* Number of Spark workers for the cluster
* Name of the service plugin
* Directory location for the log file
* Callback to server for posting application status

##TODO

* Provide time stamps for application start and finish.
* Automatically email user when application finishes.
* Provide interfaces for query previous application jobs.
* (Optional) Persist application status to disk.



### Customization for Non-Janelia Cluster



