package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
	kext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type deploymentListTypeAlias kext.DeploymentList

//go:generate gotemplate "./templates/resource" "deploymentsClass(\"Deployments\", deployments, deploymentListTypeAlias)"

func (c *deploymentsClass) getList(ns string, listOptions kapi.ListOptions) (*kext.DeploymentList, error) {
	return c.rk.clientset.Extensions().Deployments(ns).List(listOptions)
}

func (c *deploymentsClass) getItem(deployments deploymentListTypeAlias, index int) (*deploymentClassInstance, error) {
	newDeploymentObj, err := c.rk.classes.Deployment.New()
	if err != nil {
		return nil, err
	}
	deployment := deployments.Items[index]
	newDeploymentObj.vars.deployment = deploymentTypeAlias(deployment)
	return newDeploymentObj, nil
}

//go:generate gotemplate "./templates/resource/list" "deploymentsListModule(deploymentsClass, \"Deployments\", deployments, deploymentListTypeAlias)"

func (c *deploymentsClass) defineOwnMethods() {
	c.defineListMethods()
}

func (o *deploymentsClassInstance) Update(args ...*mruby.MrbValue) (mruby.Value, error) {
	return call(o.self, "get!", args...)
}
