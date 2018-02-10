package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
)

type deploymentListTypeAlias = appsv1.DeploymentList

//go:generate gotemplate "./templates/resource" "deploymentsClass(\"Deployments\", deployments, deploymentListTypeAlias)"

func (c *deploymentsClass) getList(ns string, listOptions metav1.ListOptions) (*appsv1.DeploymentList, error) {
	return c.rk.clientset.Apps().Deployments(ns).List(listOptions)
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
