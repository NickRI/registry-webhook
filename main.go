package main

import (
	"log"
	"net/http"

	"github.com/NickRI/registry-webhook/webhook"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Configuration in cluster error: %+v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicf("Clietnset get error: %+v", err)
	}

	http.HandleFunc("/webhook", webhook.WebHookHandlerWrapper(clientset))
	log.Println("Registry-webhook service listen on :8089")
	log.Fatal(http.ListenAndServe(":8089", nil))
}
