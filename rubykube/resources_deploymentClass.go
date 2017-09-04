package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type deploymentTypeAlias kext.Deployment

//go:generate gotemplate "./templates/resource" "deploymentClass(\"Deployment\", deployment, deploymentTypeAlias)"

func (c *deploymentClass) getSingleton(ns, name string) (*kext.Deployment, error) {
	return c.rk.clientset.Extensions().Deployments(ns).Get(name)
}

//go:generate gotemplate "./templates/resource/singleton" "deploymentSingletonModule(deploymentClass, \"deployment\", deployment, deploymentTypeAlias)"

//go:generate gotemplate "./templates/resource/podfinder" "deploymentPodFinderModule(deploymentClass, \"deployment\", deployment, deploymentTypeAlias)"

func (c *deploymentClass) defineOwnMethods() {
	c.defineSingletonMethods()
	c.definePodFinderMethods()
}

func (o *deploymentClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
