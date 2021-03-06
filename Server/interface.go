package Server

// String representing interface for DVID Spark services
const ramlInterface = `#%%RAML 0.8
title: Interface for DVID Services
baseUri: /
/services:
  get:
    description: "List services available to server" 
    responses:
      200:
        body:
          application/json:
            schema: |
              { "$schema": "http://json-schema.org/schema#",
                "title": "List of DVID services",
                "type": "array",
                "items": {"type": "string"}
              }
/service/{servicename}:
  get:
    description: "Retrieve JSON schema for requested service" 
    responses:
      200:
        body:
          application/json: !!null
  post:
    description: "Launch service with posted JSON (schema not validated on server)" 
    body:
      application/json: !!null
    responses:
      200:
        body:
          application/json: 
            schema: |
              { "$schema": "http://json-schema.org/schema#",
                "title": "Response to Job submission",
                "type": "object",
                "properties": {
                  "callBack": {
                    "description": "URL for job status (embeds job ID)",
                    "type": "string"
                  }
                },
                "required": ["callback", "sparkAddr"]
              }
/jobid/{jobid}:
  get:
    description: "Retrieves job status"
    responses:
      200:
        body:
          application/json:
            schema: |
              { "$schema": "http://json-schema.org/schema#",
                "title": "Spark job status",
                "type": "object",
                "properties": {
                  "job_status": {
                    "description": "State of the job",
                    "type": "string",
                    "enum": [ "Waiting", "Running", "Finished", "Error" ]
                  },
                  "job_message": {
                    "description": "Information related to the job status",
                    "type": "string"
                  },
                  "sparkAddr": {
                    "description": "Address for monitoring spark job (can be used to access REST api for Spark >=1.4",
                    "type": "string"
                  },
		  "runtime": {
		    "description": "Current runtime for job in seconds",
		    "type": "string"
		  },
                  "config" : {
                    "description": "Configuration file",
                    "type": "object"
                  }
                },
                "required": ["job_status", "job_message", "sparkAddr", "runtime", "config"]
              }
  post:
    description: "Set job status (should only be done by Spark driver program)"
    body:
      application/json:
        schema: |
          { "$schema": "http://json-schema.org/schema#",
            "title": "Spark job status",
            "type": "object",
            "properties": {
              "job_status": {
                "description": "State of the job",
                "type": "string",
                "enum": [ "Started", "Finished", "Error" ]
              },
              "job_message": {
                "description": "Information related to the job status",
                "type": "string"
              },
              "sparkAddr": {
                "description": "Address for monitoring spark job (can be used to access REST api for Spark >=1.4",
                "type": "string"
              }
            },
            "required": ["job_status", "sparkAddr"]
          }
/interface:
  get:
    description: "Get the interface for the cluster calculate label service"
    responses:
      200:
        body:
          application/raml+yaml:
`
