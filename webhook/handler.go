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
		namespace, err := splitRepository(event.Target.Repository)
		if err != nil {
			return err
		}

		deploymentIface := clientset.Apps().Deployments(namespace)

		return findAndUpdate(deploymentIface, newImage)
	default:
		log.Printf("Action %s is not handled now", event.Action)
	}

	return nil
}

func findAndUpdate(deploymentIface cliv1.DeploymentInterface, newImage string) error {
	var needUpdate bool
	deployments, err := deploymentIface.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, deployment := range deployments.Items {
		for i := 0; i < len(deployment.Spec.Template.Spec.Containers); i++ {
			if imagesBaseEqual(deployment.Spec.Template.Spec.Containers[i].Image, newImage) {
				deployment.Spec.Template.Spec.Containers[i].Image = newImage
				log.Printf("Scheduled updating image %s for %s", newImage, deployment.Spec.Template.Spec.Containers[i].Name)
				needUpdate = true
			}
		}

		for i := 0; i < len(deployment.Spec.Template.Spec.InitContainers); i++ {
			if imagesBaseEqual(deployment.Spec.Template.Spec.InitContainers[i].Image, newImage) {
				deployment.Spec.Template.Spec.InitContainers[i].Image = newImage
				log.Printf("Scheduled updating image %s for %s", newImage, deployment.Spec.Template.Spec.InitContainers[i].Name)
				needUpdate = true
			}
		}

		if needUpdate {
			if _, err := deploymentIface.Update(&deployment); err != nil {
				return err
			}
		}
	}
	return nil
}

func imagesBaseEqual(img1, img2 string) bool {
	img1sl := strings.Split(img1, ":")
	if len(img1sl) < 2 {
		return false
	}

	tags1 := strings.Split(img1sl[1], "-")

	img2sl := strings.Split(img2, ":")
	if len(img2sl) < 2 {
		return false
	}

	tags2 := strings.Split(img2sl[1], "-")

	return img1sl[0] == img2sl[0] && tags1[0] == tags2[0]
}

func splitRepository(repository string) (string, error) {
	repChunks := strings.Split(repository, "/")
	if len(repChunks) < 2 {
		return "", ErrBadChunksSize
	}
	return repChunks[0], nil
}
