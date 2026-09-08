package kubernetes

import (
	"fmt"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	kubeclient "k8s.io/client-go/kubernetes"
)

const (
	kubeConfigHint = `Make sure to either:
  - Set the environment variable: export KUBECONFIG=/path/to/config
  - Or use: --kubeconfig=/path/to/config`
)

// isRunningInCluster checks if the process is running inside a Kubernetes cluster
func isRunningInCluster() bool {
	// Check for environment variables that indicate in-cluster config
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
		os.Getenv("KUBERNETES_SERVICE_PORT") != ""
}

func NewKubeClient(masterUrl string, kubeconfigPath string) (kubeClient *kubeclient.Clientset, err error) {
	if masterUrl == "" && kubeconfigPath == "" {
		// Determine the appropriate config to use based on where we're running
		kubeconfig, err := GetKubeConfigFromPath(kubeconfigPath)
		if err != nil {
			return nil, err
		}

		kubeClient, err = kubeclient.NewForConfig(kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes client: %w\n\n%s", err, kubeConfigHint)
		}

		return kubeClient, nil
	}

	kubeconfig, err := GetKubeConfigFromPath(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	kubeClient, err = kubeclient.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w\n\n%s", err, kubeConfigHint)
	}

	return kubeClient, nil
}

func GetKubeConfigFromPath(kubeconfigPath string) (*rest.Config, error) {
	// If a specific kubeconfig path is provided, check if it exists
	if kubeconfigPath != "" {
		if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("provided kubeconfig path does not exist: %s\n\n%s", kubeconfigPath, kubeConfigHint)
		}
	}

	// Determine if we're running in-cluster
	inCluster := isRunningInCluster()

	// If running in-cluster and no kubeconfig is specified, try in-cluster config first
	if inCluster && kubeconfigPath == "" {
		kubeconfig, err := rest.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}
		// Fall through to try default kubeconfig
	}

	// Build config from flags
	// Empty kubeconfigPath will use default config loading (in-cluster or ~/.kube/config)
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		var hint string
		if kubeconfigPath == "" {
			if inCluster {
				hint = "Failed to load in-cluster configuration. Ensure KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT are defined."
			} else {
				hint = fmt.Sprintf("Failed to load default kubeconfig. Make sure KUBECONFIG is set or ~/.kube/config exists.\n\n%s", kubeConfigHint)
			}
			return nil, fmt.Errorf("%s", hint)
		}
		return nil, fmt.Errorf("failed to load kubeconfig from path '%s': %v\n\n%s", kubeconfigPath, err, kubeConfigHint)
	}

	return kubeconfig, nil
}
