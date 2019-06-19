package v1alpha1

import (
	"strconv"

	"github.com/knative/pkg/kmeta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	BuildNumberLabel = "image.build.pivotal.io/buildNumber"
	ImageLabel       = "image.build.pivotal.io/image"
)

func (im *Image) BuildNeeded(lastBuild *Build, builder *Builder) bool {
	if lastBuild == nil {
		return true
	}

	if im.configMatches(lastBuild) && builtWithBuilderBuildpacks(builder, lastBuild) {
		return false
	}

	return true
}

func builtWithBuilderBuildpacks(builder *Builder, build *Build) bool {
	for _, bp := range build.Status.BuildMetadata {
		if !builder.Status.BuilderMetadata.Include(bp) {
			return false
		}
	}

	return true
}

func (im *Image) configMatches(build *Build) bool {
	return im.Spec.Image == build.Spec.Image &&
		im.Spec.Source == build.Spec.Source
}

func (im *Image) CreateBuild(builder *Builder) *Build {
	return &Build{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: im.generateBuildName(),
			OwnerReferences: []v1.OwnerReference{
				*kmeta.NewControllerRef(im),
			},
			Labels: map[string]string{
				BuildNumberLabel: im.nextBuildNumber(),
				ImageLabel:       im.Name,
			},
		},
		Spec: BuildSpec{
			BuildNumber:    im.nextBuildNumber(),
			Image:          im.Spec.Image,
			Builder:        builder.Spec.Image,
			ServiceAccount: im.Spec.ServiceAccount,
			Source:         im.Spec.Source,
		},
	}
}

func (im *Image) nextBuildNumber() string {
	return strconv.Itoa(int(im.Status.BuildCounter + 1))
}

func (im *Image) generateBuildName() string {
	name := im.Name + "-build-" + im.nextBuildNumber() + "-"
	if len(name) > 64 {
		return name[:63]
	}

	return name
}
