package configchange

import (
	"fmt"

	"github.com/golang/glog"

	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
)

// DeploymentConfigChangeController increments the version of a
// DeploymentConfig which has a config change trigger when a pod template
// change is detected.
//
// Use the DeploymentConfigChangeControllerFactory to create this controller.
type DeploymentConfigChangeController struct {
	// changeStrategy knows how to generate and update DeploymentConfigs.
	changeStrategy changeStrategy
	// decodeConfig knows how to decode the deploymentConfig from a deployment's annotations.
	decodeConfig func(deployment *kapi.ReplicationController) (*deployapi.DeploymentConfig, error)
}

// fatalError is an error which can't be retried.
type fatalError string

func (e fatalError) Error() string {
	return fmt.Sprintf("fatal error handling configuration: %s", string(e))
}

// Handle processes change triggers for config.
func (c *DeploymentConfigChangeController) Handle(config *deployapi.DeploymentConfig) error {
	if !deployutil.HasChangeTrigger(config) {
		glog.V(5).Infof("Ignoring DeploymentConfig %s; no change triggers detected", deployutil.LabelForDeploymentConfig(config))
		return nil
	}

	if config.Status.LatestVersion == 0 {
		_, _, abort, err := c.generateDeployment(config)
		if err != nil {
			if kerrors.IsConflict(err) {
				return fatalError(fmt.Sprintf("DeploymentConfig %s updated since retrieval; aborting trigger: %v", deployutil.LabelForDeploymentConfig(config), err))
			}
			glog.V(4).Infof("Couldn't create initial deployment for deploymentConfig %q: %v", deployutil.LabelForDeploymentConfig(config), err)
			return nil
		}
		if !abort {
			glog.V(4).Infof("Created initial deployment for deploymentConfig %q", deployutil.LabelForDeploymentConfig(config))
		}
		return nil
	}

	latestDeploymentName := deployutil.LatestDeploymentNameForConfig(config)
	deployment, err := c.changeStrategy.getDeployment(config.Namespace, latestDeploymentName)
	if err != nil {
		// If there's no deployment for the latest config, we have no basis of
		// comparison. It's the responsibility of the deployment config controller
		// to make the deployment for the config, so return early.
		if kerrors.IsNotFound(err) {
			glog.V(5).Infof("Ignoring change for DeploymentConfig %s; no existing Deployment found", deployutil.LabelForDeploymentConfig(config))
			return nil
		}
		return fmt.Errorf("couldn't retrieve Deployment for DeploymentConfig %s: %v", deployutil.LabelForDeploymentConfig(config), err)
	}

	deployedConfig, err := c.decodeConfig(deployment)
	if err != nil {
		return fatalError(fmt.Sprintf("error decoding DeploymentConfig from Deployment %s for DeploymentConfig %s: %v", deployutil.LabelForDeployment(deployment), deployutil.LabelForDeploymentConfig(config), err))
	}

	// Detect template diffs, and return early if there aren't any changes.
	if kapi.Semantic.DeepEqual(config.Spec.Template, deployedConfig.Spec.Template) {
		glog.V(5).Infof("Ignoring DeploymentConfig change for %s (latestVersion=%d); same as Deployment %s", deployutil.LabelForDeploymentConfig(config), config.Status.LatestVersion, deployutil.LabelForDeployment(deployment))
		return nil
	}

	// There was a template diff, so generate a new config version.
	fromVersion, toVersion, abort, err := c.generateDeployment(config)
	if err != nil {
		if kerrors.IsConflict(err) {
			return fatalError(fmt.Sprintf("DeploymentConfig %s updated since retrieval; aborting trigger: %v", deployutil.LabelForDeploymentConfig(config), err))
		}
		return fmt.Errorf("couldn't generate deployment for DeploymentConfig %s: %v", deployutil.LabelForDeploymentConfig(config), err)
	}
	if !abort {
		glog.V(4).Infof("Updated DeploymentConfig %s from version %d to %d for existing deployment %s", deployutil.LabelForDeploymentConfig(config), fromVersion, toVersion, deployutil.LabelForDeployment(deployment))
	}
	return nil
}

func (c *DeploymentConfigChangeController) generateDeployment(config *deployapi.DeploymentConfig) (int, int, bool, error) {
	newConfig, err := c.changeStrategy.generateDeploymentConfig(config.Namespace, config.Name)
	if err != nil {
		return -1, -1, false, err
	}

	// The generator returns a cause only when there is an image change. If the configchange
	// controller detects an image change, it should just quit, otherwise it is racing with
	// the imagechange controller.
	if newConfig.Status.LatestVersion != config.Status.LatestVersion &&
		deployutil.CauseFromAutomaticImageChange(newConfig) {
		return -1, -1, true, nil
	}

	if newConfig.Status.LatestVersion == config.Status.LatestVersion {
		newConfig.Status.LatestVersion++
	}

	// set the trigger details for the new deployment config
	causes := []deployapi.DeploymentCause{
		{
			Type: deployapi.DeploymentTriggerOnConfigChange,
		},
	}
	newConfig.Status.Details = &deployapi.DeploymentDetails{
		Causes: causes,
	}

	// This update is atomic. If it fails because a newer resource was already persisted, that's
	// okay - we can just ignore the update for the old resource and any changes to the more
	// current config will be captured in future events.
	updatedConfig, err := c.changeStrategy.updateDeploymentConfig(config.Namespace, newConfig)
	if err != nil {
		return config.Status.LatestVersion, newConfig.Status.LatestVersion, false, err
	}

	return config.Status.LatestVersion, updatedConfig.Status.LatestVersion, false, nil
}

// changeStrategy knows how to generate and update DeploymentConfigs.
type changeStrategy interface {
	getDeployment(namespace, name string) (*kapi.ReplicationController, error)
	generateDeploymentConfig(namespace, name string) (*deployapi.DeploymentConfig, error)
	updateDeploymentConfig(namespace string, config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
}

// changeStrategyImpl is a pluggable changeStrategy.
type changeStrategyImpl struct {
	getDeploymentFunc            func(namespace, name string) (*kapi.ReplicationController, error)
	generateDeploymentConfigFunc func(namespace, name string) (*deployapi.DeploymentConfig, error)
	updateDeploymentConfigFunc   func(namespace string, config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error)
}

func (i *changeStrategyImpl) getDeployment(namespace, name string) (*kapi.ReplicationController, error) {
	return i.getDeploymentFunc(namespace, name)
}

func (i *changeStrategyImpl) generateDeploymentConfig(namespace, name string) (*deployapi.DeploymentConfig, error) {
	return i.generateDeploymentConfigFunc(namespace, name)
}

func (i *changeStrategyImpl) updateDeploymentConfig(namespace string, config *deployapi.DeploymentConfig) (*deployapi.DeploymentConfig, error) {
	return i.updateDeploymentConfigFunc(namespace, config)
}
