package controller

import (
	"github.com/loodse/operator-workshop/podhealth/operatorsdk/pkg/controller/podhealth"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, podhealth.Add)
}
