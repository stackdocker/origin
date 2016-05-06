'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:ProjectsController
 * @description
 * # ProjectsController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('ProjectsController', function ($scope, $route, $timeout, $filter, $location, DataService, AuthService, AlertMessageService, Logger, hashSizeFilter) {
    $scope.projects = {};
    $scope.alerts = $scope.alerts || {};
    $scope.showGetStarted = false;
    $scope.canCreate = undefined;

    AlertMessageService.getAlerts().forEach(function(alert) {
      $scope.alerts[alert.name] = alert.data;
    });
    AlertMessageService.clearAlerts();

    $timeout(function() {
      $('#openshift-logo').on('click.projectsPage', function() {
        // Force a reload. Angular doesn't reload the view when the URL doesn't change.
        $route.reload();
      });
    });

    $scope.$on('deleteProject', function() {
      loadProjects();
    });

    $scope.$on('$destroy', function(){
      // The click handler is only necessary on the projects page.
      $('#openshift-logo').off('click.projectsPage');
    });

    AuthService.withUser().then(function() {
      loadProjects();
    });

    // Test if the user can submit project requests. Handle error notifications
    // ourselves because 403 responses are expected.
    DataService.get("projectrequests", null, $scope, { errorNotification: false})
    .then(function() {
      $scope.canCreate = true;
    }, function(result) {
      $scope.canCreate = false;

      var data = result.data || {};

      // 403 Forbidden indicates the user doesn't have authority.
      // Any other failure status is an unexpected error.
      if (result.status !== 403) {
        var msg = 'Failed to determine create project permission';
        if (result.status !== 0) {
          msg += " (" + result.status + ")";
        }
        Logger.warn(msg);
        return;
      }

      // Check if there are detailed messages. If there are, show them instead of our default message.
      if (data.details) {
        var messages = [];
        _.forEach(data.details.causes || [], function(cause) {
          if (cause.message) { messages.push(cause.message); }
        });
        if (messages.length > 0) {
          $scope.newProjectMessage = messages.join("\n");
        }
      }
    });

    function loadProjects() {
      DataService.list("projects", $scope, function(projects) {
        $scope.projects = projects.by("metadata.name");
        $scope.showGetStarted = hashSizeFilter($scope.projects) === 0;
      });
    }
  });
