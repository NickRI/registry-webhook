package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	cliv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

//HandlerWrapper wrap and executes webhooks
func HandlerWrapper(clientset *kubernetes.Clientset) func(http.ResponseWriter, *http.Request) {
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

//UpdateImage executes image updating for deployments
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

		deploymentIface := clientset.Apps().Deployments(namespace)

		deployments, err := deploymentIface.List(metav1.ListOptions{})
		if err != nil {
			return err
		}

		return findAndUpdate(deployments.Items, deploymentIface, project, newImage)
	default:
		log.Printf("Action %s is not handled now", event.Action)
	}

	return nil
}

func findAndUpdate(deployments []v1.Deployment, deploymentIface cliv1.DeploymentInterface, project, newImage string) error {
	for _, deployment := range deployments {
		for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
			if deployment.Spec.Template.Spec.Containers[i].Name == project {
				deployment.Spec.Template.Spec.Containers[i].Image = newImage
				if _, err := deploymentIface.Update(&deployment); err != nil {
					return err
				}
				log.Printf("Start updating image %s for %s", newImage, project)
			}
		}

		for i := 0; i < len(deployment.Spec.Template.Spec.InitContainers); i++ {
			if deployment.Spec.Template.Spec.InitContainers[i].Name == project {
				deployment.Spec.Template.Spec.InitContainers[i].Image = newImage
				if _, err := deploymentIface.Update(&deployment); err != nil {
					return err
				}
				log.Printf("Start updating image %s for %s", newImage, project)
			}
		}
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
