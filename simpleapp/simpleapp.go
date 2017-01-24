package main // simpleapp

import (
	"encoding/json"
	"fmt"
	"strings"

	kapi "k8s.io/client-go/pkg/api/v1"
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

func (app *AppService) Build(opts AppServiceBuildOpts) *kapi.Pod {
	container := kapi.Container{}
	var name string

	container.Image = app.Image
	imageParts := strings.Split(strings.Split(app.Image, ":")[0], "/")
	name = imageParts[len(imageParts)-1]

	if app.Name != "" {
		name = app.Name
	}

	container.Name = name
	labels := map[string]string{"name": name}

	pod := kapi.Pod{
		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: kapi.PodSpec{
			Containers: []kapi.Container{container},
		},
	}

	if opts.Namespace != "" {
		pod.ObjectMeta.Namespace = opts.Namespace
	}

	return &pod
}

func (app *App) Build() *kapi.PodList {
	opts := AppServiceBuildOpts{Namespace: app.Name}
	list := &kapi.PodList{}

	for _, service := range app.Services {
		list.Items = append(list.Items, *service.Build(opts))
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

	data, err := json.MarshalIndent(app.Build(), "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
