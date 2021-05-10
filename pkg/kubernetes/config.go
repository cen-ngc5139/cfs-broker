package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	traitv1 "github.com/ghostbaby/cfs-broker/pkg/api/v1"
	"github.com/ghostbaby/cfs-broker/pkg/g"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewKubeClient() (*kubernetes.Clientset, *rest.Config, error) {

	var restConf *rest.Config
	var err error
	restConf, err = rest.InClusterConfig()
	if err != nil {
		var kubeConfig string

		if len(g.Config().KubeConfig) > 0 {
			kubeConfig = g.Config().KubeConfig
		} else {
			home := HomeDir()
			kubeConfig = filepath.Join(home, ".kube", "config")
		}

		restConf, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			fmt.Println("get kubeconfig err", err)
			return nil, nil, err
		}
	}

	kubeClient, err := kubernetes.NewForConfig(restConf)
	if err != nil {
		return nil, nil, err
	}
	return kubeClient, restConf, nil
}

func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func GetKubernetesClient() (client.Client, kubernetes.Interface, *rest.RESTClient, error) {

	sc := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sc)
	_ = traitv1.AddToScheme(sc)

	kubeClient, config, err := NewKubeClient()
	if err != nil {
		return nil, nil, nil, err
	}

	cli, err := client.New(config, client.Options{Scheme: sc})
	if err != nil {
		fmt.Println("create client from kubeconfig err", err)
		return nil, nil, nil, err
	}

	crdClient, err := newCRDClient(config)
	if err != nil {
		return nil, nil, nil, err
	}

	return cli, kubeClient, crdClient, nil
}

func configureConfig(cfg *rest.Config) (*rest.Config, error) {

	traitv1.AddToScheme(scheme.Scheme)

	config := *cfg

	config.GroupVersion = &traitv1.GroupVersion
	config.APIPath = "/apis"
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)

	return &config, nil
}

func newCRDClient(config *rest.Config) (*rest.RESTClient, error) {

	cfg, err := configureConfig(config)
	if err != nil {
		return nil, err
	}

	crdClient, err := rest.RESTClientFor(cfg)
	if err != nil {
		return nil, err
	}

	return crdClient, nil
}
