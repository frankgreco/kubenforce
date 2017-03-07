package main

import (
	"fmt"
	"time"
    "log"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
    "k8s.io/client-go/tools/cache"
    "k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/rest"
    "k8s-audit/issue"
    "k8s-audit/config"
)

var clientset *kubernetes.Clientset

func main() {
    // grab the k8s configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
    // create a new client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
    // starting watching resources
    watchResources(clientset)
    // keep alive <-- perhaps there's a better way to do this
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func watchResources(clientset *kubernetes.Clientset) {
    watchlist := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())
    _, controller := cache.NewInformer(
        watchlist,
        &v1.Pod{},
        time.Second * 30,
        cache.ResourceEventHandlerFuncs{
            AddFunc: podAdded,
            UpdateFunc: podUpdated,
        },
    )
    stop := make(chan struct{})
    go controller.Run(stop)
}

func podAdded(obj interface{}) {
    pod := obj.(*v1.Pod)

    if pod.ObjectMeta.Namespace == "default" || pod.ObjectMeta.Namespace == "kube-system" {
        return
    }

    if _, ok := pod.ObjectMeta.Labels["source"]; !ok {
        //delete pod
        // opts := v1.DeleteOptions{}
        // err := clientset.Core().Pods(pod.ObjectMeta.Namespace).Delete(pod.ObjectMeta.Name, &opts)
		// if err != nil {
        //     fmt.Print(pod)
		// 	panic(err.Error())
		// }
        return
    }

    fmt.Print(pod.Spec.HostNetwork)

    if pod.Spec.HostNetwork && pod.Spec.HostNetwork == true {
        title := "Pod enables hostNetwork"
        body := "*NOTE: this issue was automatically generated*\n\n**Issue:**\nDue to security reasons, Pods cannot utilize the `hostNetwork` option in this Namespace.\n\n**Offending Code:**\n```yaml\nhostNetwork: true\n```\n**How to Fix:**\nPlease disable this options and redeploy"
        state := "open"

        issueTemplate := config.Issue{
            Owner: pod.ObjectMeta.Labels["source"],
            Repo: pod.ObjectMeta.Labels["app"],
            Title: &title,
            Body: &body,
            State: &state,
        }
        issue.Create(&issueTemplate)
    }

    // err := clientset.Core().Pods(pod.ObjectMeta.Namespace).Delete(pod.ObjectMeta.Name, &v1.DeleteOptions{})
    // if err != nil {
    //     panic(err.Error())
    // }
}

func podUpdated(oldObj, newObj interface{}) {
    pod := newObj.(*v1.Pod)
    fmt.Println("Pod created: " + pod.ObjectMeta.Name)
}
