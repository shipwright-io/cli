package flags

import (
	"fmt"
	"strings"

	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// StepResourcesArrayValue implements the pflag.Value interface, storing per-step resource
// requirement overrides used on Shipwright's BuildRunSpec.StepResources.
type StepResourcesArrayValue struct {
	stepResources *[]buildv1beta1.StepResourceOverride // pointer to the slice of overrides
}

// String prints out the string representation of the slice of overrides.
func (s *StepResourcesArrayValue) String() string {
	slice := []string{}
	for _, o := range *s.stepResources {
		slice = append(slice, o.Name)
	}
	csv, _ := writeAsCSV(slice)
	return fmt.Sprintf("[%s]", csv)
}

// Set receives an entry in the format "<step>=<limits|requests>.<resource>=<quantity>", for example
// "build-and-push=limits.memory=2Gi". Repeated entries for the same step are merged.
func (s *StepResourcesArrayValue) Set(value string) error {
	step, spec, err := splitKeyValue(value)
	if err != nil {
		return fmt.Errorf("step resources value %q is not in <step>=<limits|requests>.<resource>=<quantity> format: %w", value, err)
	}

	resKey, qtyStr, err := splitKeyValue(spec)
	if err != nil {
		return fmt.Errorf("step resources value %q is not in <step>=<limits|requests>.<resource>=<quantity> format: %w", value, err)
	}

	kind, resName, found := strings.Cut(resKey, ".")
	if !found || resName == "" {
		return fmt.Errorf("step resources value %q must specify the resource as <limits|requests>.<resource>", value)
	}
	if kind != "limits" && kind != "requests" {
		return fmt.Errorf("step resources value %q must use %q or %q, got %q", value, "limits", "requests", kind)
	}

	qty, err := resource.ParseQuantity(qtyStr)
	if err != nil {
		return fmt.Errorf("step resources value %q has an invalid quantity %q: %w", value, qtyStr, err)
	}

	override := s.getOrCreate(step)
	if kind == "limits" {
		if override.Resources.Limits == nil {
			override.Resources.Limits = corev1.ResourceList{}
		}
		override.Resources.Limits[corev1.ResourceName(resName)] = qty
	} else {
		if override.Resources.Requests == nil {
			override.Resources.Requests = corev1.ResourceList{}
		}
		override.Resources.Requests[corev1.ResourceName(resName)] = qty
	}
	return nil
}

// getOrCreate returns a pointer to the override for the given step name, appending a new entry when
// one does not exist yet.
func (s *StepResourcesArrayValue) getOrCreate(step string) *buildv1beta1.StepResourceOverride {
	for i := range *s.stepResources {
		if (*s.stepResources)[i].Name == step {
			return &(*s.stepResources)[i]
		}
	}
	*s.stepResources = append(*s.stepResources, buildv1beta1.StepResourceOverride{Name: step})
	return &(*s.stepResources)[len(*s.stepResources)-1]
}

// Type analogous to the pflag "stringArray" type, where each flag entry is translated to a single
// slice entry.
func (s *StepResourcesArrayValue) Type() string {
	return "stringArray"
}

// NewStepResourcesArrayValue instantiates a StepResourcesArrayValue sharing the slice pointer.
func NewStepResourcesArrayValue(stepResources *[]buildv1beta1.StepResourceOverride) *StepResourcesArrayValue {
	return &StepResourcesArrayValue{stepResources: stepResources}
}
