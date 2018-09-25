package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WebHookHandlerWrapper(clientset *kubernetes.Clientset) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var hook = new(RegistyHook)
		defer req.Body.Close()
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := json.NewDecoder(req.Body).Decode(&hook); err != nil {
			log.Printf("Post decode error: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		repr, err := json.MarshalIndent(hook, "", "  ")
		if err != nil {
			log.Printf("Post marshal error: %+v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Recieved webhook from %s:\n%s", req.RemoteAddr, string(repr))

		for _, event := range hook.Events {
			if err := UpdateImage(clientset, &event); err != nil {
				log.Printf("Failed to update %s with %+v", event.Target.Repository, err)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateImage(clientset *kubernetes.Clientset, event *Event) error {
	evtURL, err := url.Parse(event.Target.URL)
	if err != nil {
		return err
	}

	newImage := fmt.Sprintf("%s/%s:%s", evtURL.Host, event.Target.Repository, event.Target.Tag)

	switch event.Action {
	case "pull":
		log.Printf("Seems image %s for %s is updating", newImage, event.Target.Repository)
	case "push":
		namespace, project, err := splitRepository(event.Target.Repository)
		if err != nil {
			return err
		}

		deployment, err := clientset.Apps().Deployments(namespace).Get(project, metav1.GetOptions{})
		if err != nil {
			return err
		}

		deployment.Spec.Template.Spec.Containers[0].Image = newImage
		log.Printf("Start updating image %s for %s", newImage, event.Target.Repository)

		_, err = clientset.Apps().Deployments(namespace).Update(deployment)
		return err
	default:
		log.Printf("Action %s is not handled now", event.Action)
	}

	return nil
}

func splitRepository(repository string) (string, string, error) {
	repChunks := strings.Split(repository, "/")
	if len(repChunks) < 2 {
		return "", "", ErrBadChunksSize
	}
	return repChunks[0], repChunks[1], nil
}
