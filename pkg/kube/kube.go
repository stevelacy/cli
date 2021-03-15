package kube

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DopplerHQ/cli/pkg/utils"
	apiv1 "k8s.io/api/core/v1"
	errorsv1 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var client *kubernetes.Clientset
var dopplerKey = "secret"

func KubeClient() (*kubernetes.Clientset, error) {
	if client != nil {
		return client, nil
	}
	var kubeconfig string
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		k, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(k)
	}
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = filepath.Join(home, ".kube/config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
		utils.HandleError(errors.New("Unable to connect to the kubernetes API"))
	}

	return kubernetes.NewForConfig(config)
}

func SyncKubeSecret(kubernetesSecretName, kubernetesNamespace, encryptedResponse string) {
	stringData := map[string]string{
		dopplerKey: encryptedResponse,
	}
	clientset, err := KubeClient()
	if err != nil {
		fmt.Println(err.Error())
		utils.HandleError(errors.New("Unable to connect to the kubernetes API"))
	}
	value := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubernetesSecretName,
		},
		StringData: stringData,
	}

	_, err = clientset.CoreV1().Secrets(kubernetesNamespace).Get(context.TODO(), kubernetesSecretName, metav1.GetOptions{})
	if errorsv1.IsNotFound(err) {
		_, err = clientset.CoreV1().Secrets(kubernetesNamespace).Create(context.TODO(), value, metav1.CreateOptions{})
		if err != nil {
			fmt.Println(err.Error())
			utils.HandleError(errors.New("Unable to create kubernetes secret"))
		}
	} else {
		_, err = clientset.CoreV1().Secrets(kubernetesNamespace).Update(context.TODO(), value, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println(err.Error())
			utils.HandleError(errors.New("Unable to update kubernetes secret"))
		}
	}

}

// GetKubeSecret returns the encrypted values from k8s
func GetKubeSecret(kubernetesSecretName, kubernetesNamespace string) ([]byte, error) {
	clientset, err := KubeClient()
	if err != nil {
		fmt.Println(err.Error())
		utils.HandleError(errors.New("Unable to connect to the kubernetes API"))
	}
	ks, err := clientset.CoreV1().Secrets(kubernetesNamespace).Get(context.TODO(), kubernetesSecretName, metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}
	return ks.Data[dopplerKey], nil
}
