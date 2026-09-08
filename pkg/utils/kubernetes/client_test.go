package kubernetes

import (
	"os"
	"testing"
)

// TestIsRunningInCluster tests the isRunningInCluster function
func TestIsRunningInCluster(t *testing.T) {
	// Save original values
	origHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	origPort := os.Getenv("KUBERNETES_SERVICE_PORT")

	// Clean slate
	os.Unsetenv("KUBERNETES_SERVICE_HOST")
	os.Unsetenv("KUBERNETES_SERVICE_PORT")

	// Should be false when unset
	if isRunningInCluster() {
		t.Error("Expected false when KUBERNETES_SERVICE_HOST is unset")
	}

	// Set both variables
	os.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")

	// Should be true
	if !isRunningInCluster() {
		t.Error("Expected true when KUBERNETES_SERVICE_HOST is set")
	}

	// Set only host
	os.Setenv("KUBERNETES_SERVICE_PORT", "")
	if isRunningInCluster() {
		t.Error("Expected false when only KUBERNETES_SERVICE_HOST is set")
	}

	// Set only port
	os.Setenv("KUBERNETES_SERVICE_HOST", "")
	os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	if isRunningInCluster() {
		t.Error("Expected false when only KUBERNETES_SERVICE_PORT is set")
	}

	// Restore originals
	if origHost != "" {
		os.Setenv("KUBERNETES_SERVICE_HOST", origHost)
	} else {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
	}

	if origPort != "" {
		os.Setenv("KUBERNETES_SERVICE_PORT", origPort)
	} else {
		os.Unsetenv("KUBERNETES_SERVICE_PORT")
	}
}

// TestGetKubeConfigFromPath tests various kubeconfig loading scenarios
func TestGetKubeConfigFromPath(t *testing.T) {
	// Test with explicit non-existent path
	_, err := GetKubeConfigFromPath("/nonexistent/path/to/config")
	if err == nil {
		t.Error("Expected error for non-existent kubeconfig path")
	}

	// Test with empty path (should use default or in-cluster)
	_, err = GetKubeConfigFromPath("")
	if err == nil {
		t.Error("Expected error when kubeconfig path is empty and no default config exists")
	}

	// Clean up KUBECONFIG env var for the test
	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Unsetenv("KUBECONFIG")

	// Test with existing home kubeconfig (~/.kube/config if it exists)
	// This should not fail (either succeed or give a different error)
	_, err = GetKubeConfigFromPath("")
	// Either succeeds or gives a meaningful error about kubeconfig
	if err != nil && !containsString(err.Error(), "kubeconfig") && !containsString(err.Error(), "in-cluster") {
		t.Errorf("Unexpected error type: %v", err)
	}

	// Restore KUBECONFIG
	if origKubeconfig != "" {
		os.Setenv("KUBECONFIG", origKubeconfig)
	}
}

// TestNewKubeClient tests the main NewKubeClient function
func TestNewKubeClient(t *testing.T) {
	// Test with empty master and kubeconfig (should use default)
	_, err := NewKubeClient("", "")
	if err == nil {
		t.Error("Expected error when both master and kubeconfig are empty")
	}

	// Clean up KUBECONFIG
	origKubeconfig := os.Getenv("KUBECONFIG")
	os.Unsetenv("KUBECONFIG")

	// Test with empty arguments (should handle gracefully)
	_, err = NewKubeClient("", "")
	// Should get a meaningful error about kubeconfig
	if err == nil {
		t.Error("Expected error when both master and kubeconfig are empty")
	}

	// Restore KUBECONFIG
	if origKubeconfig != "" {
		os.Setenv("KUBECONFIG", origKubeconfig)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
