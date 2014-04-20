package js

func Controllers() string {
	return `
var mailhogApp = angular.module('mailhogApp', []);

mailhogApp.controller('MailCtrl', function ($scope, $http) {
  $scope.refresh = function() {
    $http.get('/api/v1/messages').success(function(data) {
      $scope.messages = data;
    });
  }
  $scope.refresh();

  $scope.date = function(timestamp) {
  	return (new Date(timestamp)).toString();
  };

  $scope.selectMessage = function(message) {
  	$scope.preview = message;
  }

  $scope.deleteAll = function() {
  	$('#confirm-delete-all').modal('show');
  }

  $scope.deleteAllConfirm = function() {
  	$('#confirm-delete-all').modal('hide');
  	$http.post('/api/v1/messages/delete').success(function() {
  		$scope.refresh();
  		$scope.preview = null;
  	});
  }

  $scope.deleteOne = function(message) {
  	$http.post('/api/v1/messages/' + message._id + '/delete').success(function() {
  		if($scope.preview._id == message._id) $scope.preview = null;
  		$scope.refresh();
  	});
  }
});
`;
}