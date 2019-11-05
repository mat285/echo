package kube

import (
	"encoding/json"

	exception "github.com/blend/go-sdk/exception"

	yaml "gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeyaml "k8s.io/client-go/kubernetes/scheme"
)

// YAMLEncode encodes an object as kube yaml.
func YAMLEncode(v interface{}) (string, error) {
	j, err := json.Marshal(v)
	if err != nil {
		return "", exception.New(err)
	}
	// Convert the JSON to an object.
	var jsonObj interface{}
	err = yaml.Unmarshal(j, &jsonObj)
	if err != nil {
		return "", exception.New(err)
	}

	// Marshal this object into YAML.
	final, err := yaml.Marshal(jsonObj)
	return string(final), exception.New(err)
}

// YAMLDecode decodes an object from yaml.
func YAMLDecode(yamlContents []byte) (runtime.Object, *schema.GroupVersionKind, error) {
	obj, kind, err := kubeyaml.Codecs.UniversalDeserializer().Decode(yamlContents, nil, nil)
	return obj, kind, exception.New(err)
}
