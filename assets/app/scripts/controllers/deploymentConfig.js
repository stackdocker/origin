'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:DeploymentConfigController
 * @description
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('DeploymentConfigController',
              function ($scope,
                        $filter,
                        $routeParams,
                        AlertMessageService,
                        DataService,
                        DeploymentsService,
                        HPAService,
                        ImageStreamResolver,
                        Navigate,
                        ProjectsService,
                        LabelFilter) {
    $scope.projectName = $routeParams.project;
    $scope.deploymentConfigName = $routeParams.deploymentconfig;
    $scope.deploymentConfig = null;
    $scope.deployments = {};
    $scope.unfilteredDeployments = {};
    $scope.imageStreams = {};
    $scope.imagesByDockerReference = {};
    $scope.imageStreamImageRefByDockerReference = {}; // lets us determine if a particular container's docker image reference belongs to an imageStream
    $scope.builds = {};
    $scope.labelSuggestions = {};
    // TODO we should add this back in and show the pod template on this page
    //$scope.podTemplates = {};
    //$scope.imageStreams = {};
    //$scope.imagesByDockerReference = {};
    //$scope.imageStreamImageRefByDockerReference = {}; // lets us determine if a particular container's docker image reference belongs to an imageStream
    //$scope.builds = {};
    $scope.alerts = {};
    $scope.breadcrumbs = [
      {
        title: "Deployments",
        link: "project/" + $routeParams.project + "/browse/deployments"
      },
      {
        title: $routeParams.deploymentconfig
      }
    ];
    $scope.emptyMessage = "Loading...";
    $scope.healthCheckURL = Navigate.healthCheckURL($routeParams.project,
                                                    "DeploymentConfig",
                                                    $routeParams.deploymentconfig);

    // Check for a ?tab=<name> query param to allow linking directly to a tab.
    if ($routeParams.tab) {
      $scope.selectedTab = {};
      $scope.selectedTab[$routeParams.tab] = true;
    }

    // get and clear any alerts
    AlertMessageService.getAlerts().forEach(function(alert) {
      $scope.alerts[alert.name] = alert.data;
    });
    AlertMessageService.clearAlerts();

    var watches = [];
    var hashSize = $filter('hashSize');

    ProjectsService
      .get($routeParams.project)
      .then(_.spread(function(project, context) {
        $scope.project = project;
        $scope.projectContext = context;

        var limitRanges = {};

        var updateHPAWarnings = function() {
            HPAService.getHPAWarnings($scope.deploymentConfig, $scope.autoscalers, limitRanges, project)
                      .then(function(warnings) {
              $scope.hpaWarnings = warnings;
            });
        };

        DataService.get("deploymentconfigs", $routeParams.deploymentconfig, context).then(
          // success
          function(deploymentConfig) {
            $scope.loaded = true;
            $scope.deploymentConfig = deploymentConfig;
            updateHPAWarnings();
            ImageStreamResolver.fetchReferencedImageStreamImages([deploymentConfig.spec.template], $scope.imagesByDockerReference, $scope.imageStreamImageRefByDockerReference, context);

            // If we found the item successfully, watch for changes on it
            watches.push(DataService.watchObject("deploymentconfigs", $routeParams.deploymentconfig, context, function(deploymentConfig, action) {
              if (action === "DELETED") {
                $scope.alerts["deleted"] = {
                  type: "warning",
                  message: "This deployment configuration has been deleted."
                };
              }
              $scope.deploymentConfig = deploymentConfig;
              updateHPAWarnings();
              ImageStreamResolver.fetchReferencedImageStreamImages([deploymentConfig.spec.template], $scope.imagesByDockerReference, $scope.imageStreamImageRefByDockerReference, context);
            }));
          },
          // failure
          function(e) {
            $scope.loaded = true;
            $scope.alerts["load"] = {
              type: "error",
              message: e.status === 404 ? "This deployment configuration can not be found, it may have been deleted." : "The deployment configuration details could not be loaded.",
              details: e.status === 404 ? "Any remaining deployment history for this deployment will be shown." : "Reason: " + $filter('getErrorDetails')(e)
            };
          }
        );

        // TODO we should add this back in and show the pod template on this page
        // function extractPodTemplates() {
        //   angular.forEach($scope.deployments, function(deployment, deploymentId){
        //     $scope.podTemplates[deploymentId] = deployment.spec.template;
        //   });
        // }

        // List limit ranges in this project to determine if there is a default
        // CPU request for autoscaling.
        DataService.list("limitranges", context, function(response) {
          limitRanges = response.by("metadata.name");
          updateHPAWarnings();
        });

        watches.push(DataService.watch("replicationcontrollers", context, function(deployments, action, deployment) {
          // TODO we should add this back in and show the pod template on this page
          // extractPodTemplates();
          // ImageStreamResolver.fetchReferencedImageStreamImages($scope.podTemplates, $scope.imagesByDockerReference, $scope.imageStreamImageRefByDockerReference, $scope);
          $scope.emptyMessage = "No deployments to show";

          if (!action) {
            var deploymentsByDeploymentConfig = DeploymentsService.associateDeploymentsToDeploymentConfig(deployments.by("metadata.name"));
            $scope.unfilteredDeployments = deploymentsByDeploymentConfig[$routeParams.deploymentconfig] || {};
            angular.forEach($scope.unfilteredDeployments, function(deployment) {
              deployment.causes = $filter('deploymentCauses')(deployment);
            });
            // Loading of the page that will create deploymentConfigDeploymentsInProgress structure, which will associate running deployment to his deploymentConfig.
            $scope.deploymentConfigDeploymentsInProgress = DeploymentsService.associateRunningDeploymentToDeploymentConfig(deploymentsByDeploymentConfig);
          } else if (DeploymentsService.deploymentBelongsToConfig(deployment, $routeParams.deploymentconfig)) {
            var deploymentName = deployment.metadata.name;
            var deploymentConfigName = $routeParams.deploymentconfig;
            switch (action) {
              case 'ADDED':
              case 'MODIFIED':
                $scope.unfilteredDeployments[deploymentName] = deployment;
                // When deployment is retried, associate him to his deploymentConfig and add him into deploymentConfigDeploymentsInProgress structure.
                if ($filter('deploymentIsInProgress')(deployment)){
                  $scope.deploymentConfigDeploymentsInProgress[deploymentConfigName] = $scope.deploymentConfigDeploymentsInProgress[deploymentConfigName] || {};
                  $scope.deploymentConfigDeploymentsInProgress[deploymentConfigName][deploymentName] = deployment;
                } else if ($scope.deploymentConfigDeploymentsInProgress[deploymentConfigName]) { // After the deployment ends remove him from the deploymentConfigDeploymentsInProgress structure.
                  delete $scope.deploymentConfigDeploymentsInProgress[deploymentConfigName][deploymentName];
                }
                deployment.causes = $filter('deploymentCauses')(deployment);
                break;
              case 'DELETED':
                delete $scope.unfilteredDeployments[deploymentName];
                if ($scope.deploymentConfigDeploymentsInProgress[deploymentConfigName]) {
                  delete $scope.deploymentConfigDeploymentsInProgress[deploymentConfigName][deploymentName];
                }
                break;
            }
          }

          $scope.deployments = LabelFilter.getLabelSelector().select($scope.unfilteredDeployments);
          updateFilterWarning();
          LabelFilter.addLabelSuggestionsFromResources($scope.unfilteredDeployments, $scope.labelSuggestions);
          LabelFilter.setLabelSuggestions($scope.labelSuggestions);
        }));

        watches.push(DataService.watch("imagestreams", context, function(imageStreams) {
          $scope.imageStreams = imageStreams.by("metadata.name");
          ImageStreamResolver.buildDockerRefMapForImageStreams($scope.imageStreams, $scope.imageStreamImageRefByDockerReference);
          // If the dep config has been loaded already
          if ($scope.deploymentConfig) {
            ImageStreamResolver.fetchReferencedImageStreamImages([$scope.deploymentConfig.spec.template], $scope.imagesByDockerReference, $scope.imageStreamImageRefByDockerReference, context);
          }
          Logger.log("imagestreams (subscribe)", $scope.imageStreams);
        }));

        watches.push(DataService.watch("builds", context, function(builds) {
          $scope.builds = builds.by("metadata.name");
          Logger.log("builds (subscribe)", $scope.builds);
        }));

        watches.push(DataService.watch({
          group: "extensions",
          resource: "horizontalpodautoscalers"
        }, context, function(hpa) {
          $scope.autoscalers =
            HPAService.hpaForDC(hpa.by("metadata.name"), $routeParams.deploymentconfig);
          updateHPAWarnings();
        }));

        function updateFilterWarning() {
          if (!LabelFilter.getLabelSelector().isEmpty() && $.isEmptyObject($scope.deployments) && !$.isEmptyObject($scope.unfilteredDeployments)) {
            $scope.alerts["deployments"] = {
              type: "warning",
              details: "The active filters are hiding all deployments."
            };
          }
          else {
            delete $scope.alerts["deployments"];
          }
        }

        LabelFilter.onActiveFiltersChanged(function(labelSelector) {
          // trigger a digest loop
          $scope.$apply(function() {
            $scope.deployments = labelSelector.select($scope.unfilteredDeployments);
            updateFilterWarning();
          });
        });

        $scope.canDeploy = function() {
          if (!$scope.deploymentConfig) {
            return false;
          }

          if ($scope.deploymentConfig.metadata.deletionTimestamp) {
            return false;
          }

          if ($scope.deploymentConfigDeploymentsInProgress &&
              hashSize($scope.deploymentConfigDeploymentsInProgress[$scope.deploymentConfig.metadata.name]) > 0) {
            return false;
          }

          return true;
        };

        $scope.startLatestDeployment = function() {
          if ($scope.canDeploy()) {
            DeploymentsService.startLatestDeployment($scope.deploymentConfig, context, $scope);
          }
        };

        $scope.scale = function(replicas) {
          var showScalingError = function(result) {
            $scope.alerts = $scope.alerts || {};
            $scope.alerts["scale"] = {
              type: "error",
              message: "An error occurred scaling the deployment config.",
              details: $filter('getErrorDetails')(result)
            };
          };

          DeploymentsService.scaleDC($scope.deploymentConfig, replicas).then(_.noop, showScalingError);
        };

        $scope.$on('$destroy', function(){
          DataService.unwatchAll(watches);
        });
    }));
  });
