package e2e

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const nginxDeploymentConf = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 1
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
`

const nginxServiceConf = `
kind: Service
apiVersion: v1
metadata:
  name: nginx
spec:
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: LoadBalancer
`

var clientset *kubernetes.Clientset

func initK8SClient() {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func getNodes() ([]corev1.Node, error) {
	nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

func createNamespace(name string) error {
	_, err := clientset.Core().Namespaces().Create(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return err
}

func deleteNamespace(name string) error {
	return clientset.Core().Namespaces().Delete(name, nil)
}

func createNginxDeployment() error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(nginxDeploymentConf), nil, nil)
	if err != nil {
		return err
	}

	deploymentsClient := clientset.AppsV1().Deployments("lbtest")
	result, err := deploymentsClient.Create(obj.(*appsv1.Deployment))
	fmt.Printf("Created deployment %s\n", result.GetObjectMeta().GetName())
	return err
}

func createNginxService() error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(nginxServiceConf), nil, nil)
	if err != nil {
		return err
	}

	servicesClient := clientset.CoreV1().Services("lbtest")
	result, err := servicesClient.Create(obj.(*corev1.Service))
	fmt.Printf("Created service %s\n", result.GetObjectMeta().GetName())
	return err

}

func getSvcExternalIP(namespace, serviceName string, timeout time.Duration) (string, error) {
	svcListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "services", namespace,
		fields.SelectorFromSet(map[string]string{"metadata.name": serviceName}))
	evt, err := cache.ListWatchUntil(timeout, svcListWatcher, func(evt watch.Event) (bool, error) {
		lb := evt.Object.(*corev1.Service).Status.LoadBalancer
		return (len(lb.Ingress) > 0 && lb.Ingress[0].IP != ""), nil
	})
	if err != nil {
		return "", nil
	}
	return evt.Object.(*corev1.Service).Status.LoadBalancer.Ingress[0].IP, nil
}
