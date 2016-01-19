# Launching a Spark Job on the Janelia SGE

sparklaunch_google is the entry program and can technically be called standalone by
replacing the callback address with the config json file path for the spark job.
sparklaunch_google_int is called by sparklaunch_google.

## Plugin Configuration

The script uses bdutil to deploy a google cluster, run a spark job, and delete the cluster.

To run these scripts, ensure the following:

* add bdutil to the executable path
* modify the environment variables defined in custombdutil.env
* add sparklaunch_google and sparklaunch_google_int to the executable path
* specify the location of custombdutil.env in sparklaunch_google_int
* create a custom image with DVIDSparkServices installed and modify the custom environment as appropriate

## Starting VM Instances

### Server

The user should install the DVIDServicesServer and supporting packages by executing the commands in server_vm_setup as root.  This script should run correctly Debian and Ubuntu distributions.  In addition, to installing the server, this commmand also automatically launches the service server on boot and also starts a local DVID point to google bucket storage on port 8000.  The user should ensure that the VM is configured to allow traffic on ports 80 and 8000.
### Workers

The server_vm_setup script can be used to create the workers as well.  It is recommended that one creates a custom image with this script and then point to this image in the custombdutil.env.  The script embeds a configuration for dvid.  This should be modified to point to the google bucket that will store dvid data.
