package main // simpleapp

import (
	"encoding/json"
	"fmt"
	"strings"

	meta "k8s.io/client-go/pkg/api/unversioned" // Should eventually migrate to "k8s.io/apimachinery/pkg/apis/meta/v1"?
	kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type AppServiceOpts struct {
	PrometheusPath string
	StandardProbes bool
	OnlyDeployment bool
}

type AppService struct {
	Image    string
	Name     string
	Port     int
	Opts     AppServiceOpts
	Env      map[string]string
	Override interface{}
}

type AppServiceBuildOpts struct {
	Namespace string
}

type App struct {
	Name     string
	Services []AppService
}

// TODO figure out how to use kapi.List here, if we can
type AppServiceResourcePair struct {
	Deployment *kext.Deployment
	Service    *kapi.Service
}

const (
	DEFAULT_REPLICAS = 1
)

func (app *AppService) GetNameAndLabels() (string, map[string]string) {
	var name string

	imageParts := strings.Split(strings.Split(app.Image, ":")[0], "/")
	name = imageParts[len(imageParts)-1]

	if app.Name != "" {
		name = app.Name
	}

	labels := map[string]string{"name": name}

	return name, labels
}

func (app *AppService) BuildPod(opts AppServiceBuildOpts) *kapi.PodTemplateSpec {
	name, labels := app.GetNameAndLabels()
	container := kapi.Container{Name: name, Image: app.Image}

	pod := kapi.PodTemplateSpec{
		ObjectMeta: kapi.ObjectMeta{
			Labels: labels,
		},
		Spec: kapi.PodSpec{
			Containers: []kapi.Container{container},
		},
	}

	return &pod
}

func (app *AppService) BuildDeployment(opts AppServiceBuildOpts, pod *kapi.PodTemplateSpec) *kext.Deployment {
	if pod == nil {
		return nil
	}

	replicas := int32(DEFAULT_REPLICAS)

	name, labels := app.GetNameAndLabels()
	deploymentSpec := kext.DeploymentSpec{
		Replicas: &replicas,
		Selector: &meta.LabelSelector{MatchLabels: labels},
		Template: *pod,
	}

	deployment := &kext.Deployment{
		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: deploymentSpec,
	}

	if opts.Namespace != "" {
		deployment.ObjectMeta.Namespace = opts.Namespace
	}

	return deployment
}

func (app *AppService) BuildService(opts AppServiceBuildOpts) *kapi.Service {
	return nil
}

func (app *AppService) Build(opts AppServiceBuildOpts) AppServiceResourcePair {
	pod := app.BuildPod(opts)

	return AppServiceResourcePair{
		app.BuildDeployment(opts, pod),
		app.BuildService(opts),
	}
}

func (app *App) Build() []AppServiceResourcePair {
	opts := AppServiceBuildOpts{Namespace: app.Name}
	list := []AppServiceResourcePair{}

	for _, service := range app.Services {
		list = append(list, service.Build(opts))
	}

	return list
}

func main() {
	altPromPath := AppServiceOpts{PrometheusPath: "/prometheus"}
	noStandardProbes := AppServiceOpts{StandardProbes: false}

	app := App{
		Name: "sock-shop",
		Services: []AppService{
			{
				Image: "weaveworksdemos/cart:0.4.0",
				Opts:  altPromPath,
			},
			{
				Image: "weaveworksdemos/catalogue-db:0.3.0",
				Port:  3306,
				Opts:  noStandardProbes,
				Env: map[string]string{
					"MYSQL_ROOT_PASSWORD": "fake_password",
					"MYSQL_DATABASE":      "socksdb",
				},
			},
		},
	}
	for _, resources := range app.Build() {
		deployment, err := json.MarshalIndent(resources.Deployment, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(deployment))

		service, err := json.MarshalIndent(resources.Service, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(service))
	}
}
