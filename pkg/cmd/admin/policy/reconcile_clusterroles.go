package policy

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	kapi "k8s.io/kubernetes/pkg/api"
	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/sets"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
	"github.com/openshift/origin/pkg/authorization/rulevalidation"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/server/bootstrappolicy"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	osutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

// ReconcileClusterRolesRecommendedName is the recommended command name
const ReconcileClusterRolesRecommendedName = "reconcile-cluster-roles"

type ReconcileClusterRolesOptions struct {
	// RolesToReconcile says which roles should be reconciled.  An empty or nil slice means
	// reconcile all of them.
	RolesToReconcile []string

	Confirmed bool
	Union     bool

	Out    io.Writer
	Output string

	RoleClient client.ClusterRoleInterface
}

const (
	reconcileLong = `
Replace cluster roles to match the recommended bootstrap policy

This command will inspect the cluster roles against the recommended bootstrap policy.  Any cluster role
that does not match will be replaced by the recommended bootstrap role.  This command will not remove
any additional cluster role.

You can see which cluster role have recommended changed by choosing an output type.`

	reconcileExample = `  # Display the cluster roles that would be modified
  $ %[1]s

  # Replace cluster roles that don't match the current defaults
  $ %[1]s --confirm

  # Display the union of the default and modified cluster roles
  $ %[1]s --additive-only`
)

// NewCmdReconcileClusterRoles implements the OpenShift cli reconcile-cluster-roles command
func NewCmdReconcileClusterRoles(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	o := &ReconcileClusterRolesOptions{Out: out}

	cmd := &cobra.Command{
		Use:     name + " [ClusterRoleName]...",
		Short:   "Replace cluster roles to match the recommended bootstrap policy",
		Long:    reconcileLong,
		Example: fmt.Sprintf(reconcileExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd, f, args); err != nil {
				kcmdutil.CheckErr(err)
			}

			if err := o.Validate(); err != nil {
				kcmdutil.CheckErr(kcmdutil.UsageError(cmd, err.Error()))
			}

			if err := o.RunReconcileClusterRoles(cmd, f); err != nil {
				kcmdutil.CheckErr(err)
			}
		},
	}

	cmd.Flags().BoolVar(&o.Confirmed, "confirm", o.Confirmed, "Specify that cluster roles should be modified. Defaults to false, displaying what would be replaced but not actually replacing anything.")
	cmd.Flags().BoolVar(&o.Union, "additive-only", o.Union, "Preserves modified cluster roles.")
	kcmdutil.AddPrinterFlags(cmd)
	cmd.Flags().Lookup("output").DefValue = "yaml"
	cmd.Flags().Lookup("output").Value.Set("yaml")

	return cmd
}

func (o *ReconcileClusterRolesOptions) Complete(cmd *cobra.Command, f *clientcmd.Factory, args []string) error {
	oclient, _, err := f.Clients()
	if err != nil {
		return err
	}
	o.RoleClient = oclient.ClusterRoles()

	o.Output = kcmdutil.GetFlagString(cmd, "output")

	mapper, _ := f.Object(false)
	for _, resourceString := range args {
		resource, name, err := osutil.ResolveResource(authorizationapi.Resource("clusterroles"), resourceString, mapper)
		if err != nil {
			return err
		}
		if resource != authorizationapi.Resource("clusterroles") {
			return fmt.Errorf("%v is not a valid resource type for this command", resource)
		}
		if len(name) == 0 {
			return fmt.Errorf("%s did not contain a name", resourceString)
		}

		o.RolesToReconcile = append(o.RolesToReconcile, name)
	}

	return nil
}

func (o *ReconcileClusterRolesOptions) Validate() error {
	if o.RoleClient == nil {
		return errors.New("a role client is required")
	}
	if o.Output != "yaml" && o.Output != "json" && o.Output != "" {
		return fmt.Errorf("unknown output specified: %s", o.Output)
	}
	return nil
}

// RunReconcileClusterRoles contains all the necessary functionality for the OpenShift cli reconcile-cluster-roles command
func (o *ReconcileClusterRolesOptions) RunReconcileClusterRoles(cmd *cobra.Command, f *clientcmd.Factory) error {
	changedClusterRoles, err := o.ChangedClusterRoles()
	if err != nil {
		return err
	}

	if len(changedClusterRoles) == 0 {
		return nil
	}

	if (len(o.Output) != 0) && !o.Confirmed {
		list := &kapi.List{}
		for _, item := range changedClusterRoles {
			list.Items = append(list.Items, item)
		}
		mapper, _ := f.Object(false)
		fn := cmdutil.VersionedPrintObject(f.PrintObject, cmd, mapper, o.Out)
		if err := fn(list); err != nil {
			return err
		}
	}

	if o.Confirmed {
		return o.ReplaceChangedRoles(changedClusterRoles)
	}

	return nil
}

// ChangedClusterRoles returns the roles that must be created and/or updated to
// match the recommended bootstrap policy
func (o *ReconcileClusterRolesOptions) ChangedClusterRoles() ([]*authorizationapi.ClusterRole, error) {
	changedRoles := []*authorizationapi.ClusterRole{}

	rolesToReconcile := sets.NewString(o.RolesToReconcile...)
	rolesNotFound := sets.NewString(o.RolesToReconcile...)
	bootstrapClusterRoles := bootstrappolicy.GetBootstrapClusterRoles()
	for i := range bootstrapClusterRoles {
		expectedClusterRole := &bootstrapClusterRoles[i]
		if (len(rolesToReconcile) > 0) && !rolesToReconcile.Has(expectedClusterRole.Name) {
			continue
		}
		rolesNotFound.Delete(expectedClusterRole.Name)

		actualClusterRole, err := o.RoleClient.Get(expectedClusterRole.Name)
		if kapierrors.IsNotFound(err) {
			changedRoles = append(changedRoles, expectedClusterRole)
			continue
		}
		if err != nil {
			return nil, err
		}

		// Copy any existing labels/annotations, so the displayed update is correct
		// This assumes bootstrap roles will not set any labels/annotations
		// These aren't actually used during update; the latest labels/annotations are pulled from the existing object again
		expectedClusterRole.Labels = actualClusterRole.Labels
		expectedClusterRole.Annotations = actualClusterRole.Annotations

		if !kapi.Semantic.DeepEqual(expectedClusterRole.Rules, actualClusterRole.Rules) {
			if o.Union {
				_, missingRules := rulevalidation.Covers(expectedClusterRole.Rules, actualClusterRole.Rules)
				expectedClusterRole.Rules = append(expectedClusterRole.Rules, missingRules...)
			}
			changedRoles = append(changedRoles, expectedClusterRole)
		}
	}

	if len(rolesNotFound) != 0 {
		// return the known changes and the error so that a caller can decide if he wants a partial update
		return changedRoles, fmt.Errorf("did not find requested cluster role %s", rolesNotFound.List())
	}

	return changedRoles, nil
}

// ReplaceChangedRoles will reconcile all the changed roles back to the recommended bootstrap policy
func (o *ReconcileClusterRolesOptions) ReplaceChangedRoles(changedRoles []*authorizationapi.ClusterRole) error {
	for i := range changedRoles {
		role, err := o.RoleClient.Get(changedRoles[i].Name)
		if err != nil && !kapierrors.IsNotFound(err) {
			return err
		}

		if kapierrors.IsNotFound(err) {
			createdRole, err := o.RoleClient.Create(changedRoles[i])
			if err != nil {
				return err
			}

			fmt.Fprintf(o.Out, "clusterrole/%s\n", createdRole.Name)
			continue
		}

		role.Rules = changedRoles[i].Rules
		updatedRole, err := o.RoleClient.Update(role)
		if err != nil {
			return err
		}

		fmt.Fprintf(o.Out, "clusterrole/%s\n", updatedRole.Name)
	}

	return nil
}
