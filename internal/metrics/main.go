package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/demizer/go-logs/src/logs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// creataKubeDynamicClient()
func createKubeDynamicClient(kubeConfigPath *string) (dynamic.Interface, error) {
	var config *rest.Config
	var err error

	if config, err = rest.InClusterConfig(); err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeConfigPath)
		if err != nil {
			logs.Criticalf("failed parsing kubeconfig %v", err.Error())
			return nil, err
		}
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return dynClient, nil
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

func podsCounterTest(client *kubernetes.Clientset, namespace string) {
	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logs.Criticalf("Failed retrieving pods %v", err)
	}
	for _, pod := range podList.Items {
		logs.Printf("Pod %s \n", pod.Name)
	}
}

func podsInformerTest(client *kubernetes.Clientset, namespace string) {
	podSharedInformer := informers.NewPodInformer(client, namespace, 0, cache.Indexers{})
	pods := []*v1.Pod{}
	for _, p := range podSharedInformer.GetStore().List() {
		p := p.(*v1.Pod)
		logs.Println(p)
		pods = append(pods, p)
	}
	logs.Printf("pods: %d", len(pods))
}

var podResource = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "pods",
}

func ListPods(ctx context.Context, client dynamic.Interface, namespace string) ([]unstructured.Unstructured, error) {
	list, err := client.Resource(podResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Criticalln(err)
		return nil, err
	}
	for _, pod := range list.Items {
		logs.Println(pod.GetName())
	}
	return list.Items, nil
}

func countResources(ctx context.Context, client dynamic.Interface, namespace string) (int, error) {
	var counter int

	list, err := client.Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Criticalln(err)
		return 0, err
	}
	counter = len(list.Items)
	return counter, nil
}

func findResources(kubeconfigPath *string) {
	kubeconfig, _ := ioutil.ReadFile(*kubeconfigPath)
	restconfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		logs.Criticalln(err)
		os.Exit(1)
	}
	dc := discovery.NewDiscoveryClientForConfigOrDie(restconfig)
	groupResources, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		logs.Criticalln(err)
		os.Exit(1)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := mapper.RESTMapping(schema.ParseGroupKind("Deployment.apps"), "v1")
	if err != nil {
		logs.Criticalln(err)
		os.Exit(1)
	}
	fmt.Println(mapping.Resource)
}

// out of band CRDs

func main() {
	// ctx := context.Background()
	// client, err := createKubeDynamicClient(getKubeConfigPath())
	// if err != nil {
	// 	logs.Criticalln(err)
	// 	os.Exit(1)
	// }
	// size, err := countResources(ctx, client, "")
	// if err != nil {
	// 	logs.Criticalln(err)
	// 	os.Exit(1)
	// }
	// logs.Printf("Objects found: %d\n", size)
	findResources(getKubeConfigPath())
}
