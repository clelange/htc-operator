# Examples

This folder includes examples/tests that display the functionality of HTCJob resources.

## Retcode

This directory contains two examples.

File `zero.yaml` containsa script, which returns a `0` return status.
This is reflected in the status of the HTCJob - when the job is completes in HTCondor,
the HTCJob is marked as `succeeded`.

```yaml
  status:
    active: 0
    failed: 0
    jobid:
    - "891628.0"
    succeeded: 1
```

File `notzero.yaml` is similar, the only difference is that its script returns a non-`0` status,
which marks the HTCJob status as `failed`.

```yaml
  status:
    active: 0
    failed: 1
    jobid:
    - "891629.0"
    succeeded: 0
```

## Command

The two files `R.yaml` and `python.yaml` display the execution of non-shell scripts.

## h2t

This example shows potential usage of an HTCJob resource in a real world analysis.
The workflow used is __HiggsTauTau__ analysis taken from [here](https://awesome-workshop.github.io/awesome-htautau-analysis/).
File `wf.yaml` contains the analysis in `Argo Workflows` without the use of HTCJob resources.

File `wf-htcjob.yaml` includes the same workflow, but utilizes an HTCJob resource.

Folder `root18` includes a Dockerfile, which builds the required image, since the one used in the source material
could not assure the correct version of `root`.

## wf-htcjobs

The `skim` steps is sent to HTCondor for execution.
In other to ensure that the HTCJob is also deleted when the Argo workflow is deleted,
the following is added to the metadata of HTCJob:

```yaml
          ownerReferences:
          - apiVersion: argoproj.io/v1alpha1
            blockOwnerDeletion: true
            kind: Workflow
            name: "{{workflow.name}}"
            uid: "{{workflow.uid}}"
```

Also, while the `ceph` volume storage is used for all other steps, the code that is executed inside HTCondor
environment has access to AFS and EOS only. Because of this, two additional jobs are added,
which transfer the data between the ceph volume and EOS; they use a special container for that.
Information about the EOS container can be found [here](https://clouddocs.web.cern.ch/containers/tutorials/eos.html).
