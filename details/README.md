# Core files

Descriptions of the main files that comprise the operator

## pkg/apis/

- `htc/v1alpha1/htcjob_types.go` - definition of the `HTCJob` type

## pkg/controller/

- `htcjob/htcjob_controller.go` - main file describing the logic of the controller
- `htcjob/db.go` - functions to interact with the database
- `htcjob/send.go` - functions to create HTCondor jobs

## build/bin/

- `ensureDB` - script that ensures the existance of the DB
- `subCondor` and `rmCondor` - scripts used to send and remove HTCondor jobs
- `sender` and `receiver` - cloudevents applications
- `main` - script that is run inside the operator Pod, combines all the previous scripts
