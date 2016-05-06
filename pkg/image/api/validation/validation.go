package validation

import (
	"fmt"
	"regexp"

	"github.com/docker/distribution/reference"
	"github.com/golang/glog"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/util/diff"
	"k8s.io/kubernetes/pkg/util/validation/field"

	oapi "github.com/openshift/origin/pkg/api"
	"github.com/openshift/origin/pkg/image/api"
)

// RepositoryNameComponentRegexp restricts registry path component names to
// start with at least one letter or number, with following parts able to
// be separated by one period, dash or underscore.
// Copied from github.com/docker/distribution/registry/api/v2/names.go v2.1.1
var RepositoryNameComponentRegexp = regexp.MustCompile(`[a-z0-9]+(?:[._-][a-z0-9]+)*`)

// RepositoryNameComponentAnchoredRegexp is the version of
// RepositoryNameComponentRegexp which must completely match the content
// Copied from github.com/docker/distribution/registry/api/v2/names.go v2.1.1
var RepositoryNameComponentAnchoredRegexp = regexp.MustCompile(`^` + RepositoryNameComponentRegexp.String() + `$`)

// RepositoryNameRegexp builds on RepositoryNameComponentRegexp to allow
// multiple path components, separated by a forward slash.
// Copied from github.com/docker/distribution/registry/api/v2/names.go v2.1.1
var RepositoryNameRegexp = regexp.MustCompile(`(?:` + RepositoryNameComponentRegexp.String() + `/)*` + RepositoryNameComponentRegexp.String())

func ValidateImageStreamName(name string, prefix bool) (bool, string) {
	if ok, reason := oapi.MinimalNameRequirements(name, prefix); !ok {
		return ok, reason
	}

	if !RepositoryNameComponentAnchoredRegexp.MatchString(name) {
		return false, fmt.Sprintf("must match %q", RepositoryNameComponentRegexp.String())
	}
	return true, ""
}

// ValidateImage tests required fields for an Image.
func ValidateImage(image *api.Image) field.ErrorList {
	return validateImage(image, nil)
}

func validateImage(image *api.Image, fldPath *field.Path) field.ErrorList {
	result := validation.ValidateObjectMeta(&image.ObjectMeta, false, oapi.MinimalNameRequirements, fldPath.Child("metadata"))

	if len(image.DockerImageReference) == 0 {
		result = append(result, field.Required(fldPath.Child("dockerImageReference"), ""))
	} else {
		if _, err := api.ParseDockerImageReference(image.DockerImageReference); err != nil {
			result = append(result, field.Invalid(fldPath.Child("dockerImageReference"), image.DockerImageReference, err.Error()))
		}
	}

	return result
}

func ValidateImageUpdate(newImage, oldImage *api.Image) field.ErrorList {
	result := validation.ValidateObjectMetaUpdate(&newImage.ObjectMeta, &oldImage.ObjectMeta, field.NewPath("metadata"))
	result = append(result, ValidateImage(newImage)...)

	return result
}

// ValidateImageStream tests required fields for an ImageStream.
func ValidateImageStream(stream *api.ImageStream) field.ErrorList {
	result := validation.ValidateObjectMeta(&stream.ObjectMeta, true, ValidateImageStreamName, field.NewPath("metadata"))

	// Ensure we can generate a valid docker image repository from namespace/name
	if len(stream.Namespace+"/"+stream.Name) > reference.NameTotalLengthMax {
		result = append(result, field.Invalid(field.NewPath("metadata", "name"), stream.Name, fmt.Sprintf("'namespace/name' cannot be longer than %d characters", reference.NameTotalLengthMax)))
	}

	if len(stream.Spec.DockerImageRepository) != 0 {
		dockerImageRepositoryPath := field.NewPath("spec", "dockerImageRepository")
		if ref, err := api.ParseDockerImageReference(stream.Spec.DockerImageRepository); err != nil {
			result = append(result, field.Invalid(dockerImageRepositoryPath, stream.Spec.DockerImageRepository, err.Error()))
		} else {
			if len(ref.Tag) > 0 {
				result = append(result, field.Invalid(dockerImageRepositoryPath, stream.Spec.DockerImageRepository, "the repository name may not contain a tag"))
			}
			if len(ref.ID) > 0 {
				result = append(result, field.Invalid(dockerImageRepositoryPath, stream.Spec.DockerImageRepository, "the repository name may not contain an ID"))
			}
		}
	}
	for tag, tagRef := range stream.Spec.Tags {
		path := field.NewPath("spec", "tags").Key(tag)
		result = append(result, ValidateImageStreamTagReference(tagRef, path)...)
	}
	for tag, history := range stream.Status.Tags {
		for i, tagEvent := range history.Items {
			if len(tagEvent.DockerImageReference) == 0 {
				result = append(result, field.Required(field.NewPath("status", "tags").Key(tag).Child("items").Index(i).Child("dockerImageReference"), ""))
			}
		}
	}

	return result
}

// ValidateImageStreamTagReference ensures that a given tag reference is valid.
func ValidateImageStreamTagReference(tagRef api.TagReference, fldPath *field.Path) field.ErrorList {
	var errs field.ErrorList
	if tagRef.From != nil {
		if len(tagRef.From.Name) == 0 {
			errs = append(errs, field.Required(fldPath.Child("from", "name"), "name is required"))
		}
		switch tagRef.From.Kind {
		case "DockerImage":
			if ref, err := api.ParseDockerImageReference(tagRef.From.Name); err == nil && tagRef.ImportPolicy.Scheduled && len(ref.ID) > 0 {
				errs = append(errs, field.Invalid(fldPath.Child("from", "name"), tagRef.From.Name, "only tags can be scheduled for import"))
			}
		case "ImageStreamImage", "ImageStreamTag":
			if tagRef.ImportPolicy.Scheduled {
				errs = append(errs, field.Invalid(fldPath.Child("importPolicy", "scheduled"), tagRef.ImportPolicy.Scheduled, "only tags pointing to Docker repositories may be scheduled for background import"))
			}
		default:
			errs = append(errs, field.Required(fldPath.Child("from", "kind"), "valid values are 'DockerImage', 'ImageStreamImage', 'ImageStreamTag'"))
		}
	}
	return errs
}

func ValidateImageStreamUpdate(newStream, oldStream *api.ImageStream) field.ErrorList {
	result := validation.ValidateObjectMetaUpdate(&newStream.ObjectMeta, &oldStream.ObjectMeta, field.NewPath("metadata"))
	result = append(result, ValidateImageStream(newStream)...)

	return result
}

// ValidateImageStreamStatusUpdate tests required fields for an ImageStream status update.
func ValidateImageStreamStatusUpdate(newStream, oldStream *api.ImageStream) field.ErrorList {
	result := validation.ValidateObjectMetaUpdate(&newStream.ObjectMeta, &oldStream.ObjectMeta, field.NewPath("metadata"))
	return result
}

// ValidateImageStreamMapping tests required fields for an ImageStreamMapping.
func ValidateImageStreamMapping(mapping *api.ImageStreamMapping) field.ErrorList {
	result := validation.ValidateObjectMeta(&mapping.ObjectMeta, true, oapi.MinimalNameRequirements, field.NewPath("metadata"))

	hasRepository := len(mapping.DockerImageRepository) != 0
	hasName := len(mapping.Name) != 0
	switch {
	case hasRepository:
		if _, err := api.ParseDockerImageReference(mapping.DockerImageRepository); err != nil {
			result = append(result, field.Invalid(field.NewPath("dockerImageRepository"), mapping.DockerImageRepository, err.Error()))
		}
	case hasName:
	default:
		result = append(result, field.Required(field.NewPath("name"), ""))
		result = append(result, field.Required(field.NewPath("dockerImageRepository"), ""))
	}

	if ok, msg := validation.ValidateNamespaceName(mapping.Namespace, false); !ok {
		result = append(result, field.Invalid(field.NewPath("metadata", "namespace"), mapping.Namespace, msg))
	}
	if len(mapping.Tag) == 0 {
		result = append(result, field.Required(field.NewPath("tag"), ""))
	}
	if errs := validateImage(&mapping.Image, field.NewPath("image")); len(errs) != 0 {
		result = append(result, errs...)
	}
	return result
}

// ValidateImageStreamTag validates a mutation of an image stream tag, which can happen on PUT
func ValidateImageStreamTag(ist *api.ImageStreamTag) field.ErrorList {
	result := validation.ValidateObjectMeta(&ist.ObjectMeta, true, oapi.MinimalNameRequirements, field.NewPath("metadata"))
	if ist.Tag != nil {
		result = append(result, ValidateImageStreamTagReference(*ist.Tag, field.NewPath("tag"))...)
		if ist.Tag.Annotations != nil && !kapi.Semantic.DeepEqual(ist.Tag.Annotations, ist.ObjectMeta.Annotations) {
			result = append(result, field.Invalid(field.NewPath("tag", "annotations"), "<map>", "tag annotations must not be provided or must be equal to the object meta annotations"))
		}
	}

	return result
}

// ValidateImageStreamTagUpdate ensures that only the annotations of the IST have changed
func ValidateImageStreamTagUpdate(newIST, oldIST *api.ImageStreamTag) field.ErrorList {
	result := validation.ValidateObjectMetaUpdate(&newIST.ObjectMeta, &oldIST.ObjectMeta, field.NewPath("metadata"))

	if newIST.Tag != nil {
		result = append(result, ValidateImageStreamTagReference(*newIST.Tag, field.NewPath("tag"))...)
		if newIST.Tag.Annotations != nil && !kapi.Semantic.DeepEqual(newIST.Tag.Annotations, newIST.ObjectMeta.Annotations) {
			result = append(result, field.Invalid(field.NewPath("tag", "annotations"), "<map>", "tag annotations must not be provided or must be equal to the object meta annotations"))
		}
	}

	// ensure that only tag and annotations have changed
	newISTCopy := *newIST
	oldISTCopy := *oldIST
	newISTCopy.Annotations, oldISTCopy.Annotations = nil, nil
	newISTCopy.Tag, oldISTCopy.Tag = nil, nil
	newISTCopy.Generation = oldISTCopy.Generation
	if !kapi.Semantic.Equalities.DeepEqual(&newISTCopy, &oldISTCopy) {
		glog.Infof("objects differ: ", diff.ObjectDiff(oldISTCopy, newISTCopy))
		result = append(result, field.Invalid(field.NewPath("metadata"), "", "may not update fields other than metadata.annotations"))
	}

	return result
}

func ValidateImageStreamImport(isi *api.ImageStreamImport) field.ErrorList {
	specPath := field.NewPath("spec")
	imagesPath := specPath.Child("images")
	repoPath := specPath.Child("repository")

	errs := field.ErrorList{}
	for i, spec := range isi.Spec.Images {
		from := spec.From
		switch from.Kind {
		case "DockerImage":
			if spec.To != nil && len(spec.To.Name) == 0 {
				errs = append(errs, field.Invalid(imagesPath.Index(i).Child("to", "name"), spec.To.Name, "the name of the target tag must be specified"))
			}
			if len(spec.From.Name) == 0 {
				errs = append(errs, field.Required(imagesPath.Index(i).Child("from", "name"), ""))
			} else {
				if ref, err := api.ParseDockerImageReference(spec.From.Name); err != nil {
					errs = append(errs, field.Invalid(imagesPath.Index(i).Child("from", "name"), spec.From.Name, err.Error()))
				} else {
					if len(ref.ID) > 0 && spec.ImportPolicy.Scheduled {
						errs = append(errs, field.Invalid(imagesPath.Index(i).Child("from", "name"), spec.From.Name, "only tags can be scheduled for import"))
					}
				}
			}
		default:
			errs = append(errs, field.Invalid(imagesPath.Index(i).Child("from", "kind"), from.Kind, "only DockerImage is supported"))
		}
	}

	if spec := isi.Spec.Repository; spec != nil {
		from := spec.From
		switch from.Kind {
		case "DockerImage":
			if len(spec.From.Name) == 0 {
				errs = append(errs, field.Required(repoPath.Child("from", "name"), ""))
			} else {
				if ref, err := api.ParseDockerImageReference(from.Name); err != nil {
					errs = append(errs, field.Invalid(repoPath.Child("from", "name"), from.Name, err.Error()))
				} else {
					if len(ref.ID) > 0 || len(ref.Tag) > 0 {
						errs = append(errs, field.Invalid(repoPath.Child("from", "name"), from.Name, "you must specify an image repository, not a tag or ID"))
					}
				}
			}
		default:
			errs = append(errs, field.Invalid(repoPath.Child("from", "kind"), from.Kind, "only DockerImage is supported"))
		}
	}
	if len(isi.Spec.Images) == 0 && isi.Spec.Repository == nil {
		errs = append(errs, field.Invalid(imagesPath, nil, "you must specify at least one image or a repository import"))
	}

	errs = append(errs, validation.ValidateObjectMeta(&isi.ObjectMeta, true, ValidateImageStreamName, field.NewPath("metadata"))...)
	return errs
}
