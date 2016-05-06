'use strict';

angular.module('openshiftConsole')
  .directive('overviewMonopod', function(Navigate, $location) {
    return {
      restrict: 'E',
      scope: {
        pod: '='
      },
      templateUrl: 'views/_overview-monopod.html',
      link: function(scope) {
        scope.viewPod = function() {
          var url = Navigate.resourceURL(scope.pod, "Pod", scope.pod.metadata.namespace);
          $location.url(url);
        };
      }
    };
  })
  .directive('podTemplate', function() {
    return {
      restrict: 'E',
      scope: {
        podTemplate: '=',
        imagesByDockerReference: '=',
        builds: '=',
        detailed: '=?',
        // Optional URL for setting health checks on the resource when missing.
        addHealthCheckUrl: '@?'
      },
      templateUrl: 'views/_pod-template.html'
    };
  })

  /*
   * This directive is not currently used since we've switched to a donut chart on the overview.
   */
  //.directive('pods', function() {
  //  return {
  //    restrict: 'E',
  //    scope: {
  //      pods: '=',
  //      projectName: '@?' //TODO optional for now
  //    },
  //    templateUrl: 'views/_pods.html',
  //    controller: function($scope) {
  //      $scope.phases = [
  //        "Failed",
  //        "Pending",
  //        "Running",
  //        "Succeeded",
  //        "Unknown"
  //      ];
  //      $scope.expandedPhase = null;
  //      $scope.warningsExpanded = false;
  //      $scope.expandPhase = function(phase, warningsExpanded, $event) {
  //        $scope.expandedPhase = phase;
  //        $scope.warningsExpanded = warningsExpanded;
  //        if ($event) {
  //          $event.stopPropagation();
  //        }
  //      };
  //    }
  //  };
  //})

  /*
   * This directive is not currently used since we've switched to a donut chart on the overview.
   */
  //.directive('podContent', function() {
  //  // sub-directive used by the pods directive
  //  return {
  //    restrict: 'E',
  //    scope: {
  //      pod: '=',
  //      troubled: '='
  //    },
  //    templateUrl: 'views/directives/_pod-content.html'
  //  };
  //})

  .directive('triggers', function() {
    var hideBuildKey = function(build) {
      return 'hide/build/' + build.metadata.uid;
    };
    return {
      restrict: 'E',
      scope: {
        triggers: '=',
        buildsByOutputImage: '=',
        namespace: '='
      },
      link: function(scope) {
        scope.isBuildHidden = function(build) {
          var key = hideBuildKey(build);
          return sessionStorage.getItem(key) === 'true';
        };
        scope.hideBuild = function(build) {
          var key = hideBuildKey(build);
          sessionStorage.setItem(key, 'true');
        };
      },
      templateUrl: 'views/_triggers.html'
    };
  })
  .directive('deploymentConfigMetadata', function() {
    return {
      restrict: 'E',
      scope: {
        deploymentConfigId: '=',
        exists: '=',
        differentService: '='
      },
      templateUrl: 'views/_deployment-config-metadata.html'
    };
  })
  .directive('annotations', function() {
    return {
      restrict: 'E',
      scope: {
        annotations: '='
      },
      templateUrl: 'views/directives/annotations.html',
      link: function(scope) {
        scope.expandAnnotations = false;
        scope.toggleAnnotations = function() {
          scope.expandAnnotations = !scope.expandAnnotations;
        };
      }
    };
  })
  .directive('volumes', function() {
    return {
      restrict: 'E',
      scope: {
        volumes: '=',
        namespace: '='
      },
      templateUrl: 'views/_volumes.html'
    };
  })
  .directive('environment', function() {
    return {
      restrict: 'E',
      scope: {
        envVars: '='
      },
      templateUrl: 'views/directives/environment.html',
      controller: function($scope) {
        $scope.expanded = {};
      }
    };
  })
  .directive('hpa', function() {
    return {
      restrict: 'E',
      scope: {
        hpa: '=',
        project: '=',
        showScaleTarget: '=?',
        alerts: '='
      },
      templateUrl: 'views/directives/hpa.html'
    };
  })
  .directive('probe', function() {
    return {
      restrict: 'E',
      scope: {
        probe: '='
      },
      templateUrl: 'views/directives/_probe.html'
    };
  });
