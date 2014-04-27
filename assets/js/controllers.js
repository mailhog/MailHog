var mailhogApp = angular.module('mailhogApp', []);

function guid() {
  function s4() {
    return Math.floor((1 + Math.random()) * 0x10000)
               .toString(16)
               .substring(1);
  }
  return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
         s4() + '-' + s4() + s4() + s4();
}

mailhogApp.controller('MailCtrl', function ($scope, $http, $sce, $timeout) {
  $scope.cache = {};
  $scope.previewAllHeaders = false;

  $scope.eventsPending = {};
  $scope.eventCount = 0;
  $scope.eventDone = 0;
  $scope.eventFailed = 0;

  $scope.startEvent = function(name, args) {
    var eID = guid();
    console.log("Starting event '" + name + "' with id '" + eID + "'")
    var e = {
      id: eID,
      name: name,
      started: new Date(),
      complete: false,
      failed: false,
      args: args,
      getClass: function() {
        // FIXME bit nasty
        if(this.failed) {
          return "bg-danger"
        }
        if(this.complete) {
          return "bg-success"
        }
        return "bg-warning"; // pending
      },
      done: function() {
        //delete $scope.eventsPending[eID]
        var e = this;
        e.complete = true;
        $scope.eventDone++;
        if(this.failed) {
          console.log("Failed event '" + e.name + "' with id '" + eID + "'")
        } else {
          console.log("Completed event '" + e.name + "' with id '" + eID + "'")
          $timeout(function() {
            e.remove();
          }, 10000);
        }
      },
      fail: function() {
        $scope.eventFailed++;
        this.failed = true;
        this.done();
      },
      remove: function() {
        console.log("Deleted event '" + e.name + "' with id '" + eID + "'")
        if(e.failed) {
          $scope.eventFailed--;
        }
        delete $scope.eventsPending[eID];
        $scope.eventDone--;
        $scope.eventCount--;
        return false;
      }
    };
    $scope.eventsPending[eID] = e;
    $scope.eventCount++;
    return e;
  }

  $scope.refresh = function() {
    var e = $scope.startEvent("Loading messages");
    $http.get('/api/v1/messages').success(function(data) {
      $scope.messages = data;
      e.done();
    });
  }
  $scope.refresh();

  $scope.selectMessage = function(message) {
  	if($scope.cache[message.Id]) {
  		$scope.preview = $scope.cache[message.Id];
      reflow();
  	} else {
  		$scope.preview = message;
      var e = $scope.startEvent("Loading message", message.Id);
	  	$http.get('/api/v1/messages/' + message.Id).success(function(data) {
	  	  $scope.cache[message.Id] = data;
	      data.previewHTML = $sce.trustAsHtml($scope.getMessageHTML(data));
  		  $scope.preview = data;
  		  preview = $scope.cache[message.Id];
        reflow();
        e.done();
	    });
	}
  }

  $scope.toggleHeaders = function(val) {
    $scope.previewAllHeaders = val;
    var t = window.setInterval(function() {
      if(val) {
        if($('#hide-headers').length) {
          window.clearInterval(t);
          reflow();
        }
      } else {
        if($('#show-headers').length) {
          window.clearInterval(t);
          reflow();
        }
      }
    }, 10);
  }

  $scope.getMessagePlain = function(message) {
  	var part;

  	if(message.MIME) {
  		for(var p in message.MIME.Parts) {
        if("Content-Type" in message.MIME.Parts[p].Headers) {
          if(message.MIME.Parts[p].Headers["Content-Type"].length > 0) {
      			if(message.MIME.Parts[p].Headers["Content-Type"][0].match(/text\/plain;?.*/)) {
      				part = message.MIME.Parts[p];
      				break;
      			}
          }
        }
  		}
	}

	if(!part) part = message.Content;

	return part.Body;
  }
  $scope.getMessageHTML = function(message) {
  	var part;
  	
  	if(message.MIME) {
  		for(var p in message.MIME.Parts) {
        if("Content-Type" in message.MIME.Parts[p].Headers) {
          if(message.MIME.Parts[p].Headers["Content-Type"].length > 0) {
      			if(message.MIME.Parts[p].Headers["Content-Type"][0].match(/text\/html;?.*/)) {
      				part = message.MIME.Parts[p];
      				break;
      			}
          }
        }
  		}
	}

	if(!part) part = message.Content;

	return part.Body;
  }

  $scope.date = function(timestamp) {
  	return (new Date(timestamp)).toString();
  };

  $scope.deleteAll = function() {
  	$('#confirm-delete-all').modal('show');
  }

  $scope.releaseOne = function(message) {
    $scope.releasing = message;
    $('#release-one').modal('show');
  }
  $scope.confirmReleaseMessage = function() {
    $('#release-one').modal('hide');
    var message = $scope.releasing;
    $scope.releasing = null;

    var e = $scope.startEvent("Releasing message", message.Id);

    $http.post('/api/v1/messages/' + message.Id + '/release', {
      email: $('#release-message-email').val(),
      host: $('#release-message-smtp-host').val(),
      port: $('#release-message-smtp-port').val(),
    }).success(function() {
      e.done();
    }).error(function(err) {
      e.fail();
      e.error = err;
    });
  }

  $scope.getSource = function(message) {
  	var source = "";
  	$.each(message.Content.Headers, function(k, v) {
  		source += k + ": " + v + "\n";
  	});
	source += "\n";
	source += message.Content.Body;
	return source;
  }

  $scope.deleteAllConfirm = function() {
  	$('#confirm-delete-all').modal('hide');
    var e = $scope.startEvent("Deleting all messages");
  	$http.post('/api/v1/messages/delete').success(function() {
  		$scope.refresh();
  		$scope.preview = null;
      e.done()
  	});
  }

  $scope.deleteOne = function(message) {
    var e = $scope.startEvent("Deleting message", message.Id);
  	$http.post('/api/v1/messages/' + message.Id + '/delete').success(function() {
  		if($scope.preview._id == message._id) $scope.preview = null;
  		$scope.refresh();
      e.done();
  	});
  }
});