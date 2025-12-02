package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	openshiftClientset "github.com/openshift/client-go/operator/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Global K8s Client
var ocClient *openshiftClientset.Clientset

func main() {
	// 1. Initialize In-Cluster Config (Connect to K8s API)
	config, err := rest.InClusterConfig()
	if err != nil {
		panic("Failed to connect to K8s. Are we running inside a pod?")
	}
	ocClient, err = openshiftClientset.NewForConfig(config)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*") // Load HTML files

	// 2. Serve the Dashboard
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"status": "Ready",
		})
	})

	// 3. Action: Modify Ingress Controller (Elevated Action)
	// Example: Changing the number of replicas or tuning parameters
	r.POST("/router/scale", func(c *gin.Context) {
		ctx := context.TODO()
		
		// Get the default IngressController
		ingress, err := ocClient.OperatorV1().IngressControllers("openshift-ingress-operator").Get(ctx, "default", metav1.GetOptions{})
		if err != nil {
			c.String(500, "Error getting ingress: "+err.Error())
			return
		}

		// Example Logic: Toggle between 2 and 3 replicas
		if ingress.Spec.Replicas != nil && *ingress.Spec.Replicas == 2 {
			replicas := int32(3)
			ingress.Spec.Replicas = &replicas
		} else {
			replicas := int32(2)
			ingress.Spec.Replicas = &replicas
		}

		_, err = ocClient.OperatorV1().IngressControllers("openshift-ingress-operator").Update(ctx, ingress, metav1.UpdateOptions{})
		if err != nil {
			c.String(500, "Failed to update router: "+err.Error())
			return
		}

		c.String(200, fmt.Sprintf("Router Scaled to %d replicas", *ingress.Spec.Replicas))
	})

	// 4. Action: Start Load Test (Launch k6 Job)
	r.POST("/test/start", func(c *gin.Context) {
		// logic to use k8s client to spawn a `batch/v1 Job` running k6
		c.String(200, "K6 Job Launched! 🚀")
	})

	r.Run(":8080")
}
