package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	netestv1alpha1 "github.com/terloo/kubenetest-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func parseTemplate(path string, paramObj netestv1alpha1.Netest) ([]byte, error) {
	fpath := filepath.Join("template", path)

	_, err := os.Stat(fpath)
	if err != nil {
		return nil, err
	}

	parser, err := template.ParseFiles(fpath)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = parser.Execute(buf, paramObj)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func RenderDaemonSet(netest netestv1alpha1.Netest, sheme *runtime.Scheme) (*appsv1.DaemonSet, error) {
	b, err := parseTemplate("ds.yaml", netest)
	if err != nil {
		return nil, err
	}

	ds := &appsv1.DaemonSet{}
	err = yaml.Unmarshal(b, ds)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
