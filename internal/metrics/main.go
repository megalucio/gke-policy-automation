package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/demizer/go-logs/src/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

func createKubeClient(kubeConfigPath *string) *kubernetes.Clientset {
	var config *rest.Config
	var err error

	// try in-cluster config, and then default to kubeconfig
	if config, err = rest.InClusterConfig(); err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfigPath)
		if err != nil {
			logs.Criticalf("failed parsing kubeconfig %v", err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logs.Criticalf("failed creating the clientset %v", err.Error())
	}
	return clientset
}

func getKubeConfigPath() *string {
	var kubeconfigPath *string
	dirname, err := os.UserHomeDir()
	if err != nil {
		logs.Criticalf("unable to get userhome")
		return kubeconfigPath
	}
	kubeconfigPath = flag.String("kubeconfig", fmt.Sprintf("%s/.kube/config", dirname), "kubeconfig location")
	return kubeconfigPath
}

func main() {
	var defaultNamespace = flag.String("namespace", "kube-system", "namespace to use for k8s")
	client := createKubeClient(getKubeConfigPath())
	podList, err := client.CoreV1().Pods(*defaultNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logs.Criticalf("Failed retrieving pods %v", err)
	}
	for _, pod := range podList.Items {
		logs.Printf("Pod %s \n", pod.Name)
	}
}
