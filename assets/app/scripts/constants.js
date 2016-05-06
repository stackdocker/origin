// Assigns global constants to things like external documentation, links to external resources, annotations and naming, etc.
// Can be customized using custom scripts in the master config file that override one or multiple of these objects.
// Reference: https://docs.openshift.org/latest/install_config/web_console_customization.html#loading-custom-scripts-and-stylesheets

// NOTE: Update extensions/examples/online-extensions.js if you add a new help link to this map.

window.OPENSHIFT_CONSTANTS = {
  // Maps links to specific topics in external documentation.
  HELP: {
    "cli":                     "https://docs.openshift.org/latest/cli_reference/overview.html",
    "get_started_cli":         "https://docs.openshift.org/latest/cli_reference/get_started_cli.html",
    "basic_cli_operations":    "https://docs.openshift.org/latest/cli_reference/basic_cli_operations.html",
    "webhooks":                "https://docs.openshift.org/latest/dev_guide/builds.html#webhook-triggers",
    "new_app":                 "https://docs.openshift.org/latest/dev_guide/new_app.html",
    "start-build":             "https://docs.openshift.org/latest/dev_guide/builds.html#starting-a-build",
    "deployment-operations":   "https://docs.openshift.org/latest/cli_reference/basic_cli_operations.html#build-and-deployment-cli-operations",
    "route-types":             "https://docs.openshift.org/latest/architecture/core_concepts/routes.html#route-types",
    "persistent_volumes":      "https://docs.openshift.org/latest/dev_guide/persistent_volumes.html",
    "compute_resources":       "https://docs.openshift.org/latest/dev_guide/compute_resources.html",
    "pod_autoscaling":         "https://docs.openshift.org/latest/dev_guide/pod_autoscaling.html",
    "application_health":      "https://docs.openshift.org/latest/dev_guide/application_health.html",
    "default":                 "https://docs.openshift.org/latest/welcome/index.html"
  },
  // Maps links names to URL's where the CLI tools can be downloaded, may point directly to files or to external pages in a CDN, for example.
  CLI: {
    "Latest Release":          "https://github.com/openshift/origin/releases/latest"
  },
  // The default CPU target percentage for horizontal pod autoscalers created or edited in the web console.
  // This value is set in the HPA when the input is left blank.
  DEFAULT_HPA_CPU_TARGET_PERCENT: 80,
  
  // This blacklist hides certain kinds from the "Other Resources" page because they are unpersisted, disallowed for most end users, or not supported by openshift but exist in kubernetes
  AVAILABLE_KINDS_BLACKLIST: [
      // These are k8s kinds that are not supported in the current release of OpenShift
      "PodTemplate",
      "Binding",
      "Ingress",
      "ReplicaSet",
   
      // These are things like SARs that aren't actually persisted resources
      "LocalResourceAccessReview",
      "LocalSubjectAccessReview",
      "ResourceAccessReview",
      "SubjectAccessReview",
      "ReplicationControllerDummy",
      "DeploymentConfigRollback"    
  ]
};
