package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

		deployments, err := clientset.Apps().Deployments(namespace).List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		dep, container := findCorrectDeploymentContainer(deployments.Items, project)
		if dep != nil {
			container.Image = newImage
			fmt.Println(dep.Spec.Template.Spec.InitContainers[0].Image, container.Image, newImage)
			if _, err = clientset.Apps().Deployments(namespace).Update(dep); err != nil {
				return err
			}
			log.Printf("Start updating image %s for %s", newImage, event.Target.Repository)
		}
	default:
		log.Printf("Action %s is not handled now", event.Action)
	}

	return nil
}

func findCorrectDeploymentContainer(deployments []v1.Deployment, project string) (*v1.Deployment, *corev1.Container) {
	for _, deployment := range deployments {
		for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
			if deployment.Spec.Template.Spec.Containers[i].Name == project {
				return &deployment, &deployment.Spec.Template.Spec.Containers[i]
			}
		}

		for i := 0; i < len(deployment.Spec.Template.Spec.InitContainers); i++ {
			if deployment.Spec.Template.Spec.InitContainers[i].Name == project {
				return &deployment, &deployment.Spec.Template.Spec.InitContainers[i]
			}
		}
	}
	return nil, nil
}

func splitRepository(repository string) (string, string, error) {
	repChunks := strings.Split(repository, "/")
	if len(repChunks) < 2 {
		return "", "", ErrBadChunksSize
	}
	return repChunks[0], repChunks[1], nil
}
