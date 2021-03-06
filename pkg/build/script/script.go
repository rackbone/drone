package script

import (
	"io/ioutil"
	"strings"

	"launchpad.net/goyaml"

	"github.com/drone/drone/pkg/build/buildfile"
	"github.com/drone/drone/pkg/build/script/deployment"
	"github.com/drone/drone/pkg/build/script/notification"
	"github.com/drone/drone/pkg/build/script/publish"
)

func ParseBuild(data []byte) (*Build, error) {
	build := Build{}

	// parse the build configuration file
	err := goyaml.Unmarshal(data, &build)
	return &build, err
}

func ParseBuildFile(filename string) (*Build, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ParseBuild(data)
}

// Build stores the configuration details for
// building, testing and deploying code.
type Build struct {
	// Image specifies the Docker Image that will be
	// used to virtualize the Build process.
	Image string

	// Name specifies a user-defined label used
	// to identify the build.
	Name string

	// Script specifies the build and test commands.
	Script []string

	// Env specifies the environment of the build.
	Env []string

	// Services specifies external services, such as
	// database or messaging queues, that should be
	// linked to the build environment.
	Services []string

	Deploy        *deployment.Deploy         `yaml:"deploy,omitempty"`
	Publish       *publish.Publish           `yaml:"publish,omitempty"`
	Notifications *notification.Notification `yaml:"notify,omitempty"`
}

// Write adds all the steps to the build script, including
// build commands, deploy and publish commands.
func (b *Build) Write(f *buildfile.Buildfile) {
	// append build commands
	b.WriteBuild(f)

	// write publish commands
	if b.Publish != nil {
		b.Publish.Write(f)
	}

	// write deployment commands
	if b.Deploy != nil {
		b.Deploy.Write(f)
	}
}

// WriteBuild adds only the build steps to the build script,
// omitting publish and deploy steps. This is important for
// pull requests, where deployment would be undesirable.
func (b *Build) WriteBuild(f *buildfile.Buildfile) {
	// append environment variables
	for _, env := range b.Env {
		parts := strings.Split(env, "=")
		if len(parts) != 2 {
			continue
		}
		f.WriteEnv(parts[0], parts[1])
	}

	// append build commands
	for _, cmd := range b.Script {
		f.WriteCmd(cmd)
	}
}

type Publish interface {
	Write(f *buildfile.Buildfile)
}

type Deployment interface {
	Write(f *buildfile.Buildfile)
}

type Notification interface {
	Set(c Context)
}

type Context interface {
	Host() string
	Owner() string
	Name() string

	Branch() string
	Hash() string
	Status() string
	Message() string
	Author() string
	Gravatar() string

	Duration() int64
	HumanDuration() string

	//Settings
}
