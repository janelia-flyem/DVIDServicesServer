# Launching a Spark Job on the Janelia SGE

sparklaunch_janelia is the entry program and can technically be called standalone by
replacing the callback address with the config json file path for the spark job.
sparklaunch_janelia_int is called by sparklaunch_janelia, which is executed on a cluster node

To run these scripts, ensure the following

* add sparklaunch_janelia  and sparklaunch_janelia_int to the executable path
* install drmaa into python used for dvidsparkservices (easy_install drmaa)
* sparklaunch_janelia should be run on a login SGE node (like login1 or login2)
* Specify the following parameters in sparklaunch_janelia_int
    * SPARK_HOME: location of spark distribution
    * CONF_DIR: location of spark configuration settings (defaults in the SPARK_HOME/conf are likely okay)
    * LOG_DIR: path for the log files
