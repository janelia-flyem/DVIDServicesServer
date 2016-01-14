# Launching a Spark Job on the Janelia SGE

sparklaunch_google is the entry program and can technically be called standalone by
replacing the callback address with the config json file path for the spark job.
sparklaunch_google_int is called by sparklaunch_google.

The script uses bdutil to deploy a google cluster, run a spark job, and delete the cluster.

To run these scripts, ensure the following:

* add bdutil to the executable path
* modify the environment variables defined in custombdutil.env
* add sparklaunch_google and sparklaunch_google_int to the executable path
* specify the lcoation of custombdutil.env in sparklaunch_google_int
* create a custom image with DVIDSparkServices installed and modify the custom environment as appropriate
