package rubykube

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"

	mruby "github.com/mitchellh/go-mruby"
	kapi "k8s.io/client-go/pkg/api/v1"
)

type podLogsClassInstanceVars struct {
	pods []kapi.Pod
	logs map[string]*bytes.Buffer
}

func newPodLogsClassInstanceVars(c *podLogsClass, s *mruby.MrbValue, args ...mruby.Value) (*podLogsClassInstanceVars, error) {
	return &podLogsClassInstanceVars{logs: make(map[string]*bytes.Buffer)}, nil
}

//go:generate gotemplate "./templates/basic" "podLogsClass(\"PodLogs\", newPodLogsClassInstanceVars, podLogsClassInstanceVars)"

func (c *podLogsClass) defineOwnMethods() {
	c.rk.appendMethods(c.class, map[string]methodDefintion{
		"get!": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				for _, pod := range vars.pods {
					logBuffer := bytes.Buffer{}

					for _, container := range pod.Spec.Containers {
						name := fmt.Sprintf("%s/%s:%s", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, container.Name)
						stream, err := c.rk.clientset.Core().Pods(pod.ObjectMeta.Namespace).GetLogs(pod.ObjectMeta.Name, &kapi.PodLogOptions{Container: container.Name}).Stream()
						if err != nil {
							return nil, createException(m, err.Error())
						}
						defer stream.Close()
						io.Copy(&logBuffer, stream)
						vars.logs[name] = &logBuffer
					}
				}
				return self, nil
			},
			instanceMethod,
		},
		"to_s": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				//vars, err := c.LookupVars(self)
				//if err != nil {
				//	return nil, createException(m, err.Error())
				//}

				return self, nil
			},
			instanceMethod,
		},
		"puts": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				for name, logBuffer := range vars.logs {
					scanner := bufio.NewScanner(logBuffer)
					for scanner.Scan() {
						fmt.Printf("[%s] %s\n", name, scanner.Text())
					}
					if err := scanner.Err(); err != nil {
						return nil, createException(m, err.Error())
					}
				}

				return self, nil
			},
			instanceMethod,
		},
		"grep": {
			mruby.ArgsNone(), func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, mruby.Value) {
				vars, err := c.LookupVars(self)
				if err != nil {
					return nil, createException(m, err.Error())
				}

				args := m.GetArgs()

				if len(args) == 0 {
					return nil, createException(m, "At least one argument must be specified")
				}

				matchAgainst := []*regexp.Regexp{}
				for _, arg := range args {
					if arg.Type() == mruby.TypeString {
						re, err := regexp.Compile(arg.String())
						if err != nil {
							return nil, createException(m, err.Error())
						}
						matchAgainst = append(matchAgainst, re)
					}
				}

				grep := func(name string, logBuffer *bytes.Buffer) error {
					scanner := bufio.NewScanner(logBuffer)
					for scanner.Scan() {
						line := scanner.Text()
						for _, re := range matchAgainst {
							if re.MatchString(line) {
								fmt.Printf("[%s] %s\n", name, line)
							}
						}
					}

					if err := scanner.Err(); err != nil {
						return err
					}
					return nil
				}

				for name, logBuffer := range vars.logs {
					if err := grep(name, logBuffer); err != nil {
						return nil, createException(m, err.Error())
					}
				}

				return self, nil
			},
			instanceMethod,
		},
	})
}

func (o *podLogsClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
