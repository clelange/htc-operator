package apis

import (
	"gitlab.cern.ch/cms-cloud/htc-operator/pkg/apis/htc/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and bac
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
