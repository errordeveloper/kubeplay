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
	logBuffer bytes.Buffer
	pod       *kapi.Pod
}

func newPodLogsClassInstanceVars(c *podLogsClass, s *mruby.MrbValue, args ...mruby.Value) (*podLogsClassInstanceVars, error) {
	return &podLogsClassInstanceVars{}, nil
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

				stream, err := c.rk.clientset.Core().Pods(vars.pod.ObjectMeta.Namespace).GetLogs(vars.pod.ObjectMeta.Name, &kapi.PodLogOptions{}).Stream()
				if err != nil {
					return nil, createException(m, err.Error())
				}
				defer stream.Close()
				io.Copy(&vars.logBuffer, stream)
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

				scanner := bufio.NewScanner(&vars.logBuffer)
				for scanner.Scan() {
					fmt.Println(scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					return nil, createException(m, err.Error())
				}
				fmt.Println(vars.logBuffer.Len())

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

				commonPrefix := fmt.Sprintf("%s/%s", vars.pod.ObjectMeta.Namespace, vars.pod.ObjectMeta.Name)

				scanner := bufio.NewScanner(&vars.logBuffer)
				for scanner.Scan() {
					line := scanner.Text()
					for _, re := range matchAgainst {
						if re.MatchString(line) {
							fmt.Printf("[%s] %s\n", commonPrefix, line)
						}
					}
				}

				if err := scanner.Err(); err != nil {
					return nil, createException(m, err.Error())
				}
				fmt.Println(vars.logBuffer.Len())

				return self, nil
			},
			instanceMethod,
		},
	})
}

func (o *podLogsClassInstance) Update() (mruby.Value, error) {
	return call(o.self, "get!")
}
