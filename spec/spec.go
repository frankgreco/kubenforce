package spec

import (
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/frankgreco/kubenforce/issue"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

type ConfigPolicyList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is a list of third party objects
	Items []*ConfigPolicy `json:"items"`
}

type ConfigPolicy struct {
	unversioned.TypeMeta `json:",inline"`
	api.ObjectMeta       `json:"metadata,omitempty"`
	Spec                 ConfigPolicySpec `json:"spec"`
}

type ConfigPolicySpec struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Rules      []Rule `json:"rules"`
}

type Rule struct {
	Remove bool          `json:"remove,omitempty"`
	Issue  IssueTemplate `json:"issue"`
	Policy Policy        `json:policy`
}

type IssueTemplate struct {
	Title string `json:"title"`
	Body  Body   `json:body`
}

type Policy struct {
	Template string `json:"template"`
	Regex    string `json:"regex"`
}

type Body struct {
	Issue      string `json:"issue"`
	Code       string `json:"code"`
	Resolution string `json:"resolution"`
}

func (cp *ConfigPolicy) RetroFit(c *client.Client) {
	// hard coded for services right now <- CHANGE LATER
	services, err := c.Services(cp.ObjectMeta.Namespace).List(api.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, svc := range services.Items {
		for _, rule := range cp.Spec.Rules {
			if svc.Spec.Type == "NodePort" {
				source := svc.ObjectMeta.Annotations["source"]
				logrus.Infof("Annotation is: %s", source)
				state := "open"
				body := constructIssueBody(&rule.Issue.Body)
				issue := issue.Issue{
					Owner: regexp.MustCompile(":\\/\\/|\\/").Split(source, 4)[2],
					Repo:  regexp.MustCompile(":\\/\\/|\\/").Split(source, 4)[3],
					Title: &rule.Issue.Title,
					Body:  &body,
					State: &state,
				}
				issue.Create()
				c.Services(svc.ObjectMeta.Namespace).Delete(svc.ObjectMeta.Name)
			}
		}

	}
}

func constructIssueBody(body *Body) string {
	header := "*NOTE: this issue was automatically generated*"
	issue := "**Issue:**\n" + body.Issue
	code := "**Offending Code:**\n```yaml\n" + body.Code + "\n```"
	resolution := "**How to Fix:**\n" + body.Resolution
	return header + "\n\n" + issue + "\n\n" + code + "\n" + resolution
}
