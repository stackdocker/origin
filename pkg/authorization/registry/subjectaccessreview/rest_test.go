package subjectaccessreview

import (
	"errors"
	"reflect"
	"testing"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/diff"
	"k8s.io/kubernetes/pkg/util/sets"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	"github.com/openshift/origin/pkg/authorization/authorizer"
)

type subjectAccessTest struct {
	authorizer    *testAuthorizer
	reviewRequest *authorizationapi.SubjectAccessReview
}

type testAuthorizer struct {
	allowed          bool
	reason           string
	err              string
	deniedNamespaces sets.String

	actualAttributes authorizer.DefaultAuthorizationAttributes
}

func (a *testAuthorizer) Authorize(ctx kapi.Context, passedAttributes authorizer.AuthorizationAttributes) (allowed bool, reason string, err error) {
	// allow the initial check for "can I run this SAR at all"
	if passedAttributes.GetResource() == "localsubjectaccessreviews" {
		if len(a.deniedNamespaces) != 0 && a.deniedNamespaces.Has(kapi.NamespaceValue(ctx)) {
			return false, "denied initial check", nil
		}

		return true, "", nil
	}

	attributes, ok := passedAttributes.(authorizer.DefaultAuthorizationAttributes)
	if !ok {
		return false, "ERROR", errors.New("unexpected type for test")
	}

	a.actualAttributes = attributes

	if len(a.err) == 0 {
		return a.allowed, a.reason, nil
	}
	return a.allowed, a.reason, errors.New(a.err)
}
func (a *testAuthorizer) GetAllowedSubjects(ctx kapi.Context, passedAttributes authorizer.AuthorizationAttributes) (sets.String, sets.String, error) {
	return sets.String{}, sets.String{}, nil
}

func TestDeniedNamespace(t *testing.T) {
	test := &subjectAccessTest{
		authorizer: &testAuthorizer{
			allowed:          false,
			err:              "denied initial check",
			deniedNamespaces: sets.NewString("foo"),
		},
		reviewRequest: &authorizationapi.SubjectAccessReview{
			Action: authorizationapi.AuthorizationAttributes{
				Namespace: "foo",
				Verb:      "get",
				Resource:  "pods",
			},
			User:   "foo",
			Groups: sets.NewString(),
		},
	}

	test.runTest(t)
}

func TestEmptyReturn(t *testing.T) {
	test := &subjectAccessTest{
		authorizer: &testAuthorizer{
			allowed: false,
			reason:  "because reasons",
		},
		reviewRequest: &authorizationapi.SubjectAccessReview{
			Action: authorizationapi.AuthorizationAttributes{
				Verb:     "get",
				Resource: "pods",
			},
			User:   "foo",
			Groups: sets.NewString(),
		},
	}

	test.runTest(t)
}

func TestNoErrors(t *testing.T) {
	test := &subjectAccessTest{
		authorizer: &testAuthorizer{
			allowed: true,
			reason:  "because good things",
		},
		reviewRequest: &authorizationapi.SubjectAccessReview{
			Action: authorizationapi.AuthorizationAttributes{
				Verb:     "delete",
				Resource: "deploymentConfigs",
			},
			Groups: sets.NewString("not-master"),
		},
	}

	test.runTest(t)
}

func TestErrors(t *testing.T) {
	test := &subjectAccessTest{
		authorizer: &testAuthorizer{
			err: "some-random-failure",
		},
		reviewRequest: &authorizationapi.SubjectAccessReview{
			Action: authorizationapi.AuthorizationAttributes{
				Verb:     "get",
				Resource: "pods",
			},
			User:   "foo",
			Groups: sets.NewString("first", "second"),
		},
	}

	test.runTest(t)
}

func (r *subjectAccessTest) runTest(t *testing.T) {
	storage := REST{r.authorizer}

	expectedResponse := &authorizationapi.SubjectAccessReviewResponse{
		Namespace: r.reviewRequest.Action.Namespace,
		Allowed:   r.authorizer.allowed,
		Reason:    r.authorizer.reason,
	}

	expectedAttributes := authorizer.ToDefaultAuthorizationAttributes(r.reviewRequest.Action)

	ctx := kapi.WithNamespace(kapi.NewContext(), kapi.NamespaceAll)
	obj, err := storage.Create(ctx, r.reviewRequest)
	if err != nil && len(r.authorizer.err) == 0 {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.authorizer.err) != 0 {
		if err == nil {
			t.Fatalf("unexpected non-error: %v", err)
		}
		if e, a := r.authorizer.err, err.Error(); e != a {
			t.Fatalf("expected %v, got %v", e, a)
		}

		return
	}

	switch obj.(type) {
	case *authorizationapi.SubjectAccessReviewResponse:
		if !reflect.DeepEqual(expectedResponse, obj) {
			t.Errorf("diff %v", diff.ObjectGoPrintDiff(expectedResponse, obj))
		}
	case nil:
		if len(r.authorizer.err) == 0 {
			t.Fatal("unexpected nil object")
		}

	default:
		t.Errorf("Unexpected obj type: %v", obj)
	}

	if !reflect.DeepEqual(expectedAttributes, r.authorizer.actualAttributes) {
		t.Errorf("diff %v", diff.ObjectGoPrintDiff(expectedAttributes, r.authorizer.actualAttributes))
	}
}
