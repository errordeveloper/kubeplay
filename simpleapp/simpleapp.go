package main // simpleapp

import (
	"encoding/json"
	"fmt"
	"strings"

	meta "k8s.io/client-go/pkg/api/unversioned" // Should eventually migrate to "k8s.io/apimachinery/pkg/apis/meta/v1"?
	kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type AppComponentOpts struct {
	PrometheusPath string
	StandardProbes bool
	OnlyDeployment bool
}

type AppComponent struct {
	Image    string
	Name     string
	Port     int
	Replicas int32
	Opts     AppComponentOpts
	Env      map[string]string
	Override interface{}
}

type AppComponentBuildOpts struct {
	Namespace string
}

type App struct {
	Name  string
	Group []AppComponent
}

// TODO figure out how to use kapi.List here, if we can
type AppComponentResourcePair struct {
	Deployment *kext.Deployment
	Service    *kapi.Service
}

const (
	DEFAULT_REPLICAS = 1
)

func (i *AppComponent) GetNameAndLabels() (string, map[string]string) {
	var name string

	imageParts := strings.Split(strings.Split(i.Image, ":")[0], "/")
	name = imageParts[len(imageParts)-1]

	if i.Name != "" {
		name = i.Name
	}

	labels := map[string]string{"name": name}

	return name, labels
}

func (i *AppComponent) GetMeta() kapi.ObjectMeta {
	name, labels := i.GetNameAndLabels()
	return kapi.ObjectMeta{
		Name:   name,
		Labels: labels,
	}
}

func (i *AppComponent) BuildPod(opts AppComponentBuildOpts) *kapi.PodTemplateSpec {
	name, labels := i.GetNameAndLabels()
	container := kapi.Container{Name: name, Image: i.Image}

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

func (i *AppComponent) BuildDeployment(opts AppComponentBuildOpts, pod *kapi.PodTemplateSpec) *kext.Deployment {
	if pod == nil {
		return nil
	}

	replicas := int32(DEFAULT_REPLICAS)

	if i.Replicas != nil {
		replicas = i.Replicas
	}

	name, labels := i.GetNameAndLabels()
	deploymentSpec := kext.DeploymentSpec{
		Replicas: &replicas,
		Selector: &meta.LabelSelector{MatchLabels: labels},
		Template: *pod,
	}

	deployment := &kext.Deployment{
		ObjectMeta: i.GetMeta(),
		Spec:       deploymentSpec,
	}

	if opts.Namespace != "" {
		deployment.ObjectMeta.Namespace = opts.Namespace
	}

	return deployment
}

func (i *AppComponent) BuildService(opts AppComponentBuildOpts) *kapi.Service {

	service := &kapi.Service{
		ObjectMeta: i.GetMeta(),
	}
	return nil
}

func (i *AppComponent) Build(opts AppComponentBuildOpts) AppComponentResourcePair {
	pod := i.BuildPod(opts)

	return AppComponentResourcePair{
		i.BuildDeployment(opts, pod),
		i.BuildService(opts),
	}
}

func (i *App) Build() []AppComponentResourcePair {
	opts := AppComponentBuildOpts{Namespace: i.Name}
	list := []AppComponentResourcePair{}

	for _, service := range i.Components {
		list = append(list, service.Build(opts))
	}

	return list
}

func main() {
	altPromPath := AppComponentOpts{PrometheusPath: "/prometheus"}
	noStandardProbes := AppComponentOpts{StandardProbes: false}

	app := App{
		Name: "sock-shop",
		Services: []AppComponent{
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
