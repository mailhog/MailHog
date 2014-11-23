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

  $scope.hasEventSource = false;
  $scope.source = null;

  $(function() {
    $scope.openStream();
  });

  $scope.toggleStream = function() {
    $scope.source == null ? $scope.openStream() : $scope.closeStream();
  }
  $scope.openStream = function() {
    $scope.source = new EventSource('/api/v1/events');
    $scope.source.addEventListener('message', function(e) {
      $scope.$apply(function() {
        $scope.messages.push(JSON.parse(e.data));
      });
    }, false);
    $scope.source.addEventListener('open', function(e) {
      $scope.$apply(function() {
        $scope.hasEventSource = true;
      });
    }, false);
    $scope.source.addEventListener('error', function(e) {
      //if(e.readyState == EventSource.CLOSED) {
        $scope.$apply(function() {
          $scope.hasEventSource = false;
        });
      //}
    }, false);
  }
  $scope.closeStream = function() {
    $scope.source.close();
    $scope.source = null;
    $scope.hasEventSource = false;
  }

  $scope.tryDecodeMime = function(str) {
    return unescapeFromMime(str)
  }

  $scope.startEvent = function(name, args, glyphicon) {
    var eID = guid();
    //console.log("Starting event '" + name + "' with id '" + eID + "'")
    var e = {
      id: eID,
      name: name,
      started: new Date(),
      complete: false,
      failed: false,
      args: args,
      glyphicon: glyphicon,
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
          // console.log("Failed event '" + e.name + "' with id '" + eID + "'")
        } else {
          // console.log("Completed event '" + e.name + "' with id '" + eID + "'")
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
        // console.log("Deleted event '" + e.name + "' with id '" + eID + "'")
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

  $scope.messagesDisplayed = function() {
    return $('#messages-container table tbody tr').length
  }

  $scope.refresh = function() {
    var e = $scope.startEvent("Loading messages", null, "glyphicon-download");
    $http.get('/api/v1/messages').success(function(data) {
      $scope.messages = data;
      e.done();
    });
  }
  $scope.refresh();

  $scope.selectMessage = function(message) {
  	if($scope.cache[message.ID]) {
  		$scope.preview = $scope.cache[message.ID];
      reflow();
  	} else {
  		$scope.preview = message;
      var e = $scope.startEvent("Loading message", message.ID, "glyphicon-download-alt");
	  	$http.get('/api/v1/messages/' + message.ID).success(function(data) {
	  	  $scope.cache[message.ID] = data;
	      data.previewHTML = $sce.trustAsHtml($scope.getMessageHTML(data));
  		  $scope.preview = data;
  		  preview = $scope.cache[message.ID];
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

  $scope.tryDecodeContent = function(message, content) {
    var charset = "UTF-8"
    if(message.Content.Headers["Content-Type"][0]) {
      // TODO
    }

    if(message.Content.Headers["Content-Transfer-Encoding"][0]) {
      if(message.Content.Headers["Content-Transfer-Encoding"][0] == "quoted-printable") {
        content = unescapeFromQuotedPrintable(content, charset)
      }
    }

    return content
  }

  $scope.getMessagePlain = function(message) {
    var l = $scope.findMatchingMIME(message, "text/plain");
    if(l != null && l !== "undefined") {
      return l.Body;
    }
    return message.Content.Body;
	}

  $scope.findMatchingMIME = function(part, mime) {
    // TODO cache results
    if(part.MIME) {
      for(var p in part.MIME.Parts) {
        if("Content-Type" in part.MIME.Parts[p].Headers) {
          if(part.MIME.Parts[p].Headers["Content-Type"].length > 0) {
            if(part.MIME.Parts[p].Headers["Content-Type"][0].match(mime + ";?.*")) {
              return part.MIME.Parts[p];
            } else if (part.MIME.Parts[p].Headers["Content-Type"][0].match(/multipart\/.*/)) {
              var f = $scope.findMatchingMIME(part.MIME.Parts[p], mime);
              if(f != null) {
                return f;
              }
            }
          }
        }
      }
    }
    return null;
  }
  $scope.hasHTML = function(message) {
    // TODO cache this
    var l = $scope.findMatchingMIME(message, "text/html");
    if(l != null && l !== "undefined") {
      return true
    }
    return false;
  }
  $scope.getMessageHTML = function(message) {
    var l = $scope.findMatchingMIME(message, "text/html");
    if(l != null && l !== "undefined") {
      return l.Body;
    }
  	return "<HTML not found>";
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

    var e = $scope.startEvent("Releasing message", message.ID, "glyphicon-share");

    $http.post('/api/v1/messages/' + message.ID + '/release', {
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
    var e = $scope.startEvent("Deleting all messages", null, "glyphicon-remove-circle");
  	$http.delete('/api/v1/messages').success(function() {
  		$scope.refresh();
  		$scope.preview = null;
      e.done()
  	});
  }

  $scope.deleteOne = function(message) {
    var e = $scope.startEvent("Deleting message", message.ID, "glyphicon-remove");
  	$http.delete('/api/v1/messages/' + message.ID).success(function() {
  		if($scope.preview._id == message._id) $scope.preview = null;
  		$scope.refresh();
      e.done();
  	});
  }
});
