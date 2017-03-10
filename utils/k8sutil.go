package utils

import (
	"fmt"
	"net/http"

	apierrors "k8s.io/kubernetes/pkg/api/errors"
	unversionedAPI "k8s.io/kubernetes/pkg/api/unversioned"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func GetFactory() *cmdutil.Factory {
	factory := cmdutil.NewFactory(nil)
	return factory
}

func IsKubernetesResourceAlreadyExistError(err error) bool {
	se, ok := err.(*apierrors.StatusError)
	if !ok {
		return false
	}
	if se.Status().Code == http.StatusConflict && se.Status().Reason == unversionedAPI.StatusReasonAlreadyExists {
		return true
	}
	return false
}

func WatchResources(host, ns string, httpClient *http.Client) (*http.Response, error) {
    return httpClient.Get(fmt.Sprintf("%s/apis/k8s.io/v1/configpolicies?watch=true",host))
}
