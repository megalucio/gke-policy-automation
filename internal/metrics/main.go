package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/demizer/go-logs/src/logs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

// creataKubeDynamicClient()
func createKubeDynamicClient(kubeConfigPath string) (dynamic.Interface, error) {
	var config *rest.Config
	var err error

	if config, err = rest.InClusterConfig(); err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
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

// getKubeConfigPath retrieves the current kubeconfig path
func getKubeConfigPath() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		logs.Criticalf("unable to get userhome")
		return ""
	}
	// return flag.String("kubeconfig", fmt.Sprintf("%s/.kube/config", dirname), "kubeconfig location")
	return fmt.Sprintf("%s/.kube/config", dirname)
}

// FindResourcesGroups extracts Resource groups and names from the APIs
func findResourcesGroups(ctx context.Context, kubeconfigPath string) ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	kubeconfig, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, nil, err
	}
	restconfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		logs.Criticalln(err)
		return nil, nil, err
	}
	dc := discovery.NewDiscoveryClientForConfigOrDie(restconfig)
	apiGroup, apiResources, err := dc.ServerGroupsAndResources()
	if err != nil {
		logs.Criticalln(err)
		return nil, nil, err
	}
	return apiGroup, apiResources, nil
}

// printResourceGroups is a test
func printResourceGroups(apiGroup []*metav1.APIGroup, apiResources []*metav1.APIResourceList) {
	fmt.Println("apiGroup -------------------- ")
	for i := 0; i < len(apiGroup); i++ {
		println(apiGroup[i].Name)
	}

	fmt.Printf("ApiResourcesList count: %d\n", len(apiResources))
	for i := 0; i < len(apiResources); i++ {
		fmt.Printf("Group %s resources: %d\n", apiResources[i].GroupVersion, len(apiResources[i].APIResources))
		for k := 0; k < len(apiResources[i].APIResources); k++ {
			fmt.Printf("     %s\n", apiResources[i].APIResources[k].Name)
		}
	}
}

// countResources returns the number of instances of a given resource, by type, group and version
func countResources(ctx context.Context, client dynamic.Interface, namespace string, group string, version string, resource string) (int, error) {
	var counter int
	schemaResource := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	// fmt.Printf("%s %s %s\n", schemaResource.Group, schemaResource.Version, schemaResource.Resource)
	list, err := client.Resource(schemaResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	counter = len(list.Items)
	return counter, nil
}

// we should put an informer and count the existing
func main() {
	discoveryCtx := context.Background()
	_, apiResources, err := findResourcesGroups(discoveryCtx, getKubeConfigPath())
	if err != nil {
		logs.Criticalln(err)
		os.Exit(1)
	}
	// printResourceGroups(apiGroup, apiResources)
	ctx := context.Background()
	client, err := createKubeDynamicClient(getKubeConfigPath())
	if err != nil {
		logs.Criticalln(err)
		os.Exit(1)
	}
	for i := 0; i < len(apiResources); i++ {
		for k := 0; k < len(apiResources[i].APIResources); k++ {
			if strings.Contains(apiResources[i].GroupVersion, "/") {
				group := strings.Split(apiResources[i].GroupVersion, "/")[0]
				version := strings.Split(apiResources[i].GroupVersion, "/")[1]
				if counter, err := countResources(ctx, client, "", group, version, apiResources[i].APIResources[k].Name); err == nil {
					fmt.Printf("%s/%s %s found %d\n", group, version, apiResources[i].APIResources[k].Name, counter)
				}
			} else {
				if counter, err := countResources(ctx, client, "", apiResources[i].GroupVersion, "", apiResources[i].APIResources[k].Name); err == nil {
					fmt.Printf("%s %s found %d\n", apiResources[i].GroupVersion, apiResources[i].APIResources[k].Name, counter)
				}
			}
		}
	}
}

///////// - ARCHIVE

// func podsCounterTest(client *kubernetes.Clientset, namespace string) {
// 	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
// 	if err != nil {
// 		logs.Criticalf("Failed retrieving pods %v", err)
// 	}
// 	for _, pod := range podList.Items {
// 		logs.Printf("Pod %s \n", pod.Name)
// 	}
// }

// func podsInformerTest(client *kubernetes.Clientset, namespace string) {
// 	podSharedInformer := informers.NewPodInformer(client, namespace, 0, cache.Indexers{})
// 	pods := []*v1.Pod{}
// 	for _, p := range podSharedInformer.GetStore().List() {
// 		p := p.(*v1.Pod)
// 		logs.Println(p)
// 		pods = append(pods, p)
// 	}
// 	logs.Printf("pods: %d", len(pods))
// }

// var podResource = schema.GroupVersionResource{
// 	Group:    "",
// 	Version:  "v1",
// 	Resource: "pods",
// }

// func ListPods(ctx context.Context, client dynamic.Interface, namespace string) ([]unstructured.Unstructured, error) {
// 	list, err := client.Resource(podResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
// 	if err != nil {
// 		logs.Criticalln(err)
// 		return nil, err
// 	}
// 	for _, pod := range list.Items {
// 		logs.Println(pod.GetName())
// 	}
// 	return list.Items, nil
// }
