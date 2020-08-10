package k8s

import (
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var once sync.Once
var client *kubernetes.Clientset
var clientErr error

// Client returns k8s client
func Client() (*kubernetes.Clientset, error) {
	once.Do(func() {
		path := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			clientErr = err
			return
		}
		client, clientErr = kubernetes.NewForConfig(config)
	})
	return client, clientErr
}
