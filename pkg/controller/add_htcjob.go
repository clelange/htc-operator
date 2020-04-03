package controller

import (
    "htc-operator/pkg/controller/htcjob"
)

func init() {
    // AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
    AddToManagerFuncs = append(AddToManagerFuncs, htcjob.Add)
}
