package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/seldonio/seldon-core/operator/api/v1alpha2"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const readyTimeout = 100 * time.Second

var (
	customResourceFile string
	namespace          string
)

func init() {
	// add seldon-core custom resource definition to kubernetes Scheme so that
	// it can decode and encode seldon-core custom resource.
	err := v1alpha2.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("failed to add seldon-core api to kubernetes scheme: %v", err)
	}

	// parse flags
	flag.StringVar(&customResourceFile, "cr_file", "", "cr_file is the file path to seldon custom resource")
	flag.StringVar(&namespace, "namespace", "default", "namespace is the namespace where you deploy the seldon custom resource; must be same namespace as the seldon operator")
	flag.Parse()
}

func main() {
	seldonCR := mustCreateSeldonObjFromFile(customResourceFile)
	applyDefaults(seldonCR)
	client := mustCreateClient()

	log.Printf("Creating custom resource (%v) ...", seldonCR.Name)
	err := client.Create(context.TODO(), seldonCR)
	if err != nil {
		log.Fatalf("failed to create custom resource (%v): %v", seldonCR.Name, err)
	}
	log.Printf("Custom resource (%v) created", seldonCR.Name)

	log.Println("Checking if Seldon deployment is already ...")
	done := false
	timeout := time.NewTimer(readyTimeout).C
	for !done {
		ready := &v1alpha2.SeldonDeployment{}
		err = client.Get(context.TODO(), types.NamespacedName{Name: seldonCR.Name, Namespace: seldonCR.Namespace}, ready)
		if err != nil {
			log.Fatalf("failed to get Seldon custom resource (%v): %v", seldonCR.Name, err)
		}
		log.Printf("Deployment state (%v)", ready.Status.State)
		if ready.Status.State == "Available" {
			done = true
			continue
		}
		select {
		case <-time.Tick(time.Second):
		case <-timeout:
			log.Printf("Deployment is not ready after timeout %v: exiting...", readyTimeout.String())
			done = true
		}
	}
	log.Println("Seldon deployment is already")

	log.Printf("Deleting Seldon custom resource (%v)", seldonCR.Name)
	err = client.Delete(context.TODO(), seldonCR)
	if err != nil {
		log.Fatalf("failed to delete custom resource (%v): %v", seldonCR.Name, err)
	}
	log.Printf("Seldon custom resource (%v) deleted", seldonCR.Name)
}

func mustCreateSeldonObjFromFile(filename string) *v1alpha2.SeldonDeployment {
	// create a universal decode from scheme pkg to decode bytes from file into SeldonDeployment obj.
	decode := scheme.Codecs.UniversalDeserializer().Decode
	jsonByt, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read Seldon custom resource file (%v): %v", filename, err)
	}

	sd := &v1alpha2.SeldonDeployment{}
	_, _, err = decode(jsonByt, nil, sd)
	if err != nil {
		log.Fatalf("failed to decode Seldon custom resource bytes into SeldonDeployment type: %v", err)
	}
	return sd
}

func applyDefaults(sd *v1alpha2.SeldonDeployment) {
	if sd == nil {
		return
	}
	// namespace must not be empty or else the client errors out.
	if len(sd.Namespace) == 0 {
		sd.Namespace = "default"
	}
}

func mustCreateClient() client.Client {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("failed to create Kubernetes config: %v", err)
	}

	client, err := client.New(cfg, client.Options{})
	if err != nil {
		log.Fatalf("failed to create Kubernetes client: %v", err)
	}
	return client
}
