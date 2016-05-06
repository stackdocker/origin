'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:CreateProjectController
 * @description
 * # CreateProjectController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('CreateProjectController', function ($scope, $location, AuthService, DataService, AlertMessageService) {
    $scope.alerts = {};

    AuthService.withUser();

    AlertMessageService.getAlerts().forEach(function(alert) {
      $scope.alerts[alert.name] = alert.data;
    });
    AlertMessageService.clearAlerts();

    $scope.createProject = function() {
      $scope.disableInputs = true;
      if ($scope.createProjectForm.$valid) {
        DataService.create('projectrequests', null, {
          apiVersion: "v1",
          kind: "ProjectRequest",
          metadata: {
            name: $scope.name
          },
          displayName: $scope.displayName,
          description: $scope.description
        }, $scope).then(function(data) { // Success
          // Take the user directly to the create page to add content.
          $location.path("project/" + encodeURIComponent(data.metadata.name) + "/create");
        }, function(result) { // Failure
          $scope.disableInputs = false;
          var data = result.data || {};
          if (data.reason === 'AlreadyExists') {
            $scope.nameTaken = true;
          } else {
            var msg = data.message || 'An error occurred creating the project.';
            $scope.alerts['error-creating-project'] = {type: 'error', message: msg};
          }
        });
      }
    };
  });
