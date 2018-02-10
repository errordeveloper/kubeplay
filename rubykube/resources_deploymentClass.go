package rubykube

import (
	mruby "github.com/mitchellh/go-mruby"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
)

type deploymentTypeAlias = appsv1.Deployment

//go:generate gotemplate "./templates/resource" "deploymentClass(\"Deployment\", deployment, deploymentTypeAlias)"

func (c *deploymentClass) getSingleton(ns, name string) (*appsv1.Deployment, error) {
	return c.rk.clientset.Apps().Deployments(ns).Get(name, metav1.GetOptions{})
}

//go:generate gotemplate "./templates/resource/singleton" "deploymentSingletonModule(deploymentClass, \"deployment\", deployment, deploymentTypeAlias)"

//go:generate gotemplate "./templates/resource/podfinder" "deploymentPodFinderModule(deploymentClass, \"deployment\", deployment, deploymentTypeAlias)"

func (c *deploymentClass) defineOwnMethods() {
	c.defineSingletonMethods()
	c.definePodFinderMethods()

	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"replicasets": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				ns := vars.deployment.ObjectMeta.Namespace

				listOptions := metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(vars.deployment.Spec.Selector)}

				replicaSets, err := c.rk.clientset.Apps().ReplicaSets(ns).List(listOptions)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				newReplicaSetsObj, err := c.rk.classes.ReplicaSets.New()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				newReplicaSetsObj.vars.replicaSets = replicaSetListTypeAlias(*replicaSets)
				return newReplicaSetsObj.self, nil
			},
			instanceMethod,
		},
	})
}

func (o *deploymentClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
