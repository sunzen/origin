package install

import (
	"reflect"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestDescriptions(t *testing.T) {
	scheme := runtime.NewScheme()
	InstallInternalOpenShift(scheme)
	InstallInternalKube(scheme)

	for _, version := range scheme.PrioritizedVersionsAllGroups() {
		seen := map[reflect.Type]bool{}

		for _, apiType := range scheme.KnownTypes(version) {
			checkDescriptions(apiType, &seen, t)
		}
	}
}

func checkDescriptions(objType reflect.Type, seen *map[reflect.Type]bool, t *testing.T) {
	if _, exists := (*seen)[objType]; exists {
		return
	}
	(*seen)[objType] = true
	if !strings.Contains(objType.PkgPath(), "github.com/openshift/origin/pkg") {
		return
	}

	for i := 0; i < objType.NumField(); i++ {
		structField := objType.FieldByIndex([]int{i})

		// these fields don't need descriptions
		if structField.Name == "TypeMeta" || structField.Name == "ObjectMeta" || structField.Name == "ListMeta" {
			continue
		}
		if structField.Type == reflect.TypeOf(metav1.Time{}) || structField.Type == reflect.TypeOf(time.Time{}) || structField.Type == reflect.TypeOf(runtime.RawExtension{}) {
			continue
		}

		descriptionTag := structField.Tag.Get("description")
		if len(descriptionTag) > 0 {
			t.Errorf("%v", structField.Tag)
			t.Errorf("%v.%v should not have a description tag", objType, structField.Name)
		}

		switch structField.Type.Kind() {
		case reflect.Struct:
			checkDescriptions(structField.Type, seen, t)
		}
	}
}

func TestInternalJsonTags(t *testing.T) {
	scheme := runtime.NewScheme()
	InstallInternalOpenShift(scheme)
	InstallInternalKube(scheme)

	seen := map[reflect.Type]bool{}
	seenGroups := sets.String{}

	for _, version := range scheme.PrioritizedVersionsAllGroups() {
		if seenGroups.Has(version.Group) {
			continue
		}
		seenGroups.Insert(version.Group)

		internalVersion := schema.GroupVersion{Group: version.Group, Version: runtime.APIVersionInternal}
		for _, apiType := range scheme.KnownTypes(internalVersion) {
			checkInternalJsonTags(apiType, &seen, t)
		}
	}
}

// internalTypesWithAllowedJsonTags is the list of special structs that have a particular need to have json tags on their
// internal types.  Do not add to this list without having you paperwork checked in triplicate.
var internalTypesWithAllowedJsonTags = sets.NewString("DockerConfig", "DockerImage")

func checkInternalJsonTags(objType reflect.Type, seen *map[reflect.Type]bool, t *testing.T) {
	if objType.Kind() != reflect.Struct {
		return
	}
	if _, exists := (*seen)[objType]; exists {
		return
	}
	(*seen)[objType] = true
	if !strings.Contains(objType.PkgPath(), "github.com/openshift/origin/pkg") {
		return
	}
	if internalTypesWithAllowedJsonTags.Has(objType.Name()) {
		return
	}
	if objType.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < objType.NumField(); i++ {
		structField := objType.FieldByIndex([]int{i})

		jsonTag := structField.Tag.Get("json")
		if len(jsonTag) != 0 {
			t.Errorf("%v.%v should not have a json tag", objType, structField.Name)
		}

		switch structField.Type.Kind() {
		case reflect.Struct:
			checkInternalJsonTags(structField.Type, seen, t)
		case reflect.Ptr:
			checkInternalJsonTags(structField.Type.Elem(), seen, t)
		}
	}
}

func TestExternalJsonTags(t *testing.T) {
	scheme := runtime.NewScheme()
	InstallInternalOpenShift(scheme)
	InstallInternalKube(scheme)

	seen := map[reflect.Type]bool{}

	for _, version := range scheme.PrioritizedVersionsAllGroups() {
		for _, apiType := range scheme.KnownTypes(version) {
			checkExternalJsonTags(apiType, &seen, t)
		}
	}
}

func checkExternalJsonTags(objType reflect.Type, seen *map[reflect.Type]bool, t *testing.T) {
	if objType.Kind() != reflect.Struct {
		return
	}
	if _, exists := (*seen)[objType]; exists {
		return
	}
	(*seen)[objType] = true
	if !strings.Contains(objType.PkgPath(), "github.com/openshift/origin/pkg") {
		return
	}

	if objType.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < objType.NumField(); i++ {
		structField := objType.FieldByIndex([]int{i})

		jsonTag := structField.Tag.Get("json")
		if len(jsonTag) == 0 {
			t.Errorf("%v.%v should have a json tag", objType, structField.Name)
		}

		switch structField.Type.Kind() {
		case reflect.Struct:
			checkExternalJsonTags(structField.Type, seen, t)
		case reflect.Ptr:
			checkExternalJsonTags(structField.Type.Elem(), seen, t)
		}
	}
}
