package flags

import (
	"errors"
	"testing"

	"github.com/onsi/gomega"
	buildv1beta1 "github.com/shipwright-io/build/pkg/apis/build/v1beta1"
)

func TestNewParamArrayValue(t *testing.T) {
	g := gomega.NewWithT(t)

	testCases := map[string]struct {
		paramPassed string
		paramKey    string
		paramValue  string
		expectedErr error
	}{
		"simpleKeyValPair": {
			paramPassed: "dockerfile=Dockerfile",
			paramKey:    "dockerfile",
			paramValue:  "Dockerfile",
			expectedErr: nil,
		},
		"specialCharKeyValPair": {
			paramPassed: "b=cd=e",
			paramKey:    "b",
			paramValue:  "cd=e",
			expectedErr: nil,
		},
		"noEquals": {
			paramPassed: "bc",
			expectedErr: errors.New("informed value 'bc' is not in key=value format"),
		},
		"withSpaceVal": {
			paramPassed: "b=c d",
			paramKey:    "b",
			paramValue:  "c d",
			expectedErr: nil,
		},
	}
	for tName, tCase := range testCases {
		t.Run(tName, func(_ *testing.T) {
			buildSpec := buildv1beta1.BuildSpec{}
			buildParamVal := NewParamArrayValue(&buildSpec.ParamValues)

			err := buildParamVal.Set(tCase.paramPassed)
			if tCase.expectedErr != nil {
				g.Expect(err).To(gomega.Equal(tCase.expectedErr))
				return
			}
			g.Expect(err).To(gomega.BeNil())

			paramVal := tCase.paramValue
			g.Expect(buildSpec.ParamValues[0].Name).To(gomega.Equal(tCase.paramKey))
			g.Expect(buildSpec.ParamValues[0].Value).To(gomega.Equal(&paramVal))
		})
	}
}
