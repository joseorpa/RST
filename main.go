package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	openshiftClientset "github.com/openshift/client-go/operator/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
)

// Global K8s Standard Client
var kubeClient *kubernetes.Clientset

func main() {
    // ... config loading (same as before)
    config, _ := rest.InClusterConfig()
    
    // Initialize Standard K8s Client (For Jobs)
    kubeClient, _ = kubernetes.NewForConfig(config)
    
    // Initialize OpenShift Client (For Routes - from previous step)
    ocClient, _ = openshiftClientset.NewForConfig(config)

    // ... Gin Setup

    // 4. Action: Start Load Test
    r.POST("/test/start", func(c *gin.Context) {
        // Hardcoded examples - in real life, get these from the Web Form (c.PostForm)
        target := "http://my-target-service/target"
        vus := 50
        
        // Launch the Job!
        // We assume the controller runs in "rst-namespace"
        err := launchK6Job(kubeClient, "rst-namespace", target, vus)
        
        if err != nil {
            c.String(500, "Failed to launch k6: " + err.Error())
            return
        }
        
        c.String(200, fmt.Sprintf("🚀 Attack launched with %d Users against %s", vus, target))
    })

    r.Run(":8080")
}
