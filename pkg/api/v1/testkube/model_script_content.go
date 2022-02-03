/*
 * TestKube API
 *
 * TestKube provides a Kubernetes-native framework for test definition, execution and results
 *
 * API version: 1.0.0
 * Contact: testkube@kubeshop.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package testkube

type ScriptContent struct {
	// script type
	Type_      string      `json:"type,omitempty"`
	Repository *Repository `json:"repository,omitempty"`
	// script content data as string
	Data string `json:"data,omitempty"`
	// script content
	Uri string `json:"uri,omitempty"`
}