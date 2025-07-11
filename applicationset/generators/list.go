package generators

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	argoprojiov1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
)

var _ Generator = (*ListGenerator)(nil)

type ListGenerator struct{}

func NewListGenerator() Generator {
	g := &ListGenerator{}
	return g
}

func (g *ListGenerator) GetRequeueAfter(_ *argoprojiov1alpha1.ApplicationSetGenerator) time.Duration {
	return NoRequeueAfter
}

func (g *ListGenerator) GetTemplate(appSetGenerator *argoprojiov1alpha1.ApplicationSetGenerator) *argoprojiov1alpha1.ApplicationSetTemplate {
	return &appSetGenerator.List.Template
}

func (g *ListGenerator) GenerateParams(appSetGenerator *argoprojiov1alpha1.ApplicationSetGenerator, appSet *argoprojiov1alpha1.ApplicationSet, _ client.Client) ([]map[string]any, error) {
	if appSetGenerator == nil {
		return nil, ErrEmptyAppSetGenerator
	}

	if appSetGenerator.List == nil {
		return nil, ErrEmptyAppSetGenerator
	}

	res := make([]map[string]any, len(appSetGenerator.List.Elements))

	for i, tmpItem := range appSetGenerator.List.Elements {
		params := map[string]any{}
		var element map[string]any
		err := json.Unmarshal(tmpItem.Raw, &element)
		if err != nil {
			return nil, fmt.Errorf("error unmarshling list element %w", err)
		}

		if appSet.Spec.GoTemplate {
			res[i] = element
		} else {
			for key, value := range element {
				if key == "values" {
					values, ok := (value).(map[string]any)
					if !ok {
						return nil, errors.New("error parsing values map")
					}
					for k, v := range values {
						value, ok := v.(string)
						if !ok {
							return nil, fmt.Errorf("error parsing value as string %w", err)
						}
						params["values."+k] = value
					}
				} else {
					v, ok := value.(string)
					if !ok {
						return nil, fmt.Errorf("error parsing value as string %w", err)
					}
					params[key] = v
				}
				res[i] = params
			}
		}
	}

	// Append elements from ElementsYaml to the response
	if appSetGenerator.List.ElementsYaml != "" {
		var yamlElements []map[string]any
		err := yaml.Unmarshal([]byte(appSetGenerator.List.ElementsYaml), &yamlElements)
		if err != nil {
			return nil, fmt.Errorf("error unmarshling decoded ElementsYaml %w", err)
		}
		res = append(res, yamlElements...)
	}

	return res, nil
}
