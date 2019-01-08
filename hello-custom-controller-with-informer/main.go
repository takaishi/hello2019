package main

import (
	clientset "github.com/takaishi/hello2019/hello-custom-controller-with-informer/pkg/client/clientset/versioned"
	informers "github.com/takaishi/hello2019/hello-custom-controller-with-informer/pkg/client/informers/externalversions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	app := cli.NewApp()
	app.Flags = []cli.Flag{}

	app.Action = func(c *cli.Context) error {
		return action(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Printf(err.Error())
	}

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Printf(err.Error())
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(client, time.Second*30, informers.WithNamespace("default"))
	informer := informerFactory.Samplecontroller().V1alpha().Foos()
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	})

	stopCh := SetupSignalHandler()
	go informerFactory.Start(stopCh)

	log.Printf("[DEBUG] start")
	<-stopCh
	log.Printf("[DEBUG] shutdown")
	return nil
}

var onlyOneSignalHandler = make(chan struct{})
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func SetupSignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler)

	stop := make(chan struct{})
	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		close(stop)
		<-c
		os.Exit(1)
	}()

	return stop
}

func addFunc(obj interface{}) {
	log.Printf("[DEBUG] addFunc")
	log.Printf("[DEBUG] obj: %+v\n", obj)
}

func updateFunc(old, obj interface{}) {
	log.Printf("[DEBUG] updateFunc")
	log.Printf("[DEBUG] old: %+v\n", old)
	log.Printf("[DEBUG] obj: %+v\n", obj)
}

func deleteFunc(obj interface{}) {
	log.Printf("[DEBUG] deleteFunc")
	log.Printf("[DEBUG] obj: %+v\n", obj)
}
